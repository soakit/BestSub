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
  "https://raw.githubusercontent.com/ripaojiedian/freenode/main/clash"
  "https://raw.githubusercontent.com/vxiaov/free_proxies/main/clash/clash.provider.yaml"
  "https://raw.githubusercontent.com/chengaopan/AutoMergePublicNodes/master/list.yml"
  "https://raw.githubusercontent.com/peasoft/NoMoreWalls/refs/heads/master/list.yml"
)

# 具名订阅（name|url，用于 gist 等需自定义名称的源）
SUB_NAMED_URLS=(
  "gist-xuehu2319|https://gist.fshare.wang/xuehu2319/all.yaml?key=q8317831"
  "gist-tdison|https://gist.fshare.wang/Tdison/all.yaml?key=q8317831"
  "gist-wlget|https://gist.fshare.wang/WLget/all.yaml?key=q8317831"
)

# 链接列表型订阅（每行一个 URL，BestSub 会逐个拉取）
SUB_LIST_SOURCES=(
  "gist-subscribes|https://gist.githubusercontent.com/sucan2233/7c426c2d0494ce074d41852e509be155/raw/03c689dbba7e90b70f04f55b976f885f7fcb6497/subscribes.txt"
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

  echo "添加具名订阅..."
  for entry in "${SUB_NAMED_URLS[@]}"; do
    local name="${entry%%|*}" url="${entry#*|}"
    if echo "$existing" | python3 -c "import json,sys; subs=json.load(sys.stdin)['data']; sys.exit(0 if any('$url' in s.get('url',[]) for s in subs) else 1)" 2>/dev/null; then
      echo "  跳过已存在: $name"
      continue
    fi
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

  echo "添加链接列表订阅..."
  for entry in "${SUB_LIST_SOURCES[@]}"; do
    local name="${entry%%|*}" list_url="${entry#*|}"
    if echo "$existing" | python3 -c "import json,sys; subs=json.load(sys.stdin)['data']; sys.exit(0 if any('$list_url' in s.get('url',[]) for s in subs) else 1)" 2>/dev/null; then
      echo "  跳过已存在: $name"
      continue
    fi
    api POST /api/v1/sub/create -d "{
      \"name\": \"$name\",
      \"url\": [\"$list_url\"],
      \"url_type\": 1,
      \"enable\": 1,
      \"auto_update\": 1,
      \"cron_expr\": \"0 */6 * * *\",
      \"proxy_mode\": 0
    }" >/dev/null
    echo "  已添加列表: $name ($(curl -sS "$list_url" | grep -cE '^https?://' || echo '?') 条链接)"
  done
}

refresh_subscriptions() {
  echo "刷新所有订阅（后台异步执行，大列表需较长时间）..."
  local ids count_before
  ids=$(api GET /api/v1/sub/list | python3 -c "import json,sys; print(' '.join(s['id'] for s in json.load(sys.stdin)['data']))")
  count_before=$(api GET /api/v1/sub/list | python3 -c "import json,sys; print(sum(s.get('node_num',0) for s in json.load(sys.stdin)['data']))")
  for id in $ids; do
    api POST "/api/v1/sub/refresh/$id" >/dev/null 2>&1 || true
  done
  echo "  已触发刷新，当前节点池: $count_before 个（会继续增长）"
  echo "  含 gist 大列表时建议稍后在 WebUI 查看订阅进度"
}

