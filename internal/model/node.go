package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Node struct {
	ID         string    `gorm:"column:id;primaryKey;type:varchar(36)" json:"id"` // ID
	NodeConfig           // 节点配置
	NodeInfo             // 节点测试信息
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"` // 创建时间
}

type NodeConfig struct {
	Name    string `gorm:"column:name;type:varchar(255)" json:"name"`                  // 节点名称
	Tags    []Tag  `gorm:"many2many:tag_nodes" json:"tags"`                            // 标签
	Content string `gorm:"column:content;type:text" json:"content" binding:"required"` // 节点内容
}

type NodeInfo struct {
	Delay         uint16    `gorm:"column:delay;default:0" json:"delay"`                   // 延迟，单位毫秒；0 表示未知或未测试
	DownloadSpeed uint64    `gorm:"column:download_speed;default:0" json:"download_speed"` // 下载速度，单位 bytes/s；0 表示未知或未测试
	CountryCode   string    `gorm:"column:country_code;type:char(2)" json:"country_code"`  // 落地国家，ISO 3166-1 alpha-2 两位字母代码
	TestedAt      time.Time `gorm:"column:tested_at;default:null" json:"tested_at"`        // 最近一次测试时间
}

func (n *Node) BeforeCreate(tx *gorm.DB) error {
	if n.ID == "" {
		n.ID = uuid.New().String()
	}
	return nil
}
