#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
BACKEND_DIR="$PROJECT_DIR/backend"
FRONTEND_DIR="$PROJECT_DIR/frontend"

echo "=== AI News Website 一键部署脚本 ==="

# 1. 构建 Go 后端
echo ""
echo "[1/4] 构建 Go 后端..."
cd "$BACKEND_DIR"
go mod tidy
go build -o bin/crawler ./cmd/crawler/
echo "后端构建完成: bin/crawler"

# 2. 运行采集服务（采集今天的数据）
echo ""
echo "[2/4] 运行数据采集..."
if [ -n "$1" ]; then
  ./bin/crawler -config configs/config.yaml -date "$1"
else
  ./bin/crawler -config configs/config.yaml
fi
echo "数据采集完成"

# 3. 构建前端静态站点
echo ""
echo "[3/4] 构建前端静态站点..."
cd "$FRONTEND_DIR"
npm install
npm run build
echo "前端构建完成: out/"

# 4. 部署提示
echo ""
echo "[4/4] 部署完成！"
echo ""
echo "下一步："
echo "  1. 将 frontend/out/ 目录的内容部署到 Nginx 静态托管"
echo "  2. 配置 cron 定时任务，例如："
echo "     0 8 * * * cd $BACKEND_DIR && ./bin/crawler -config configs/config.yaml && cd $FRONTEND_DIR && npm run build"
echo ""
echo "Nginx 配置示例："
echo "  server {"
echo "    listen 80;"
echo "    server_name your-domain.com;"
echo "    root $FRONTEND_DIR/out;"
echo "    index index.html;"
echo "    location / { try_files \$uri \$uri.html \$uri/ =404; }"
echo "  }"
