package node

import (
	"strings"
	"sync"

	"github.com/bestruirui/bestsub/internal/model"
)

type poolNode struct { // 订阅节点池内部状态，来源由外层 subscriptionID 表达。
	Raw  *Raw           // 节点原文及指纹。
	Info model.NodeInfo // 节点附加信息。
}

var (
	poolMu sync.RWMutex                        // 保护 pool 的并发读写。
	pool   = map[string]map[uint64]*poolNode{} // 按订阅 ID 和节点指纹分组的内存节点池。
)

type PoolFilter struct { // 节点池筛选条件，0 值表示不限制对应条件。
	MinDelay            uint16   // 最小延迟，单位毫秒；0 表示不限制。
	MaxDelay            uint16   // 最大延迟，单位毫秒；0 表示不限制。
	MinDownloadSpeed    uint32   // 最小下载速度，单位 kb/s；0 表示不限制。
	MaxDownloadSpeed    uint32   // 最大下载速度，单位 kb/s；0 表示不限制。
	IncludeCountryCodes []string // 非空时只保留这些 ISO 3166-1 alpha-2 国家代码。
	ExcludeCountryCodes []string // 非空时排除这些 ISO 3166-1 alpha-2 国家代码。
}

// PoolAdd 向指定订阅写入节点；重复指纹只刷新原文和倍率并返回 false。
func PoolAdd(subscriptionID string, proxy []byte) bool {
	raw := string(proxy)
	fingerprint := Fingerprint(proxy)
	multiplier := ParseTrafficMultiplier(raw)

	poolMu.Lock()
	defer poolMu.Unlock()

	if pool[subscriptionID] == nil {
		pool[subscriptionID] = map[uint64]*poolNode{}
	}
	if current := pool[subscriptionID][fingerprint]; current != nil {
		// 刷新同一节点时替换不可变原文和派生倍率，同时保留已有检测信息。
		current.Raw = &Raw{Text: raw, Fingerprint: fingerprint}
		current.Info.TrafficMultiplier = multiplier
		return false
	}
	pool[subscriptionID][fingerprint] = &poolNode{
		Raw:  &Raw{Text: raw, Fingerprint: fingerprint},
		Info: model.NodeInfo{TrafficMultiplier: multiplier},
	}
	return true
}

// PoolUpdateInfo 更新节点全部附加信息。
func PoolUpdateInfo(subscriptionID string, fingerprint uint64, info model.NodeInfo) bool {
	poolMu.Lock()
	defer poolMu.Unlock()
	if current := pool[subscriptionID][fingerprint]; current != nil {
		current.Info = info
		return true
	}
	return false
}

// PoolUpdateDelay 更新节点延迟，单位毫秒。
func PoolUpdateDelay(subscriptionID string, fingerprint uint64, delay uint16) bool {
	poolMu.Lock()
	defer poolMu.Unlock()
	if current := pool[subscriptionID][fingerprint]; current != nil {
		current.Info.Delay = delay
		return true
	}
	return false
}

// PoolUpdateDownloadSpeed 更新节点下载速度，单位 kb/s。
func PoolUpdateDownloadSpeed(subscriptionID string, fingerprint uint64, speed uint32) bool {
	poolMu.Lock()
	defer poolMu.Unlock()
	if current := pool[subscriptionID][fingerprint]; current != nil {
		current.Info.DownloadSpeed = speed
		return true
	}
	return false
}

// PoolUpdateCountryCode 更新节点落地国家。
func PoolUpdateCountryCode(subscriptionID string, fingerprint uint64, countryCode string) bool {
	poolMu.Lock()
	defer poolMu.Unlock()
	if current := pool[subscriptionID][fingerprint]; current != nil {
		current.Info.CountryCode = countryCode
		return true
	}
	return false
}

// PoolDelete 从指定订阅删除指定指纹的节点。
func PoolDelete(subscriptionID string, fingerprint uint64) bool {
	poolMu.Lock()
	defer poolMu.Unlock()

	nodes := pool[subscriptionID]
	if nodes[fingerprint] == nil {
		return false
	}
	delete(nodes, fingerprint)
	if len(nodes) == 0 {
		delete(pool, subscriptionID)
	}
	return true
}

// PoolClear 清空指定订阅的所有节点。
func PoolClear(subscriptionID string) {
	poolMu.Lock()
	defer poolMu.Unlock()
	delete(pool, subscriptionID)
}

// PoolCount 返回指定订阅的节点数量。
func PoolCount(subscriptionID string) int {
	poolMu.RLock()
	defer poolMu.RUnlock()
	return len(pool[subscriptionID])
}

// PoolListBySubscription 返回指定订阅节点的独立切片。
func PoolListBySubscription(subscriptionID string) []Node {
	poolMu.RLock()
	defer poolMu.RUnlock()

	result := make([]Node, 0, len(pool[subscriptionID]))
	for _, current := range pool[subscriptionID] {
		result = append(result, Node{SubscriptionID: subscriptionID, Raw: current.Raw, Info: current.Info})
	}
	return result
}

// PoolFilterNodes 按节点检测信息筛选指定订阅的节点原文。
func PoolFilterNodes(subscriptionID string, filter PoolFilter) []*Raw {
	includeCountries := map[string]struct{}{}
	for _, countryCode := range filter.IncludeCountryCodes {
		includeCountries[strings.ToUpper(countryCode)] = struct{}{}
	}
	excludeCountries := map[string]struct{}{}
	for _, countryCode := range filter.ExcludeCountryCodes {
		excludeCountries[strings.ToUpper(countryCode)] = struct{}{}
	}

	poolMu.RLock()
	defer poolMu.RUnlock()

	nodes := pool[subscriptionID]
	result := make([]*Raw, 0, len(nodes))
	for _, current := range nodes {
		if filter.MinDelay > 0 && current.Info.Delay < filter.MinDelay {
			continue
		}
		if filter.MaxDelay > 0 && current.Info.Delay > filter.MaxDelay {
			continue
		}
		if filter.MinDownloadSpeed > 0 && current.Info.DownloadSpeed < filter.MinDownloadSpeed {
			continue
		}
		if filter.MaxDownloadSpeed > 0 && current.Info.DownloadSpeed > filter.MaxDownloadSpeed {
			continue
		}
		if len(includeCountries) > 0 {
			if _, ok := includeCountries[strings.ToUpper(current.Info.CountryCode)]; !ok {
				continue
			}
		}
		if _, ok := excludeCountries[strings.ToUpper(current.Info.CountryCode)]; ok {
			continue
		}
		result = append(result, current.Raw)
	}
	return result
}
