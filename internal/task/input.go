package task

import (
	"fmt"

	"github.com/bestruirui/bestsub/internal/model"
	"github.com/bestruirui/bestsub/internal/store"
)

func expandTaskInput(task model.Task) ([]stepNode, error) {
	nodes := []stepNode{}

	if task.AllInputEnable == 1 {
		subscriptions := store.SubscriptionList()
		task.Subscriptions = make([]model.SubscriptionRef, 0, len(subscriptions))
		for _, sub := range subscriptions {
			task.Subscriptions = append(task.Subscriptions, model.SubscriptionRef{ID: sub.ID})
		}

		standaloneNodes := store.NodeList()
		task.Nodes = make([]model.NodeRef, 0, len(standaloneNodes))
		for _, node := range standaloneNodes {
			task.Nodes = append(task.Nodes, model.NodeRef{ID: node.ID})
		}
	} else if len(task.Tags) > 0 {
		tags := make([]model.Tag, 0, len(task.Tags))
		for _, tag := range task.Tags {
			tags = append(tags, model.Tag{ID: tag.ID})
		}
		subIDs, nodeIDs := store.TagResourceIDs(tags)
		// Tags 只扩展输入来源，实际节点读取统一走下面的订阅和单独节点流程。
		seenSubIDs := map[string]struct{}{}
		for _, sub := range task.Subscriptions {
			seenSubIDs[sub.ID] = struct{}{}
		}
		for _, id := range subIDs {
			if _, ok := seenSubIDs[id]; ok {
				continue
			}
			seenSubIDs[id] = struct{}{}
			task.Subscriptions = append(task.Subscriptions, model.SubscriptionRef{ID: id})
		}
		seenNodeIDs := map[string]struct{}{}
		for _, node := range task.Nodes {
			seenNodeIDs[node.ID] = struct{}{}
		}
		for _, id := range nodeIDs {
			if _, ok := seenNodeIDs[id]; ok {
				continue
			}
			seenNodeIDs[id] = struct{}{}
			task.Nodes = append(task.Nodes, model.NodeRef{ID: id})
		}
	}

	for _, sub := range task.Subscriptions {
		if sub.ID == "" {
			continue
		}
		for _, item := range store.NodePoolListBySubscription(sub.ID) {
			nodes = append(nodes, stepNode{
				SubscriptionID: sub.ID,
				Raw:            item.Raw,
				Info:           item.Info,
			})
		}
	}
	for _, node := range task.Nodes {
		if node.ID == "" {
			return nil, fmt.Errorf("node id is required")
		}
		full, ok := store.NodeGet(node.ID)
		if !ok {
			return nil, fmt.Errorf("node not found: %s", node.ID)
		}
		nodes = append(nodes, stepNode{NodeID: full.ID, Raw: &model.NodeRaw{Text: full.Content, Fingerprint: store.NodeFingerprint([]byte(full.Content))}, Info: full.NodeInfo})
	}

	resultMu.RLock()
	defer resultMu.RUnlock()
	for _, resultTask := range task.ResultTasks {
		for _, node := range resultNodes[resultTask.ID] {
			// ResultTasks 只复用上次通过的 raw 节点，不继承原任务的写回来源。
			nodes = append(nodes, stepNode{Raw: node.Raw, Info: node.Info})
		}
	}
	return nodes, nil
}

func expandLandingInput(task model.Task) ([]stepNode, error) {
	if task.CustomLandingNodeEnable != 1 {
		return nil, nil
	}
	nodes := []stepNode{}
	for _, sub := range task.LandingSubscriptions {
		for _, item := range store.NodePoolListBySubscription(sub.ID) {
			nodes = append(nodes, stepNode{SubscriptionID: sub.ID, Raw: item.Raw})
		}
	}
	for _, node := range task.LandingNodes {
		if node.ID == "" {
			return nil, fmt.Errorf("landing node id is required")
		}
		full, ok := store.NodeGet(node.ID)
		if !ok {
			return nil, fmt.Errorf("landing node not found: %s", node.ID)
		}
		nodes = append(nodes, stepNode{NodeID: full.ID, Raw: &model.NodeRaw{Text: full.Content, Fingerprint: store.NodeFingerprint([]byte(full.Content))}, Info: full.NodeInfo})
	}
	return nodes, nil
}
