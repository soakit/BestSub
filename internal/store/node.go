package store

import (
	"fmt"

	"bestsub/internal/model"
	"bestsub/internal/utils/cache"

	"gorm.io/gorm"
)

var nodeCache = cache.New[string, model.Node](16)

func initNode() error {
	nodes := []model.Node{}
	if err := db.Find(&nodes).Error; err != nil {
		return fmt.Errorf("failed to load nodes: %w", err)
	}
	for _, node := range nodes {
		nodeCache.Set(node.ID, node)
	}
	return nil
}

func NodeCreate(node *model.Node) error {
	if err := db.Create(node).Error; err != nil {
		return err
	}
	nodeCache.Set(node.ID, *node)
	return nil
}

func NodeDelete(id string) error {
	if err := db.Delete(&model.Node{}, "id = ?", id).Error; err != nil {
		return err
	}
	nodeCache.Del(id)
	return nil
}

func NodeList() []model.Node {
	nodes := make([]model.Node, 0, nodeCache.Len())
	for _, node := range nodeCache.GetAll() {
		nodes = append(nodes, node)
	}
	return nodes
}

func NodeUpdate(id string, node *model.Node) error {
	tags := node.Tags
	node.Tags = nil

	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Node{}).Where("id = ?", id).Updates(map[string]interface{}{
			"name":    node.Name,
			"content": node.Content,
		}).Error; err != nil {
			return err
		}
		n := &model.Node{ID: id}
		if err := tx.Model(n).Association("Tags").Replace(tags); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	if n, ok := nodeCache.Get(id); ok {
		n.Name = node.Name
		n.Content = node.Content
		n.Tags = tags
		nodeCache.Set(id, n)
	}
	return nil
}
