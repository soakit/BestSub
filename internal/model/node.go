package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Node struct {
	ID         string    `gorm:"column:id;primaryKey;type:varchar(36)" json:"id"` // ID
	NodeConfig           // 节点配置
	NodeInfo             // 节点附加信息
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"` // 创建时间
}

// NodeRef 是输入来源关联单独节点的轻量模型。
type NodeRef struct {
	ID string `gorm:"column:id;primaryKey;type:varchar(36)" json:"id"` // 单独节点 ID
}

func (NodeRef) TableName() string {
	return "nodes"
}

type NodeConfig struct {
	Name     string   `gorm:"column:name;type:varchar(255)" json:"name"`                  // 节点名称
	TagNames []string `gorm:"-" json:"tag_names"`                                         // 标签名称，仅接口展示。
	Content  string   `gorm:"column:content;type:text" json:"content" binding:"required"` // 节点内容
}

type NodeInfo struct { // 保存节点检测结果和从名称提取的附加信息
	Delay             uint16  `gorm:"column:delay;default:0" json:"delay"`                           // 延迟，单位毫秒；0 表示未知或未测试
	DownloadSpeed     uint32  `gorm:"column:download_speed;default:0" json:"download_speed"`         // 下载速度，单位 kb/s；0 表示未知或未测试
	CountryCode       string  `gorm:"column:country_code;type:char(2)" json:"country_code"`          // 落地国家，ISO 3166-1 alpha-2 两位字母代码
	TrafficMultiplier float32 `gorm:"column:traffic_multiplier;default:1" json:"traffic_multiplier"` // 流量扣费倍率；节点名称未标注时为 1
}

// NodeFilter 保存任务步骤和分享共同使用的节点筛选条件。
type NodeFilter struct {
	Limit               int      `json:"limit,omitempty"`                 // 节点的数量上限，0 表示不限制
	MinDelay            uint16   `json:"min_delay,omitempty"`             // 最小延迟，单位毫秒
	MaxDelay            uint16   `json:"max_delay,omitempty"`             // 最大延迟，单位毫秒
	MinDownloadSpeed    uint32   `json:"min_download_speed,omitempty"`    // 最小下载速度，单位 kb/s
	MaxDownloadSpeed    uint32   `json:"max_download_speed,omitempty"`    // 最大下载速度，单位 kb/s
	IncludeCountryCodes []string `json:"include_country_codes,omitempty"` // 只保留这些国家代码
	ExcludeCountryCodes []string `json:"exclude_country_codes,omitempty"` // 排除这些国家代码
}

type NodeRaw struct { // 可在节点池、任务和分享之间复用的不可变节点原文
	Text        string // 单条 Mihomo YAML 节点内容，创建后不再修改。
	Fingerprint uint64 // 按既有 NodeFingerprint 算法生成的节点指纹。
}

type NodeSnapshot struct { // 节点原文与一组附加信息的运行时快照
	Raw  *NodeRaw // 可共享的不可变节点原文。
	Info NodeInfo // 当前或任务完成时的节点附加信息。
}

func (n *Node) BeforeCreate(tx *gorm.DB) error {
	if n.ID == "" {
		n.ID = uuid.New().String()
	}
	return nil
}
