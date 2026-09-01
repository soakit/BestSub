package node

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	clashGroupProxy = "PROXY"
	clashGroupAuto  = "♻️ 自动选择"
	loyalsoldierCDN = "https://cdn.jsdelivr.net/gh/Loyalsoldier/clash-rules@release"
)

// Mihomo 将单行 JSON 节点组装为 Mihomo 订阅，并按表达式重命名节点。
func Mihomo(nodes []Node, renameExpression string) ([]byte, error) {
	content, _, err := buildProxies(nodes, renameExpression)
	return content, err
}

// MihomoClashProfile 输出可直接导入 Clash/Mihomo 的完整配置（含策略组与规则）。
func MihomoClashProfile(nodes []Node, renameExpression string) ([]byte, error) {
	proxies, names, err := buildProxies(nodes, renameExpression)
	if err != nil {
		return nil, err
	}
	var content bytes.Buffer
	content.WriteString("mode: rule\n")
	content.Write(proxies)
	if err := writeProxyGroups(&content, names); err != nil {
		return nil, err
	}
	if err := writeRuleProviders(&content); err != nil {
		return nil, err
	}
	writeRules(&content)
	return content.Bytes(), nil
}

func buildProxies(nodes []Node, renameExpression string) ([]byte, []string, error) {
	var content bytes.Buffer
	content.WriteString("proxies:\n")
	renameExpression = strings.TrimSpace(renameExpression)
	seenNames := make(map[string]struct{}, len(nodes))
	names := make([]string, 0, len(nodes))
	for i, node := range nodes {
		raw := strings.TrimSpace(node.Raw.Text)
		name, err := nodeName(raw)
		if err != nil {
			return nil, nil, fmt.Errorf("node %d name: %w", i+1, err)
		}
		if renameExpression != "" {
			name, err = Rename(node, uint32(i+1), renameExpression)
			if err != nil {
				return nil, nil, fmt.Errorf("rename node %d: %w", i+1, err)
			}
		}
		name = uniqueProxyName(seenNames, name)
		raw, err = replaceNodeName(raw, name)
		if err != nil {
			return nil, nil, fmt.Errorf("node %d name: %w", i+1, err)
		}
		names = append(names, name)
		content.WriteString("  - ")
		content.WriteString(raw)
		content.WriteByte('\n')
	}
	return content.Bytes(), names, nil
}

type proxyGroup struct {
	Name     string   `yaml:"name"`
	Type     string   `yaml:"type"`
	Proxies  []string `yaml:"proxies,omitempty"`
	URL      string   `yaml:"url,omitempty"`
	Interval int      `yaml:"interval,omitempty"`
}

func writeProxyGroups(content *bytes.Buffer, names []string) error {
	if len(names) == 0 {
		content.WriteString("\nproxy-groups: []\n")
		return nil
	}
	groups := []proxyGroup{
		{
			Name:    clashGroupProxy,
			Type:    "select",
			Proxies: append([]string{clashGroupAuto}, names...),
		},
		{
			Name:     clashGroupAuto,
			Type:     "url-test",
			Proxies:  names,
			URL:      "http://www.gstatic.com/generate_204",
			Interval: 300,
		},
	}
	encoded, err := yaml.Marshal(groups)
	if err != nil {
		return fmt.Errorf("marshal proxy groups: %w", err)
	}
	content.WriteString("\nproxy-groups:\n")
	for _, line := range strings.Split(strings.TrimSuffix(string(encoded), "\n"), "\n") {
		content.WriteString("  ")
		content.WriteString(line)
		content.WriteByte('\n')
	}
	return nil
}

type ruleProvider struct {
	Type     string `yaml:"type"`
	Behavior string `yaml:"behavior"`
	URL      string `yaml:"url"`
	Path     string `yaml:"path"`
	Interval int    `yaml:"interval"`
}

