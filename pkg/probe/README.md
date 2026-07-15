# pkg/probe

`pkg/probe` 用于通过调用方提供的 `*http.Client` 执行一次探测。

它不创建代理、不解析节点、不调度任务，也不保存结果。代理链路应由调用方提前配置到 `http.Client.Transport`。

检测模块只暴露 `Run`，检测参数由各检测文件内部解析。新增检测项时，新建文件并在 `init` 中注册即可，不需要修改 `probe.go`。

## 延迟探测

```go
var delay uint16
err := probe.Run(ctx, probe.TypeDelay, client, json.RawMessage(`{
	"url": "https://www.gstatic.com/generate_204",
	"attempts": 3
}`), &delay)
if err != nil {
	return err
}
```

## 测速

```go
var speed uint32
err := probe.Run(ctx, probe.TypeSpeed, client, json.RawMessage(`{
	"url": "https://example.com/100mb.bin",
	"max_kb": 32768,
	"max_duration_ms": 10000
}`), &speed)
if err != nil {
	return err
}
```

`max_kb` 的单位为 kb，`speed` 的单位为 kb/s。

## 国家代码

```go
var countryCode string
err := probe.Run(ctx, probe.TypeCountry, client, json.RawMessage(`{
	"timeout_ms": 10000
}`), &countryCode)
if err != nil {
	return err
}
```

`ctx` 会进入 `http.Request`，请求取消和超时会传递到 `http.Transport.DialContext`。
