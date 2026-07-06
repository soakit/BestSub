package store

import (
	"fmt"

	"bestsub/internal/model"
	"bestsub/internal/utils/cache"
)

var tagCache = cache.New[uint, model.Tag](4)

func initTag() error {
	var tags []model.Tag
	if err := db.Find(&tags).Error; err != nil {
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
