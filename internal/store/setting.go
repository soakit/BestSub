package store

import (
	"fmt"

	"github.com/bestruirui/bestsub/internal/model"
)

var settingCache = make(map[string]string)

func initSetting() error {
	var settings []model.Setting
	if err := db.Find(&settings).Error; err != nil {
		return fmt.Errorf("failed to load settings: %w", err)
	}
	for _, s := range settings {
		settingCache[s.Key] = s.Value
	}
	return nil
}

func SettingGet(key string) string {
	return settingCache[key]
}

func SettingList() []model.Setting {
	settings := make([]model.Setting, 0, len(settingCache))
	for k, v := range settingCache {
		settings = append(settings, model.Setting{Key: k, Value: v})
	}
	return settings
}

func SettingSet(key, value string) error {
	if err := db.Save(&model.Setting{Key: key, Value: value}).Error; err != nil {
		return fmt.Errorf("failed to save setting %s: %w", key, err)
	}
	settingCache[key] = value
	return nil
}
