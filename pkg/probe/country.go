package probe

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const TypeCountry ProbeType = "country" // 国家代码探测类型。

func init() {
	register(TypeCountry, runCountryProbe)
}

func runCountryProbe(ctx context.Context, client *http.Client, raw json.RawMessage, result any) error {
	out, ok := result.(*string)
	if !ok || out == nil {
		return fmt.Errorf("probe %s result type mismatch", TypeCountry)
	}

	var params httpParams
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &params); err != nil {
			return err
		}
	}
	if params.TimeoutMS <= 0 {
		params.TimeoutMS = 10000
	}
	const desktopChromeUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36" // 三个来源共用的桌面版 Chrome 标识。

	// 每个来源使用独立超时，避免前一个请求失败后耗尽后续备用请求的时间。
	requestJSON := func(url string, headers map[string]string, body any) error {
		reqCtx, cancel := context.WithTimeout(ctx, time.Duration(params.TimeoutMS)*time.Millisecond)
		defer cancel()
		req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		for name, value := range headers {
			req.Header.Set(name, value)
		}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
			return fmt.Errorf("status %d", resp.StatusCode)
		}
		if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(body); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}
		return nil
	}

	var cloudflareBody struct {
		Country string `json:"country"` // Cloudflare 返回的二字母国家代码。
	}
	cloudflareErr := requestJSON(
		"https://speed.cloudflare.com/meta",
		map[string]string{
			"Referer":    "https://speed.cloudflare.com",
			"User-Agent": desktopChromeUA,
		},
		&cloudflareBody,
	) // 记录首选来源失败原因，备用来源也失败时一并返回。
	if cloudflareErr == nil && cloudflareBody.Country != "" {
		*out = cloudflareBody.Country
		return nil
	}
	if cloudflareErr == nil {
		cloudflareErr = fmt.Errorf("country is empty")
	}

	var ipSBBody struct {
		CountryCode string `json:"country_code"` // api.ip.sb 返回的二字母国家代码。
	}
	ipSBErr := requestJSON("https://api.ip.sb/geoip", map[string]string{"User-Agent": desktopChromeUA}, &ipSBBody) // 记录第一备用来源失败原因，最终失败时一并返回。
	if ipSBErr == nil && ipSBBody.CountryCode != "" {
		*out = ipSBBody.CountryCode
		return nil
	}
	if ipSBErr == nil {
		ipSBErr = fmt.Errorf("country is empty")
	}

	var myIPBody struct {
		CountryCode string `json:"cc"` // api.myip.com 返回的二字母国家代码。
	}
	if err := requestJSON("https://api.myip.com", map[string]string{"User-Agent": desktopChromeUA}, &myIPBody); err != nil {
		return fmt.Errorf("cloudflare country failed: %v; api.ip.sb country failed: %v; api.myip.com country failed: %w", cloudflareErr, ipSBErr, err)
	}
	if myIPBody.CountryCode == "" {
		return fmt.Errorf("cloudflare country failed: %v; api.ip.sb country failed: %v; api.myip.com country is empty", cloudflareErr, ipSBErr)
	}
	*out = myIPBody.CountryCode
	return nil
}
