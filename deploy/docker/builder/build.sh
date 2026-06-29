#!/bin/sh
# AI Daily 前端构建脚本（容器内执行）
set -euo pipefail

FRONTEND_DIR="/app/frontend"
SITE_DIR="/app/site"
DATA_DIR="${SITE_DIR}/data"
LOG_FILE="/app/logs/frontend-build.log"

log() {
    echo "[$(date "+%Y-%m-%d %H:%M:%S")] $*" | tee -a "$LOG_FILE"
}

mkdir -p "$(dirname "$LOG_FILE")"
log "=== Frontend build started ==="

# 1. 同步最新 JSON 数据到前端 data 目录
log "Step 1: Syncing JSON data..."
mkdir -p "$FRONTEND_DIR/data/daily"
cp -f "$DATA_DIR/index.json" "$FRONTEND_DIR/data/index.json" 2>/dev/null || true
cp -f "$DATA_DIR/daily/"*.json "$FRONTEND_DIR/data/daily/" 2>/dev/null || true
DATA_COUNT=$(ls "$FRONTEND_DIR/data/daily/"*.json 2>/dev/null | wc -l)
log "  Synced $DATA_COUNT daily JSON files"

if [ "$DATA_COUNT" -eq 0 ]; then
    log "ERROR: No JSON data files found, aborting"
    exit 1
fi

# 2. 安装依赖（node_modules 缺失、为空、或 package.json 有变更时）
log "Step 2: Checking dependencies..."
cd "$FRONTEND_DIR"
if [ ! -f "node_modules/.package-lock.json" ] || [ "package.json" -nt "node_modules/.package-lock.json" ]; then
    log "  Installing npm dependencies..."
    npm ci >> "$LOG_FILE" 2>&1
else
    log "  Dependencies up to date, skipping install"
fi

# 3. 构建 Next.js SSG
log "Step 3: Building Next.js..."
rm -rf .next out
npm run build >> "$LOG_FILE" 2>&1
if [ $? -ne 0 ]; then
    log "ERROR: Next.js build failed"
    exit 1
fi
if [ ! -d "out" ] || [ ! -f "out/index.html" ]; then
    log "ERROR: Build output missing"
    exit 1
fi
OUT_FILES=$(find out -name "*.html" | wc -l)
log "  Build successful: $OUT_FILES HTML files"

# 4. 部署到站点目录（保留 data 目录不覆盖）
log "Step 4: Deploying..."
rsync -a --delete --exclude="data/" out/ "$SITE_DIR/"

DEPLOYED_DATES=$(ls -d "$SITE_DIR/daily/2026-"* 2>/dev/null | wc -l || echo 0)
log "  Deployed: $DEPLOYED_DATES date pages"
log "=== Frontend build completed ==="
