package probe

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"
)

type Delay struct {
	HTTPParams        // 延迟检测使用的公共 HTTP 参数。
	Method     string `json:"method,omitempty"`   // 请求方法；为空使用 GET。
	Attempts   int    `json:"attempts,omitempty"` // 请求尝试次数；0 使用 1。
}

func (params Delay) Run(ctx context.Context, client *http.Client) (NodeInfoPatch, error) {
	params.withDefaults()
	if err := params.HTTPParams.validate(); err != nil {
		return NodeInfoPatch{}, err
	}

	var lastErr error
	var best uint16
	for i := 0; i < params.Attempts; i++ {
		// 多次尝试只取最快的成功结果，失败原因保留最后一次用于返回。
		delay, err := runDelayAttempt(ctx, client, params)
		if err != nil {
			lastErr = err
			continue
		}
		if best == 0 || delay < best {
			best = delay
		}
	}
	if best == 0 {
		return NodeInfoPatch{}, lastErr
	}

	return NodeInfoPatch{Delay: &best}, nil
}

func (p *Delay) withDefaults() {
	if p.Method == "" {
		p.Method = http.MethodGet
	}
	p.Method = strings.ToUpper(p.Method)
	if p.Attempts <= 0 {
		p.Attempts = 1
	}
	p.HTTPParams.withDefaults(10000)
}

func runDelayAttempt(ctx context.Context, client *http.Client, params Delay) (uint16, error) {
	reqCtx, cancel := context.WithTimeout(ctx, params.HTTPParams.timeout())
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, params.Method, params.URL, nil)
	if err != nil {
		return 0, err
	}
	startedAt := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if !params.HTTPParams.statusOK(resp.StatusCode) {
		return 0, fmt.Errorf("delay status %d", resp.StatusCode)
	}
	ms := time.Since(startedAt).Milliseconds()
	if ms <= 0 {
		ms = 1
	}
	if ms > math.MaxUint16 {
		return 0, fmt.Errorf("delay %dms exceeds uint16 range", ms)
	}
	return uint16(ms), nil
}
