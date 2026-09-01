#!/usr/bin/env bash
# BestSub 全自动配置：添加免费订阅源 → 节点检测 → 生成分享链接
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
RUNTIME="$ROOT/runtime"
DATA="$RUNTIME/data"
BIN="$RUNTIME/bestsub"
BASE_URL="${BESTSUB_URL:-http://127.0.0.1:8080}"
USER="${BESTSUB_USER:-admin}"
PASS="${BESTSUB_PASS:-admin}"
PORT="${BESTSUB_PORT:-8080}"
COOKIE_JAR="$RUNTIME/.cookies"
LOG="$RUNTIME/bestsub.log"
PID_FILE="$RUNTIME/bestsub.pid"

# 来自 getsubs.md 的公开订阅源
SUB_URLS=(
  "https://raw.githubusercontent.com/snakem982/proxypool/main/source/clash-meta.yaml"
  "https://raw.githubusercontent.com/ripaojiedian/freenode/main/clash"
  "https://raw.githubusercontent.com/vxiaov/free_proxies/main/clash/clash.provider.yaml"
  "https://raw.githubusercontent.com/chengaopan/AutoMergePublicNodes/master/list.yml"
  "https://raw.githubusercontent.com/Ruk1ng001/freeSub/main/clash_top30.yaml"
  "https://raw.githubusercontent.com/peasoft/NoMoreWalls/refs/heads/master/list.yml"
)

api() {
  local method="$1" path="$2"
  shift 2
  curl -sS -X "$method" "$BASE_URL$path" \
    -b "$COOKIE_JAR" -c "$COOKIE_JAR" \
    -H "Content-Type: application/json" \
    "$@"
}

json_get() {
  python3 -c "import json,sys; d=json.load(sys.stdin); print(d.get('data', d).$1)" 2>/dev/null || true
}

wait_ready() {
  for i in $(seq 1 30); do
    if curl -sS "$BASE_URL/api/v1/user/login" -o /dev/null -w '%{http_code}' -X POST \
      -H 'Content-Type: application/json' \
      -d '{"username":"x","password":"x"}' | grep -qE '401|400|200'; then
      return 0
    fi
    sleep 1
  done
  echo "BestSub 启动超时" >&2
  return 1
}

start_server() {
  mkdir -p "$DATA"
  if [[ -f "$PID_FILE" ]] && kill -0 "$(cat "$PID_FILE")" 2>/dev/null; then
    echo "BestSub 已在运行 (PID $(cat "$PID_FILE"))"
    return 0
  fi
  echo "启动 BestSub..."
  cd "$RUNTIME"
  # 首次运行自动生成 data/config.json
  if [[ ! -f "$DATA/config.json" ]]; then
    mkdir -p "$DATA"
    cat >"$DATA/config.json" <<'EOF'
{
  "server": { "host": "0.0.0.0", "port": 8080 },
  "logging": { "level": "info" },
  "database": { "type": "sqlite", "path": "data/data.db" }
}
EOF
  fi
  nohup "$BIN" start >"$LOG" 2>&1 &
  echo $! >"$PID_FILE"
  wait_ready
}

login() {
  rm -f "$COOKIE_JAR"
  api POST /api/v1/user/login -d "{\"username\":\"$USER\",\"password\":\"$PASS\",\"trust\":true}" \
    | python3 -c "import json,sys; r=json.load(sys.stdin); assert r['code']==200, r.get('message','login failed')"
  echo "登录成功"
}

