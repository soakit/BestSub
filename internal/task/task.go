package task

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/bestruirui/bestsub/internal/model"
	"github.com/bestruirui/bestsub/internal/node"
	"github.com/bestruirui/bestsub/internal/store"
)

var (
	ErrTaskNotFound   = errors.New("task not found")          // 任务不存在错误
	ErrTaskRunning    = errors.New("task is already running") // 任务已运行错误
	ErrTaskNotRunning = errors.New("task is not running")     // 任务未运行错误

	taskMu      sync.Mutex         // 保护 taskCtx、taskCancel 和 taskStarted 的读写
	taskCtx     context.Context    // 任务模块生命周期上下文，作为单次任务运行的父上下文
	taskCancel  context.CancelFunc // 取消任务模块生命周期上下文，用于整体停止调度和运行任务
	taskStarted bool               // 任务模块是否已启动，防止重复启动或重复停止
)

func Start(ctx context.Context) error {
	taskMu.Lock()
	defer taskMu.Unlock()
	if taskStarted {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	taskCtx, taskCancel = context.WithCancel(ctx)
	for _, task := range store.TaskList() {
		if err := syncSchedule(task); err != nil {
			taskCancel()
			taskCtx = nil
			taskCancel = nil
			return err
		}
	}
	scheduleCron.Start()
	taskStarted = true
	return nil
}

func Stop() {
	taskMu.Lock()
	if !taskStarted {
		taskMu.Unlock()
		return
	}
	cancel := taskCancel
	taskStarted = false
	taskCtx = nil
	taskCancel = nil
	taskMu.Unlock()

	cancel()
	<-scheduleCron.Stop().Done()

	runningMu.Lock()
	states := make([]*runState, 0, len(runningTasks))
	for id, state := range runningTasks {
		states = append(states, state)
		delete(runningTasks, id)
	}
	runningMu.Unlock()
	for _, state := range states {
		state.cancel()
	}
	for _, state := range states {
		<-state.done
	}
}

func Create(task *model.Task) error {
	if err := ValidateSchedule(task.TaskConfig); err != nil {
		return err
	}
	if err := store.TaskCreate(task); err != nil {
		return err
	}
	return syncSchedule(*task)
}

func Update(id string, config model.TaskConfig) error {
	if _, ok := store.TaskGet(id); !ok {
		return fmt.Errorf("%w: %s", ErrTaskNotFound, id)
	}
	if err := ValidateSchedule(config); err != nil {
		return err
	}
	if err := store.TaskUpdateConfig(id, config); err != nil {
		return err
	}
	task, ok := store.TaskGet(id)
	if !ok {
		return fmt.Errorf("%w: %s", ErrTaskNotFound, id)
	}
	return syncSchedule(task)
}

func Delete(id string) error {
	if _, ok := store.TaskGet(id); !ok {
		return fmt.Errorf("%w: %s", ErrTaskNotFound, id)
	}
	if err := store.TaskDelete(id); err != nil {
		return err
	}
	removeSchedule(id)
	cancelRunningAndWait(id)
	node.DeleteResult(id)
	return nil
}
