package node

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/bestruirui/bestsub/internal/model"
	"github.com/bestruirui/bestsub/internal/store"

	"github.com/klauspost/compress/zstd"
)

const (
	poolSnapshotPath    = "data/bestsub.zst"
	poolSnapshotVersion = uint16(1)
)

type poolSnapshot struct {
	Version       uint16                     // 快照格式版本。
	Subscriptions []poolSnapshotSubscription // 按订阅分组的节点数据。
}

type poolSnapshotSubscription struct {
	ID    string             // 订阅 ID。
	Nodes []poolSnapshotNode // 该订阅保存的节点。
}

type poolSnapshotNode struct {
	Fingerprint       uint64  // 节点身份指纹。
	Text              string  // 单条 Mihomo YAML 节点内容。
	Delay             uint16  // 延迟，单位毫秒。
	DownloadSpeed     uint32  // 下载速度，单位 kb/s。
	CountryCode       string  // 落地国家代码。
	TrafficMultiplier float32 // 流量扣费倍率。
}

// PoolLoad 从退出快照恢复节点池；快照不存在时保持空池。
func PoolLoad() error {
	file, err := os.Open(poolSnapshotPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("open node pool snapshot: %w", err)
	}
	defer file.Close()

	reader, err := zstd.NewReader(file)
	if err != nil {
		return fmt.Errorf("open node pool snapshot decoder: %w", err)
	}
	defer reader.Close()

	var snapshot poolSnapshot
	if err := gob.NewDecoder(reader).Decode(&snapshot); err != nil {
		return fmt.Errorf("decode node pool snapshot: %w", err)
	}
	if _, err := io.Copy(io.Discard, reader); err != nil {
		return fmt.Errorf("verify node pool snapshot: %w", err)
	}
	if snapshot.Version != poolSnapshotVersion {
		return fmt.Errorf("unsupported node pool snapshot version: %d", snapshot.Version)
	}

	// 先完整构建新池，避免损坏快照把当前节点池替换成半成品。
	restored := make(map[string]map[uint64]*poolNode, len(snapshot.Subscriptions))
	for _, subscription := range snapshot.Subscriptions {
		if _, ok := store.SubscriptionGet(subscription.ID); !ok {
			continue
		}
		nodes := make(map[uint64]*poolNode, len(subscription.Nodes))
		for _, saved := range subscription.Nodes {
			nodes[saved.Fingerprint] = &poolNode{
				Raw: &Raw{Text: saved.Text, Fingerprint: saved.Fingerprint},
				Info: model.NodeInfo{
					Delay:             saved.Delay,
					DownloadSpeed:     saved.DownloadSpeed,
					CountryCode:       saved.CountryCode,
					TrafficMultiplier: saved.TrafficMultiplier,
				},
			}
		}
		if len(nodes) > 0 {
			restored[subscription.ID] = nodes
		}
	}
	poolMu.Lock()
	pool = restored
	poolMu.Unlock()
	return nil
}

// PoolSave 将当前节点池直接覆盖写入压缩快照。
func PoolSave() error {
	// 字符串只复制只读引用，额外内存主要是快照切片，不复制十几万份节点原文。
	poolMu.RLock()
	snapshot := poolSnapshot{Version: poolSnapshotVersion, Subscriptions: make([]poolSnapshotSubscription, 0, len(pool))}
	for subscriptionID, nodes := range pool {
		subscription := poolSnapshotSubscription{ID: subscriptionID, Nodes: make([]poolSnapshotNode, 0, len(nodes))}
		for _, current := range nodes {
			subscription.Nodes = append(subscription.Nodes, poolSnapshotNode{
				Fingerprint:       current.Raw.Fingerprint,
				Text:              current.Raw.Text,
				Delay:             current.Info.Delay,
				DownloadSpeed:     current.Info.DownloadSpeed,
				CountryCode:       current.Info.CountryCode,
				TrafficMultiplier: current.Info.TrafficMultiplier,
			})
		}
		snapshot.Subscriptions = append(snapshot.Subscriptions, subscription)
	}
	poolMu.RUnlock()

	file, err := os.OpenFile(poolSnapshotPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("open node pool snapshot: %w", err)
	}
	defer file.Close()

	writer, err := zstd.NewWriter(file, zstd.WithEncoderLevel(zstd.SpeedFastest), zstd.WithEncoderCRC(true))
	if err != nil {
		return fmt.Errorf("open node pool snapshot encoder: %w", err)
	}
	if err := gob.NewEncoder(writer).Encode(snapshot); err != nil {
		writer.Close()
		return fmt.Errorf("encode node pool snapshot: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("compress node pool snapshot: %w", err)
	}
	return nil
}
