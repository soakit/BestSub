package model

type Tag struct {
	ID            uint              `gorm:"column:id;primaryKey;autoIncrement" json:"id"`                                            // ID
	Name          string            `gorm:"column:name;uniqueIndex;type:varchar(64);not null" json:"name"`                           // 标签名称
	Subscriptions []TagSubscription `gorm:"many2many:tag_subscriptions;joinForeignKey:TagID;joinReferences:SubscriptionID" json:"-"` // 关联订阅 ID。
	Nodes         []TagNode         `gorm:"many2many:tag_nodes;joinForeignKey:TagID;joinReferences:NodeID" json:"-"`                 // 关联单独节点 ID。
}

type TagSubscription struct {
	ID string `gorm:"column:id;primaryKey;type:varchar(36)" json:"id"` // 订阅 ID
}

func (TagSubscription) TableName() string {
	return "subscriptions"
}

type TagNode struct {
	ID string `gorm:"column:id;primaryKey;type:varchar(36)" json:"id"` // 单独节点 ID
}

func (TagNode) TableName() string {
	return "nodes"
}
