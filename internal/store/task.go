package store

import (
	"fmt"

	"github.com/bestruirui/bestsub/internal/model"
	"github.com/bestruirui/bestsub/internal/utils/cache"

	"gorm.io/gorm"
)

var taskCache = cache.New[string, model.Task](16)

func initTask() error {
	tasks := []model.Task{}
	if err := db.Preload("Subscriptions").Preload("Nodes").Preload("Tags").Preload("ResultTasks").Preload("LandingSubscriptions").Preload("LandingNodes").Preload("Storage").Find(&tasks).Error; err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
	}
	for _, task := range tasks {
		taskCache.Set(task.ID, task)
	}
	return nil
}

func TaskCreate(task *model.Task) error {
	subscriptions := task.Subscriptions
	nodes := task.Nodes
	tags := task.Tags
	resultTasks := task.ResultTasks
	landingSubscriptions := task.LandingSubscriptions
	landingNodes := task.LandingNodes
	task.Subscriptions = nil
	task.Nodes = nil
	task.Tags = nil
	task.ResultTasks = nil
	task.LandingSubscriptions = nil
	task.LandingNodes = nil

	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(task).Error; err != nil {
			return err
		}
		if err := tx.Model(task).Association("Subscriptions").Replace(subscriptions); err != nil {
			return err
		}
		if err := tx.Model(task).Association("Nodes").Replace(nodes); err != nil {
			return err
		}
		if err := tx.Model(task).Association("Tags").Replace(tags); err != nil {
			return err
		}
		if err := tx.Model(task).Association("ResultTasks").Replace(resultTasks); err != nil {
			return err
		}
		if err := tx.Model(task).Association("LandingSubscriptions").Replace(landingSubscriptions); err != nil {
			return err
		}
		return tx.Model(task).Association("LandingNodes").Replace(landingNodes)
	}); err != nil {
		return err
	}

	task.Subscriptions = subscriptions
	task.Nodes = nodes
	task.Tags = tags
	task.ResultTasks = resultTasks
	task.LandingSubscriptions = landingSubscriptions
	task.LandingNodes = landingNodes
	taskCache.Set(task.ID, *task)
	return nil
}

func TaskDelete(id string) error {
	if err := db.Delete(&model.Task{}, "id = ?", id).Error; err != nil {
		return err
	}
	taskCache.Del(id)
	return nil
}

func TaskList() []model.Task {
	tasks := make([]model.Task, 0, taskCache.Len())
	for _, task := range taskCache.GetAll() {
		tasks = append(tasks, task)
	}
	return tasks
}

func TaskGet(id string) (model.Task, bool) {
	return taskCache.Get(id)
}

func TaskUpdateConfig(id string, config model.TaskConfig) error {
	subscriptions := config.Subscriptions
	nodes := config.Nodes
	tags := config.Tags
	resultTasks := config.ResultTasks
	landingSubscriptions := config.LandingSubscriptions
	landingNodes := config.LandingNodes
	config.Subscriptions = nil
	config.Nodes = nil
	config.Tags = nil
	config.ResultTasks = nil
	config.LandingSubscriptions = nil
	config.LandingNodes = nil

	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Task{}).
			Where("id = ?", id).
			Select("*").
			Omit("Subscriptions", "Nodes", "Tags", "ResultTasks", "LandingSubscriptions", "LandingNodes", "Storage").
			Updates(config).Error; err != nil {
			return err
		}
		task := &model.Task{ID: id}
		if err := tx.Model(task).Association("Subscriptions").Replace(subscriptions); err != nil {
			return err
		}
		if err := tx.Model(task).Association("Nodes").Replace(nodes); err != nil {
			return err
		}
		if err := tx.Model(task).Association("Tags").Replace(tags); err != nil {
			return err
		}
		if err := tx.Model(task).Association("ResultTasks").Replace(resultTasks); err != nil {
			return err
		}
		if err := tx.Model(task).Association("LandingSubscriptions").Replace(landingSubscriptions); err != nil {
			return err
		}
		return tx.Model(task).Association("LandingNodes").Replace(landingNodes)
	}); err != nil {
		return err
	}

	if task, ok := taskCache.Get(id); ok {
		config.Subscriptions = subscriptions
		config.Nodes = nodes
		config.Tags = tags
		config.ResultTasks = resultTasks
		config.LandingSubscriptions = landingSubscriptions
		config.LandingNodes = landingNodes
		task.TaskConfig = config
		taskCache.Set(id, task)
	}
	return nil
}
