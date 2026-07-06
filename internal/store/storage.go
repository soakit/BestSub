package store

import (
	"fmt"

	"github.com/bestruirui/bestsub/internal/model"
	"github.com/bestruirui/bestsub/internal/utils/cache"
)

var storageCache = cache.New[string, model.Storage](4)

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
