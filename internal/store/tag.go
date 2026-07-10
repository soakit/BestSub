package store

import (
	"fmt"
	"slices"

	"github.com/bestruirui/bestsub/internal/model"
	"github.com/bestruirui/bestsub/internal/utils/cache"

	"gorm.io/gorm"
)

var tagCache = cache.New[uint, model.Tag](4) // 标签缓存，key 为 tag ID。

func initTag() error {
	var tags []model.Tag
	if err := db.
		Preload("Subscriptions", func(tx *gorm.DB) *gorm.DB { return tx.Select("id") }).
		Preload("Nodes", func(tx *gorm.DB) *gorm.DB { return tx.Select("id") }).
		Find(&tags).Error; err != nil {
		return fmt.Errorf("failed to load tags: %w", err)
	}
	for _, tag := range tags {
		tagCache.Set(tag.ID, tag)
	}
	return nil
}

func TagList() []model.Tag {
	tags := make([]model.Tag, 0, tagCache.Len())
	for _, tag := range tagCache.GetAll() {
		tags = append(tags, tag)
	}
	return tags
}

// TagResourceIDs 从 tag 缓存返回关联订阅 ID 和单独节点 ID。
func TagResourceIDs(tags []model.Tag) ([]string, []string) {
	if len(tags) == 0 {
		return nil, nil
	}

	subIDs := []string{}
	nodeIDs := []string{}
	seenSub := map[string]struct{}{}
	seenNode := map[string]struct{}{}
	for _, tag := range tags {
		full, ok := tagCache.Get(tag.ID)
		if !ok {
			continue
		}
		// 同一资源可能命中多个 tag，只返回一次，避免任务重复检测。
		for _, sub := range full.Subscriptions {
			if _, ok := seenSub[sub.ID]; ok {
				continue
			}
			seenSub[sub.ID] = struct{}{}
			subIDs = append(subIDs, sub.ID)
		}
		for _, node := range full.Nodes {
			if _, ok := seenNode[node.ID]; ok {
				continue
			}
			seenNode[node.ID] = struct{}{}
			nodeIDs = append(nodeIDs, node.ID)
		}
	}
	return subIDs, nodeIDs
}

func TagNamesBySubscription(id string) []string {
	names := []string{}
	for _, tag := range tagCache.GetAll() {
		for _, sub := range tag.Subscriptions {
			if sub.ID == id {
				names = append(names, tag.Name)
				break
			}
		}
	}
	return names
}

func TagNamesByNode(id string) []string {
	names := []string{}
	for _, tag := range tagCache.GetAll() {
		for _, node := range tag.Nodes {
			if node.ID == id {
				names = append(names, tag.Name)
				break
			}
		}
	}
	return names
}

func TagSetSubscriptionNames(id string, names []string) error {
	selected := map[string]struct{}{}
	for _, name := range names {
		selected[name] = struct{}{}
	}
	for _, tag := range tagCache.GetAll() {
		_, want := selected[tag.Name]
		has := slices.ContainsFunc(tag.Subscriptions, func(sub model.TagSubscription) bool { return sub.ID == id })
		if want == has {
			continue
		}
		if want {
			if err := db.Omit("Subscriptions.*").Model(&model.Tag{ID: tag.ID}).Association("Subscriptions").Append(&model.TagSubscription{ID: id}); err != nil {
				return err
			}
			tag.Subscriptions = append(tag.Subscriptions, model.TagSubscription{ID: id})
		} else {
			if err := db.Model(&model.Tag{ID: tag.ID}).Association("Subscriptions").Delete(&model.TagSubscription{ID: id}); err != nil {
				return err
			}
			tag.Subscriptions = slices.DeleteFunc(tag.Subscriptions, func(sub model.TagSubscription) bool { return sub.ID == id })
		}
		tagCache.Set(tag.ID, tag)
	}
	return nil
}

func TagSetNodeNames(id string, names []string) error {
	selected := map[string]struct{}{}
	for _, name := range names {
		selected[name] = struct{}{}
	}
	for _, tag := range tagCache.GetAll() {
		_, want := selected[tag.Name]
		has := slices.ContainsFunc(tag.Nodes, func(node model.TagNode) bool { return node.ID == id })
		if want == has {
			continue
		}
		if want {
			if err := db.Omit("Nodes.*").Model(&model.Tag{ID: tag.ID}).Association("Nodes").Append(&model.TagNode{ID: id}); err != nil {
				return err
			}
			tag.Nodes = append(tag.Nodes, model.TagNode{ID: id})
		} else {
			if err := db.Model(&model.Tag{ID: tag.ID}).Association("Nodes").Delete(&model.TagNode{ID: id}); err != nil {
				return err
			}
			tag.Nodes = slices.DeleteFunc(tag.Nodes, func(node model.TagNode) bool { return node.ID == id })
		}
		tagCache.Set(tag.ID, tag)
	}
	return nil
}

func TagCreate(tag *model.Tag) error {
	if err := db.Create(tag).Error; err != nil {
		return err
	}
	tagCache.Set(tag.ID, *tag)
	return nil
}

func TagDelete(id uint) error {
	if err := db.Delete(&model.Tag{}, id).Error; err != nil {
		return err
	}
	tagCache.Del(id)
	return nil
}
