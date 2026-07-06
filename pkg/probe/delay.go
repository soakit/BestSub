package probe

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const TypeDelay ProbeType = "delay"

func init() {
	register(TypeDelay, runDelay)
}

type delayParams struct {
	httpParams        // 延迟检测使用的公共 HTTP 参数。
	Method     string `json:"method,omitempty"`   // 请求方法；为空使用 GET。
	Attempts   int    `json:"attempts,omitempty"` // 请求尝试次数；0 使用 1。
}

func runDelay(ctx context.Context, client *http.Client, raw json.RawMessage, result any) error {
	out, ok := result.(*uint64)
	if !ok || out == nil {
		return fmt.Errorf("probe %s result type mismatch", TypeDelay)
	}

	var params delayParams
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &params); err != nil {
			return err
		}
	}
	params.withDefaults()
	if err := params.httpParams.validate(); err != nil {
		return err
	}

	var lastErr error
	var best uint64
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
		return lastErr
	}

	*out = best
	return nil
}

func (p *delayParams) withDefaults() {
	if p.Method == "" {
		p.Method = http.MethodGet
	}
	p.Method = strings.ToUpper(p.Method)
	if p.Attempts <= 0 {
		p.Attempts = 1
	}
	p.httpParams.withDefaults(10000)
}

func runDelayAttempt(ctx context.Context, client *http.Client, params delayParams) (uint64, error) {
	reqCtx, cancel := context.WithTimeout(ctx, params.httpParams.timeout())
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
	if !params.httpParams.statusOK(resp.StatusCode) {
		return 0, fmt.Errorf("delay status %d", resp.StatusCode)
	}
	ms := time.Since(startedAt).Milliseconds()
	if ms <= 0 {
		ms = 1
	}
	return uint64(ms), nil
}
