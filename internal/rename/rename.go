package rename

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/bestruirui/bestsub/internal/model"
	"github.com/bestruirui/bestsub/pkg/countries"
)

type templateData struct { // 重命名模板的渲染上下文。
	Index             uint32            // 最终输出顺序中的序号。
	Delay             uint32            // 延迟，单位毫秒。
	DownloadSpeed     uint32            // 下载速度，单位 kb/s。
	TrafficMultiplier float32           // 流量扣费倍率。
	Country           countries.Country // 节点落地国家信息。
}

// Rename 按 Go 模板渲染节点名称。
func Rename(info model.NodeInfo, index uint32, expression string) (string, error) {
	tmpl, err := template.New("node").Funcs(template.FuncMap{
		"add": func(x, y uint32) uint32 { return x + y },
		"sub": func(x, y uint32) uint32 {
			if x < y {
				return 0
			}
			return x - y
		},
		"mul": func(x, y uint32) uint32 { return x * y },
		"div": func(x, y uint32) uint32 {
			if y == 0 {
				return 0
			}
			return x / y
		},
		"mod": func(x, y uint32) uint32 {
			if y == 0 {
				return 0
			}
			return x % y
		},
	}).Option("missingkey=error").Parse(expression)
	if err != nil {
		return "", fmt.Errorf("parse rename template: %w", err)
	}

	var result bytes.Buffer
	if err := tmpl.Execute(&result, templateData{
		Index:             index,
		Delay:             uint32(info.Delay),
		DownloadSpeed:     info.DownloadSpeed,
		TrafficMultiplier: info.TrafficMultiplier,
		Country:           countries.Get(info.CountryCode),
	}); err != nil {
		return "", fmt.Errorf("execute rename template: %w", err)
	}
	return result.String(), nil
}
