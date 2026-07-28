package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// URLType 订阅 URL 类型
type URLType uint8

const (
	URLTypeDirect URLType = iota // 直接获取节点
	URLTypeList                  // 响应体为 URL 列表，再逐个拉取
)

// ProxyMode 代理模式
type ProxyMode uint8

const (
	ProxyModeAuto     ProxyMode = iota // 自动：先直连，失败后使用全局代理
	ProxyModeDisabled                  // 禁用：始终直连
	ProxyModeEnabled                   // 启用：始终使用全局代理
)

// NodeType 表示 Mihomo 支持的节点协议类型。
type NodeType string

const (
	NodeTypeSS        NodeType = "ss"        // Shadowsocks 节点。
	NodeTypeSSR       NodeType = "ssr"       // ShadowsocksR 节点。
	NodeTypeSOCKS5    NodeType = "socks5"    // SOCKS5 节点。
	NodeTypeHTTP      NodeType = "http"      // HTTP 代理节点。
	NodeTypeVMess     NodeType = "vmess"     // VMess 节点。
	NodeTypeVLESS     NodeType = "vless"     // VLESS 节点。
	NodeTypeSnell     NodeType = "snell"     // Snell 节点。
	NodeTypeTrojan    NodeType = "trojan"    // Trojan 节点。
	NodeTypeHysteria  NodeType = "hysteria"  // Hysteria 节点。
	NodeTypeHysteria2 NodeType = "hysteria2" // Hysteria 2 节点。
	NodeTypeWireGuard NodeType = "wireguard" // WireGuard 节点。
	NodeTypeTUIC      NodeType = "tuic"      // TUIC 节点。
	NodeTypeSSH       NodeType = "ssh"       // SSH 节点。
	NodeTypeMieru     NodeType = "mieru"     // Mieru 节点。
	NodeTypeAnyTLS    NodeType = "anytls"    // AnyTLS 节点。
)

type Subscription struct {
	ID string `gorm:"column:id;primaryKey;type:varchar(36)" json:"id"` // ID
	SubscriptionConfig
	SubscriptionStatus
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"` // 创建时间
}

// SubscriptionRef 是输入来源关联订阅的轻量模型。
type SubscriptionRef struct {
	ID string `gorm:"column:id;primaryKey;type:varchar(36)" json:"id"` // 订阅 ID
}

func (SubscriptionRef) TableName() string {
	return "subscriptions"
}

type SubscriptionConfig struct {
	UrlType            URLType           `gorm:"column:url_type;default:0" json:"url_type" binding:"oneof=0 1"`            // URL 类型: 0=直接获取, 1=链接列表
	Url                []string          `gorm:"column:url;not null;type:json;serializer:json" json:"url" binding:"min=1"` // 订阅链接列表
	Enable             uint8             `gorm:"column:enable;default:1" json:"enable"`                                    // 是否启用
	Name               string            `gorm:"column:name;type:varchar(255)" json:"name"`                                // 名称备注
	TagNames           []string          `gorm:"-" json:"tag_names"`                                                       // 标签名称，仅接口展示。
	Header             map[string]string `gorm:"column:header;type:json;serializer:json" json:"header"`                    // 拉取订阅时的请求头
	AutoUpdate         uint8             `gorm:"column:auto_update;default:1" json:"auto_update"`                          // 是否启用自动更新
	CronExpr           string            `gorm:"column:cron_expr;type:varchar(64)" json:"cron_expr"`                       // Cron更新表达式
	ProxyMode          ProxyMode         `gorm:"column:proxy_mode;default:0" json:"proxy_mode"`                            // 代理模式: 0=自动, 1=禁用, 2=启用
	ProtocolFilterMode uint8             `gorm:"column:protocol_filter_mode;default:0" json:"protocol_filter_mode"`        // 节点过滤模式: 0=不启用, 1=包含, 2=排除
	ProtocolFilter     []NodeType        `gorm:"column:protocol_filter;type:json;serializer:json" json:"protocol_filter"`  // 更新时节点过滤
}

type SubscriptionStatus struct {
	NodeNum      uint32    `gorm:"-" json:"node_num"`                                    // 节点数量，实时来自内存节点池，不落库。
	TrafficTotal int64     `gorm:"column:traffic_total;default:0" json:"traffic_total"`  // 总流量
	TrafficUsed  int64     `gorm:"column:traffic_used;default:0" json:"traffic_used"`    // 已使用的流量
	ExpiresAt    time.Time `gorm:"column:expires_at;default:null" json:"expires_at"`     // 到期时间
	RefreshedAt  time.Time `gorm:"column:refreshed_at;default:null" json:"refreshed_at"` // 刷新时间
}

func (s *Subscription) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}
