package store

import (
	"fmt"

	"github.com/bestruirui/bestsub/internal/model"
	"github.com/bestruirui/bestsub/internal/utils/cache"
)

var nodeCache = cache.New[string, model.Node](16) // 单独节点缓存，key 为节点 ID。

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
	if err := TagSetNodeNames(node.ID, node.TagNames); err != nil {
		return err
	}
	nodeCache.Set(node.ID, *node)
	return nil
}

func NodeDelete(id string) error {
	if err := db.Delete(&model.Node{}, "id = ?", id).Error; err != nil {
		return err
	}
	if err := TagSetNodeNames(id, nil); err != nil {
		return err
	}
	nodeCache.Del(id)
	return nil
}

func NodeList() []model.Node {
	nodes := make([]model.Node, 0, nodeCache.Len())
	for _, node := range nodeCache.GetAll() {
		node.TagNames = TagNamesByNode(node.ID)
		nodes = append(nodes, node)
	}
	return nodes
}

func NodeLen() int { return nodeCache.Len() }

func NodeGet(id string) (model.Node, bool) {
	node, ok := nodeCache.Get(id)
	if ok {
		node.TagNames = TagNamesByNode(node.ID)
	}
	return node, ok
}

func NodeUpdateInfo(id string, info model.NodeInfo) error {
	if err := db.Model(&model.Node{}).Where("id = ?", id).Select("*").Updates(info).Error; err != nil {
		return err
	}
	if node, ok := nodeCache.Get(id); ok {
		node.NodeInfo = info
		nodeCache.Set(id, node)
	}
	return nil
}

func NodeUpdate(id string, node *model.Node) error {
	if err := db.Model(&model.Node{}).Where("id = ?", id).Select("name", "content", "traffic_multiplier", "landing_only").Updates(node).Error; err != nil {
		return err
	}
	if err := TagSetNodeNames(id, node.TagNames); err != nil {
		return err
	}

	if n, ok := nodeCache.Get(id); ok {
		n.Name = node.Name
		n.Content = node.Content
		n.TrafficMultiplier = node.TrafficMultiplier
		n.LandingOnly = node.LandingOnly
		nodeCache.Set(id, n)
	}
	return nil
}
