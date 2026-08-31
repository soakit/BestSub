package migrate

import (
	"fmt"

	"github.com/bestruirui/bestsub/internal/model"

	"gorm.io/gorm"
)

func init() {
	RegisterBeforeAutoMigration(Migration{
		Version: 1,
		Up:      tagSubscriptionsCascade,
	})
}

// tagSubscriptionRow 仅用于迁移时备份与恢复 tag_subscriptions 关联数据。
type tagSubscriptionRow struct {
	TagID          uint   `gorm:"column:tag_id"`
	SubscriptionID string `gorm:"column:subscription_id"`
}

// tagSubscriptionsCascade 重建 tag_subscriptions 关联表，使外键带 ON DELETE CASCADE，
// 删除订阅或标签时能自动清理关联记录，避免外键约束导致删除失败。
func tagSubscriptionsCascade(db *gorm.DB) error {
	// 新库尚无该表，AutoMigrate 会按当前模型创建，无需迁移。
	if !db.Migrator().HasTable("tag_subscriptions") {
		return nil
	}

	return db.Transaction(func(tx *gorm.DB) error {
		var rows []tagSubscriptionRow
		if err := tx.Table("tag_subscriptions").Find(&rows).Error; err != nil {
			return fmt.Errorf("backup tag_subscriptions: %w", err)
		}
		// 先挪开旧表，让 AutoMigrate 按当前模型重建缺失的 join 表（含 CASCADE 外键）。
		if err := tx.Migrator().RenameTable("tag_subscriptions", "tag_subscriptions_old"); err != nil {
			return fmt.Errorf("rename tag_subscriptions: %w", err)
		}
		if err := tx.AutoMigrate(&model.Tag{}); err != nil {
			return fmt.Errorf("recreate tag_subscriptions: %w", err)
		}
		if len(rows) > 0 {
			if err := tx.Table("tag_subscriptions").Create(&rows).Error; err != nil {
				return fmt.Errorf("restore tag_subscriptions: %w", err)
			}
		}
		if err := tx.Migrator().DropTable("tag_subscriptions_old"); err != nil {
			return fmt.Errorf("drop tag_subscriptions_old: %w", err)
		}
		return nil
	})
}
