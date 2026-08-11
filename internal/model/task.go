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
	CreateAt   time.Time `gorm:"column:create_at;autoCreateTime" json:"create_at"` // 创建时间
	FinishedAt time.Time `gorm:"column:finished_at" json:"finished_at"`            // 最近一次运行完成时间
}

// TaskConfig 保存任务基础配置和外部输入引用。
type TaskConfig struct {
	Name     string     `gorm:"column:name;type:varchar(255)" json:"name" binding:"required"`             // 任务名称
	AutoRun  uint8      `gorm:"column:auto_run;default:0" json:"auto_run"`                                // 是否自动运行
	CronExpr string     `gorm:"column:cron_expr;type:varchar(64)" json:"cron_expr"`                       // Cron 表达式
	Steps    []TaskStep `gorm:"column:steps;type:json;serializer:json" json:"steps" binding:"min=1,dive"` // 线性步骤列表
	TaskInput
	StorageConfig
}

// TaskInput 保存任务输入来源，关联项只保留 ID，避免任务缓存重复保存完整订阅和节点内容。
type TaskInput struct {
	AllInputEnable          uint8             `gorm:"column:all_input_enable;default:0" json:"all_input_enable"`                                                                      // 是否动态使用全部订阅节点池和全部单独节点
	Subscriptions           []SubscriptionRef `gorm:"many2many:task_input_subscriptions;constraint:OnDelete:CASCADE" json:"subscriptions"`                                            // 指定订阅的内存节点池
	Nodes                   []NodeRef         `gorm:"many2many:task_input_nodes;constraint:OnDelete:CASCADE" json:"nodes"`                                                            // 指定单独节点
	Tags                    []TagRef          `gorm:"many2many:task_input_tags;constraint:OnDelete:CASCADE" json:"tags"`                                                              // 指定 tag 下的订阅和单独节点
	ResultTasks             []TaskRef         `gorm:"many2many:task_input_results;joinForeignKey:TaskID;joinReferences:ResultTaskID;constraint:OnDelete:CASCADE" json:"result_tasks"` // 其他任务最近一次内存结果
	CustomLandingNodeEnable uint8             `gorm:"column:custom_landing_node_enable;default:0" json:"custom_landing_node_enable"`                                                  // 是否使用自定义落地节点检测前置节点
	LandingNodeID           *string           `gorm:"column:landing_node_id;type:varchar(36)" json:"-"`                                                                               // 自定义落地节点 ID，未启用时为空
	LandingNode             NodeRef           `gorm:"foreignKey:LandingNodeID;references:ID;constraint:OnDelete:SET NULL" json:"landing_node"`                                        // 自定义落地节点引用，接口只读写节点 ID
}

// TaskRef 是输入来源关联结果任务的轻量模型。
type TaskRef struct {
	ID string `gorm:"column:id;primaryKey;type:varchar(36)" json:"id"` // 结果来源任务 ID
}

func (TaskRef) TableName() string {
	return "tasks"
}

// OrderType 表示任务步骤的结果排序方式。
type OrderType uint8

const (
	OrderNone  OrderType = iota // 不排序
	OrderDelay                  // 按延迟排序,优先处理延迟小的
	OrderSpeed                  // 按下载速度降序,优先处理下载速度大的
)

// TaskStep 保存单个检测步骤。
type TaskStep struct {
	Type           probe.ProbeType `json:"type" binding:"oneof=delay speed country"` // 步骤类型: delay/speed/country
	Params         json.RawMessage `json:"params,omitempty"`                         // 检测模块参数，由 pkg/probe 内部按类型解析
	Concurrency    int             `json:"concurrency,omitempty"`                    // 本步骤并发数
	Pass           NodeFilter      `json:"pass,omitempty"`                           // 通过条件
	Order          OrderType       `json:"order" binding:"oneof=0 1 2"`              // 处理顺序：0=不排序，1=延迟升序，2=下载速度降序
	NodePoolDelete uint8           `json:"node_pool_delete,omitempty"`               // 探测失败时是否从订阅节点池删除节点
	SkipExisting   uint8           `json:"skip_existing,omitempty"`                  // 对应检测结果已存在时是否跳过本步骤探测
}

// StorageConfig 保存任务完成后的储存配置。
type StorageConfig struct {
	StorageEnable        uint8    `gorm:"column:storage_enable;default:0" json:"storage_enable"`                 // 是否在任务完成后储存
	StorageID            *string  `gorm:"column:storage_id;type:varchar(36)" json:"storage_id"`                  // 储存目标 ID，未启用储存时为空
	Storage              Storage  `gorm:"foreignKey:StorageID;references:ID" json:"storage,omitempty" binding:"-"` // 储存目标，仅用于接口展示，不参与请求校验
	SaveFormat           string   `gorm:"column:save_format;type:varchar(32)" json:"save_format"`                // 保存格式，对应订阅转换目标
	SavePath             string   `gorm:"column:save_path;type:text" json:"save_path"`                           // 储存路径
	NodeRenameExpression string   `gorm:"column:node_rename_expression;type:text" json:"node_rename_expression"` // 节点重命名表达式
}

func (t *Task) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}
