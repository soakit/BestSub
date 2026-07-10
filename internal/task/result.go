package task

import (
	"fmt"
	"sync"

	"github.com/bestruirui/bestsub/internal/model"
)

var (
	resultMu    sync.RWMutex                           // 保护 resultNodes 和 taskResults 的并发读写
	resultNodes = map[string][]stepNode{}              // 记录任务最近一次通过检测的节点，用于 ResultTasks 输入
	taskResults = map[string][]model.TaskResultGroup{} // 记录任务最近一次结果，用于接口查询
)

func Result(id string) ([]model.TaskResultGroup, error) {
	resultMu.RLock()
	defer resultMu.RUnlock()
	result, ok := taskResults[id]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrTaskNotRunning, id)
	}
	return result, nil
}

func ResultCount(id string) int {
	resultMu.RLock()
	defer resultMu.RUnlock()
	return len(resultNodes[id])
}

func saveTaskNodes(id string, nodes []stepNode) {
	resultMu.Lock()
	defer resultMu.Unlock()

	resultNodes[id] = make([]stepNode, 0, len(nodes))
	for _, node := range nodes {
		resultNodes[id] = append(resultNodes[id], cloneStepNode(node))
	}

	groupMap := map[string][]uint64{}
	for _, node := range nodes {
		if node.SubscriptionID != "" {
			groupMap[node.SubscriptionID] = append(groupMap[node.SubscriptionID], node.Fingerprint)
		}
	}
	groups := make([]model.TaskResultGroup, 0, len(groupMap))
	for subID, fingerprints := range groupMap {
		groups = append(groups, model.TaskResultGroup{
			SubscriptionID: subID,
			Fingerprints:   fingerprints,
		})
	}
	taskResults[id] = groups
}

func cloneStepNode(node stepNode) stepNode {
	node.Proxy = append([]byte(nil), node.Proxy...)
	return node
}
