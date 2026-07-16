package node

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/bestruirui/bestsub/internal/model"
	"github.com/bestruirui/bestsub/internal/rhttp"
	"github.com/bestruirui/bestsub/internal/store"
)

type ConvertTarget string // 节点订阅转换的目标客户端格式。

const (
	ConvertTargetQuantumX     ConvertTarget = "QuantumultX"  // Quantumult X 格式。
	ConvertTargetSurge        ConvertTarget = "Surge"        // Surge 格式。
	ConvertTargetLoon         ConvertTarget = "Loon"         // Loon 格式。
	ConvertTargetSurgeMac     ConvertTarget = "SurgeMac"     // Surge Mac 格式。
	ConvertTargetMihomo       ConvertTarget = "Mihomo"       // Mihomo 格式。
	ConvertTargetURI          ConvertTarget = "URI"          // 通用 URI 格式。
	ConvertTargetV2Ray        ConvertTarget = "V2Ray"        // V2Ray 格式。
	ConvertTargetShadowRocket ConvertTarget = "ShadowRocket" // Shadowrocket 格式。
	ConvertTargetSurfboard    ConvertTarget = "Surfboard"    // Surfboard 格式。
	ConvertTargetSingbox      ConvertTarget = "singbox"      // sing-box 格式。
	ConvertTargetEgern        ConvertTarget = "Egern"        // Egern 格式。
)

// Convert 通过转换服务将订阅内容转为目标格式，并允许调用方取消请求。
func Convert(ctx context.Context, content []byte, target ConvertTarget) ([]byte, error) {
	convertURL := store.SettingGet(model.SettingSubConvertUrl)
	if convertURL == "" {
		return nil, fmt.Errorf("sub_convert_url not configured")
	}

	proxy := "direct"
	if store.SettingGet(model.SettingSubConvertProxy) == "1" {
		// 空值表示让 rhttp 使用全局代理配置，而不是把开关值当成代理地址。
		proxy = ""
	}

	requestBody, _ := json.Marshal(map[string]string{"client": string(target), "data": string(content)})
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, convertURL, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("convert request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := rhttp.New(proxy).Do(request)
	if err != nil {
		return nil, fmt.Errorf("convert: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		errorBody, _ := io.ReadAll(response.Body)
		return nil, fmt.Errorf("convert status: %d, body: %s", response.StatusCode, string(errorBody))
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read convert: %w", err)
	}

	var result struct {
		Status string `json:"status"`
		Data   struct {
			ParseResult string `json:"par_res"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse convert response: %w", err)
	}
	if result.Status != "success" {
		return nil, fmt.Errorf("convert status: %s", result.Status)
	}
	return []byte(result.Data.ParseResult), nil
}
