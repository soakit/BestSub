package model

import (
	"encoding/json"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Storage struct { // 储存目标配置，任务通过 StorageID 复用同一套连接参数
	ID                  string `gorm:"column:id;primaryKey;type:varchar(36)" json:"id"` // ID
	StorageTargetConfig        // 储存目标基础配置。
	Status              string `gorm:"column:status;type:varchar(32)" json:"status"` // 储存目标状态，由任务写入结果更新。
}

type StorageTargetConfig struct { // 储存目标基础配置，各类型参数由 internal/storage 按 Type 解析
	Name   string          `gorm:"column:name;type:varchar(255)" json:"name"`             // 显示名称
	Type   string          `gorm:"column:type;type:varchar(32);not null" json:"type"`     // 储存类型: local/webdav/gist
	Params json.RawMessage `gorm:"column:params;type:json;serializer:json" json:"params"` // 储存类型私有参数
}

func (s *Storage) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}
