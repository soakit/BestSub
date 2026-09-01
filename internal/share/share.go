package share

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/bestruirui/bestsub/internal/model"
	"github.com/bestruirui/bestsub/internal/node"
	"github.com/bestruirui/bestsub/internal/store"
)

// NormalizeConfig 规范并校验管理接口提交的分享配置。
func NormalizeConfig(config *model.ShareConfig) error {
	if config == nil {
		return fmt.Errorf("share config is required")
	}
	config.Name = strings.TrimSpace(config.Name)
	config.NodeRenameExpression = strings.TrimSpace(config.NodeRenameExpression)
	if config.Name == "" {
		return fmt.Errorf("share name is required")
	}
	if len(config.Subscriptions)+len(config.Nodes)+len(config.Tags)+len(config.ResultTasks) == 0 {
		return fmt.Errorf("share input is required")
	}
	if config.Filter.MinDelay > 0 && config.Filter.MaxDelay > 0 && config.Filter.MinDelay > config.Filter.MaxDelay {
		return fmt.Errorf("min delay cannot exceed max delay")
	}
	if config.Filter.MinDownloadSpeed > 0 && config.Filter.MaxDownloadSpeed > 0 && config.Filter.MinDownloadSpeed > config.Filter.MaxDownloadSpeed {
		return fmt.Errorf("min download speed cannot exceed max download speed")
	}

	for _, sub := range config.Subscriptions {
		if sub.ID == "" {
			return fmt.Errorf("subscription id is required")
		}
		if _, ok := store.SubscriptionGet(sub.ID); !ok {
			return fmt.Errorf("subscription not found: %s", sub.ID)
		}
	}
	config.Subscriptions = slices.DeleteFunc(config.Subscriptions, func(sub model.SubscriptionRef) bool {
		return node.PoolCount(sub.ID) == 0
	})
	for _, node := range config.Nodes {
		if node.ID == "" {
			return fmt.Errorf("node id is required")
		}
		if _, ok := store.NodeGet(node.ID); !ok {
			return fmt.Errorf("node not found: %s", node.ID)
		}
	}
	for _, resultTask := range config.ResultTasks {
		if resultTask.ID == "" {
			return fmt.Errorf("result task id is required")
		}
		if _, ok := store.TaskGet(resultTask.ID); !ok {
			return fmt.Errorf("task not found: %s", resultTask.ID)
		}
	}
	tagIDs := map[uint]struct{}{}
	for _, tag := range store.TagList() {
		tagIDs[tag.ID] = struct{}{}
	}
	for _, tag := range config.Tags {
		if _, ok := tagIDs[tag.ID]; !ok {
			return fmt.Errorf("tag not found: %d", tag.ID)
		}
	}
	return nil
}

func Count(share model.Share) (int, error) {
	nodes, err := resolveNodes(share)
	return len(nodes), err
}

// Write 按配置重命名并写出 Mihomo 订阅。
func Write(share model.Share, writer io.Writer) error {
	nodes, err := resolveNodes(share)
	if err != nil {
		return err
	}
	if len(nodes) == 0 {
		content, err := node.MihomoClashProfile(nil, share.NodeRenameExpression)
		if err != nil {
			return err
		}
		_, err = writer.Write(content)
		return err
	}
	content, err := node.MihomoClashProfile(nodes, share.NodeRenameExpression)
	if err != nil {
		return fmt.Errorf("build share Mihomo subscription: %w", err)
	}
	_, err = writer.Write(content)
	return err
}

// resolveNodes 统一解析分享输入并应用分享筛选条件。
func resolveNodes(share model.Share) ([]node.Node, error) {
	nodes, err := node.ResolveInput(node.Input{
		Subscriptions: share.Subscriptions,
		Nodes:         share.Nodes,
		Tags:          share.Tags,
		ResultTasks:   share.ResultTasks,
	}, 0)
	if err != nil {
		return nil, err
	}
	result := nodes[:0]
	for _, node := range nodes {
		if passFilter(node.Info, share.Filter) {
			result = append(result, node)
		}
	}
	return result, nil
}

// passFilter 在启用某维筛选时拒绝未测试值，不启用时允许未知值通过。
func passFilter(info model.NodeInfo, filter model.NodeFilter) bool {
	if (filter.MinDelay > 0 || filter.MaxDelay > 0) && info.Delay == 0 {
		return false
	}
	if (filter.MinDelay > 0 && info.Delay < filter.MinDelay) || (filter.MaxDelay > 0 && info.Delay > filter.MaxDelay) {
		return false
	}
	if (filter.MinDownloadSpeed > 0 || filter.MaxDownloadSpeed > 0) && info.DownloadSpeed == 0 {
		return false
	}
	if (filter.MinDownloadSpeed > 0 && info.DownloadSpeed < filter.MinDownloadSpeed) || (filter.MaxDownloadSpeed > 0 && info.DownloadSpeed > filter.MaxDownloadSpeed) {
		return false
	}
	if (len(filter.IncludeCountryCodes) > 0 || len(filter.ExcludeCountryCodes) > 0) && info.CountryCode == "" {
		return false
	}
	if len(filter.IncludeCountryCodes) > 0 && !slices.Contains(filter.IncludeCountryCodes, info.CountryCode) {
		return false
	}
	return !slices.Contains(filter.ExcludeCountryCodes, info.CountryCode)
}
