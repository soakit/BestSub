#!/usr/bin/env bash
# 一键启动 BestSub + MiniSubConvert
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
RUNTIME="$ROOT/runtime"
DATA="$RUNTIME/data"
BIN="$RUNTIME/bestsub"
MSC_DIR="$RUNTIME/minisubconvert"
MSC_SRC="$MSC_DIR/dist/minisubconvert.js"
BASE_URL="${BESTSUB_URL:-http://127.0.0.1:8080}"
USER="${BESTSUB_USER:-admin}"
PASS="${BESTSUB_PASS:-admin}"
SECRET="${MSC_SECRET:-minisubconvert}"
MSC_PORT="${MSC_PORT:-3001}"
COOKIE_JAR="$RUNTIME/.cookies"

BESTSUB_PID="$RUNTIME/bestsub.pid"
BESTSUB_LOG="$RUNTIME/bestsub.log"
MSC_PID="$RUNTIME/minisubconvert.pid"
MSC_LOG="$RUNTIME/minisubconvert.log"

find_node() {
  for candidate in \
    "${NODE_BIN:-}" \
    "$HOME/.nvm/versions/node/v20.20.2/bin/node" \
    "$HOME/.nvm/versions/node/v20.18.1/bin/node" \
    "$(command -v node 2>/dev/null || true)"; do
    [[ -n "$candidate" && -x "$candidate" ]] || continue
    if "$candidate" -e 'process.exit(Number(process.versions.node.split(".")[0]) >= 18 ? 0 : 1)' 2>/dev/null; then
      echo "$candidate"
      return 0
    fi
  done
  return 1
}

is_running() {
  local pid_file="$1"
  [[ -f "$pid_file" ]] && kill -0 "$(cat "$pid_file")" 2>/dev/null
}

port_free() {
  ! lsof -iTCP:"$1" -sTCP:LISTEN >/dev/null 2>&1
}

pick_msc_port() {
  local port="$MSC_PORT"
  while ! port_free "$port"; do
    port=$((port + 1))
  done
  echo "$port"
}

ensure_minisubconvert() {
  if [[ -f "$MSC_SRC" ]]; then
    return 0
  fi
  # 复用之前已构建好的副本
  if [[ -f /tmp/MiniSubConvert/dist/minisubconvert.js ]]; then
    echo "[1/4] 复用已构建的 MiniSubConvert..."
    mkdir -p "$MSC_DIR/dist"
    cp /tmp/MiniSubConvert/dist/minisubconvert.js "$MSC_SRC"
    return 0
  fi

  echo "[1/4] 首次运行，正在准备 MiniSubConvert..."
  local node_bin node_dir
  node_bin="$(find_node)" || { echo "需要 Node.js 18+，请安装后重试" >&2; exit 1; }
  node_dir="$(dirname "$node_bin")"
  export PATH="$node_dir:$PATH"

  mkdir -p "$MSC_DIR"
  if [[ ! -d "$MSC_DIR/.git" ]]; then
    git clone --depth 1 https://github.com/bestruirui/MiniSubConvert.git "$MSC_DIR"
  fi
  (
    cd "$MSC_DIR"
    "$node_bin" --version
    if [[ ! -d node_modules ]]; then
      npm install --silent
    fi
    npm run build --silent
  )
  [[ -f "$MSC_SRC" ]] || { echo "MiniSubConvert 构建失败" >&2; exit 1; }
}

start_minisubconvert() {
  if is_running "$MSC_PID"; then
    local port
    port="$(lsof -Pan -p "$(cat "$MSC_PID")" -iTCP -sTCP:LISTEN 2>/dev/null | awk '{print $9}' | head -1 | sed 's/.*://')"
    MSC_PORT="${port:-$MSC_PORT}"
    echo "[1/4] MiniSubConvert 已在运行 (PID $(cat "$MSC_PID"), 端口 $MSC_PORT)"
    return 0
  fi

  ensure_minisubconvert
  local node_bin port
  node_bin="$(find_node)"
  port="$(pick_msc_port)"
  MSC_PORT="$port"

  echo "[1/4] 启动 MiniSubConvert (端口 $MSC_PORT)..."
  cd "$MSC_DIR"
  nohup env SECRET="$SECRET" PORT="$MSC_PORT" HOST="0.0.0.0" \
    "$node_bin" dist/minisubconvert.js >>"$MSC_LOG" 2>&1 &
  echo $! >"$MSC_PID"
  sleep 1

  for _ in $(seq 1 15); do
    if curl -sS --connect-timeout 1 -X POST \
      "http://127.0.0.1:$MSC_PORT/$SECRET/api/proxy/parse" \
      -H 'Content-Type: application/json' \
      -d '{"client":"Mihomo","data":"proxies:\n  - {name: t, type: ss, server: 1.1.1.1, port: 1, cipher: aes-256-gcm, password: t}"}' \
      | grep -q '"status":"success"'; then
      echo "      MiniSubConvert 就绪"
      return 0
    fi
    sleep 1
  done
  echo "MiniSubConvert 启动失败，查看日志: $MSC_LOG" >&2
  exit 1
}

