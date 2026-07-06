package probe

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type ProbeType string

type runner func(context.Context, *http.Client, json.RawMessage, any) error

var runners = map[ProbeType]runner{}

type httpParams struct {
	URL              string `json:"url"`                          // 检测请求目标地址。
	TimeoutMS        int    `json:"timeout_ms,omitempty"`         // 单次请求超时时间，单位毫秒；0 使用模块默认值。
	SuccessStatusMin int    `json:"success_status_min,omitempty"` // 认为请求成功的最小 HTTP 状态码；0 使用 200。
	SuccessStatusMax int    `json:"success_status_max,omitempty"` // 认为请求成功的最大 HTTP 状态码；0 使用 399。
}

func register(typ ProbeType, run runner) {
	if typ == "" {
		panic("probe type is required")
	}
	if run == nil {
		panic("probe runner is required")
	}
	if _, ok := runners[typ]; ok {
		panic("probe type registered: " + string(typ))
	}
	runners[typ] = run
}

func Run(ctx context.Context, typ ProbeType, client *http.Client, params json.RawMessage, result any) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if client == nil {
		return fmt.Errorf("client is required")
	}
	if result == nil {
		return fmt.Errorf("result is required")
	}
	run, ok := runners[typ]
	if !ok {
		return fmt.Errorf("unknown probe type: %s", typ)
	}
	err := run(ctx, client, params, result)
	if err := ctx.Err(); err != nil {
		return err
	}
	return err
}

func (p *httpParams) withDefaults(timeoutMS int) {
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

func (p httpParams) timeout() time.Duration {
	return time.Duration(p.TimeoutMS) * time.Millisecond
}

func (p httpParams) validate() error {
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

func (p httpParams) statusOK(code int) bool {
	return code >= p.SuccessStatusMin && code <= p.SuccessStatusMax
}
