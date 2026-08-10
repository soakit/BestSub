<p align="center">
  <img src="web/public/favicon.svg" width="96" alt="BESTSUB">
</p>

<h1 align="center">BESTSUB</h1>

<p align="center">Best Sub, Best for Your Net</p>

## 界面截图

<p align="center">
  <img src="web/public/screenshot/dashboard.png" width="32%" alt="仪表盘">
  <img src="web/public/screenshot/sub.png" width="32%" alt="订阅">
  <img src="web/public/screenshot/node.png" width="32%" alt="节点">
  <img src="web/public/screenshot/task.png" width="32%" alt="任务">
  <img src="web/public/screenshot/share.png" width="32%" alt="分享">
  <img src="web/public/screenshot/setting.png" width="32%" alt="设置">
</p>

## 使用方式

BESTSUB 必须依赖 [MiniSubConvert](https://github.com/bestruirui/MiniSubConvert) 或 [Sub-Store](https://github.com/sub-store-org/Sub-Store) 完成订阅转换。使用前请先部署其中任意一个，然后登录 BESTSUB WebUI，在“设置”中配置订阅转换地址。

### 直接运行


从 [Releases](https://github.com/bestruirui/bestsub/releases) 下载对应平台的压缩包，解压后运行，默认用户名/密码：`admin` / `admin`。

```bash
./bestsub start
```

Windows 使用：

```powershell
.\bestsub.exe start
```

访问 <http://127.0.0.1:8080>。

### Docker

```bash
docker run -d \
  --name bestsub \
  --restart unless-stopped \
  -p 8080:8080 \
  -v "$PWD/data:/app/data" \
  ghcr.io/bestruirui/bestsub:latest
```

或者

```yaml
services:
  bestsub:
    image: ghcr.io/bestruirui/bestsub:latest
    container_name: bestsub
    restart: unless-stopped
    network_mode: bridge
    ports:
      - "8080:8080"
    volumes:
      - ./data:/app/data
```

```bash
docker compose up -d
```

访问 `http://服务器地址:8080`。

