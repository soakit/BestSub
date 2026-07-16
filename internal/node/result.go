package node

import "sync"

var (
	resultMu    sync.RWMutex          // 保护 resultNodes 的并发读写。
	resultNodes = map[string][]Node{} // 记录任务最近一次通过检测的节点。
)

// ResultCount 返回指定任务的结果节点数量。
func ResultCount(taskID string) int {
	resultMu.RLock()
	defer resultMu.RUnlock()
	return len(resultNodes[taskID])
}

// ResultNodes 返回指定任务结果的独立切片。
func ResultNodes(taskID string) []Node {
	resultMu.RLock()
	defer resultMu.RUnlock()
	return append([]Node(nil), resultNodes[taskID]...)
}

// SaveResult 保存任务结果，不继承输入节点的写回来源。
func SaveResult(taskID string, nodes []Node) {
	resultMu.Lock()
	defer resultMu.Unlock()

	resultNodes[taskID] = make([]Node, 0, len(nodes))
	for _, current := range nodes {
		resultNodes[taskID] = append(resultNodes[taskID], Node{Raw: current.Raw, Info: current.Info})
	}
}

// DeleteResult 删除指定任务的结果节点。
func DeleteResult(taskID string) {
	resultMu.Lock()
	delete(resultNodes, taskID)
	resultMu.Unlock()
}
