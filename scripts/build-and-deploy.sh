#!/bin/bash
# build-and-deploy.sh — 前端自动构建部署脚本
# 在爬虫生成 JSON 数据后调用，自动重建 Next.js 静态页面并部署到站点目录。
#
# 用法：
#   /data/system/aidaily/scripts/build-and-deploy.sh
#
# 目录约定：
#   前端源码：/data/system/aidaily/frontend
#   JSON 数据：/data/web/www/ai-daily/data
#   站点目录：/data/web/www/ai-daily

set -euo pipefail

# === 配置 ===
FRONTEND_DIR="/data/system/aidaily/frontend"
DATA_DIR="/data/web/www/ai-daily/data"
SITE_DIR="/data/web/www/ai-daily"
LOG_FILE="/data/system/aidaily/logs/frontend-build.log"

# === 日志函数 ===
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*" | tee -a "$LOG_FILE"
}

# 确保日志目录存在
mkdir -p "$(dirname "$LOG_FILE")"

log "=== Frontend build started ==="

# === 1. 同步最新 JSON 数据到前端 data 目录 ===
log "Step 1: Syncing JSON data..."
mkdir -p "$FRONTEND_DIR/data/daily"
cp -f "$DATA_DIR/index.json" "$FRONTEND_DIR/data/index.json" 2>/dev/null || true
cp -f "$DATA_DIR/daily/"*.json "$FRONTEND_DIR/data/daily/" 2>/dev/null || true

DATA_COUNT=$(ls "$FRONTEND_DIR/data/daily/"*.json 2>/dev/null | wc -l)
log "  Synced $DATA_COUNT daily JSON files"

if [ "$DATA_COUNT" -eq 0 ]; then
    log "ERROR: No JSON data files found, aborting build"
    exit 1
fi

# === 2. 构建 Next.js 静态页面 ===
log "Step 2: Building Next.js..."
cd "$FRONTEND_DIR"

# 清理上次构建缓存
rm -rf .next out

npm run build >> "$LOG_FILE" 2>&1
BUILD_EXIT=$?

if [ $BUILD_EXIT -ne 0 ]; then
    log "ERROR: Next.js build failed (exit code: $BUILD_EXIT)"
    exit 1
fi

# 验证构建产物
if [ ! -d "out" ] || [ ! -f "out/index.html" ]; then
    log "ERROR: Build output missing (out/index.html not found)"
    exit 1
fi

OUT_FILES=$(find out -name "*.html" | wc -l)
log "  Build successful: $OUT_FILES HTML files generated"

# === 3. 部署到站点目录（保留 data 目录不覆盖）===
log "Step 3: Deploying to site directory..."

# 部署策略：
#   - HTML/非静态文件：rsync --delete 确保最新
#   - _next/static/：只追加不删除旧 chunk，防止浏览器缓存的旧 HTML 引用失效
#   - data/：由爬虫直接写入，不覆盖
if command -v rsync &>/dev/null; then
    # 第一步：同步除 _next/static 和 data 之外的文件（删除旧文件）
    rsync -a --delete --exclude='data/' --exclude='_next/static/' out/ "$SITE_DIR/"
    # 第二步：_next/static 只追加，不删除旧 chunk（保证旧 HTML 仍能加载）
    rsync -a out/_next/static/ "$SITE_DIR/_next/static/"
else
    # 没有 rsync 则手动同步（保留 data 目录）
    cd out
    for item in *; do
        if [ "$item" != "data" ]; then
            cp -rf "$item" "$SITE_DIR/"
        fi
    done
    cd ..
fi

# === 3.1 清理过期的旧 chunk 文件（保留 7 天内的）===
log "Step 3.1: Cleaning old static chunks (older than 7 days)..."
OLD_CHUNKS=$(find "$SITE_DIR/_next/static" -type f -mtime +7 2>/dev/null | wc -l)
if [ "$OLD_CHUNKS" -gt 0 ]; then
    find "$SITE_DIR/_next/static" -type f -mtime +7 -delete 2>/dev/null || true
    # 清理空目录
    find "$SITE_DIR/_next/static" -type d -empty -delete 2>/dev/null || true
    log "  Cleaned $OLD_CHUNKS old chunk files (>7 days)"
else
    log "  No old chunks to clean"
fi

# 验证部署
DEPLOYED_DATES=$(ls -d "$SITE_DIR/daily/2026-"* 2>/dev/null | wc -l)
log "  Deployed: $DEPLOYED_DATES date pages available"

log "=== Frontend build completed successfully ==="
