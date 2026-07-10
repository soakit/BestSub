package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/bestruirui/bestsub/internal/rhttp"
)

const TypeWebDAV Type = "webdav"

type webDAVBackend struct{}

func init() {
	register(TypeWebDAV, webDAVBackend{})
}

type webDAVParams struct { // WebDAV 储存参数
	Endpoint string `json:"endpoint"`           // WebDAV 根地址。
	Username string `json:"username,omitempty"` // Basic Auth 用户名。
	Password string `json:"password,omitempty"` // Basic Auth 密码。
}

func parseWebDAVParams(raw json.RawMessage) (webDAVParams, error) {
	var params webDAVParams
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &params); err != nil {
			return params, err
		}
	}
	return params, nil
}

func webDAVEndpoint(endpoint string) (*url.URL, error) {
	u, err := url.Parse(strings.TrimSpace(endpoint))
	if err != nil {
		return nil, err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("url scheme must be http or https")
	}
	if u.Host == "" {
		return nil, fmt.Errorf("url host is required")
	}
	return u, nil
}

func (webDAVBackend) Put(ctx context.Context, req Request) error {
	params, err := parseWebDAVParams(req.Params)
	if err != nil {
		return err
	}
	u, err := webDAVEndpoint(params.Endpoint)
	if err != nil {
		return err
	}
	if strings.HasSuffix(u.Path, "/") {
		u.Path += strings.TrimPrefix(req.Path, "/")
	} else if strings.HasPrefix(req.Path, "/") {
		u.Path += req.Path
	} else {
		u.Path += "/" + req.Path
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPut, u.String(), bytes.NewReader(req.Content))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "text/plain; charset=utf-8")
	if params.Username != "" {
		httpReq.SetBasicAuth(params.Username, params.Password)
	}
	resp, err := rhttp.New("").Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("webdav put status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}

func (webDAVBackend) Test(ctx context.Context, raw json.RawMessage) error {
	params, err := parseWebDAVParams(raw)
	if err != nil {
		return err
	}
	u, err := webDAVEndpoint(params.Endpoint)
	if err != nil {
		return err
	}
	httpReq, err := http.NewRequestWithContext(ctx, "PROPFIND", u.String(), nil)
	if err != nil {
		return err
	}
	httpReq.Header.Set("Depth", "0")
	if params.Username != "" {
		httpReq.SetBasicAuth(params.Username, params.Password)
	}
	resp, err := rhttp.New("").Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("webdav test status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}
