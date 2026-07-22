package store

import (
	"fmt"

	"github.com/bestruirui/bestsub/internal/model"
	"github.com/bestruirui/bestsub/internal/utils/cache"
)

var subCache = cache.New[string, model.Subscription](16) // 订阅缓存，key 为订阅 ID。

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
	if err := TagSetSubscriptionNames(sub.ID, sub.TagNames); err != nil {
		return err
	}
	subCache.Set(sub.ID, *sub)
	return nil
}

func SubscriptionDelete(id string) error {
	if err := db.Delete(&model.Subscription{}, "id = ?", id).Error; err != nil {
		return err
	}
	if err := TagSetSubscriptionNames(id, nil); err != nil {
		return err
	}
	subCache.Del(id)
	return nil
}

func SubscriptionUpdateStatus(id string, status model.SubscriptionStatus) error {
	if err := db.Model(&model.Subscription{}).Where("id = ?", id).Select("*").Updates(status).Error; err != nil {
		return err
	}
	if sub, ok := subCache.Get(id); ok {
		sub.SubscriptionStatus = status
		subCache.Set(id, sub)
	}
	return nil
}

func SubscriptionUpdateConfig(id string, config model.SubscriptionConfig) error {
	if err := db.Model(&model.Subscription{}).Where("id = ?", id).Select("*").Updates(config).Error; err != nil {
		return err
	}
	if err := TagSetSubscriptionNames(id, config.TagNames); err != nil {
		return err
	}

	if sub, ok := subCache.Get(id); ok {
		sub.SubscriptionConfig = config
		subCache.Set(id, sub)
	}
	return nil
}

func SubscriptionList() []model.Subscription {
	subs := make([]model.Subscription, 0, subCache.Len())
	for _, sub := range subCache.GetAll() {
		sub.TagNames = TagNamesBySubscription(sub.ID)
		subs = append(subs, sub)
	}
	return subs
}

func SubscriptionLen() int { return subCache.Len() }

func SubscriptionGet(id string) (model.Subscription, bool) {
	sub, ok := subCache.Get(id)
	if ok {
		sub.TagNames = TagNamesBySubscription(sub.ID)
	}
	return sub, ok
}
