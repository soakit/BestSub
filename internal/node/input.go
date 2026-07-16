package node

import (
	"fmt"

	"github.com/bestruirui/bestsub/internal/model"
	"github.com/bestruirui/bestsub/internal/store"
)

type Input struct { // 任务和分享共同使用的节点来源。
	Subscriptions []model.SubscriptionRef // 指定订阅的内存节点池。
	Nodes         []model.NodeRef         // 指定单独节点。
	Tags          []model.TagRef          // 指定标签下的订阅和单独节点。
	ResultTasks   []model.TaskRef         // 指定任务的最近一次内存结果。
}

// ResolveInput 按来源顺序获取节点，并排除与订阅节点重复的任务结果。
func ResolveInput(input Input, allInput uint8) ([]Node, error) {
	nodes := []Node{}

	if allInput == 1 {
		subscriptions := store.SubscriptionList()
		input.Subscriptions = make([]model.SubscriptionRef, 0, len(subscriptions))
		for _, subscription := range subscriptions {
			input.Subscriptions = append(input.Subscriptions, model.SubscriptionRef{ID: subscription.ID})
		}

		standaloneNodes := store.NodeList()
		input.Nodes = make([]model.NodeRef, 0, len(standaloneNodes))
		for _, current := range standaloneNodes {
			input.Nodes = append(input.Nodes, model.NodeRef{ID: current.ID})
		}
	} else if len(input.Tags) > 0 {
		subscriptionIDs, nodeIDs := store.TagResourceIDs(input.Tags)
		// 标签只扩展输入来源，实际节点读取统一走订阅和单独节点流程。
		seenSubscriptionIDs := map[string]struct{}{}
		for _, subscription := range input.Subscriptions {
			seenSubscriptionIDs[subscription.ID] = struct{}{}
		}
		for _, id := range subscriptionIDs {
			if _, ok := seenSubscriptionIDs[id]; ok {
				continue
			}
			seenSubscriptionIDs[id] = struct{}{}
			input.Subscriptions = append(input.Subscriptions, model.SubscriptionRef{ID: id})
		}
		seenNodeIDs := map[string]struct{}{}
		for _, current := range input.Nodes {
			seenNodeIDs[current.ID] = struct{}{}
		}
		for _, id := range nodeIDs {
			if _, ok := seenNodeIDs[id]; ok {
				continue
			}
			seenNodeIDs[id] = struct{}{}
			input.Nodes = append(input.Nodes, model.NodeRef{ID: id})
		}
	}

	seenFingerprints := map[uint64]struct{}{}
	for _, subscription := range input.Subscriptions {
		for _, current := range PoolListBySubscription(subscription.ID) {
			nodes = append(nodes, current)
			seenFingerprints[current.Raw.Fingerprint] = struct{}{}
		}
	}
	for _, current := range input.Nodes {
		if current.ID == "" {
			return nil, fmt.Errorf("node id is required")
		}
		stored, ok := store.NodeGet(current.ID)
		if !ok {
			return nil, fmt.Errorf("node not found: %s", current.ID)
		}
		nodes = append(nodes, Node{
			NodeID: stored.ID,
			Raw:    &Raw{Text: stored.Content, Fingerprint: Fingerprint([]byte(stored.Content))},
			Info:   stored.NodeInfo,
		})
	}

	for _, resultTask := range input.ResultTasks {
		for _, current := range ResultNodes(resultTask.ID) {
			if _, ok := seenFingerprints[current.Raw.Fingerprint]; ok {
				continue
			}
			seenFingerprints[current.Raw.Fingerprint] = struct{}{}
			// 任务结果只复用原文和检测信息，不继承原任务的写回来源。
			nodes = append(nodes, Node{Raw: current.Raw, Info: current.Info})
		}
	}
	return nodes, nil
}
