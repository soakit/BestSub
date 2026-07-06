package rhttp

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"bestsub/internal/mihomo"
	"bestsub/internal/model"
	"bestsub/internal/store"
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

	raw := fmt.Appendf(nil, "{type: %s, server: %s, port: %s", proxyType, u.Hostname(), u.Port())
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
	transport, err := mihomo.NewTransport(raw, iface)
	if err != nil {
		return client
	}
	client.Transport = transport
	return client
}
