package rhttp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/bestruirui/bestsub/internal/model"
	"github.com/bestruirui/bestsub/internal/store"
	"github.com/bestruirui/bestsub/pkg/mihomo"
)

// New 返回一个 HTTP 客户端。proxy 参数控制代理行为：
//   - "" 使用系统全局代理设置
//   - "direct" 直连，不走代理
//   - 其他值作为代理 URL（如 "socks5://127.0.0.1:1080"）
func New(proxy string) *http.Client {
	client := &http.Client{Timeout: 30 * time.Second}

	if proxy == "direct" {
		return client
	}

	proxyUrl := proxy
	if proxyUrl == "" {
		if store.SettingGet(model.SettingProxyEnable) != "1" {
			return client
		}
		proxyUrl = store.SettingGet(model.SettingProxyURL)
	}
	if proxyUrl == "" {
		return client
	}

	u, err := url.Parse(proxyUrl)
	if err != nil {
		return client
	}

	proxyType := "http"
	tls := false
	switch u.Scheme {
	case "socks5":
		proxyType = "socks5"
	case "https":
		tls = true
	}

	raw := fmt.Appendf(nil, "{name: bestsub, type: %s, server: %s, port: %s", proxyType, u.Hostname(), u.Port())
	if tls {
		raw = append(raw, ", tls: true"...)
	}
	if u.User != nil {
		raw = fmt.Appendf(raw, ", username: %s", u.User.Username())
		if p, ok := u.User.Password(); ok {
			raw = fmt.Appendf(raw, ", password: %s", p)
		}
	}
	raw = append(raw, '}')

	iface := store.SettingGet(model.SettingBindInterface)
	transport, _, err := mihomo.NewTransport(raw, nil, iface)
	if err != nil {
		return client
	}
	client.Transport = transport
	return client
}

// Get 使用指定代理和请求头发起 GET 请求，返回响应体与响应头。
func Get(ctx context.Context, rawURL, proxy string, header map[string]string) ([]byte, http.Header, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("create request: %w", err)
	}
	for key, value := range header {
		req.Header.Set(key, value)
	}

	resp, err := New(proxy).Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("fetch: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("read body: %w", err)
	}
	return body, resp.Header, nil
}
