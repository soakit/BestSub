package task

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/bestruirui/bestsub/internal/model"
	"github.com/bestruirui/bestsub/internal/store"
	"github.com/bestruirui/bestsub/pkg/mihomo"
	"github.com/bestruirui/bestsub/pkg/probe"

	"github.com/charmbracelet/log"
)

func runStep(ctx context.Context, taskID string, step model.TaskStep, nodes, landingNodes []stepNode) ([]stepNode, error) {
	if err := validateStep(step); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, nil
	}
	for i := range nodes {
		nodes[i].Index = i
	}
	concurrency := step.Concurrency
	if concurrency <= 0 {
		concurrency = 1
	}
	if concurrency > len(nodes) {
		concurrency = len(nodes)
	}

	jobs := make(chan stepNode)
	results := make(chan stepNode)
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for node := range jobs {
				next, err := runStepNode(ctx, step, node, landingNodes)
				if err != nil {
					if ctx.Err() == nil {
						handleStepFailure(step, node)
						addTaskProgressDone(taskID)
					}
					continue
				}
				if err := writeStepSuccess(step, next); err != nil {
					log.Errorf("write task node result error: %v", err)
					addTaskProgressDone(taskID)
					continue
				}
				addTaskProgressDone(taskID)
				select {
				case <-ctx.Done():
					return
				case results <- next:
				}
			}
		}()
	}

	go func() {
		defer close(jobs)
		for _, node := range nodes {
			select {
			case <-ctx.Done():
				return
			case jobs <- node:
			}
		}
	}()
	go func() {
		wg.Wait()
		close(results)
	}()

	out := make([]stepNode, 0, len(nodes))
	for node := range results {
		out = append(out, node)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return applyStepPassOrder(step, out)
}

func runStepNode(ctx context.Context, step model.TaskStep, node stepNode, landingNodes []stepNode) (stepNode, error) {
	if step.SkipExisting == 1 {
		switch step.Type {
		case probe.TypeDelay:
			if node.Info.Delay > 0 {
				return node, nil
			}
		case probe.TypeSpeed:
			if node.Info.DownloadSpeed > 0 {
				return node, nil
			}
		case probe.TypeCountry:
			if strings.TrimSpace(node.Info.CountryCode) != "" {
				return node, nil
			}
		}
	}
	if len(landingNodes) == 0 {
		transport, closeProxy, err := mihomo.NewTransport([]byte(node.Raw.Text), nil, store.SettingGet(model.SettingBindInterface))
		if err != nil {
			return node, err
		}
		defer func() {
			transport.CloseIdleConnections()
			if err := closeProxy(); err != nil {
				log.Errorf("close task node proxy error: %v", err)
			}
		}()
		return runProbe(ctx, step, node, &http.Client{Transport: transport})
	}

	var best stepNode
	var lastErr error
	for _, landing := range landingNodes {
		if err := ctx.Err(); err != nil {
			return node, err
		}
		transport, closeProxy, err := mihomo.NewTransport([]byte(node.Raw.Text), []byte(landing.Raw.Text), store.SettingGet(model.SettingBindInterface))
		if err != nil {
			lastErr = err
			continue
		}
		next, err := runProbe(ctx, step, node, &http.Client{Transport: transport})
		// 每个候选节点探测结束后立即释放内外层代理资源。
		transport.CloseIdleConnections()
		if err := closeProxy(); err != nil {
			log.Errorf("close task node proxy error: %v", err)
		}
		if err != nil {
			lastErr = err
			continue
		}
		if best.Raw == nil || betterStepNode(step, next, best) {
			best = next
		}
	}
	if best.Raw != nil {
		return best, nil
	}
	return node, lastErr
}

func runProbe(ctx context.Context, step model.TaskStep, node stepNode, client *http.Client) (stepNode, error) {
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
	default:
		return node, fmt.Errorf("unknown task step type: %s", step.Type)
	}
	return node, nil
}

