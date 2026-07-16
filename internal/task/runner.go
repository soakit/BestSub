package task

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"

	"github.com/bestruirui/bestsub/internal/model"
	"github.com/bestruirui/bestsub/internal/node"
	"github.com/bestruirui/bestsub/internal/store"

	"github.com/charmbracelet/log"
)

var (
	runningMu    sync.Mutex               // 保护 runningTasks 的读写
	runningTasks = map[string]*runState{} // 记录正在运行的任务 ID 到取消状态的映射，用于按任务停止
)

type runState struct { // 单个任务运行状态
	cancel context.CancelFunc // 取消当前任务运行上下文。
	done   chan struct{}      // 任务 goroutine 退出后关闭，删除任务时等待结果写入结束。
}

func Run(id string) error {
	taskMu.Lock()
	parent := taskCtx
	taskMu.Unlock()
	if parent == nil {
		parent = context.Background()
	}

	task, ok := store.TaskGet(id)
	if !ok {
		return fmt.Errorf("%w: %s", ErrTaskNotFound, id)
	}

	ctx, cancel := context.WithCancel(parent)
	runningMu.Lock()
	if _, ok := runningTasks[task.ID]; ok {
		runningMu.Unlock()
		cancel()
		return fmt.Errorf("%w: %s", ErrTaskRunning, task.ID)
	}
	state := &runState{cancel: cancel, done: make(chan struct{})}
	runningTasks[task.ID] = state
	runningMu.Unlock()

	go func() {
		defer close(state.done)
		defer runtime.GC()
		defer cancel()
		defer clearRunning(task.ID, state)
		if err := runTask(ctx, task); err != nil && !errors.Is(err, context.Canceled) {
			log.Errorf("task %s error: %v", task.ID, err)
		}
	}()
	return nil
}

func StopTask(id string) error {
	if !cancelRunning(id) {
		return fmt.Errorf("%w: %s", ErrTaskNotRunning, id)
	}
	return nil
}

func runTask(ctx context.Context, task model.Task) error {
	defer deleteTaskProgress(task.ID)
	nodes, err := node.ResolveInput(node.Input{
		Subscriptions: task.Subscriptions,
		Nodes:         task.Nodes,
		Tags:          task.Tags,
		ResultTasks:   task.ResultTasks,
	}, task.AllInputEnable)
	if err != nil {
		return err
	}
	var landingNode *node.Raw
	if task.CustomLandingNodeEnable == 1 {
		// 落地配置只有一个 NodeRef，复用统一输入解析后取唯一节点原文。
		resolved, err := node.ResolveInput(node.Input{Nodes: []model.NodeRef{task.LandingNode}}, 0)
		if err != nil {
			return err
		}
		landingNode = resolved[0].Raw
	}
	for i, step := range task.Steps {
		if err := ctx.Err(); err != nil {
			return err
		}
		setTaskProgress(TaskProgress{
			TaskID: task.ID,
			Step:   i + 1,
			Total:  len(nodes),
		})
		nodes, err = runStep(ctx, task.ID, step, nodes, landingNode)
		if err != nil {
			return err
		}
		if len(nodes) == 0 {
			break
		}
	}
	// 检测结果先落入内存并更新时间，储存失败时仍保留本次已经完成的节点结果。
	store.TaskUpdateFinishedAt(task.ID)
	node.SaveResult(task.ID, nodes)
	if task.StorageEnable == 1 {
		return saveTaskOutput(ctx, task, nodes)
	}
	return nil
}

func cancelRunning(id string) bool {
	runningMu.Lock()
	state, ok := runningTasks[id]
	if ok {
		delete(runningTasks, id)
	}
	runningMu.Unlock()
	if ok {
		state.cancel()
	}
	return ok
}

// cancelRunningAndWait 确保删除任务前运行协程已退出，避免其随后重新写入结果缓存。
func cancelRunningAndWait(id string) {
	runningMu.Lock()
	state, ok := runningTasks[id]
	if ok {
		delete(runningTasks, id)
	}
	runningMu.Unlock()
	if ok {
		state.cancel()
		<-state.done
	}
}

func clearRunning(id string, state *runState) {
	runningMu.Lock()
	if runningTasks[id] == state {
		delete(runningTasks, id)
	}
	runningMu.Unlock()
}
