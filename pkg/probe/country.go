package probe

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
)

const TypeCountry ProbeType = "country"

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
	params.URL = "https://speed.cloudflare.com/meta"

	reqCtx, cancel := context.WithTimeout(ctx, time.Duration(params.TimeoutMS)*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, params.URL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Referer", "https://speed.cloudflare.com")
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("resp %v", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		log.Errorf("status %d", resp.StatusCode)
		return fmt.Errorf("cloudflare country status %d", resp.StatusCode)
	}

	var body struct {
		Country string `json:"country"` // Cloudflare 返回的二字母国家代码。
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&body); err != nil {
		return fmt.Errorf("parse cloudflare country response: %w", err)
	}
	log.Infof("cloudflare country %s", body.Country)
	*out = body.Country
	return nil
}
