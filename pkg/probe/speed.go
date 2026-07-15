package probe

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const TypeSpeed ProbeType = "speed"

func init() {
	register(TypeSpeed, runSpeedProbe)
}

// speedParams 保存下载测速参数。
type speedParams struct {
	httpParams           // 下载测速使用的公共 HTTP 参数。
	MaxKB         uint64 `json:"max_kb,omitempty"`          // 最大读取量，单位 kb；0 使用模块默认值。
	MaxDurationMS uint   `json:"max_duration_ms,omitempty"` // 最大测速读取时长，单位毫秒；0 使用模块默认值。
}

func runSpeedProbe(ctx context.Context, client *http.Client, raw json.RawMessage, result any) error {
	out, ok := result.(*uint32)
	if !ok || out == nil {
		return fmt.Errorf("probe %s result type mismatch", TypeSpeed)
	}

	var params speedParams
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &params); err != nil {
			return err
		}
	}
	if params.MaxKB == 0 {
		params.MaxKB = 32 << 10
	}
	if params.MaxDurationMS == 0 {
		params.MaxDurationMS = 10000
	}
	if params.TimeoutMS <= 0 {
		params.TimeoutMS = 30000
	}
	if params.MaxKB > ^uint64(0)/1024 {
		return fmt.Errorf("max kb exceeds supported limit")
	}
	// time.Duration 使用有符号纳秒值，先限制毫秒数避免转换和乘法溢出。
	if uint64(params.MaxDurationMS) > (^uint64(0)>>1)/uint64(time.Millisecond) {
		return fmt.Errorf("max duration exceeds supported limit")
	}
	reqCtx, cancel := context.WithTimeout(ctx, time.Duration(params.TimeoutMS)*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, params.URL, nil)
	if err != nil {
		return err
	}
	startedAt := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("speed status %d", resp.StatusCode)
	}

	buf := make([]byte, 32*1024)
	var bytesRead uint64 // 记录本次测速已读取的总字节数。
	// 测速只读取到任一上限，避免小文件读完或大文件无限下载影响任务调度。
	for bytesRead < params.MaxKB*1024 && time.Since(startedAt) < time.Duration(params.MaxDurationMS)*time.Millisecond {
		if err := reqCtx.Err(); err != nil {
			return err
		}
		need := uint64(len(buf))
		if remain := params.MaxKB*1024 - bytesRead; remain < need {
			need = remain
		}
		n, err := resp.Body.Read(buf[:need])
		bytesRead += uint64(n)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	elapsed := time.Since(startedAt)
	if bytesRead == 0 {
		return fmt.Errorf("speed read 0 kb")
	}
	if elapsed <= 0 {
		elapsed = time.Millisecond
	}
	// 下载速度统一按 kb/s 输出，截断结果与已有整数速度保持一致。
	speedFloat := float64(bytesRead) / elapsed.Seconds() / 1024
	if speedFloat > float64(^uint32(0)) {
		return fmt.Errorf("speed overflow")
	}
	*out = uint32(speedFloat)
	return nil
}
