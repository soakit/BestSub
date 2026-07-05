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

type Country struct {
	HTTPParams          // 国家检测使用的公共 HTTP 参数。
	CountryField string `json:"country_field,omitempty"` // 响应 JSON 中的国家代码字段名；为空使用 country_code。
}

func (params Country) Run(ctx context.Context, client *http.Client) (NodeInfoPatch, error) {
	params.withDefaults()
	if err := params.HTTPParams.validate(); err != nil {
		return NodeInfoPatch{}, err
	}

	code, err := runCountry(ctx, client, params)
	if err != nil {
		return NodeInfoPatch{}, err
	}
	return NodeInfoPatch{CountryCode: &code}, nil
}

func (p *Country) withDefaults() {
	if p.CountryField == "" {
		p.CountryField = "country_code"
	}
	p.HTTPParams.withDefaults(10000)
}

func runCountry(ctx context.Context, client *http.Client, params Country) (string, error) {
	reqCtx, cancel := context.WithTimeout(ctx, params.HTTPParams.timeout())
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
	if !params.HTTPParams.statusOK(resp.StatusCode) {
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
