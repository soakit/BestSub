package task

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/bestruirui/bestsub/internal/model"
	"github.com/bestruirui/bestsub/internal/store"

	"github.com/charmbracelet/log"
)

var (
	runningMu    sync.Mutex               // 保护 runningTasks 的读写
	runningTasks = map[string]*runState{} // 记录正在运行的任务 ID 到取消状态的映射，用于按任务停止
)

type runState struct { // 单个任务运行状态，当前只保存取消函数
	cancel context.CancelFunc // 取消当前任务运行上下文
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
	if task.ID == "" {
		return fmt.Errorf("task id is required")
	}

	ctx, cancel := context.WithCancel(parent)
	runningMu.Lock()
	if _, ok := runningTasks[task.ID]; ok {
		runningMu.Unlock()
		cancel()
		return fmt.Errorf("%w: %s", ErrTaskRunning, task.ID)
	}
	state := &runState{cancel: cancel}
	runningTasks[task.ID] = state
	runningMu.Unlock()

	go func() {
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
	if len(task.Steps) == 0 {
		return fmt.Errorf("task steps is required")
	}
	nodes, err := expandTaskInput(task)
	if err != nil {
		return err
	}
	landingNodes, err := expandLandingInput(task)
	if err != nil {
		return err
	}
	if task.CustomLandingNodeEnable == 1 && len(landingNodes) == 0 {
		return fmt.Errorf("landing nodes is required")
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
		nodes, err = runStep(ctx, task.ID, step, nodes, landingNodes)
		if err != nil {
			return err
		}
		if len(nodes) == 0 {
			break
		}
	}
	store.TaskUpdateFinishedAt(task.ID)
	saveTaskNodes(task.ID, nodes)
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

func clearRunning(id string, state *runState) {
	runningMu.Lock()
	if runningTasks[id] == state {
		delete(runningTasks, id)
	}
	runningMu.Unlock()
}
