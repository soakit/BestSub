package node

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// Mihomo 将单行 JSON 节点组装为 Mihomo 订阅，并按表达式重命名节点。
func Mihomo(nodes []Node, renameExpression string) ([]byte, error) {
	var content bytes.Buffer // 订阅内容在全部节点处理成功后一次性交给调用方。
	content.WriteString("proxies:\n")
	renameExpression = strings.TrimSpace(renameExpression)
	for i, node := range nodes {
		raw := strings.TrimSpace(node.Raw.Text)
		if renameExpression != "" {
			name, err := Rename(node, uint32(i+1), renameExpression)
			if err != nil {
				return nil, fmt.Errorf("rename node %d: %w", i+1, err)
			}
			nameField := namePattern.FindStringIndex(raw)
			if nameField == nil {
				return nil, fmt.Errorf("node %d name is required", i+1)
			}
			raw = raw[:nameField[0]] + `"name":` + strconv.Quote(name) + raw[nameField[1]:]
		}
		content.WriteString("  - ")
		content.WriteString(raw)
		content.WriteByte('\n')
	}
	return content.Bytes(), nil
}
