package model

import (
	"encoding/json"
	"time"

	"github.com/bestruirui/bestsub/pkg/probe"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Task struct {
	ID string `gorm:"column:id;primaryKey;type:varchar(36)" json:"id"` // ID
	TaskConfig
}

type TaskConfig struct {
	Name     string     `gorm:"column:name;type:varchar(255)" json:"name"`                              // 任务名称
	AutoRun  uint8      `gorm:"column:auto_run;default:0" json:"auto_run"`                              // 是否自动运行
	CronExpr string     `gorm:"column:cron_expr;type:varchar(64)" json:"cron_expr"`                     // Cron 表达式
	Steps    []TaskStep `gorm:"column:steps;type:json;serializer:json" json:"steps" binding:"required"` // 线性步骤列表
	TaskInput
	StorageConfig
}

type TaskInput struct {
	Subscriptions           []Subscription `gorm:"many2many:task_input_subscriptions;constraint:OnDelete:CASCADE" json:"subscriptions"`                                            // 指定订阅的内存节点池
	Nodes                   []Node         `gorm:"many2many:task_input_nodes;constraint:OnDelete:CASCADE" json:"nodes"`                                                            // 指定单独节点
	Tags                    []Tag          `gorm:"many2many:task_input_tags;constraint:OnDelete:CASCADE" json:"tags"`                                                              // 指定 tag 下的订阅和单独节点
	ResultTasks             []Task         `gorm:"many2many:task_input_results;joinForeignKey:TaskID;joinReferences:ResultTaskID;constraint:OnDelete:CASCADE" json:"result_tasks"` // 其他任务最近一次内存结果
	CustomLandingNodeEnable uint8          `gorm:"column:custom_landing_node_enable;default:0" json:"custom_landing_node_enable"`                                                  // 是否使用自定义落地节点检测前置节点
	LandingSubscriptions    []Subscription `gorm:"many2many:task_landing_subscriptions;constraint:OnDelete:CASCADE" json:"landing_subscriptions"`                                  // 自定义落地订阅来源
	LandingNodes            []Node         `gorm:"many2many:task_landing_nodes;constraint:OnDelete:CASCADE" json:"landing_nodes"`                                                  // 自定义落地节点来源
}

type TaskStep struct {
	Type        probe.ProbeType `json:"type"`                  // 步骤类型: delay/download/country
	Params      json.RawMessage `json:"params,omitempty"`      // 检测模块参数，由 pkg/probe 内部按类型解析
	Concurrency int             `json:"concurrency,omitempty"` // 本步骤并发数
	Pass        TaskPass        `json:"pass,omitempty"`        // 通过条件
	Order       string          `json:"order,omitempty"`       // 处理顺序: none/delay/speed
}

type TaskPass struct {
	Limit               int      `json:"limit,omitempty"`                 // 本步骤通过节点达到该数量后停止处理剩余节点
	MinDelay            uint16   `json:"min_delay,omitempty"`             // 最小延迟，单位毫秒
	MaxDelay            uint16   `json:"max_delay,omitempty"`             // 最大延迟，单位毫秒
	MinDownloadSpeed    uint64   `json:"min_download_speed,omitempty"`    // 最小下载速度，单位 bytes/s
	MaxDownloadSpeed    uint64   `json:"max_download_speed,omitempty"`    // 最大下载速度，单位 bytes/s
	IncludeCountryCodes []string `json:"include_country_codes,omitempty"` // 只保留这些国家代码
	ExcludeCountryCodes []string `json:"exclude_country_codes,omitempty"` // 排除这些国家代码
}

type StorageConfig struct {
	StorageEnable        uint8   `gorm:"column:storage_enable;default:0" json:"storage_enable"`                 // 是否在任务完成后储存
	StorageID            string  `gorm:"column:storage_id;type:varchar(36)" json:"storage_id"`                  // 储存目标 ID
	Storage              Storage `gorm:"foreignKey:StorageID;references:ID" json:"storage,omitempty"`           // 储存目标
	SavePath             string  `gorm:"column:save_path;type:text" json:"save_path"`                           // 储存路径
	NodeRenameExpression string  `gorm:"column:node_rename_expression;type:text" json:"node_rename_expression"` // 节点重命名表达式
}

type TaskResult struct {
	StartedAt  time.Time         `json:"started_at"`  // 开始时间
	FinishedAt time.Time         `json:"finished_at"` // 完成时间
	Groups     []TaskResultGroup `json:"groups"`      // 按订阅分组的节点指纹
}

type TaskResultGroup struct {
	SubscriptionID string   `json:"subscription_id"` // 订阅 ID
	Fingerprints   []uint64 `json:"fingerprints"`    // 该订阅内通过检测的节点指纹
}

func (t *Task) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}
