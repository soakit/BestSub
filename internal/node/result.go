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

// ResultNodes 返回指定任务结果，调用方只读使用。
func ResultNodes(taskID string) []Node {
	resultMu.RLock()
	defer resultMu.RUnlock()
	return resultNodes[taskID]
}

// SaveResult 保存任务结果及其原始来源，供下游任务回写来源数据。
func SaveResult(taskID string, nodes []Node) {
	resultMu.Lock()
	defer resultMu.Unlock()

	resultNodes[taskID] = nodes
}

// DeleteResult 删除指定任务的结果节点。
func DeleteResult(taskID string) {
	resultMu.Lock()
	delete(resultNodes, taskID)
	resultMu.Unlock()
}
