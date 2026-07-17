package store

import (
	"fmt"

	"github.com/bestruirui/bestsub/internal/model"
	"github.com/bestruirui/bestsub/internal/utils/cache"

	"gorm.io/gorm"
)

var shareCache = cache.New[string, model.Share](16) // 分享缓存，key 为内部分享 ID。
var shareTokenIndex = cache.New[string, string](16) // 公开 Token 索引，value 为内部分享 ID。

func initShare() error {
	shares := []model.Share{}
	if err := db.
		Preload("Subscriptions", func(tx *gorm.DB) *gorm.DB { return tx.Select("id") }).
		Preload("Nodes", func(tx *gorm.DB) *gorm.DB { return tx.Select("id") }).
		Preload("Tags", func(tx *gorm.DB) *gorm.DB { return tx.Select("id") }).
		Preload("ResultTasks", func(tx *gorm.DB) *gorm.DB { return tx.Select("id") }).
		Find(&shares).Error; err != nil {
		return fmt.Errorf("failed to load shares: %w", err)
	}
	for _, share := range shares {
		shareCache.Set(share.ID, share)
		shareTokenIndex.Set(share.Token, share.ID)
	}
	return nil
}

func ShareList() []model.Share {
	shares := make([]model.Share, 0, shareCache.Len())
	for _, share := range shareCache.GetAll() {
		shares = append(shares, share)
	}
	return shares
}

func ShareGet(id string) (model.Share, bool) {
	return shareCache.Get(id)
}

func ShareGetByToken(token string) (model.Share, bool) {
	id, ok := shareTokenIndex.Get(token)
	if !ok {
		return model.Share{}, false
	}
	return shareCache.Get(id)
}

func ShareCreate(share *model.Share) error {
	if share == nil {
		return fmt.Errorf("share is required")
	}
	input := share.ShareInput
	share.ShareInput = model.ShareInput{}
	// 主表和四类来源必须同时成功，避免缓存与关联表出现半成品。
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(share).Error; err != nil {
			return err
		}
		if err := tx.Omit("Subscriptions.*").Model(share).Association("Subscriptions").Replace(input.Subscriptions); err != nil {
			return err
		}
		if err := tx.Omit("Nodes.*").Model(share).Association("Nodes").Replace(input.Nodes); err != nil {
			return err
		}
		if err := tx.Omit("Tags.*").Model(share).Association("Tags").Replace(input.Tags); err != nil {
			return err
		}
		return tx.Omit("ResultTasks.*").Model(share).Association("ResultTasks").Replace(input.ResultTasks)
	})
	share.ShareInput = input
	if err != nil {
		return err
	}
	shareCache.Set(share.ID, *share)
	shareTokenIndex.Set(share.Token, share.ID)
	return nil
}

func ShareUpdateConfig(id string, config model.ShareConfig) error {
	input := config.ShareInput
	config.ShareInput = model.ShareInput{}
	// 配置更新与来源替换共用事务，缓存只接收完整的新配置。
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Share{}).Where("id = ?", id).Select("*").Omit("Subscriptions", "Nodes", "Tags", "ResultTasks").Updates(config).Error; err != nil {
			return err
		}
		share := &model.Share{ID: id}
		if err := tx.Omit("Subscriptions.*").Model(share).Association("Subscriptions").Replace(input.Subscriptions); err != nil {
			return err
		}
		if err := tx.Omit("Nodes.*").Model(share).Association("Nodes").Replace(input.Nodes); err != nil {
			return err
		}
		if err := tx.Omit("Tags.*").Model(share).Association("Tags").Replace(input.Tags); err != nil {
			return err
		}
		return tx.Omit("ResultTasks.*").Model(share).Association("ResultTasks").Replace(input.ResultTasks)
	})
	if err != nil {
		return err
	}
	if share, ok := shareCache.Get(id); ok {
		config.ShareInput = input
		share.ShareConfig = config
		shareCache.Set(id, share)
	}
	return nil
}

func ShareDelete(id string) error {
	if err := db.Delete(&model.Share{}, "id = ?", id).Error; err != nil {
		return err
	}
	if share, ok := shareCache.Get(id); ok {
		shareTokenIndex.Del(share.Token)
	}
	shareCache.Del(id)
	return nil
}
