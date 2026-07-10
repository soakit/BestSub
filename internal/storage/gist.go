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

const TypeGist Type = "gist"

type gistBackend struct{}

func init() {
	register(TypeGist, gistBackend{})
}

type gistParams struct { // GitHub Gist 储存参数
	Token  string `json:"token"`   // GitHub token，必须具备 gist 写权限。
	GistID string `json:"gist_id"` // 目标 Gist ID。
}

func parseGistParams(raw json.RawMessage) (gistParams, error) {
	var params gistParams
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &params); err != nil {
			return params, err
		}
	}
	if strings.TrimSpace(params.Token) == "" {
		return params, fmt.Errorf("gist token is required")
	}
	if strings.TrimSpace(params.GistID) == "" {
		return params, fmt.Errorf("gist id is required")
	}
	return params, nil
}

func (gistBackend) Put(ctx context.Context, req Request) error {
	params, err := parseGistParams(req.Params)
	if err != nil {
		return err
	}
	body, err := json.Marshal(map[string]any{
		"files": map[string]any{
			req.Path: map[string]string{"content": string(req.Content)},
		},
	})
	if err != nil {
		return err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPatch, "https://api.github.com/gists/"+url.PathEscape(params.GistID), bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Accept", "application/vnd.github+json")
	httpReq.Header.Set("Authorization", "Bearer "+params.Token)
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := rhttp.New("").Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("gist patch status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}

func (gistBackend) Test(ctx context.Context, raw json.RawMessage) error {
	params, err := parseGistParams(raw)
	if err != nil {
		return err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/gists/"+url.PathEscape(params.GistID), nil)
	if err != nil {
		return err
	}
	httpReq.Header.Set("Accept", "application/vnd.github+json")
	httpReq.Header.Set("Authorization", "Bearer "+params.Token)
	resp, err := rhttp.New("").Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("gist test status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}