start_bestsub() {
  mkdir -p "$DATA"
  if [[ ! -f "$DATA/config.json" ]]; then
    cat >"$DATA/config.json" <<'EOF'
{
  "server": { "host": "0.0.0.0", "port": 8080 },
  "logging": { "level": "info" },
  "database": { "type": "sqlite", "path": "data/data.db" }
}
EOF
  fi

  if is_running "$BESTSUB_PID"; then
    echo "[2/4] BestSub 已在运行 (PID $(cat "$BESTSUB_PID"))"
    return 0
  fi

  [[ -x "$BIN" ]] || { echo "未找到 $BIN，请先下载或编译 bestsub" >&2; exit 1; }
  echo "[2/4] 启动 BestSub..."
  cd "$RUNTIME"
  nohup "$BIN" start >>"$BESTSUB_LOG" 2>&1 &
  echo $! >"$BESTSUB_PID"

  for _ in $(seq 1 30); do
    if curl -sS --connect-timeout 1 "$BASE_URL/" -o /dev/null 2>/dev/null; then
      echo "      BestSub 就绪"
      return 0
    fi
    sleep 1
  done
  echo "BestSub 启动失败，查看日志: $BESTSUB_LOG" >&2
  exit 1
}

configure_bestsub() {
  echo "[3/4] 写入订阅转换地址..."
  local convert_url="http://127.0.0.1:$MSC_PORT/$SECRET/api/proxy/parse"

  curl -sS "$BASE_URL/api/v1/user/login" -X POST \
    -H 'Content-Type: application/json' \
    -d "{\"username\":\"$USER\",\"password\":\"$PASS\",\"trust\":true}" \
    -c "$COOKIE_JAR" >/dev/null

  curl -sS "$BASE_URL/api/v1/setting/update" -X PUT \
    -H 'Content-Type: application/json' \
    -b "$COOKIE_JAR" \
    -d "{\"key\":\"sub_convert_url\",\"value\":\"$convert_url\"}" >/dev/null

  # 订阅拉取：先直连，失败再走代理（仅影响获取订阅链接，不影响节点检测）
  local proxy_url=""
  for candidate in http://127.0.0.1:7890 http://127.0.0.1:7897 http://127.0.0.1:10808; do
    if curl -sS --connect-timeout 2 -x "$candidate" https://www.google.com/generate_204 -o /dev/null 2>/dev/null; then
      proxy_url="$candidate"
      break
    fi
  done
  if [[ -n "$proxy_url" ]]; then
    curl -sS "$BASE_URL/api/v1/setting/update" -X PUT \
      -H 'Content-Type: application/json' -b "$COOKIE_JAR" \
      -d "{\"key\":\"proxy_url\",\"value\":\"$proxy_url\"}" >/dev/null
    curl -sS "$BASE_URL/api/v1/setting/update" -X PUT \
      -H 'Content-Type: application/json' -b "$COOKIE_JAR" \
      -d '{"key":"proxy_enable","value":"1"}' >/dev/null
    echo "      订阅拉取: 直连优先，失败走代理 ($proxy_url)"
  else
    curl -sS "$BASE_URL/api/v1/setting/update" -X PUT \
      -H 'Content-Type: application/json' -b "$COOKIE_JAR" \
      -d '{"key":"proxy_enable","value":"0"}' >/dev/null
    echo "      订阅拉取: 仅直连（未检测到本地代理）"
  fi
  # 订阅转换走本地服务，不走代理
  curl -sS "$BASE_URL/api/v1/setting/update" -X PUT \
    -H 'Content-Type: application/json' -b "$COOKIE_JAR" \
    -d '{"key":"sub_convert_proxy","value":"0"}' >/dev/null

  echo "      转换地址: $convert_url"
  echo "      节点检测: 本地直连（经节点本身探测，不受全局代理影响）"
}

print_status() {
  echo "[4/4] 完成"
  echo ""
  echo "========================================"
  echo " BestSub 已启动"
  echo "========================================"
  echo "WebUI:  $BASE_URL"
  echo "账号:   $USER / $PASS"
  echo ""
  echo "常用命令:"
  echo "  完整配置(订阅+检测+分享): bash scripts/auto-setup.sh"
  echo "  停止服务:                 bash scripts/stop.sh"
  echo "  查看日志:                 tail -f runtime/bestsub.log"
  echo "========================================"
}

main() {
  start_minisubconvert
  start_bestsub
  configure_bestsub
  print_status

  if [[ "${1:-}" == "--setup" ]]; then
    echo ""
    exec bash "$ROOT/scripts/auto-setup.sh"
  fi
}

main "$@"
