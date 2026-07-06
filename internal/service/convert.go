package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"bestsub/internal/model"
	"bestsub/internal/rhttp"
	"bestsub/internal/store"
)

// Convert 通过转换服务将订阅内容转为目标格式
func Convert(content []byte, target string) ([]byte, error) {
	convertUrl := store.SettingGet(model.SettingSubConvertUrl)
	if convertUrl == "" {
		return nil, fmt.Errorf("sub_convert_url not configured")
	}

	proxy := "direct"
	if store.SettingGet(model.SettingSubConvertProxy) == "1" {
		proxy = store.SettingGet(model.SettingSubConvertProxy)
	}

	reqBody, _ := json.Marshal(map[string]string{"client": target, "data": string(content)})

	req, err := http.NewRequest(http.MethodPost, convertUrl, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("convert request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := rhttp.New(proxy).Do(req)
	if err != nil {
		return nil, fmt.Errorf("convert: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("convert status: %d, body: %s", resp.StatusCode, string(errBody))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read convert: %w", err)
	}

	var result struct {
		Status string `json:"status"`
		Data   struct {
			ParRes string `json:"par_res"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse convert response: %w", err)
	}
	if result.Status != "success" {
		return nil, fmt.Errorf("convert status: %s", result.Status)
	}
	return []byte(result.Data.ParRes), nil
}
