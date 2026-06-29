#!/bin/bash
set -e

echo "=== AI News Website 环境初始化 ==="

# 检查 Go
if ! command -v go &> /dev/null; then
    echo "错误: 未安装 Go，请安装 Go 1.22+"
    exit 1
fi
echo "Go 版本: $(go version)"

# 检查 Node.js
if ! command -v node &> /dev/null; then
    echo "错误: 未安装 Node.js，请安装 Node.js 18+"
    exit 1
fi
echo "Node.js 版本: $(node --version)"

# 检查 npm
if ! command -v npm &> /dev/null; then
    echo "错误: 未安装 npm"
    exit 1
fi
echo "npm 版本: $(npm --version)"

# 安装后端依赖
echo ""
echo "[1/3] 安装 Go 后端依赖..."
cd "$(dirname "$0")/../backend"
go mod tidy

# 安装前端依赖
echo ""
echo "[2/3] 安装前端依赖..."
cd "$(dirname "$0")/../frontend"
npm install

# 创建必要目录
echo ""
echo "[3/3] 创建数据目录..."
mkdir -p backend/data
mkdir -p frontend/data/daily

# 提示配置 API Key
echo ""
echo "=== 初始化完成 ==="
echo ""
echo "下一步配置："
echo "  1. 复制 backend/configs/config.yaml 并填写你的 LLM API Key"
echo "  2. 或通过环境变量设置：export OPENAI_API_KEY=your_key"
echo "  3. 运行部署脚本：bash scripts/deploy.sh"
echo ""
echo "可选：手动运行采集服务测试"
echo "  cd backend && go run ./cmd/crawler/ -config configs/config.yaml"
