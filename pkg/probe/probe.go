package probe

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type ProbeType string

type runner func(context.Context, *http.Client, json.RawMessage, any) error

var runners = map[ProbeType]runner{} // 保存探测类型到执行函数的注册关系。

// httpParams 保存各 HTTP 探测共用的请求参数。
type httpParams struct {
	URL       string `json:"url"`                  // 检测请求目标地址。
	TimeoutMS int    `json:"timeout_ms,omitempty"` // 单次请求超时时间，单位毫秒；0 使用模块默认值。
}

func register(typ ProbeType, run runner) {
	if typ == "" {
		panic("probe type is required")
	}
	if run == nil {
		panic("probe runner is required")
	}
	if _, ok := runners[typ]; ok {
		panic("probe type registered: " + string(typ))
	}
	runners[typ] = run
}

func Run(ctx context.Context, typ ProbeType, client *http.Client, params json.RawMessage, result any) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if client == nil {
		return fmt.Errorf("client is required")
	}
	if result == nil {
		return fmt.Errorf("result is required")
	}
	run, ok := runners[typ]
	if !ok {
		return fmt.Errorf("unknown probe type: %s", typ)
	}
	err := run(ctx, client, params, result)
	if err := ctx.Err(); err != nil {
		return err
	}
	return err
}
