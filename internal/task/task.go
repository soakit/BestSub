package task

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/bestruirui/bestsub/internal/model"
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

type stepNode struct { // 单次任务步骤间传递的节点状态
	SubscriptionID string         // 订阅节点来源 ID，非空时用 Fingerprint 写回节点池
	Fingerprint    uint64         // 订阅节点池指纹，用于定位订阅内唯一节点
	NodeID         string         // 单独节点来源 ID，非空时写回 node 表检测信息
	Proxy          []byte         // 单条 mihomo YAML 节点内容
	Info           model.NodeInfo // 当前任务执行过程中累计的检测信息
	Index          int            // 当前步骤输入顺序，用于 Order=none 时恢复稳定顺序
}

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
	cancels := make([]context.CancelFunc, 0, len(runningTasks))
	for id, state := range runningTasks {
		cancels = append(cancels, state.cancel)
		delete(runningTasks, id)
	}
	runningMu.Unlock()
	for _, cancel := range cancels {
		cancel()
	}
}

func Create(task *model.Task) error {
	if task == nil {
		return fmt.Errorf("task is required")
	}
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
	cancelRunning(id)
	return nil
}
