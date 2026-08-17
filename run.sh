#!/usr/bin/env bash
# 一键本地跑通生产形态：构建前端 + 启动单服务（托管静态文件与 API）。
set -e
cd "$(dirname "$0")"

echo "==> 构建前端"
cd web
[ -d node_modules ] || npm install
npm run build
cd ..

echo "==> 启动服务（默认 http://localhost:8080）"
cd server
go run .