func writeStepSuccess(step model.TaskStep, node stepNode) error {
	if node.SubscriptionID != "" {
		var ok bool
		switch step.Type {
		case probe.TypeDelay:
			ok = store.NodePoolUpdateDelay(node.SubscriptionID, node.Raw.Fingerprint, node.Info.Delay)
		case probe.TypeSpeed:
			ok = store.NodePoolUpdateDownloadSpeed(node.SubscriptionID, node.Raw.Fingerprint, node.Info.DownloadSpeed)
		case probe.TypeCountry:
			ok = store.NodePoolUpdateCountryCode(node.SubscriptionID, node.Raw.Fingerprint, node.Info.CountryCode)
		}
		if !ok {
			return fmt.Errorf("node pool item not found: %s %d", node.SubscriptionID, node.Raw.Fingerprint)
		}
	}
	if node.NodeID != "" {
		return store.NodeUpdateInfo(node.NodeID, node.Info)
	}
	return nil
}

func handleStepFailure(step model.TaskStep, node stepNode) {
	if node.SubscriptionID != "" && step.NodePoolDelete == 1 {
		store.NodePoolDelete(node.SubscriptionID, node.Raw.Fingerprint)
		return
	}
	if node.NodeID == "" {
		return
	}
	switch step.Type {
	case probe.TypeDelay:
		node.Info.Delay = 0
	case probe.TypeSpeed:
		node.Info.DownloadSpeed = 0
	case probe.TypeCountry:
		node.Info.CountryCode = ""
	}
	if err := store.NodeUpdateInfo(node.NodeID, node.Info); err != nil {
		log.Errorf("update failed node info error: %v", err)
	}
}

func applyStepPassOrder(step model.TaskStep, nodes []stepNode) ([]stepNode, error) {
	out := nodes[:0]
	for _, node := range nodes {
		if passStep(step, node) {
			out = append(out, node)
		}
	}
	switch strings.ToLower(strings.TrimSpace(step.Order)) {
	case "", "none":
		sort.SliceStable(out, func(i, j int) bool { return out[i].Index < out[j].Index })
	case "delay":
		sort.SliceStable(out, func(i, j int) bool { return out[i].Info.Delay < out[j].Info.Delay })
	case "speed":
		sort.SliceStable(out, func(i, j int) bool { return out[i].Info.DownloadSpeed > out[j].Info.DownloadSpeed })
	default:
		return nil, fmt.Errorf("unknown task step order: %s", step.Order)
	}
	if step.Pass.Limit > 0 && len(out) > step.Pass.Limit {
		out = out[:step.Pass.Limit]
	}
	return out, nil
}

func passStep(step model.TaskStep, node stepNode) bool {
	switch step.Type {
	case probe.TypeDelay:
		return (step.Pass.MinDelay == 0 || node.Info.Delay >= step.Pass.MinDelay) &&
			(step.Pass.MaxDelay == 0 || node.Info.Delay <= step.Pass.MaxDelay)
	case probe.TypeSpeed:
		return (step.Pass.MinDownloadSpeed == 0 || node.Info.DownloadSpeed >= step.Pass.MinDownloadSpeed) &&
			(step.Pass.MaxDownloadSpeed == 0 || node.Info.DownloadSpeed <= step.Pass.MaxDownloadSpeed)
	case probe.TypeCountry:
		countryCode := strings.ToUpper(strings.TrimSpace(node.Info.CountryCode))
		if len(step.Pass.IncludeCountryCodes) > 0 && !countryInList(countryCode, step.Pass.IncludeCountryCodes) {
			return false
		}
		return !countryInList(countryCode, step.Pass.ExcludeCountryCodes)
	}
	return false
}

func validateStep(step model.TaskStep) error {
	switch step.Type {
	case probe.TypeDelay, probe.TypeSpeed, probe.TypeCountry:
	default:
		return fmt.Errorf("unknown task step type: %s", step.Type)
	}
	switch strings.ToLower(strings.TrimSpace(step.Order)) {
	case "", "none", "delay", "speed":
		return nil
	default:
		return fmt.Errorf("unknown task step order: %s", step.Order)
	}
}

func betterStepNode(step model.TaskStep, next, best stepNode) bool {
	switch step.Type {
	case probe.TypeDelay:
		return next.Info.Delay < best.Info.Delay
	case probe.TypeSpeed:
		return next.Info.DownloadSpeed > best.Info.DownloadSpeed
	}
	return false
}

func countryInList(countryCode string, list []string) bool {
	for _, item := range list {
		if countryCode == strings.ToUpper(strings.TrimSpace(item)) {
			return true
		}
	}
	return false
}