configure_settings() {
  echo "配置全局设置..."
  local convert_url="${SUB_CONVERT_URL:-http://127.0.0.1:${MSC_PORT:-3001}/${MSC_SECRET:-minisubconvert}/api/proxy/parse}"
  api PUT /api/v1/setting/update -d "{\"key\":\"sub_convert_url\",\"value\":\"$convert_url\"}" >/dev/null
  echo "  订阅转换: $convert_url"
  api PUT /api/v1/setting/update -d '{"key":"sub_convert_proxy","value":"0"}' >/dev/null
  # 订阅拉取：先直连，失败再走代理（proxy_mode=自动 时生效；节点检测不受影响）
  local proxy_url=""
  for candidate in http://127.0.0.1:7890 http://127.0.0.1:7897 http://127.0.0.1:10808; do
    if curl -sS --connect-timeout 2 -x "$candidate" https://www.google.com/generate_204 -o /dev/null 2>/dev/null; then
      proxy_url="$candidate"
      break
    fi
  done
  if [[ -n "$proxy_url" ]]; then
    api PUT /api/v1/setting/update -d "{\"key\":\"proxy_url\",\"value\":\"$proxy_url\"}" >/dev/null
    api PUT /api/v1/setting/update -d '{"key":"proxy_enable","value":"1"}' >/dev/null
    echo "  订阅拉取: 直连优先，失败走代理 ($proxy_url)"
  else
    api PUT /api/v1/setting/update -d '{"key":"proxy_enable","value":"0"}' >/dev/null
    echo "  订阅拉取: 仅直连（未检测到本地代理）"
  fi
  echo "  节点检测: 本地直连（经节点本身探测）"
  api PUT /api/v1/setting/update -d '{"key":"max_concurrent","value":"100"}' >/dev/null
  api PUT /api/v1/setting/update -d '{"key":"health_check_url","value":"https://gstatic.com/generate_204"}' >/dev/null
  api PUT /api/v1/setting/update -d '{"key":"speed_test_url","value":"https://speed.cloudflare.com/__down?during=download&bytes=999999"}' >/dev/null
}

add_subscriptions() {
  echo "添加订阅源..."
  local existing
  existing=$(api GET /api/v1/sub/list)
  for url in "${SUB_URLS[@]}"; do
    if echo "$existing" | python3 -c "import json,sys; subs=json.load(sys.stdin)['data']; sys.exit(0 if any('$url' in s.get('url',[]) for s in subs) else 1)" 2>/dev/null; then
      echo "  跳过已存在: $url"
      continue
    fi
    name="$(basename "$url" .yaml | sed 's/[_.]/-/g')"
    api POST /api/v1/sub/create -d "{
      \"name\": \"$name\",
      \"url\": [\"$url\"],
      \"url_type\": 0,
      \"enable\": 1,
      \"auto_update\": 1,
      \"cron_expr\": \"0 */6 * * *\",
      \"proxy_mode\": 0
    }" >/dev/null
    echo "  已添加: $name"
  done
}

refresh_subscriptions() {
  echo "刷新所有订阅（可能需要几分钟）..."
  local ids
  ids=$(api GET /api/v1/sub/list | python3 -c "import json,sys; print(' '.join(s['id'] for s in json.load(sys.stdin)['data']))")
  for id in $ids; do
    api POST "/api/v1/sub/refresh/$id" --max-time 120 >/dev/null || true
  done
  local count
  count=$(api GET /api/v1/sub/list | python3 -c "import json,sys; print(sum(s.get('node_num',0) for s in json.load(sys.stdin)['data']))")
  echo "  当前节点池共 $count 个节点"
}

create_task() {
  echo "创建/更新检测任务..."
  local tasks task_id
  tasks=$(api GET /api/v1/task/list)
  task_id=$(echo "$tasks" | python3 -c "
import json,sys
tasks=json.load(sys.stdin)['data']
for t in tasks:
    if t.get('name')=='高质量节点筛选':
        print(t['id']); break
" 2>/dev/null || true)

  local payload='{
    "name": "高质量节点筛选",
    "auto_run": 1,
    "cron_expr": "0 */2 * * *",
    "all_input_enable": 1,
    "subscriptions": [],
    "nodes": [],
    "tags": [],
    "result_tasks": [],
    "custom_landing_node_enable": 0,
    "landing_node": {"id": ""},
    "storage_enable": 0,
    "steps": [
      {
        "type": "delay",
        "params": {"url": "https://gstatic.com/generate_204", "timeout_ms": 2000, "attempts": 1},
        "concurrency": 100,
        "pass": {"max_delay": 2000},
        "order": 1,
        "node_pool_delete": 0,
        "skip_existing": 1
      },
      {
        "type": "speed",
        "params": {"url": "https://speed.cloudflare.com/__down?during=download&bytes=999999", "timeout_ms": 10000},
        "concurrency": 10,
        "pass": {"min_download_speed": 2048},
        "order": 2,
        "node_pool_delete": 0,
        "skip_existing": 0
      }
    ]
  }'

  if [[ -n "$task_id" ]]; then
    api PUT "/api/v1/task/update/$task_id" -d "$payload" >/dev/null
    echo "  已更新任务: $task_id"
  else
    task_id=$(api POST /api/v1/task/create -d "$payload" | python3 -c "import json,sys; print(json.load(sys.stdin)['data']['id'])")
    echo "  已创建任务: $task_id"
  fi
  echo "$task_id"
}

