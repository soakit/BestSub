package node

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/bestruirui/bestsub/internal/model"
	"github.com/cespare/xxhash/v2"
	"gopkg.in/yaml.v3"
)

var (
	namePattern              = regexp.MustCompile(`"name":"(?:\\.|[^"\\])*"`)                                                                             // 匹配转换后的单行 JSON 节点名称字段。
	trafficMultiplierPattern = regexp.MustCompile(`(?i)(?:倍率\s*([0-9]+(?:\.[0-9]+)?)\b|\bx\s*([0-9]+(?:\.[0-9]+)?)\b|([0-9]+(?:\.[0-9]+)?)\s*(?:x\b|倍))`) // 匹配 0.5x、x0.5、倍率 0.5 和 0.5倍。
)

type Raw struct { // 节点池、任务结果和分享共用的不可变节点原文。
	Text        string // 单条 Mihomo YAML 节点内容。
	Fingerprint uint64 // 按节点身份字段生成的节点指纹。
}

type Node struct { // 节点在输入、检测和输出流程中传递的运行时状态。
	SubscriptionID string         // 订阅节点来源 ID，非空时允许将检测结果写回节点池。
	NodeID         string         // 单独节点来源 ID，非空时允许将检测结果写回数据库。
	Raw            *Raw           // 节点原文及指纹。
	Info           model.NodeInfo // 节点附加信息。
}

// Fingerprint 从 YAML 行内节点中解析身份字段并计算指纹。
func Fingerprint(raw []byte) uint64 {
	var key struct {
		Type       string `yaml:"type"`
		Server     string `yaml:"server"`
		Port       any    `yaml:"port"`
		Password   string `yaml:"password"`
		UUID       string `yaml:"uuid"`
		ServerName string `yaml:"servername"`
	}
	_ = yaml.Unmarshal(raw, &key)
	return xxhash.Sum64String(fmt.Sprintf("%v", key))
}

// ParseTrafficMultiplier 从规范化节点名称提取流量倍率，无有效标注时返回 1。
func ParseTrafficMultiplier(raw string) float32 {
	matches := trafficMultiplierPattern.FindStringSubmatch(namePattern.FindString(raw))
	if len(matches) == 0 {
		return 1
	}
	for _, value := range matches[1:] {
		if value == "" {
			continue
		}
		multiplier, err := strconv.ParseFloat(value, 32)
		if err == nil && multiplier > 0 {
			return float32(multiplier)
		}
		return 1
	}
	return 1
}