# 删除已刷新完成且可用节点为 0 的订阅（仍在刷新中的暂不删除）
prune_empty_subscriptions() {
  echo "剔除可用节点为 0 的订阅..."
  api GET /api/v1/sub/list | COOKIE_JAR="$COOKIE_JAR" BASE_URL="$BASE_URL" python3 <<'PY'
import json, os, subprocess, sys

cookie = os.environ["COOKIE_JAR"]
base = os.environ["BASE_URL"]
removed = 0
subs = json.load(sys.stdin)["data"]
for s in subs:
    if s.get("node_num", 0) != 0:
        continue
    refreshed = s.get("refreshed_at") or ""
    name = s.get("name", s["id"])
    if not refreshed or refreshed.startswith("0001"):
        print(f"  跳过（尚未刷新完成）: {name}", file=sys.stderr)
        continue
    r = subprocess.run(
        ["curl", "-sS", f"{base}/api/v1/sub/del/{s['id']}", "-X", "DELETE", "-b", cookie],
        capture_output=True, text=True,
    )
    try:
        payload = json.loads(r.stdout or "{}")
    except json.JSONDecodeError:
        payload = {}
    if payload.get("code") == 200:
        print(f"  已删除: {name}")
        removed += 1
    else:
        msg = payload.get("message", r.stdout.strip())
        print(f"  删除失败: {name} - {msg}", file=sys.stderr)
if removed == 0:
    print("  无需剔除")
PY
}

create_task() {
  echo "创建/更新检测任务..." >&2
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
        "params": {"url": "https://gstatic.com/generate_204", "timeout_ms": 3000, "attempts": 1},
        "concurrency": 100,
        "pass": {"max_delay": 3000},
        "order": 1,
        "node_pool_delete": 0,
        "skip_existing": 1
      },
      {
        "type": "speed",
        "params": {"url": "https://speed.cloudflare.com/__down?during=download&bytes=999999", "timeout_ms": 10000},
        "concurrency": 10,
        "pass": {"min_download_speed": 1024},
        "order": 2,
        "node_pool_delete": 0,
        "skip_existing": 0
      }
    ]
  }'

  if [[ -n "$task_id" ]]; then
    api PUT "/api/v1/task/update/$task_id" -d "$payload" >/dev/null
    echo "  已更新任务: $task_id" >&2
  else
    task_id=$(api POST /api/v1/task/create -d "$payload" | python3 -c "import json,sys; print(json.load(sys.stdin)['data']['id'])")
    echo "  已创建任务: $task_id" >&2
  fi

  if [[ ! "$task_id" =~ ^[0-9a-f-]{36}$ ]]; then
    echo "任务 ID 无效: $task_id" >&2
    return 1
  fi
  printf '%s' "$task_id"
}

