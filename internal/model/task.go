package model

import (
	"encoding/json"
	"time"

	"github.com/bestruirui/bestsub/pkg/probe"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Task 是任务配置的持久化模型。
type Task struct {
	ID string `gorm:"column:id;primaryKey;type:varchar(36)" json:"id"` // ID
	TaskConfig
	FinishedAt time.Time `gorm:"column:finished_at" json:"finished_at"` // 最近一次运行完成时间
}

// TaskConfig 保存任务基础配置和外部输入引用。
type TaskConfig struct {
	Name     string     `gorm:"column:name;type:varchar(255)" json:"name"`                              // 任务名称
	AutoRun  uint8      `gorm:"column:auto_run;default:0" json:"auto_run"`                              // 是否自动运行
	CronExpr string     `gorm:"column:cron_expr;type:varchar(64)" json:"cron_expr"`                     // Cron 表达式
	Steps    []TaskStep `gorm:"column:steps;type:json;serializer:json" json:"steps" binding:"required"` // 线性步骤列表
	TaskInput
	StorageConfig
}

// TaskInput 保存任务输入来源，关联项只保留 ID，避免任务缓存重复保存完整订阅和节点内容。
type TaskInput struct {
	Subscriptions           []SubscriptionRef `gorm:"many2many:task_input_subscriptions;constraint:OnDelete:CASCADE" json:"subscriptions"`                                            // 指定订阅的内存节点池
	Nodes                   []NodeRef         `gorm:"many2many:task_input_nodes;constraint:OnDelete:CASCADE" json:"nodes"`                                                            // 指定单独节点
	Tags                    []TagRef          `gorm:"many2many:task_input_tags;constraint:OnDelete:CASCADE" json:"tags"`                                                              // 指定 tag 下的订阅和单独节点
	ResultTasks             []TaskRef         `gorm:"many2many:task_input_results;joinForeignKey:TaskID;joinReferences:ResultTaskID;constraint:OnDelete:CASCADE" json:"result_tasks"` // 其他任务最近一次内存结果
	CustomLandingNodeEnable uint8             `gorm:"column:custom_landing_node_enable;default:0" json:"custom_landing_node_enable"`                                                  // 是否使用自定义落地节点检测前置节点
	LandingSubscriptions    []SubscriptionRef `gorm:"many2many:task_landing_subscriptions;constraint:OnDelete:CASCADE" json:"landing_subscriptions"`                                  // 自定义落地订阅来源
	LandingNodes            []NodeRef         `gorm:"many2many:task_landing_nodes;constraint:OnDelete:CASCADE" json:"landing_nodes"`                                                  // 自定义落地节点来源
}

// TaskRef 是输入来源关联结果任务的轻量模型。
type TaskRef struct {
	ID string `gorm:"column:id;primaryKey;type:varchar(36)" json:"id"` // 结果来源任务 ID
}

func (TaskRef) TableName() string {
	return "tasks"
}

// TaskStep 保存单个检测步骤。
type TaskStep struct {
	Type           probe.ProbeType `json:"type"`                       // 步骤类型: delay/speed/country
	Params         json.RawMessage `json:"params,omitempty"`           // 检测模块参数，由 pkg/probe 内部按类型解析
	Concurrency    int             `json:"concurrency,omitempty"`      // 本步骤并发数
	Pass           NodeFilter      `json:"pass,omitempty"`             // 通过条件
	Order          string          `json:"order,omitempty"`            // 处理顺序: none/delay/speed
	NodePoolDelete uint8           `json:"node_pool_delete,omitempty"` // 探测失败时是否从订阅节点池删除节点
}

// StorageConfig 保存任务完成后的储存配置。
type StorageConfig struct {
	StorageEnable        uint8   `gorm:"column:storage_enable;default:0" json:"storage_enable"`                 // 是否在任务完成后储存
	StorageID            string  `gorm:"column:storage_id;type:varchar(36)" json:"storage_id"`                  // 储存目标 ID
	Storage              Storage `gorm:"foreignKey:StorageID;references:ID" json:"storage,omitempty"`           // 储存目标
	SaveFormat           string  `gorm:"column:save_format;type:varchar(32)" json:"save_format"`                // 保存格式，对应订阅转换目标
	SavePath             string  `gorm:"column:save_path;type:text" json:"save_path"`                           // 储存路径
	NodeRenameExpression string  `gorm:"column:node_rename_expression;type:text" json:"node_rename_expression"` // 节点重命名表达式
}

func (t *Task) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}
