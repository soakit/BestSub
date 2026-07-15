package model

import (
	"crypto/rand"
	"math/big"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Share struct { // 分享配置的持久化模型，内部 ID 与公开 Token 分离
	ID          string    `gorm:"column:id;primaryKey;type:varchar(36)" json:"id"`              // 分享 ID，仅用于内部管理和数据关联。
	Token       string    `gorm:"column:token;uniqueIndex;type:char(16);not null" json:"token"` // 公开访问凭证，创建时随机生成 16 位纯小写字母。
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`           // 创建时间。
	NodeCount   int       `gorm:"-" json:"node_count"`                                          // 当前筛选后节点数，仅用于接口展示。
	ShareConfig           // 可编辑的分享配置。
}

type ShareConfig struct { // 分享名称、输入来源和筛选条件
	Name       string     `gorm:"column:name;type:varchar(255);not null" json:"name" binding:"required"` // 后台显示名称。
	Filter     NodeFilter `gorm:"column:filter;type:json;serializer:json" json:"filter"`                 // 节点筛选条件。
	ShareInput            // 与任务前置节点一致的输入来源。
}

type ShareInput struct { // 分享输入关联，只保存来源 ID
	Subscriptions []SubscriptionRef `gorm:"many2many:share_input_subscriptions;joinForeignKey:ShareID;joinReferences:SubscriptionID;constraint:OnDelete:CASCADE" json:"subscriptions"` // 指定订阅的内存节点池。
	Nodes         []NodeRef         `gorm:"many2many:share_input_nodes;joinForeignKey:ShareID;joinReferences:NodeID;constraint:OnDelete:CASCADE" json:"nodes"`                         // 指定单独节点。
	Tags          []TagRef          `gorm:"many2many:share_input_tags;joinForeignKey:ShareID;joinReferences:TagID;constraint:OnDelete:CASCADE" json:"tags"`                            // 指定标签下的订阅和单独节点。
	ResultTasks   []TaskRef         `gorm:"many2many:share_input_results;joinForeignKey:ShareID;joinReferences:ResultTaskID;constraint:OnDelete:CASCADE" json:"result_tasks"`          // 指定任务的最近一次内存结果。
}

// BeforeCreate 生成内部 UUID 和不可由客户端指定的公开访问 Token。
func (s *Share) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	token := make([]byte, 16)
	for i := range token {
		index, err := rand.Int(rand.Reader, big.NewInt(26))
		if err != nil {
			return err
		}
		token[i] = "abcdefghijklmnopqrstuvwxyz"[index.Int64()]
	}
	s.Token = string(token)
	return nil
}