run_task() {
  local task_id="$1"
  if [[ ! "$task_id" =~ ^[0-9a-f-]{36}$ ]]; then
    echo "跳过任务运行：无效的任务 ID" >&2
    return 1
  fi

  local finished_before
  finished_before=$(api GET "/api/v1/task/get/$task_id" | python3 -c "import json,sys; print(json.load(sys.stdin)['data'].get('finished_at',''))")

  echo "运行节点检测任务（节点多时需要更久，可在 WebUI 查看进度）..." >&2
  if ! api POST "/api/v1/task/run/$task_id" >/dev/null 2>&1; then
    echo "  任务可能已在运行，继续等待结果..." >&2
  fi

  for i in $(seq 1 360); do
    sleep 5
    local finished_now result_count
    finished_now=$(api GET "/api/v1/task/get/$task_id" 2>/dev/null | python3 -c "
import json,sys
try:
    print(json.load(sys.stdin)['data'].get('finished_at',''))
except: print('')
" 2>/dev/null || echo "")
    result_count=$(api GET "/api/v1/task/result/$task_id" 2>/dev/null | python3 -c "
import json,sys
try:
    print(json.load(sys.stdin)['data'])
except: print(0)
" 2>/dev/null || echo "0")

    if [[ -n "$finished_now" && "$finished_now" != "$finished_before" ]]; then
      echo "  检测完成，优质节点: ${result_count} 个" >&2
      if [[ "$result_count" == "0" ]]; then
        echo "  提示: 无节点通过筛选，可在 WebUI 调低测速/延迟条件后重跑" >&2
      fi
      return 0
    fi
    if (( i % 12 == 0 )); then
      echo "  仍在检测中... (${i}x5s，当前通过: ${result_count})" >&2
    fi
  done
  echo "  检测超时，可在 WebUI 查看进度: $BASE_URL" >&2
}

create_share() {
  local task_id="$1"
  echo "创建分享链接..." >&2
  local shares share_id token result_count
  result_count=$(api GET "/api/v1/task/result/$task_id" 2>/dev/null | python3 -c "
import json,sys
try: print(json.load(sys.stdin)['data'])
except: print(0)
" 2>/dev/null || echo "0")

  shares=$(api GET /api/v1/share/list)
  share_id=$(echo "$shares" | python3 -c "
import json,sys
for s in json.load(sys.stdin)['data']:
    if s.get('name')=='优质节点':
        print(s['id']); break
" 2>/dev/null || true)

  local payload
  if [[ "$result_count" =~ ^[0-9]+$ && "$result_count" -gt 0 ]]; then
    payload="{
      \"name\": \"优质节点\",
      \"filter\": {\"max_delay\": 3000, \"min_download_speed\": 1024, \"limit\": 100},
      \"node_rename_expression\": \"\",
      \"result_tasks\": [{\"id\": \"$task_id\"}],
      \"subscriptions\": [],
      \"nodes\": [],
      \"tags\": []
    }"
  else
    echo "  任务结果为空，改为分享全部订阅节点（延迟≤3000ms）" >&2
    payload='{
      "name": "优质节点",
      "filter": {"max_delay": 3000, "limit": 100},
      "node_rename_expression": "",
      "result_tasks": [],
      "subscriptions": [],
      "nodes": [],
      "tags": []
    }'
    # all_input 无法通过 share API 表达，改用已有且可用节点 > 0 的订阅
    payload=$(api GET /api/v1/sub/list | python3 -c "
import json,sys
subs=json.load(sys.stdin)['data']
refs=[{'id': s['id']} for s in subs if s.get('node_num',0)>0]
print(json.dumps({
    'name': '优质节点',
    'filter': {'max_delay': 3000, 'limit': 100},
    'node_rename_expression': '',
    'result_tasks': [],
    'subscriptions': refs,
    'nodes': [],
    'tags': [],
}))
")
  fi

  if [[ -n "$share_id" ]]; then
    api PUT "/api/v1/share/update/$share_id" -d "$payload" >/dev/null
    token=$(api GET /api/v1/share/list | python3 -c "
import json,sys
for s in json.load(sys.stdin)['data']:
    if s.get('name')=='优质节点':
        print(s['token']); break
")
  else
    token=$(api POST /api/v1/share/create -d "$payload" | python3 -c "
import json,sys
r=json.load(sys.stdin)
if r.get('code')!=200: raise SystemExit(r.get('message','create share failed'))
print(r['data']['token'])
")
  fi

  if [[ ! "$token" =~ ^[a-z]{16}$ ]]; then
    echo "分享创建失败，请到 WebUI 手动创建" >&2
    return 1
  fi
  printf '%s' "$token"
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
  task_id=$(create_task) || { echo "任务创建失败" >&2; exit 1; }
  run_task "$task_id" || true
  prune_empty_subscriptions
  token=$(create_share "$task_id") || token=""

  echo ""
  echo "========================================"
  echo " 配置完成！"
  echo "========================================"
  echo "WebUI:    $BASE_URL  (账号: $USER / $PASS)"
  if [[ -n "$token" ]]; then
    echo "优质节点: $BASE_URL/share/$token"
    echo ""
    echo "导入 Clash/Mihomo 客户端："
    echo "  订阅链接 → $BASE_URL/share/$token"
  else
    echo "分享链接未生成，请到 WebUI → 分享 页手动创建"
  fi
  echo ""
  echo "日志: tail -f $LOG"
  echo "停止: kill \$(cat $PID_FILE)"
}

main "$@"
