# BestSub Usage Guide

> ⚠️ **Important Notice**  
> This project is under active development. Configuration files may change frequently.

[中文文档](./README_zh.md) | English Documentation

## Quick Deploy

The `runtime/` directory provides a ready-to-use example:

```bash
cd runtime && ./start.sh
```

Results are served at `http://127.0.0.1:18989/` after each check completes.

## Direct Execution

1. Download the binary from [releases](https://github.com/bestruirui/BestSub/releases)
2. Copy [config.example.yaml](./config.example.yaml) and [rename.yaml](./rename.yaml) into `config/`
3. Edit `config.yaml` per the [Configuration Documentation](./config.md)
4. Run `./BestSub`

## Docker

```bash
docker run -itd \
    --name bestsub \
    -p 18989:18989 \
    -v /path/to/config:/app/config \
    -v /path/to/output:/app/output \
    --restart=always \
    ghcr.io/bestruirui/bestsub
```

## Build from Source

```bash
go build -o BestSub .
go run main.go -f /path/to/config.yaml -r /path/to/rename.yaml
```

## HTTP Subscription Endpoints

| File | URL |
|------|-----|
| All nodes | `http://127.0.0.1:18989/all.yaml` |
| Speed | `http://127.0.0.1:18989/speed.yaml` |
| ChatGPT | `http://127.0.0.1:18989/openai.yaml` |
| YouTube | `http://127.0.0.1:18989/youtube.yaml` |
| Netflix | `http://127.0.0.1:18989/netflix.yaml` |
| Disney+ | `http://127.0.0.1:18989/disney.yaml` |

Adjust `check.min-speed` (KB/s) to control how many nodes qualify for `speed.yaml`.

## Mihomo Integration

Use [base.yaml](./base.yaml) and point providers to local BestSub URLs:

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

### Auto-Update

```yaml
# BestSub config
mihomo-api-url: "http://127.0.0.1:9090"
mihomo-api-secret: ""
```

```yaml
# Mihomo config
external-controller: 127.0.0.1:9090
```

BestSub will refresh Mihomo providers automatically after each detection run.

## Save Methods

- **local** — `output/` directory
- **http** — HTTP server (port via `save.port`)
- **r2** — [Cloudflare R2](./r2.md)
- **gist** — [GitHub Gist](./gist.md)
- **webdav** — WebDAV server

## Custom Speed Test URL

Deploy [worker.js](./cloudflare/worker.js) to Cloudflare Workers if nodes block default test URLs.

```yaml
speed-test-url:
  - https://your-worker-url/speedtest?bytes=1000000
```
