package node

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Mihomo 将单行 JSON 节点组装为 Mihomo 订阅，并按表达式重命名节点。
func Mihomo(nodes []Node, renameExpression string) ([]byte, error) {
	var content bytes.Buffer // 订阅内容在全部节点处理成功后一次性交给调用方。
	content.WriteString("proxies:\n")
	renameExpression = strings.TrimSpace(renameExpression)
	seenNames := make(map[string]struct{}, len(nodes))
	for i, node := range nodes {
		raw := strings.TrimSpace(node.Raw.Text)
		name, err := nodeName(raw)
		if err != nil {
			return nil, fmt.Errorf("node %d name: %w", i+1, err)
		}
		if renameExpression != "" {
			name, err = Rename(node, uint32(i+1), renameExpression)
			if err != nil {
				return nil, fmt.Errorf("rename node %d: %w", i+1, err)
			}
		}
		name = uniqueProxyName(seenNames, name)
		raw, err = replaceNodeName(raw, name)
		if err != nil {
			return nil, fmt.Errorf("node %d name: %w", i+1, err)
		}
		content.WriteString("  - ")
		content.WriteString(raw)
		content.WriteByte('\n')
	}
	return content.Bytes(), nil
}

func nodeName(raw string) (string, error) {
	var item struct {
		Name string `yaml:"name"`
	}
	if err := yaml.Unmarshal([]byte(raw), &item); err != nil {
		return "", err
	}
	if item.Name == "" {
		return "", fmt.Errorf("name is required")
	}
	return item.Name, nil
}

func replaceNodeName(raw, name string) (string, error) {
	nameField := namePattern.FindStringIndex(raw)
	if nameField == nil {
		return "", fmt.Errorf("name field is required")
	}
	return raw[:nameField[0]] + `"name":` + strconv.Quote(name) + raw[nameField[1]:], nil
}

// uniqueProxyName 为 Clash/Mihomo 导出保证节点名称唯一。
func uniqueProxyName(seen map[string]struct{}, name string) string {
	base := strings.TrimSpace(name)
	if base == "" {
		base = "node"
	}
	candidate := base
	for i := 2; ; i++ {
		if _, ok := seen[candidate]; !ok {
			seen[candidate] = struct{}{}
			return candidate
		}
		candidate = fmt.Sprintf("%s-%d", base, i)
	}
}
