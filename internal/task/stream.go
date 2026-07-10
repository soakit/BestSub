package task

import (
	"sync"
	"time"

	"github.com/bestruirui/bestsub/internal/server/stream"
)

type TaskProgress struct { // 任务进度流事件，前端用 TaskID 合并多个任务状态
	TaskID string `json:"taskId"` // 任务 ID，用于前端区分多个并发任务。
	Step   int    `json:"step"`   // 当前第几个 step，从 1 开始。
	Done   int    `json:"done"`   // 当前 step 已完成测试节点数。
	Total  int    `json:"total"`  // 当前 step 总测试节点数。
}

var (
	ProgressEvents = stream.New()              // 任务进度 SSE 事件流，前端通过它订阅所有任务进度。
	progressMu     sync.Mutex                  // 保护 progressItems 的并发读写。
	progressItems  = map[string]TaskProgress{} // 运行中任务 ID 到最新进度快照的映射。
)

func init() {
	go func() {
		ticker := time.NewTicker(time.Second)
		for range ticker.C {
			progressMu.Lock()
			if len(progressItems) == 0 {
				progressMu.Unlock()
				continue
			}
			progresses := make([]TaskProgress, 0, len(progressItems))
			for _, progress := range progressItems {
				progresses = append(progresses, progress)
			}
			progressMu.Unlock()

			for _, progress := range progresses {
				ProgressEvents.Emit("progress", progress)
			}
		}
	}()
}

func setTaskProgress(progress TaskProgress) {
	progressMu.Lock()
	progressItems[progress.TaskID] = progress
	progressMu.Unlock()
}

func addTaskProgressDone(taskID string) {
	progressMu.Lock()
	if progress, ok := progressItems[taskID]; ok {
		progress.Done++
		progressItems[taskID] = progress
	}
	progressMu.Unlock()
}

func deleteTaskProgress(taskID string) {
	progressMu.Lock()
	delete(progressItems, taskID)
	progressMu.Unlock()
	ProgressEvents.Emit("done", TaskProgress{TaskID: taskID})
}
