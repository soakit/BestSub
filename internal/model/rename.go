package model

type RenameTemplate struct { // 保存可复用的节点重命名模板。
	ID         int    `gorm:"column:id;primaryKey" json:"id"`                         // 数据库自增主键。
	Preview    string `gorm:"column:preview" json:"preview"`                          // 预览
	Expression string `gorm:"column:expression" json:"expression" binding:"required"` // 交给 Go 模板解析的重命名表达式。
}
