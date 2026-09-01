# 使用方法

> ⚠️ **重要提示**  
> 本项目正在积极开发中。  
> 配置文件可能会频繁更改。  
> 请密切关注文档更新。  
> ⚠️ 注意脚本功能暂时请先不要投入过多精力，后期可能会频繁变动，导致自己编写的脚本失效

## 快速部署（推荐）

本仓库 `runtime/` 目录提供了开箱即用的示例环境：

```text
runtime/
├── BestSub          # 可执行文件（或通过 start.sh 自动编译）
├── start.sh         # 一键启动
├── base.yaml        # Mihomo 分流配置（可直接导入）
├── config/
│   ├── config.yaml
│   └── rename.yaml
└── output/          # 检测结果
```

```bash
cd runtime && ./start.sh
```

启动后访问 `http://127.0.0.1:18989` 等待检测完成。

## 直接运行

1. 根据自己系统选择 [release](https://github.com/bestruirui/BestSub/releases) 中的文件
2. 下载 [config.example.yaml](./config.example.yaml) 和 [rename.yaml](./rename.yaml) 到 `config/` 文件夹
3. 参考[配置文件说明](./config_zh.md) 修改后重命名为 `config.yaml`
4. 在可执行文件同级目录运行 `./BestSub`

目录结构：

```text
bestsub/
├── config/
│   ├── config.yaml
│   └── rename.yaml
├── BestSub
└── output/
```

## Docker

```bash
mkdir -p /path/to/config /path/to/output
```

```bash
docker run -itd \
    --name bestsub \
    -p 18989:18989 \
    -v /path/to/config:/app/config \
    -v /path/to/output:/app/output \
    --restart=always \
    ghcr.io/bestruirui/bestsub
```

## 源码编译运行

```bash
go build -o BestSub .
cd runtime   # 或你的运行目录
../BestSub   # 默认读取 ./config/config.yaml
```

指定配置文件：

```bash
go run main.go -f /path/to/config.yaml -r /path/to/rename.yaml
```

## 订阅源配置

在 `config.yaml` 的 `sub-urls` 中填入免费订阅链接，支持 clash、base64、v2ray 格式。

GitHub 搜索关键词即可找到公开源，示例：

```yaml
sub-urls:
  - https://raw.githubusercontent.com/peasoft/NoMoreWalls/refs/heads/master/list.yml
  - https://raw.githubusercontent.com/ripaojiedian/freenode/main/clash
  - https://raw.githubusercontent.com/mahdibland/V2RayAggregator/master/sub/sub_merge_yaml.yml
  - https://raw.githubusercontent.com/anaer/Sub/main/clash.yaml
```

更多来源与说明见 [getsubs.md](../getsubs.md)。

若 GitHub 无法直连，可在配置中启用代理：

```yaml
proxy:
  type: "http"
  address: "http://127.0.0.1:7890"
```

## 输出与 HTTP 订阅

配置 `save.method` 包含 `http` 和 `local` 时，检测结果同时写入本地并对外提供 HTTP 服务：

| 文件 | HTTP 地址 | 说明 |
|------|-----------|------|
| all.yaml | `http://127.0.0.1:18989/all.yaml` | 全部可用节点 |
| speed.yaml | `http://127.0.0.1:18989/speed.yaml` | 测速达标节点 |
| openai.yaml | `http://127.0.0.1:18989/openai.yaml` | 解锁 ChatGPT |
| youtube.yaml | `http://127.0.0.1:18989/youtube.yaml` | 解锁 YouTube |
| netflix.yaml | `http://127.0.0.1:18989/netflix.yaml` | 解锁 Netflix |
| disney.yaml | `http://127.0.0.1:18989/disney.yaml` | 解锁 Disney+ |

本地文件默认保存在可执行文件目录下的 `output/` 文件夹。

### 测速阈值

`check.min-speed` 控制进入 `speed.yaml` 的最低速度（KB/s），默认建议 512–2048，按需调整：

```yaml
check:
  min-speed: 512   # 值越低，高速节点越多，但质量可能下降
```

## 自建测速地址

> （可选）部分节点屏蔽常见测速地址，可自建测速服务

- 将 [worker](./cloudflare/worker.js) 部署到 Cloudflare Workers
- 将 `speed-test-url` 配置为 worker 地址

```yaml
speed-test-url:
  - https://your-worker-url/speedtest?bytes=1000000
```

## 保存方式

- 📁 **local**：保存到 `output/` 目录
- 🌐 **http**：启动 HTTP 服务提供订阅（端口由 `save.port` 控制，默认 8080，示例用 18989）
- ☁️ **r2**：Cloudflare R2 [配置方法](./r2_zh.md)
- 💾 **gist**：GitHub Gist [配置方法](./gist_zh.md)
- 🌐 **webdav**：WebDAV 服务器

## 导入 Mihomo / Clash

推荐使用 [base.yaml](./base.yaml) 作为分流配置，将 provider 链接指向 BestSub HTTP 服务：

```yaml
proxy-providers:
  ProviderALL:
    url: http://127.0.0.1:18989/all.yaml
    type: http
    interval: 600
    proxy: DIRECT
    path: ./proxy_provider/ALL.yaml
    health-check:
      enable: true
      url: http://www.gstatic.com/generate_204
      interval: 300
  ProviderOpenai:
    url: http://127.0.0.1:18989/openai.yaml
    type: http
    interval: 600
    proxy: DIRECT
    path: ./proxy_provider/Openai.yaml
    health-check:
      enable: true
      url: http://www.gstatic.com/generate_204
      interval: 300
```

若使用 `local` 保存方式，可改为 `type: file`：

```yaml
proxy-providers:
  ProviderALL:
    file: /path/to/output/all.yaml
    type: file
    interval: 600
```

Windows 裸核运行可参考 [minihomo](https://github.com/bestruirui/minihomo)。

## Mihomo 自动更新

BestSub 每次检测完成后，可通过 Mihomo API 自动刷新 proxy-provider，无需手动更新订阅。

**1. Mihomo 开启外部控制**

```yaml
external-controller: 127.0.0.1:9090
# secret: your-secret   # 可选
```

**2. BestSub 配置 API 地址**

```yaml
mihomo-api-url: "http://127.0.0.1:9090"
mihomo-api-secret: ""   # 与 Mihomo secret 一致，未设置则留空
```

检测完成后 BestSub 会自动调用 Mihomo API 更新 provider，配合 `proxy-providers` 的 `interval: 600` 实现双重保障。

> 注意：海外 VPS 检测出的节点，本地网络环境不一定可用，建议在自己实际使用的网络环境下运行 BestSub。
