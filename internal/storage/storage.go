package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type Type string // 储存类型。

type Request struct { // 单次储存写入请求，任务模块负责生成 Content
	Params  json.RawMessage // 储存类型私有参数。
	Path    string          // 储存路径或对象 key，由任务配置提供。
	Content []byte          // 完整文本文件内容。
}

type backend interface {
	Put(context.Context, Request) error
	Test(context.Context, json.RawMessage) error
}

var backends = map[Type]backend{} // 储存类型到后端实现的注册表。

func register(typ Type, b backend) {
	if typ == "" {
		panic("storage type is required")
	}
	if b == nil {
		panic("storage backend is required")
	}
	if _, ok := backends[typ]; ok {
		panic("storage type registered: " + string(typ))
	}
	backends[typ] = b
}

func Put(ctx context.Context, typ Type, req Request) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if strings.TrimSpace(req.Path) == "" {
		return fmt.Errorf("storage path is required")
	}
	b, ok := backends[typ]
	if !ok {
		return fmt.Errorf("unknown storage type: %s", typ)
	}
	err := b.Put(ctx, req)
	if err := ctx.Err(); err != nil {
		return err
	}
	return err
}

func Test(ctx context.Context, typ Type, params json.RawMessage) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	b, ok := backends[typ]
	if !ok {
		return fmt.Errorf("unknown storage type: %s", typ)
	}
	err := b.Test(ctx, params)
	if err := ctx.Err(); err != nil {
		return err
	}
	return err
}
