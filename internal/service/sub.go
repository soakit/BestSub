package service

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bestruirui/bestsub/internal/model"
	"github.com/bestruirui/bestsub/internal/rhttp"
	"github.com/bestruirui/bestsub/internal/server/stream"
	"github.com/bestruirui/bestsub/internal/store"
)

var (
	refreshMu sync.Map

	// RefreshEvents 是订阅刷新过程的实时事件流。
	RefreshEvents = stream.New()
)

// RefreshEvent 表示刷新过程中的一条事件。
type RefreshEvent struct {
	SubID   string         `json:"sub_id"`
	Type    string         `json:"type"`
	Payload map[string]any `json:"payload"`
}

// RefreshSubscription 尝试启动订阅刷新任务。已在刷新时返回错误。
func RefreshSubscription(id string) error {
	if id == "" {
		return fmt.Errorf("subscription id is required")
	}
	sub, ok := store.SubscriptionGet(id)
	if !ok {
		return fmt.Errorf("subscription %s not found", id)
	}

	// 每个订阅独立互斥，避免同一个订阅被同时刷新导致节点池和状态交错写入。
	val, _ := refreshMu.LoadOrStore(id, &sync.Mutex{})
	mu := val.(*sync.Mutex)
	if !mu.TryLock() {
		return fmt.Errorf("refresh already in progress")
	}

	go func() {
		defer mu.Unlock()

		proxy := sub.ProxyUrl
		switch sub.ProxyMode {
		case model.ProxyModeDisabled:
			proxy = "direct"
		case model.ProxyModeAuto:
			proxy = ""
		}

		var urls []string
		var err error
		switch sub.UrlType {
		case model.URLTypeDirect:
			urls = sub.Url
		case model.URLTypeList:
			if len(sub.Url) == 0 {
				RefreshEvents.Emit("failed", RefreshEvent{
					SubID: id,
					Type:  "failed",
					Payload: map[string]any{
						"message": "no url provided",
					},
				})
				return
			}
			if proxy == "" {
				urls, err = fetchUrlList(sub.Url[0], sub.Header, "direct")
				if err != nil {
					urls, err = fetchUrlList(sub.Url[0], sub.Header, sub.ProxyUrl)
				}
			} else {
				urls, err = fetchUrlList(sub.Url[0], sub.Header, proxy)
			}
			if err != nil {
				RefreshEvents.Emit("failed", RefreshEvent{
					SubID: id,
					Type:  "failed",
					Payload: map[string]any{
						"message": fmt.Sprintf("parse urls: %v", err),
					},
				})
				return
			}
		default:
			RefreshEvents.Emit("failed", RefreshEvent{
				SubID: id,
				Type:  "failed",
				Payload: map[string]any{
					"message": fmt.Sprintf("unknown url_type: %d", sub.UrlType),
				},
			})
			return
		}

		total := len(urls)
		status := model.SubscriptionStatus{RefreshedAt: time.Now()}

		for i, u := range urls {
			RefreshEvents.Emit("progress", RefreshEvent{
				SubID: id,
				Type:  "progress",
				Payload: map[string]any{
					"index":  i,
					"total":  total,
					"status": "fetching",
				},
			})

			var urlStatus model.SubscriptionStatus
			if sub.ProxyMode == model.ProxyModeAuto {
				_, urlStatus, err = refreshOne(id, u, sub.Header, "direct")
				if err != nil {
					_, urlStatus, err = refreshOne(id, u, sub.Header, sub.ProxyUrl)
				}
			} else {
				_, urlStatus, err = refreshOne(id, u, sub.Header, proxy)
			}

			if err != nil {
				RefreshEvents.Emit("progress", RefreshEvent{
					SubID: id,
					Type:  "progress",
					Payload: map[string]any{
						"index":  i,
						"total":  total,
						"status": "fail",
						"error":  err.Error(),
					},
				})
				continue
			}

			status.TrafficUsed += urlStatus.TrafficUsed
			status.TrafficTotal += urlStatus.TrafficTotal
			if urlStatus.ExpiresAt.After(status.ExpiresAt) {
				status.ExpiresAt = urlStatus.ExpiresAt
			}

			RefreshEvents.Emit("progress", RefreshEvent{
				SubID: id,
				Type:  "progress",
				Payload: map[string]any{
					"index":  i,
					"total":  total,
					"status": "ok",
				},
			})
		}

		status.NodeNum = uint32(store.NodePoolCount(id))

		if err := store.SubscriptionUpdateStatus(id, status); err != nil {
			RefreshEvents.Emit("failed", RefreshEvent{
				SubID: id,
				Type:  "failed",
				Payload: map[string]any{
					"message": fmt.Sprintf("update status: %v", err),
				},
			})
			return
		}

		RefreshEvents.Emit("done", RefreshEvent{SubID: id, Type: "done"})
	}()
	return nil
}

// fetchUrlList 拉取 URL 列表内容，按行解析为多个 URL。
func fetchUrlList(listUrl string, header map[string]string, proxy string) ([]string, error) {
	req, err := http.NewRequest(http.MethodGet, listUrl, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range header {
		req.Header.Set(k, v)
	}
	resp, err := rhttp.New(proxy).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var urls []string
	for _, line := range strings.Split(strings.TrimSpace(string(body)), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			urls = append(urls, line)
		}
	}
	if len(urls) == 0 {
		return nil, fmt.Errorf("no urls found in list")
	}
	return urls, nil
}

// refreshOne 拉取单个 URL 的订阅内容，返回节点数和流量状态。
func refreshOne(subID, rawUrl string, header map[string]string, proxy string) (uint32, model.SubscriptionStatus, error) {
	req, err := http.NewRequest(http.MethodGet, rawUrl, nil)
	if err != nil {
		return 0, model.SubscriptionStatus{}, fmt.Errorf("create request: %w", err)
	}
	for k, v := range header {
		req.Header.Set(k, v)
	}

	resp, err := rhttp.New(proxy).Do(req)
	if err != nil {
		return 0, model.SubscriptionStatus{}, fmt.Errorf("fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, model.SubscriptionStatus{}, fmt.Errorf("status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, model.SubscriptionStatus{}, fmt.Errorf("read body: %w", err)
	}

	// 解析 subscription-userinfo 响应头。
	var status model.SubscriptionStatus
	if info := resp.Header.Get("subscription-userinfo"); info != "" {
		for _, field := range strings.Split(strings.ToLower(info), ";") {
			k, v, ok := strings.Cut(strings.TrimSpace(field), "=")
			if !ok {
				continue
			}
			n, _ := strconv.ParseInt(v, 10, 64)
			switch k {
			case "upload", "download":
				status.TrafficUsed += n
			case "total":
				status.TrafficTotal = n
			case "expire":
				if n > 0 {
					status.ExpiresAt = time.Unix(n, 0)
				}
			}
		}
	}

	convBody, err := Convert(body, "mihomo")
	if err != nil {
		return 0, status, fmt.Errorf("convert: %w", err)
	}

	var nodeNum uint32
	for _, line := range bytes.Split(convBody, []byte("\n"))[1:] {
		raw := bytes.TrimPrefix(line, []byte("  - "))
		cp := make([]byte, len(raw))
		copy(cp, raw)
		if store.NodePoolAdd(subID, cp) {
			nodeNum++
		}
	}
	return nodeNum, status, nil
}
