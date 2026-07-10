package storage

import (
	"context"
	"encoding/json"
	"os"
	"path"
	"path/filepath"
)

const TypeLocal Type = "local"

type localBackend struct{}

func init() {
	register(TypeLocal, localBackend{})
}

func (localBackend) Put(ctx context.Context, req Request) error {
	target := filepath.FromSlash(req.Path)
	if !filepath.IsAbs(target) && !path.IsAbs(req.Path) {
		exe, err := os.Executable()
		if err != nil {
			return err
		}
		target = filepath.Join(filepath.Dir(exe), target)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}
	return os.WriteFile(target, req.Content, 0644)
}

func (localBackend) Test(ctx context.Context, _ json.RawMessage) error {
	return ctx.Err()
}
