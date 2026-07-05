package probe

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Request struct {
	Client *http.Client // 本次探测使用的 HTTP 客户端，由调用方负责设置代理链路和传输层。
	Probe  Probe        // 本次执行的探测项。
}

type HTTPParams struct {
	URL              string `json:"url"`                          // 检测请求目标地址。
	TimeoutMS        int    `json:"timeout_ms,omitempty"`         // 单次请求超时时间，单位毫秒；0 使用模块默认值。
	SuccessStatusMin int    `json:"success_status_min,omitempty"` // 认为请求成功的最小 HTTP 状态码；0 使用 200。
	SuccessStatusMax int    `json:"success_status_max,omitempty"` // 认为请求成功的最大 HTTP 状态码；0 使用 399。
}

type NodeInfoPatch struct {
	Delay         *uint16 `json:"delay,omitempty"`          // 延迟结果，单位毫秒；nil 表示不更新。
	DownloadSpeed *uint64 `json:"download_speed,omitempty"` // 下载速度结果，单位 bytes/s；nil 表示不更新。
	CountryCode   *string `json:"country_code,omitempty"`   // 落地国家代码，ISO 3166-1 alpha-2；nil 表示不更新。
}

type Probe interface {
	Run(context.Context, *http.Client) (NodeInfoPatch, error)
}

func Run(ctx context.Context, req Request) (NodeInfoPatch, error) {
	if err := ctx.Err(); err != nil {
		return NodeInfoPatch{}, err
	}
	if req.Client == nil {
		return NodeInfoPatch{}, fmt.Errorf("client is required")
	}
	if req.Probe == nil {
		return NodeInfoPatch{}, fmt.Errorf("probe is required")
	}
	patch, err := req.Probe.Run(ctx, req.Client)
	if err := ctx.Err(); err != nil {
		return NodeInfoPatch{}, err
	}
	return patch, err
}

func (p *HTTPParams) withDefaults(timeoutMS int) {
	if p.TimeoutMS <= 0 {
		p.TimeoutMS = timeoutMS
	}
	if p.SuccessStatusMin == 0 {
		p.SuccessStatusMin = http.StatusOK
	}
	if p.SuccessStatusMax == 0 {
		p.SuccessStatusMax = 399
	}
}

func (p HTTPParams) timeout() time.Duration {
	return time.Duration(p.TimeoutMS) * time.Millisecond
}

func (p HTTPParams) validate() error {
	u, err := url.Parse(p.URL)
	if err != nil {
		return err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("url scheme must be http or https")
	}
	if u.Host == "" {
		return fmt.Errorf("url host is required")
	}
	return nil
}

func (p HTTPParams) statusOK(code int) bool {
	return code >= p.SuccessStatusMin && code <= p.SuccessStatusMax
}
