package probe

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"
)

const TypeDelay ProbeType = "delay"

func init() {
	register(TypeDelay, runDelay)
}

// delayParams 保存延迟探测参数。
type delayParams struct {
	httpParams     // 延迟检测使用的公共 HTTP 参数。
	Attempts   int `json:"attempts,omitempty"` // 请求尝试次数；0 使用 1。
}

func runDelay(ctx context.Context, client *http.Client, raw json.RawMessage, result any) error {
	out, ok := result.(*uint16)
	if !ok || out == nil {
		return fmt.Errorf("probe %s result type mismatch", TypeDelay)
	}

	var params delayParams
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &params); err != nil {
			return err
		}
	}
	if params.Attempts <= 0 {
		params.Attempts = 1
	}
	if params.TimeoutMS <= 0 {
		params.TimeoutMS = 10000
	}
	if params.TimeoutMS > math.MaxUint16 {
		params.TimeoutMS = math.MaxUint16
	}

	var lastErr error // 保留最后一次失败原因，供全部尝试失败时返回。
	var best uint16   // 记录多次成功尝试中的最低延迟。
	for i := 0; i < params.Attempts; i++ {
		// 多次尝试只取最快的成功结果，失败原因保留最后一次用于返回。
		reqCtx, cancel := context.WithTimeout(ctx, time.Duration(params.TimeoutMS)*time.Millisecond)
		req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, params.URL, nil)
		if err != nil {
			cancel()
			lastErr = err
			continue
		}
		startedAt := time.Now()
		resp, err := client.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				cancel()
				return ctx.Err()
			}
			if reqCtx.Err() == context.DeadlineExceeded {
				cancel()
				if best == 0 {
					best = math.MaxUint16
				}
				continue
			}
			cancel()
			lastErr = err
			continue
		}
		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
			resp.Body.Close()
			cancel()
			lastErr = fmt.Errorf("delay status %d", resp.StatusCode)
			continue
		}
		delay := time.Since(startedAt).Milliseconds()
		resp.Body.Close()
		cancel()
		if delay == 0 {
			delay = 1
		}
		if delay > math.MaxUint16 {
			delay = math.MaxUint16
		}
		if best == 0 || uint16(delay) < best {
			best = uint16(delay)
		}
	}
	if best == 0 {
		return lastErr
	}

	*out = best
	return nil
}
