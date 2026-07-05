# pkg/probe

`pkg/probe` 用于通过调用方提供的 `*http.Client` 执行一次探测。

它不创建代理、不解析节点、不调度任务，也不保存结果。代理链路应由调用方提前配置到 `http.Client.Transport`。

## 延迟探测

```go
patch, err := probe.Run(ctx, probe.Request{
	Client: client,
	Probe: probe.Delay{
		HTTPParams: probe.HTTPParams{
			URL: "https://www.gstatic.com/generate_204",
		},
		Attempts: 3,
	},
})
if err != nil {
	return err
}

if patch.Delay != nil {
	node.Delay = *patch.Delay
}
```

## 下载测速

```go
patch, err := probe.Run(ctx, probe.Request{
	Client: client,
	Probe: probe.Download{
		HTTPParams: probe.HTTPParams{
			URL: "https://example.com/100mb.bin",
		},
		MaxBytes:      32 << 20,
		MaxDurationMS: 10000,
	},
})
if err != nil {
	return err
}

if patch.DownloadSpeed != nil {
	node.DownloadSpeed = *patch.DownloadSpeed
}
```

## 国家代码

```go
patch, err := probe.Run(ctx, probe.Request{
	Client: client,
	Probe: probe.Country{
		HTTPParams: probe.HTTPParams{
			URL: "https://example.com/ip",
		},
		CountryField: "country_code",
	},
})
if err != nil {
	return err
}

if patch.CountryCode != nil {
	node.CountryCode = *patch.CountryCode
}
```

`ctx` 会进入 `http.Request`，请求取消和超时会传递到 `http.Transport.DialContext`。
