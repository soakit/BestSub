package probe

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

var countryCodeRE = regexp.MustCompile(`^[A-Za-z]{2}$`)

const TypeCountry ProbeType = "country"

func init() {
	register(TypeCountry, runCountryProbe)
}

type countryParams struct {
	httpParams          // 国家检测使用的公共 HTTP 参数。
	CountryField string `json:"country_field,omitempty"` // 响应 JSON 中的国家代码字段名；为空使用 country_code。
}

func runCountryProbe(ctx context.Context, client *http.Client, raw json.RawMessage, result any) error {
	out, ok := result.(*string)
	if !ok || out == nil {
		return fmt.Errorf("probe %s result type mismatch", TypeCountry)
	}

	var params countryParams
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &params); err != nil {
			return err
		}
	}
	params.withDefaults()
	if err := params.httpParams.validate(); err != nil {
		return err
	}

	code, err := runCountry(ctx, client, params)
	if err != nil {
		return err
	}
	*out = code
	return nil
}

func (p *countryParams) withDefaults() {
	if p.CountryField == "" {
		p.CountryField = "country_code"
	}
	p.httpParams.withDefaults(10000)
}

func runCountry(ctx context.Context, client *http.Client, params countryParams) (string, error) {
	reqCtx, cancel := context.WithTimeout(ctx, params.httpParams.timeout())
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, params.URL, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if !params.httpParams.statusOK(resp.StatusCode) {
		return "", fmt.Errorf("country status %d", resp.StatusCode)
	}

	var fields map[string]any
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&fields); err != nil {
		return "", fmt.Errorf("parse country response: %w", err)
	}
	code, err := jsonStringField(fields, params.CountryField)
	if err != nil {
		return "", err
	}
	code = strings.ToUpper(strings.TrimSpace(code))
	if !countryCodeRE.MatchString(code) {
		return "", fmt.Errorf("invalid country code: %s", code)
	}

	return code, nil
}

func jsonStringField(fields map[string]any, name string) (string, error) {
	v, ok := fields[name]
	if !ok {
		return "", fmt.Errorf("json field %s is required", name)
	}
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("json field %s must be string", name)
	}
	return s, nil
}
