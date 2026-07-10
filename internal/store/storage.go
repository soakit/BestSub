package store

import (
	"fmt"

	"github.com/bestruirui/bestsub/internal/model"
	"github.com/bestruirui/bestsub/internal/utils/cache"
)

var storageCache = cache.New[string, model.Storage](4) // 储存配置缓存，key 为储存 ID。

func initStorage() error {
	storages := []model.Storage{}
	if err := db.Find(&storages).Error; err != nil {
		return fmt.Errorf("failed to load storages: %w", err)
	}
	for _, storage := range storages {
		storageCache.Set(storage.ID, storage)
	}
	return nil
}

func StorageList() []model.Storage {
	storages := make([]model.Storage, 0, storageCache.Len())
	for _, storage := range storageCache.GetAll() {
		storages = append(storages, storage)
	}
	return storages
}

func StorageGet(id string) (model.Storage, bool) {
	return storageCache.Get(id)
}

func StorageCreate(storage *model.Storage) error {
	if storage == nil {
		return fmt.Errorf("storage is required")
	}
	if err := db.Create(storage).Error; err != nil {
		return err
	}
	storageCache.Set(storage.ID, *storage)
	return nil
}

func StorageUpdateConfig(id string, config model.StorageTargetConfig) error {
	if err := db.Model(&model.Storage{}).Where("id = ?", id).Select("*").Updates(config).Error; err != nil {
		return err
	}
	if storage, ok := storageCache.Get(id); ok {
		storage.StorageTargetConfig = config
		storageCache.Set(id, storage)
	}
	return nil
}

func StorageDelete(id string) error {
	if err := db.Delete(&model.Storage{}, "id = ?", id).Error; err != nil {
		return err
	}
	storageCache.Del(id)
	return nil
}
