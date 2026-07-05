# pkg/mihomo

`pkg/mihomo` 用于把单条 mihomo 节点 YAML 转成 `http.Transport`。

它只负责代理传输层构建，不负责节点管理、任务调度、探测结果保存。

## 单节点代理

```go
transport, err := mihomo.NewTransport(raw, iface)
if err != nil {
	return err
}
defer transport.CloseIdleConnections()

client := &http.Client{Transport: transport}
```

`raw` 是单条节点的 YAML 内容。`iface` 为空时不绑定网卡。

## 链式代理

```go
transport, err := mihomo.NewChainedTransport(outer, inner, iface)
if err != nil {
	return err
}
defer transport.CloseIdleConnections()

client := &http.Client{Transport: transport}
```

链路顺序：

```text
本地 -> outer -> inner -> 目标网站
```

`outer` 是前置代理，`inner` 是后置代理。

## DNS 配置

```go
mihomo.UpdateDNSConfig(defaultServers, mainServers)
```

`defaultServers` 使用 UDP DNS，`mainServers` 使用 DoH。
