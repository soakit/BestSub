# BESTSUB

<div align="center">
  <img src="https://img.shields.io/github/v/release/bestruirui/BestSub?color=blue" alt="Version">
  <img src="https://img.shields.io/badge/Language-Go-green" alt="Language">
  <a href="./README_zh.md">
    <img src="https://img.shields.io/badge/中文文档-brightgreen" alt="中文文档">
  </a>
  <img src="https://img.shields.io/badge/License-MIT-orange" alt="License">
</div>

<div align="center">
  <h2>🚀 Stay Tuned for the New BESTSUB 🚀</h2>
  <h3 style="margin-top: 10px;">
    <a href="https://github.com/bestruirui/BestSub/tree/api">
      <img src="https://img.shields.io/badge/BESTSUB-Details-brightgreen?style=for-the-badge&logo=github" alt="BESTSUB">
    </a>
  </h3>
  <hr style="width: 50%; margin: 20px auto;">
</div>

## Preview

![preview](./doc/images/preview.png)

## Features

- ✅ Detect node availability and remove unavailable nodes
- ✅ Custom platform unlocking detection (openai, youtube, netflix, disney)
- ✅ Merge multiple subscriptions
- ✅ Convert subscriptions to clash/mihomo format
- ✅ Remove duplicate nodes and rename nodes
- ✅ Test node speed and classify results
- ✅ Auto-update Mihomo providers after each check

## Quick Start

```text
bestsub/
├── config/
│   ├── config.yaml
│   └── rename.yaml
├── BestSub
└── output/
```

```bash
# From source
go build -o BestSub .

# Or use the example runtime directory
cd runtime && ./start.sh
```

Configure `sub-urls` in `config/config.yaml`, then access results at:

- `http://127.0.0.1:18989/all.yaml` — all alive nodes
- `http://127.0.0.1:18989/speed.yaml` — speed-qualified nodes
- `http://127.0.0.1:18989/openai.yaml` — ChatGPT-unlocked nodes

See [getsubs.md](./getsubs.md) (Chinese) for a full walkthrough.

## Mihomo Auto-Update

```yaml
# BestSub config
mihomo-api-url: "http://127.0.0.1:9090"
mihomo-api-secret: ""
```

```yaml
# Mihomo config
external-controller: 127.0.0.1:9090
```

Point `proxy-providers` in [base.yaml](./doc/base.yaml) to your local BestSub HTTP endpoints.

## Documentation

- [Usage Guide](./doc/README.md)
- [Config Reference](./doc/config.md)
- [中文文档](./README_zh.md)
