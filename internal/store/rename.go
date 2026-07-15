package store

import (
	"fmt"

	"github.com/bestruirui/bestsub/internal/model"
	"github.com/bestruirui/bestsub/internal/utils/cache"
)

var renameTemplateCache = cache.New[int, model.RenameTemplate](4) // 重命名模板缓存，key 为数据库主键。

// initRenameTemplate 在数据库连接建立后把已有模板加载到缓存。
func initRenameTemplate() error {
	templates := []model.RenameTemplate{}
	if err := db.Find(&templates).Error; err != nil {
		return fmt.Errorf("failed to load rename templates: %w", err)
	}
	for _, template := range templates {
		renameTemplateCache.Set(template.ID, template)
	}
	return nil
}

// RenameTemplateList 返回缓存中的全部重命名模板。
func RenameTemplateList() []model.RenameTemplate {
	templates := make([]model.RenameTemplate, 0, renameTemplateCache.Len())
	for _, template := range renameTemplateCache.GetAll() {
		templates = append(templates, template)
	}
	return templates
}

func RenameTemplateGet(id int) (model.RenameTemplate, bool) {
	return renameTemplateCache.Get(id)
}

// RenameTemplateCreate 写库成功后同步新增缓存。
func RenameTemplateCreate(template *model.RenameTemplate) error {
	if template == nil {
		return fmt.Errorf("rename template is required")
	}
	if err := db.Create(template).Error; err != nil {
		return err
	}
	renameTemplateCache.Set(template.ID, *template)
	return nil
}

// RenameTemplateDelete 删库成功后同步删除缓存。
func RenameTemplateDelete(id int) error {
	if err := db.Delete(&model.RenameTemplate{}, id).Error; err != nil {
		return err
	}
	renameTemplateCache.Del(id)
	return nil
}
