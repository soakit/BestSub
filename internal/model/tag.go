package model

type Tag struct {
	ID   uint   `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name string `gorm:"column:name;uniqueIndex;type:varchar(64);not null" json:"name"` // 标签名称
}
