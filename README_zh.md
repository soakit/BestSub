# 订阅合并转换检测工具

<div align="center">
  <img src="https://img.shields.io/github/v/release/bestruirui/BestSub?color=blue" alt="版本">
  <img src="https://img.shields.io/badge/语言-Go-green" alt="语言">
  <a href="./README.md">
    <img src="https://img.shields.io/badge/English_Document-brightgreen" alt="英文文档">
  </a>
  <img src="https://img.shields.io/badge/许可证-MIT-orange" alt="许可证">
</div>
<div align="center">
  <h2>🚀 敬请期待全新BESTSUB 🚀</h2>
  <h3 style="margin-top: 10px;">
    <a href="https://github.com/bestruirui/BestSub/tree/api">
      <img src="https://img.shields.io/badge/BESTSUB-查看详情-brightgreen?style=for-the-badge&logo=github" alt="BESTSUB">
    </a>
  </h3>
  <hr style="width: 50%; margin: 20px auto;">
</div>

## 预览

![preview](./doc/images/preview.png)

## 功能

- ✅ 检测节点可用性，去除不可用节点
- ✅ 自定义检测平台解锁情况
    - openai
    - youtube
    - netflix
    - disney
- ✅ 根据解锁情况分类保存
- ✅ 合并多个订阅
- ✅ 将订阅转换为 clash/mihomo 格式
- ✅ 节点去重
- ✅ 节点重命名
    - API 命名
    - 自定义规则命名
- ✅ 节点测速
- ✅ 检测完成后自动更新 Mihomo provider

## 特点

- 🚀 支持多平台
- ⚡ 支持多线程
- 🍃 资源占用低

## 快速开始

### 1. 准备运行目录

```text
bestsub/
├── config/
│   ├── config.yaml    # 主配置
│   └── rename.yaml    # 重命名规则
├── BestSub            # 可执行文件
└── output/            # 检测结果（自动生成）
```

从 [releases](https://github.com/bestruirui/BestSub/releases) 下载对应平台二进制，或将 [config.example.yaml](./doc/config.example.yaml) 和 [rename.yaml](./doc/rename.yaml) 放入 `config/` 后自行编译：

```bash
go build -o BestSub .
```

本仓库已提供示例运行目录 `runtime/`，可直接使用：

```bash
cd runtime && ./start.sh
```

### 2. 修改配置

编辑 `config/config.yaml`，填入免费订阅源并选择保存方式：

```yaml
save:
  method:
    - http    # 提供 HTTP 订阅
    - local   # 保存到 output/
  port: 18989

check:
  interval: 30          # 检测间隔（分钟）
  min-speed: 512        # 高速节点最低测速（KB/s），可按需调整
  items:
    - speed
    - openai
    - youtube
    - netflix
    - disney

sub-urls:
  - https://raw.githubusercontent.com/peasoft/NoMoreWalls/refs/heads/master/list.yml
  # 更多订阅源见 getsubs.md

proxy:
  type: "http"          # 拉取订阅时使用本地代理（可选）
  address: "http://127.0.0.1:7890"
```

完整教程见 [getsubs.md](./getsubs.md)。

### 3. 获取节点

启动后等待检测完成（通常 3–10 分钟），通过以下方式获取节点：

| 方式 | 地址 |
|------|------|
| HTTP 订阅 | `http://127.0.0.1:18989/all.yaml` |
| 高速节点 | `http://127.0.0.1:18989/speed.yaml` |
| ChatGPT | `http://127.0.0.1:18989/openai.yaml` |
| YouTube | `http://127.0.0.1:18989/youtube.yaml` |
| Netflix | `http://127.0.0.1:18989/netflix.yaml` |
| Disney+ | `http://127.0.0.1:18989/disney.yaml` |
| 本地文件 | `output/all.yaml` 等 |

### 4. 导入 Mihomo / Clash

下载 [base.yaml](./doc/base.yaml)，将 `proxy-providers` 中的链接改为本地 BestSub 地址：

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
```

`runtime/base.yaml` 提供了可直接导入的简化配置（含分流规则）。

### 5. Mihomo 自动更新

在 BestSub 配置中填入 Mihomo 外部控制地址，检测完成后会自动刷新 provider：

```yaml
mihomo-api-url: "http://127.0.0.1:9090"
mihomo-api-secret: ""   # 若 Mihomo 配置了 secret 则填入
```

Mihomo 侧需开启 external-controller（默认 `9090`）。

## TODO

- [x] 适配多种订阅格式
- [ ] 支持更多的保存方式
    - [x] 本地
    - [x] cloudflare r2
    - [x] gist
    - [x] webdav
    - [x] http
    - [ ] 其他
- [ ] 储存优选节点

## 文档

- [使用方法](./doc/README_zh.md)
- [配置文件详解](./doc/config_zh.md)
- [全自动获取节点教程](./getsubs.md)
