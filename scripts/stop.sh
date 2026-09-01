#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
RUNTIME="$ROOT/runtime"

stop_one() {
  local name="$1" pid_file="$2"
  if [[ -f "$pid_file" ]] && kill -0 "$(cat "$pid_file")" 2>/dev/null; then
    kill "$(cat "$pid_file")" 2>/dev/null || true
    echo "已停止 $name (PID $(cat "$pid_file"))"
  else
    echo "$name 未在运行"
  fi
  rm -f "$pid_file"
}

stop_one "BestSub" "$RUNTIME/bestsub.pid"
stop_one "MiniSubConvert" "$RUNTIME/minisubconvert.pid"
