#!/usr/bin/env bash
# 从链接列表导入订阅（支持 gist 等 URL 列表源）
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
RUNTIME="$ROOT/runtime"
BASE_URL="${BESTSUB_URL:-http://127.0.0.1:8080}"
USER="${BESTSUB_USER:-admin}"
PASS="${BESTSUB_PASS:-admin}"
COOKIE_JAR="$RUNTIME/.cookies"
LIST_URL="${1:-$(cat "$RUNTIME/config/subscribes.url" 2>/dev/null || true)}"
NAME="${2:-gist-subscribes}"

if [[ -z "$LIST_URL" ]]; then
  echo "用法: $0 [列表URL] [订阅名称]" >&2
  exit 1
fi

if ! curl -sS --connect-timeout 2 "$BASE_URL/" -o /dev/null 2>/dev/null; then
  echo "BestSub 未运行，请先执行: bash scripts/start.sh" >&2
  exit 1
fi

curl -sS "$BASE_URL/api/v1/user/login" -X POST \
  -H 'Content-Type: application/json' \
  -d "{\"username\":\"$USER\",\"password\":\"$PASS\",\"trust\":true}" \
  -c "$COOKIE_JAR" >/dev/null

python3 <<PY
import json, subprocess, sys, urllib.parse

list_url = """$LIST_URL"""
name = """$NAME"""
cookie = """$COOKIE_JAR"""
base = """$BASE_URL"""

existing = json.loads(subprocess.check_output([
    "curl", "-sS", f"{base}/api/v1/sub/list", "-b", cookie
]))["data"]

for sub in existing:
    urls = sub.get("url") or []
    if list_url in urls or sub.get("name") == name:
        print(f"订阅已存在: {sub['name']} ({sub['id']})")
        sys.exit(0)

payload = {
    "name": name,
    "url": [list_url],
    "url_type": 1,
    "enable": 1,
    "auto_update": 1,
    "cron_expr": "0 */6 * * *",
    "proxy_mode": 0,
}
subprocess.check_output([
    "curl", "-sS", f"{base}/api/v1/sub/create",
    "-X", "POST", "-b", cookie,
    "-H", "Content-Type: application/json",
    "-d", json.dumps(payload),
])
print(f"已添加订阅列表: {name}")
print(f"  来源: {list_url}")
print("  请在 WebUI 订阅页点击刷新，或等待自动更新")
PY
