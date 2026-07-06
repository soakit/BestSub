package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Storage struct {
	ID string `gorm:"column:id;primaryKey;type:varchar(36)" json:"id"` // ID
}

func (s *Storage) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}
