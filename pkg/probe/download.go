package probe

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const TypeDownload ProbeType = "download"

func init() {
	register(TypeDownload, runDownloadProbe)
}

type downloadParams struct {
	httpParams          // 下载测速使用的公共 HTTP 参数。
	MaxBytes      int64 `json:"max_bytes,omitempty"`       // 最大读取字节数；0 使用模块默认值。
	MaxDurationMS int   `json:"max_duration_ms,omitempty"` // 最大测速读取时长，单位毫秒；0 使用模块默认值。
}

func runDownloadProbe(ctx context.Context, client *http.Client, raw json.RawMessage, result any) error {
	out, ok := result.(*uint64)
	if !ok || out == nil {
		return fmt.Errorf("probe %s result type mismatch", TypeDownload)
	}

	var params downloadParams
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &params); err != nil {
			return err
		}
	}
	params.withDefaults()
	if err := params.httpParams.validate(); err != nil {
		return err
	}

	speed, err := runDownload(ctx, client, params)
	if err != nil {
		return err
	}
	*out = speed
	return nil
}

func (p *downloadParams) withDefaults() {
	if p.MaxBytes <= 0 {
		p.MaxBytes = 32 << 20
	}
	if p.MaxDurationMS <= 0 {
		p.MaxDurationMS = 10000
	}
	p.httpParams.withDefaults(30000)
}

func runDownload(ctx context.Context, client *http.Client, params downloadParams) (uint64, error) {
	reqCtx, cancel := context.WithTimeout(ctx, params.httpParams.timeout())
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, params.URL, nil)
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
		return 0, fmt.Errorf("download status %d", resp.StatusCode)
	}

	bytesRead, err := readDownloadBody(reqCtx, resp.Body, params.MaxBytes, time.Duration(params.MaxDurationMS)*time.Millisecond, startedAt)
	if err != nil {
		return 0, err
	}
	elapsed := time.Since(startedAt)
	if bytesRead <= 0 {
		return 0, fmt.Errorf("download read 0 bytes")
	}
	if elapsed <= 0 {
		elapsed = time.Millisecond
	}
	speedFloat := float64(bytesRead) / elapsed.Seconds()
	if speedFloat > float64(^uint64(0)) {
		return 0, fmt.Errorf("download speed overflow")
	}
	return uint64(speedFloat), nil
}

func readDownloadBody(ctx context.Context, body io.Reader, maxBytes int64, maxDuration time.Duration, startedAt time.Time) (int64, error) {
	buf := make([]byte, 32*1024)
	var bytesRead int64
	// 测速只读取到任一上限，避免小文件读完或大文件无限下载影响任务调度。
	for bytesRead < maxBytes && time.Since(startedAt) < maxDuration {
		if err := ctx.Err(); err != nil {
			return bytesRead, err
		}
		need := len(buf)
		if remain := maxBytes - bytesRead; remain < int64(need) {
			need = int(remain)
		}
		n, err := body.Read(buf[:need])
		bytesRead += int64(n)
		if err == io.EOF {
			return bytesRead, nil
		}
		if err != nil {
			return bytesRead, err
		}
	}
	return bytesRead, nil
}
