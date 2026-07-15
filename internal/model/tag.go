package model

type Tag struct {
	ID            uint              `gorm:"column:id;primaryKey;autoIncrement" json:"id"`                                            // ID
	Name          string            `gorm:"column:name;uniqueIndex;type:varchar(64);not null" json:"name"`                           // 标签名称
	Subscriptions []SubscriptionRef `gorm:"many2many:tag_subscriptions;joinForeignKey:TagID;joinReferences:SubscriptionID" json:"-"` // 关联订阅 ID。
	Nodes         []NodeRef         `gorm:"many2many:tag_nodes;joinForeignKey:TagID;joinReferences:NodeID" json:"-"`                 // 关联单独节点 ID。
}

// TagRef 是输入来源关联标签的轻量模型。
type TagRef struct {
	ID uint `gorm:"column:id;primaryKey;autoIncrement" json:"id"` // 标签 ID
}

func (TagRef) TableName() string {
	return "tags"
}
