package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bestruirui/bestsub/internal/model"
	"github.com/bestruirui/bestsub/internal/node"
	"github.com/bestruirui/bestsub/internal/rhttp"
	"github.com/bestruirui/bestsub/internal/server/stream"
	"github.com/bestruirui/bestsub/internal/store"

	"github.com/charmbracelet/log"
	"github.com/robfig/cron/v3"
)

var (
	refreshMu                   sync.Map                                                                     // 按订阅 ID 保存刷新互斥锁。
	subscriptionCronParser      = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow) // 解析 5 段订阅 Cron 表达式。
	subscriptionCron            = cron.New(cron.WithParser(subscriptionCronParser))                          // 执行订阅自动更新的内存调度器。
	subscriptionScheduleMu      sync.Mutex                                                                   // 保护订阅调度项的增删改。
	subscriptionScheduleEntries = map[string]cron.EntryID{}                                                  // 记录订阅 ID 对应的 Cron 调度项。
	subscriptionLifecycleMu     sync.Mutex                                                                   // 保护订阅刷新生命周期和 WaitGroup Add/Wait 边界。
	subscriptionCtx             context.Context                                                              // 取消全部订阅刷新的生命周期上下文。
	subscriptionCancel          context.CancelFunc                                                           // 结束订阅刷新生命周期。
	subscriptionRefreshWG       sync.WaitGroup                                                               // 等待已经启动的订阅刷新退出。

	// RefreshEvents 是订阅刷新过程的实时事件流。
	RefreshEvents = stream.New()
)

// RefreshEvent 表示刷新过程中的一条事件。
type RefreshEvent struct {
	SubID   string         `json:"sub_id"`
	Type    string         `json:"type"`
	Payload map[string]any `json:"payload"`
}

// StartSubscriptionScheduler 注册已有订阅并启动自动更新调度器。
func StartSubscriptionScheduler() error {
	subscriptionLifecycleMu.Lock()
	subscriptionCtx, subscriptionCancel = context.WithCancel(context.Background())
	subscriptionLifecycleMu.Unlock()
	for _, sub := range store.SubscriptionList() {
		if err := SyncSubscriptionSchedule(sub); err != nil {
			subscriptionLifecycleMu.Lock()
			subscriptionCancel()
			subscriptionCtx = nil
			subscriptionCancel = nil
			subscriptionLifecycleMu.Unlock()
			return err
		}
	}
	subscriptionCron.Start()
	return nil
}

// StopSubscriptionScheduler 停止调度并等待已经启动的订阅刷新退出。
func StopSubscriptionScheduler() {
	subscriptionLifecycleMu.Lock()
	if subscriptionCancel != nil {
		subscriptionCancel()
	}
	subscriptionCtx = nil
	subscriptionCancel = nil
	subscriptionLifecycleMu.Unlock()
	<-subscriptionCron.Stop().Done()
	subscriptionRefreshWG.Wait()
}

// SyncSubscriptionSchedule 按订阅最新配置替换对应调度项。
func SyncSubscriptionSchedule(sub model.Subscription) error {
	if sub.AutoUpdate != 1 {
		RemoveSubscriptionSchedule(sub.ID)
		return nil
	}

	subscriptionScheduleMu.Lock()
	defer subscriptionScheduleMu.Unlock()
	if entryID, ok := subscriptionScheduleEntries[sub.ID]; ok {
		subscriptionCron.Remove(entryID)
		delete(subscriptionScheduleEntries, sub.ID)
	}
	entryID, err := subscriptionCron.AddFunc(strings.TrimSpace(sub.CronExpr), func() {
		if err := RefreshSubscription(sub.ID); err != nil {
			log.Errorf("subscription %s scheduled refresh error: %v", sub.ID, err)
		}
	})
	if err != nil {
		return err
	}
	subscriptionScheduleEntries[sub.ID] = entryID
	return nil
}

