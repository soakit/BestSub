#!/bin/bash
# BestSub 一键启动脚本
cd "$(dirname "$0")"

if [ ! -f "./BestSub" ]; then
  echo "正在编译 BestSub..."
  cd .. && go build -o runtime/BestSub . && cd runtime
fi

pkill -f "$(pwd)/BestSub" 2>/dev/null
sleep 1

echo "启动 BestSub，订阅 HTTP 服务: http://127.0.0.1:18989"
echo "配置文件: config/config.yaml"
echo "输出目录: output/"
./BestSub
