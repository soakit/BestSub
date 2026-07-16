package task

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"sort"
	"sync"

	"github.com/bestruirui/bestsub/internal/model"
	"github.com/bestruirui/bestsub/internal/node"
	"github.com/bestruirui/bestsub/internal/store"
	"github.com/bestruirui/bestsub/pkg/mihomo"
	"github.com/bestruirui/bestsub/pkg/probe"

	"github.com/charmbracelet/log"
)

type stepResult struct { // 单个节点检测完成后的结果。
	current node.Node // 成功探测或跳过后保留的节点信息。
	success bool      // 是否参与当前步骤通过条件判断。
}

// runStep 并发检测节点，在满足通过条件的节点达到 Limit 后停止派发剩余节点。
func runStep(ctx context.Context, taskID string, step model.TaskStep, nodes []node.Node, landingNode *node.Raw) ([]node.Node, error) {
	if len(nodes) == 0 {
		return nil, nil
	}
	concurrency := step.Concurrency
	if concurrency <= 0 {
		concurrency = 1
	}
	if concurrency > len(nodes) {
		concurrency = len(nodes)
	}

	jobs := make(chan node.Node)
	results := make(chan stepResult)
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for current := range jobs {
				next := current
				skip := step.SkipExisting == 1 && ((step.Type == probe.TypeDelay && current.Info.Delay > 0) ||
					(step.Type == probe.TypeSpeed && current.Info.DownloadSpeed > 0) ||
					(step.Type == probe.TypeCountry && current.Info.CountryCode != "")) // 已有当前类型检测值时直接复用，不写回。
				success := true // 是否将当前节点交给通过条件判断。
				if !skip {
					var err error
					next, err = runStepNode(ctx, step, current, landingNode)
					if err != nil {
						if ctx.Err() != nil {
							return
						}
						// 订阅节点配置删除时只从节点池清除；其余失败节点清空本步骤检测值。
						if current.SubscriptionID != "" && step.NodePoolDelete == 1 {
							node.PoolDelete(current.SubscriptionID, current.Raw.Fingerprint)
						} else if current.NodeID != "" {
							switch step.Type {
							case probe.TypeDelay:
								current.Info.Delay = 0
							case probe.TypeSpeed:
								current.Info.DownloadSpeed = 0
							case probe.TypeCountry:
								current.Info.CountryCode = ""
							}
							if err := store.NodeUpdateInfo(current.NodeID, current.Info); err != nil {
								log.Errorf("update failed node info error: %v", err)
							}
						}
						success = false
					} else if err := writeStepSuccess(step, next); err != nil {
						log.Errorf("write task node result error: %v", err)
						success = false
					}
				}
				addTaskProgressDone(taskID)
				select {
				case <-ctx.Done():
					return
				case results <- stepResult{current: next, success: success}:
				}
			}
		}()
	}
	defer func() {
		close(jobs)
		wg.Wait()
	}()

	out := make([]node.Node, 0, len(nodes))
	next, running := 0, 0
	for next < len(nodes) && running < concurrency {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case jobs <- nodes[next]:
			next++
			running++
		}
	}
	for running > 0 {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case result := <-results:
			running--
			if result.success && passStep(step, result.current) && (step.Pass.Limit <= 0 || len(out) < step.Pass.Limit) {
				out = append(out, result.current)
				if step.Pass.Limit > 0 && len(out) == step.Pass.Limit {
					setTaskProgressTotal(taskID, next)
				}
			}
			if next < len(nodes) && (step.Pass.Limit <= 0 || len(out) < step.Pass.Limit) {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case jobs <- nodes[next]:
					next++
					running++
				}
			}
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	switch step.Order {
	case model.OrderNone:
	case model.OrderDelay:
		sort.SliceStable(out, func(i, j int) bool { return out[i].Info.Delay < out[j].Info.Delay })
	case model.OrderSpeed:
		sort.SliceStable(out, func(i, j int) bool { return out[i].Info.DownloadSpeed > out[j].Info.DownloadSpeed })
	}
	return out, nil
}

func runStepNode(ctx context.Context, step model.TaskStep, node node.Node, landingNode *node.Raw) (node.Node, error) {
	var landingRaw []byte
	if landingNode != nil {
		landingRaw = []byte(landingNode.Text)
	}
	transport, closeProxy, err := mihomo.NewTransport([]byte(node.Raw.Text), landingRaw, store.SettingGet(model.SettingBindInterface))
	if err != nil {
		return node, err
	}
	// 无论是否启用落地节点，探测结束后都同时释放传输连接和代理实例。
	defer func() {
		transport.CloseIdleConnections()
		if err := closeProxy(); err != nil {
			log.Errorf("close task node proxy error: %v", err)
		}
	}()
	client := &http.Client{Transport: transport}
	switch step.Type {
	case probe.TypeDelay:
		var delay uint16
		if err := probe.Run(ctx, step.Type, client, step.Params, &delay); err != nil {
			return node, err
		}
		node.Info.Delay = delay
	case probe.TypeSpeed:
		var speed uint32
		if err := probe.Run(ctx, step.Type, client, step.Params, &speed); err != nil {
			return node, err
		}
		node.Info.DownloadSpeed = speed
	case probe.TypeCountry:
		var countryCode string
		if err := probe.Run(ctx, step.Type, client, step.Params, &countryCode); err != nil {
			return node, err
		}
		node.Info.CountryCode = countryCode
	}
	return node, nil
}

func writeStepSuccess(step model.TaskStep, current node.Node) error {
	if current.SubscriptionID != "" {
		var ok bool
		switch step.Type {
		case probe.TypeDelay:
			ok = node.PoolUpdateDelay(current.SubscriptionID, current.Raw.Fingerprint, current.Info.Delay)
		case probe.TypeSpeed:
			ok = node.PoolUpdateDownloadSpeed(current.SubscriptionID, current.Raw.Fingerprint, current.Info.DownloadSpeed)
		case probe.TypeCountry:
			ok = node.PoolUpdateCountryCode(current.SubscriptionID, current.Raw.Fingerprint, current.Info.CountryCode)
		}
		if !ok {
			return fmt.Errorf("node pool item not found: %s %d", current.SubscriptionID, current.Raw.Fingerprint)
		}
	}
	if current.NodeID != "" {
		return store.NodeUpdateInfo(current.NodeID, current.Info)
	}
	return nil
}

func passStep(step model.TaskStep, node node.Node) bool {
	switch step.Type {
	case probe.TypeDelay:
		return (step.Pass.MinDelay == 0 || node.Info.Delay >= step.Pass.MinDelay) &&
			(step.Pass.MaxDelay == 0 || node.Info.Delay <= step.Pass.MaxDelay)
	case probe.TypeSpeed:
		return (step.Pass.MinDownloadSpeed == 0 || node.Info.DownloadSpeed >= step.Pass.MinDownloadSpeed) &&
			(step.Pass.MaxDownloadSpeed == 0 || node.Info.DownloadSpeed <= step.Pass.MaxDownloadSpeed)
	case probe.TypeCountry:
		return (len(step.Pass.IncludeCountryCodes) == 0 || slices.Contains(step.Pass.IncludeCountryCodes, node.Info.CountryCode)) &&
			!slices.Contains(step.Pass.ExcludeCountryCodes, node.Info.CountryCode)
	}
	return false
}
