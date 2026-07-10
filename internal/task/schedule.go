package task

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/bestruirui/bestsub/internal/model"

	"github.com/charmbracelet/log"
	"github.com/robfig/cron/v3"
)

var (
	cronParser      = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow) // 解析 5 段 cron 表达式，不接受秒字段和 @descriptor
	scheduleMu      sync.Mutex                                                                   // 保护 scheduleEntries 和 scheduleCron 调度项变更
	scheduleCron    = cron.New(cron.WithParser(cronParser))                                      // 内存定时调度器，按任务 cron 表达式触发 Run
	scheduleEntries = map[string]cron.EntryID{}                                                  // 记录任务 ID 到 cron entry 的映射，用于更新或删除调度
)

func ValidateSchedule(config model.TaskConfig) error {
	if config.AutoRun != 1 {
		return nil
	}
	if strings.TrimSpace(config.CronExpr) == "" {
		return fmt.Errorf("cron expression is required")
	}
	if _, err := cronParser.Parse(strings.TrimSpace(config.CronExpr)); err != nil {
		return fmt.Errorf("invalid cron expression: %w", err)
	}
	return nil
}

func syncSchedule(task model.Task) error {
	if task.AutoRun != 1 || strings.TrimSpace(task.CronExpr) == "" {
		removeSchedule(task.ID)
		return nil
	}
	if _, err := cronParser.Parse(strings.TrimSpace(task.CronExpr)); err != nil {
		return fmt.Errorf("invalid cron expression: %w", err)
	}

	scheduleMu.Lock()
	defer scheduleMu.Unlock()
	if entryID, ok := scheduleEntries[task.ID]; ok {
		scheduleCron.Remove(entryID)
		delete(scheduleEntries, task.ID)
	}
	entryID, err := scheduleCron.AddFunc(strings.TrimSpace(task.CronExpr), func() {
		if err := Run(task.ID); err != nil && !errors.Is(err, ErrTaskRunning) {
			log.Errorf("task %s error: %v", task.ID, err)
		}
	})
	if err != nil {
		return err
	}
	scheduleEntries[task.ID] = entryID
	return nil
}

func removeSchedule(id string) {
	scheduleMu.Lock()
	defer scheduleMu.Unlock()
	if entryID, ok := scheduleEntries[id]; ok {
		scheduleCron.Remove(entryID)
		delete(scheduleEntries, id)
	}
}
