package store

import (
	"fmt"

	"github.com/bestruirui/bestsub/internal/model"
	"github.com/bestruirui/bestsub/internal/utils/cache"

	"gorm.io/gorm"
)

var subCache = cache.New[string, model.Subscription](16)

func initSubscription() error {
	subs := []model.Subscription{}
	if err := db.Find(&subs).Error; err != nil {
		return fmt.Errorf("failed to load subscriptions: %w", err)
	}
	for _, sub := range subs {
		subCache.Set(sub.ID, sub)
	}
	return nil
}

func SubscriptionCreate(sub *model.Subscription) error {
	if err := db.Create(sub).Error; err != nil {
		return err
	}
	subCache.Set(sub.ID, *sub)
	return nil
}

func SubscriptionDelete(id string) error {
	if err := db.Delete(&model.Subscription{}, "id = ?", id).Error; err != nil {
		return err
	}
	subCache.Del(id)
	return nil
}

func SubscriptionUpdateStatus(id string, status model.SubscriptionStatus) error {
	if err := db.Model(&model.Subscription{}).Where("id = ?", id).Updates(status).Error; err != nil {
		return err
	}
	if sub, ok := subCache.Get(id); ok {
		sub.SubscriptionStatus = status
		subCache.Set(id, sub)
	}
	return nil
}

func SubscriptionUpdateConfig(id string, config model.SubscriptionConfig) error {
	tags := config.Tags
	config.Tags = nil

	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Subscription{}).Where("id = ?", id).Updates(config).Error; err != nil {
			return err
		}
		sub := &model.Subscription{ID: id}
		if err := tx.Model(sub).Association("Tags").Replace(tags); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	if sub, ok := subCache.Get(id); ok {
		config.Tags = tags
		sub.SubscriptionConfig = config
		subCache.Set(id, sub)
	}
	return nil
}

func SubscriptionList() []model.Subscription {
	subs := make([]model.Subscription, 0, subCache.Len())
	for _, sub := range subCache.GetAll() {
		subs = append(subs, sub)
	}
	return subs
}

func SubscriptionGet(id string) (model.Subscription, bool) {
	return subCache.Get(id)
}