// RemoveSubscriptionSchedule 删除指定订阅的自动更新调度项。
func RemoveSubscriptionSchedule(id string) {
	subscriptionScheduleMu.Lock()
	defer subscriptionScheduleMu.Unlock()
	if entryID, ok := subscriptionScheduleEntries[id]; ok {
		subscriptionCron.Remove(entryID)
		delete(subscriptionScheduleEntries, id)
	}
}

// RefreshSubscription 尝试启动订阅刷新任务。已在刷新时返回错误。
func RefreshSubscription(id string) error {
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
	subscriptionLifecycleMu.Lock()
	ctx := subscriptionCtx
	if ctx == nil || ctx.Err() != nil {
		subscriptionLifecycleMu.Unlock()
		mu.Unlock()
		return fmt.Errorf("subscription scheduler is stopped")
	}
	subscriptionRefreshWG.Add(1)
	subscriptionLifecycleMu.Unlock()

	go func() {
		defer func() {
			runtime.GC()
			mu.Unlock()
			subscriptionRefreshWG.Done()
		}()

		proxy := sub.ProxyUrl
		switch sub.ProxyMode {
		case model.ProxyModeDisabled:
			proxy = "direct"
		case model.ProxyModeAuto:
			proxy = ""
		}

		var err error
		urls := sub.Url
		if sub.UrlType == model.URLTypeList {
			if proxy == "" {
				urls, err = fetchUrlList(ctx, sub.Url[0], sub.Header, "direct")
				if err != nil {
					urls, err = fetchUrlList(ctx, sub.Url[0], sub.Header, sub.ProxyUrl)
				}
			} else {
				urls, err = fetchUrlList(ctx, sub.Url[0], sub.Header, proxy)
			}
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				RefreshEvents.Emit("failed", RefreshEvent{
					SubID: id,
					Type:  "failed",
					Payload: map[string]any{
						"message": fmt.Sprintf("parse urls: %v", err),
					},
				})
				return
			}
		}

		total := len(urls)
		status := model.SubscriptionStatus{RefreshedAt: time.Now()}

		for i, u := range urls {
			if ctx.Err() != nil {
				return
			}
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
				urlStatus, err = refreshOne(ctx, id, u, sub.Header, "direct")
				if err != nil {
					urlStatus, err = refreshOne(ctx, id, u, sub.Header, sub.ProxyUrl)
				}
			} else {
				urlStatus, err = refreshOne(ctx, id, u, sub.Header, proxy)
			}

			if err != nil {
				if ctx.Err() != nil {
					return
				}
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
func fetchUrlList(ctx context.Context, listUrl string, header map[string]string, proxy string) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, listUrl, nil)
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

// refreshOne 拉取单个 URL 的订阅内容并写入节点池，返回流量状态。
func refreshOne(ctx context.Context, subID, rawUrl string, header map[string]string, proxy string) (model.SubscriptionStatus, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawUrl, nil)
	if err != nil {
		return model.SubscriptionStatus{}, fmt.Errorf("create request: %w", err)
	}
	for k, v := range header {
		req.Header.Set(k, v)
	}

	resp, err := rhttp.New(proxy).Do(req)
	if err != nil {
		return model.SubscriptionStatus{}, fmt.Errorf("fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return model.SubscriptionStatus{}, fmt.Errorf("status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.SubscriptionStatus{}, fmt.Errorf("read body: %w", err)
	}

	// 解析 subscription-userinfo 响应头。
	var status model.SubscriptionStatus
	for _, field := range strings.Split(strings.ToLower(resp.Header.Get("subscription-userinfo")), ";") {
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

	convBody, err := node.Convert(ctx, body, node.ConvertTargetMihomo)
	if err != nil {
		return status, fmt.Errorf("convert: %w", err)
	}

	for _, line := range bytes.Split(convBody, []byte("\n"))[1:] {
		raw := bytes.TrimSpace(bytes.TrimPrefix(line, []byte("  - ")))
		if len(raw) == 0 {
			continue
		}
		node.PoolAdd(subID, raw)
	}
	return status, nil
}
