package store

import (
	"fmt"
	"strings"
	"sync"

	"github.com/cespare/xxhash/v2"
	"gopkg.in/yaml.v3"

	"github.com/bestruirui/bestsub/internal/model"
)

// nodePool: map[subID]map[fingerprint]*nodePoolNode
// 外层 key 是订阅 ID，内层 key 是节点指纹，天然去重。
var (
	nodePoolLock sync.RWMutex                            // 保护 nodePool 的并发读写
	nodePool     = map[string]map[uint64]*nodePoolNode{} // 内存节点池，按订阅 ID 和节点指纹分组
)

type nodePoolNode struct { // 节点池内部节点，保存原始节点和检测信息
	Proxy []byte         // YAML 行内格式原始节点字节。
	Info  model.NodeInfo // 节点测试信息，可独立更新。
}

type NodePoolItem struct { // 节点池导出项，用于检测后按指纹写回结果
	Fingerprint uint64 // 节点指纹，用于定位节点池中的唯一节点。
	Proxy       []byte // YAML 行内格式原始节点字节。
}

type NodePoolFilter struct { // 节点池筛选条件，0 值表示不限制对应条件
	MinDelay            uint16   // 最小延迟，单位毫秒；0 表示不限制。
	MaxDelay            uint16   // 最大延迟，单位毫秒；0 表示不限制。
	MinDownloadSpeed    uint64   // 最小下载速度，单位 bytes/s；0 表示不限制。
	MaxDownloadSpeed    uint64   // 最大下载速度，单位 bytes/s；0 表示不限制。
	IncludeCountryCodes []string // 非空时只保留这些 ISO 3166-1 alpha-2 国家代码。
	ExcludeCountryCodes []string // 非空时排除这些 ISO 3166-1 alpha-2 国家代码。
}

// nodeFingerprint 从 YAML 行内格式 raw 中解析 type/server/port 计算节点指纹。
func nodeFingerprint(raw []byte) uint64 {
	var key struct {
		Type       string `yaml:"type"`
		Server     string `yaml:"server"`
		Port       any    `yaml:"port"`
		Password   string `yaml:"password"`
		UUID       string `yaml:"uuid"`
		ServerName string `yaml:"servername"`
	}
	_ = yaml.Unmarshal(raw, &key)
	return xxhash.Sum64String(fmt.Sprintf("%v", key))
}

// NodePoolAdd 向指定订阅新增节点，订阅内指纹重复时拒绝。
func NodePoolAdd(subID string, proxy []byte) bool {
	fp := nodeFingerprint(proxy)

	nodePoolLock.Lock()
	defer nodePoolLock.Unlock()

	if nodePool[subID] == nil {
		nodePool[subID] = map[uint64]*nodePoolNode{}
	}
	if _, dup := nodePool[subID][fp]; dup {
		return false
	}
	nodePool[subID][fp] = &nodePoolNode{Proxy: proxy}
	return true
}

// NodePoolUpdateInfo 更新节点测试信息。
func NodePoolUpdateInfo(subID string, fingerprint uint64, info model.NodeInfo) bool {
	nodePoolLock.Lock()
	defer nodePoolLock.Unlock()

	if n := nodePool[subID][fingerprint]; n != nil {
		n.Info = info
		return true
	}
	return false
}

// NodePoolUpdateDelay 更新节点延迟，单位毫秒。
func NodePoolUpdateDelay(subID string, fingerprint uint64, delay uint16) bool {
	nodePoolLock.Lock()
	defer nodePoolLock.Unlock()

	if n := nodePool[subID][fingerprint]; n != nil {
		n.Info.Delay = delay
		return true
	}
	return false
}

// NodePoolUpdateDownloadSpeed 更新节点下载速度，单位 bytes/s。
func NodePoolUpdateDownloadSpeed(subID string, fingerprint uint64, speed uint64) bool {
	nodePoolLock.Lock()
	defer nodePoolLock.Unlock()

	if n := nodePool[subID][fingerprint]; n != nil {
		n.Info.DownloadSpeed = speed
		return true
	}
	return false
}

// NodePoolUpdateCountryCode 更新节点落地国家。
func NodePoolUpdateCountryCode(subID string, fingerprint uint64, countryCode string) bool {
	nodePoolLock.Lock()
	defer nodePoolLock.Unlock()

	if n := nodePool[subID][fingerprint]; n != nil {
		n.Info.CountryCode = countryCode
		return true
	}
	return false
}

// NodePoolDelete 从指定订阅删除指定指纹的节点。
func NodePoolDelete(subID string, fingerprint uint64) bool {
	nodePoolLock.Lock()
	defer nodePoolLock.Unlock()

	sp := nodePool[subID]
	if sp == nil || sp[fingerprint] == nil {
		return false
	}
	delete(sp, fingerprint)
	if len(sp) == 0 {
		delete(nodePool, subID)
	}
	return true
}

// NodePoolClear 清空指定订阅的所有节点。
func NodePoolClear(subID string) {
	nodePoolLock.Lock()
	defer nodePoolLock.Unlock()
	delete(nodePool, subID)
}

// NodePoolCount 返回指定订阅的节点数量。
func NodePoolCount(subID string) int {
	nodePoolLock.RLock()
	defer nodePoolLock.RUnlock()
	return len(nodePool[subID])
}

// NodePoolListBySubscription 按订阅 ID 返回节点指纹和原始字节列表。
func NodePoolListBySubscription(subID string) []NodePoolItem {
	nodePoolLock.RLock()
	defer nodePoolLock.RUnlock()

	sp := nodePool[subID]
	if sp == nil {
		return nil
	}
	result := make([]NodePoolItem, 0, len(sp))
	for fingerprint, n := range sp {
		result = append(result, NodePoolItem{
			Fingerprint: fingerprint,
			Proxy:       n.Proxy,
		})
	}
	return result
}

// NodePoolFilterNodes 按节点测试信息筛选指定订阅的节点。
func NodePoolFilterNodes(subID string, filter NodePoolFilter) [][]byte {
	includeCountries := map[string]struct{}{}
	for _, cc := range filter.IncludeCountryCodes {
		includeCountries[strings.ToUpper(cc)] = struct{}{}
	}
	excludeCountries := map[string]struct{}{}
	for _, cc := range filter.ExcludeCountryCodes {
		excludeCountries[strings.ToUpper(cc)] = struct{}{}
	}

	nodePoolLock.RLock()
	defer nodePoolLock.RUnlock()

	sp := nodePool[subID]
	if sp == nil {
		return nil
	}
	result := make([][]byte, 0, len(sp))
	for _, n := range sp {
		if filter.MinDelay > 0 && n.Info.Delay < filter.MinDelay {
			continue
		}
		if filter.MaxDelay > 0 && n.Info.Delay > filter.MaxDelay {
			continue
		}
		if filter.MinDownloadSpeed > 0 && n.Info.DownloadSpeed < filter.MinDownloadSpeed {
			continue
		}
		if filter.MaxDownloadSpeed > 0 && n.Info.DownloadSpeed > filter.MaxDownloadSpeed {
			continue
		}
		if len(includeCountries) > 0 {
			if _, ok := includeCountries[strings.ToUpper(n.Info.CountryCode)]; !ok {
				continue
			}
		}
		if _, ok := excludeCountries[strings.ToUpper(n.Info.CountryCode)]; ok {
			continue
		}
		result = append(result, n.Proxy)
	}
	return result
}
