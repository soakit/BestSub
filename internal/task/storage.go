package task

import (
	"context"
	"fmt"

	"github.com/bestruirui/bestsub/internal/model"
	"github.com/bestruirui/bestsub/internal/node"
	"github.com/bestruirui/bestsub/internal/storage"
	"github.com/bestruirui/bestsub/internal/store"
)

// saveTaskOutput 将最终节点组装、转换后写入任务指定的储存目标。
func saveTaskOutput(ctx context.Context, task model.Task, nodes []node.Node) error {
	target, ok := store.StorageGet(task.StorageID)
	if !ok {
		return fmt.Errorf("storage target not found: %s", task.StorageID)
	}

	content, err := node.Mihomo(nodes, task.NodeRenameExpression)
	if err != nil {
		return fmt.Errorf("build task result Mihomo subscription: %w", err)
	}

	output, err := node.Convert(ctx, content, node.ConvertTarget(task.SaveFormat))
	if err != nil {
		return fmt.Errorf("convert task result: %w", err)
	}
	if err := storage.Put(ctx, storage.Type(target.Type), storage.Request{
		Params:  target.Params,
		Path:    task.SavePath,
		Content: output,
	}); err != nil {
		return fmt.Errorf("save task result: %w", err)
	}
	return nil
}