run_task() {
  local task_id="$1"
  echo "运行节点检测任务（约 3-10 分钟，取决于节点数量）..."
  api POST "/api/v1/task/run/$task_id" >/dev/null || true
  for i in $(seq 1 120); do
    sleep 5
    local result_count
    result_count=$(api GET "/api/v1/task/result/$task_id" 2>/dev/null | python3 -c "
import json,sys
try:
    print(json.load(sys.stdin)['data'])
except: print(0)
" 2>/dev/null || echo "0")
    if [[ "$result_count" =~ ^[0-9]+$ && "$result_count" -gt 0 ]]; then
      echo "  检测完成，优质节点: $result_count 个"
      return 0
    fi
    if (( i % 6 == 0 )); then
      echo "  仍在检测中... (${i}x5s)"
    fi
  done
  echo "  检测超时，可在 WebUI 查看进度: $BASE_URL"
}

create_share() {
  local task_id="$1"
  echo "创建分享链接..."
  local shares share_id token
  shares=$(api GET /api/v1/share/list)
  share_id=$(echo "$shares" | python3 -c "
import json,sys
for s in json.load(sys.stdin)['data']:
    if s.get('name')=='优质节点':
        print(s['id']); break
" 2>/dev/null || true)

  local payload="{
    \"name\": \"优质节点\",
    \"filter\": {\"max_delay\": 2000, \"min_download_speed\": 2048, \"limit\": 100},
    \"node_rename_expression\": \"\",
    \"result_tasks\": [{\"id\": \"$task_id\"}],
    \"subscriptions\": [],
    \"nodes\": [],
    \"tags\": []
  }"

  if [[ -n "$share_id" ]]; then
    api PUT "/api/v1/share/update/$share_id" -d "$payload" >/dev/null
    token=$(echo "$shares" | python3 -c "
import json,sys
for s in json.load(sys.stdin)['data']:
    if s.get('name')=='优质节点':
        print(s['token']); break
")
  else
    token=$(api POST /api/v1/share/create -d "$payload" | python3 -c "import json,sys; print(json.load(sys.stdin)['data']['token'])")
  fi
  echo "$token"
}

main() {
  echo "========================================"
  echo " BestSub 全自动获取高质量订阅节点"
  echo "========================================"
  # 确保服务已启动（含 MiniSubConvert + sub_convert_url）
  if ! curl -sS --connect-timeout 2 "$BASE_URL/" -o /dev/null 2>/dev/null; then
    bash "$(dirname "$0")/start.sh"
  fi
  login
  configure_settings
  add_subscriptions
  refresh_subscriptions
  task_id=$(create_task)
  run_task "$task_id"
  token=$(create_share "$task_id")

  echo ""
  echo "========================================"
  echo " 配置完成！"
  echo "========================================"
  echo "WebUI:    $BASE_URL  (账号: $USER / $PASS)"
  echo "优质节点: $BASE_URL/share/$token"
  echo ""
  echo "导入 Clash/Mihomo 客户端："
  echo "  订阅链接 → $BASE_URL/share/$token"
  echo ""
  echo "日志: tail -f $LOG"
  echo "停止: kill \$(cat $PID_FILE)"
}

main "$@"