func writeRuleProviders(content *bytes.Buffer) error {
	providers := map[string]ruleProvider{
		"reject": {
			Type: "http", Behavior: "domain",
			URL: loyalsoldierCDN + "/reject.txt", Path: "./ruleset/reject.yaml", Interval: 86400,
		},
		"icloud": {
			Type: "http", Behavior: "domain",
			URL: loyalsoldierCDN + "/icloud.txt", Path: "./ruleset/icloud.yaml", Interval: 86400,
		},
		"apple": {
			Type: "http", Behavior: "domain",
			URL: loyalsoldierCDN + "/apple.txt", Path: "./ruleset/apple.yaml", Interval: 86400,
		},
		"google": {
			Type: "http", Behavior: "domain",
			URL: loyalsoldierCDN + "/google.txt", Path: "./ruleset/google.yaml", Interval: 86400,
		},
		"proxy": {
			Type: "http", Behavior: "domain",
			URL: loyalsoldierCDN + "/proxy.txt", Path: "./ruleset/proxy.yaml", Interval: 86400,
		},
		"direct": {
			Type: "http", Behavior: "domain",
			URL: loyalsoldierCDN + "/direct.txt", Path: "./ruleset/direct.yaml", Interval: 86400,
		},
		"private": {
			Type: "http", Behavior: "domain",
			URL: loyalsoldierCDN + "/private.txt", Path: "./ruleset/private.yaml", Interval: 86400,
		},
		"telegramcidr": {
			Type: "http", Behavior: "ipcidr",
			URL: loyalsoldierCDN + "/telegramcidr.txt", Path: "./ruleset/telegramcidr.yaml", Interval: 86400,
		},
		"cncidr": {
			Type: "http", Behavior: "ipcidr",
			URL: loyalsoldierCDN + "/cncidr.txt", Path: "./ruleset/cncidr.yaml", Interval: 86400,
		},
		"lancidr": {
			Type: "http", Behavior: "ipcidr",
			URL: loyalsoldierCDN + "/lancidr.txt", Path: "./ruleset/lancidr.yaml", Interval: 86400,
		},
		"applications": {
			Type: "http", Behavior: "classical",
			URL: loyalsoldierCDN + "/applications.txt", Path: "./ruleset/applications.yaml", Interval: 86400,
		},
	}
	encoded, err := yaml.Marshal(providers)
	if err != nil {
		return fmt.Errorf("marshal rule providers: %w", err)
	}
	content.WriteString("\nrule-providers:\n")
	for _, line := range strings.Split(strings.TrimSuffix(string(encoded), "\n"), "\n") {
		content.WriteString("  ")
		content.WriteString(line)
		content.WriteByte('\n')
	}
	return nil
}

func writeRules(content *bytes.Buffer) {
	// Loyalsoldier/clash-rules 白名单模式：https://github.com/Loyalsoldier/clash-rules
	content.WriteString("\nrules:\n")
	for _, rule := range []string{
		"RULE-SET,applications,DIRECT",
		"DOMAIN,clash.razord.top,DIRECT",
		"DOMAIN,yacd.haishan.me,DIRECT",
		"RULE-SET,private,DIRECT",
		"RULE-SET,reject,REJECT",
		"RULE-SET,icloud,DIRECT",
		"RULE-SET,apple,DIRECT",
		"RULE-SET,google," + clashGroupProxy,
		"RULE-SET,proxy," + clashGroupProxy,
		"RULE-SET,direct,DIRECT",
		"RULE-SET,lancidr,DIRECT,no-resolve",
		"RULE-SET,cncidr,DIRECT,no-resolve",
		"RULE-SET,telegramcidr," + clashGroupProxy + ",no-resolve",
		"GEOIP,LAN,DIRECT,no-resolve",
		"GEOIP,CN,DIRECT,no-resolve",
		"MATCH," + clashGroupProxy,
	} {
		content.WriteString("  - ")
		content.WriteString(rule)
		content.WriteByte('\n')
	}
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
