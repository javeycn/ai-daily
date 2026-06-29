#!/bin/bash
# AI Daily 前端构建部署脚本（宿主机直接运行版本）
# 增量构建优化：只生成最近 7 天 daily 页面，历史页面保留不覆盖
set -euo pipefail

AIDAILY_DIR="/data/aidaily"
FRONTEND_DIR="${AIDAILY_DIR}/frontend"
DATA_DIR="${AIDAILY_DIR}/www/ai-daily/data"
SITE_DIR="${AIDAILY_DIR}/www/ai-daily"
LOG_FILE="${AIDAILY_DIR}/logs/frontend-build.log"

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

# 2. 安装依赖
log "Step 2: Checking dependencies..."
cd "$FRONTEND_DIR"
if [ ! -f "node_modules/.package-lock.json" ] || [ "package.json" -nt "node_modules/.package-lock.json" ]; then
    log "  Installing npm dependencies..."
    npm ci >> "$LOG_FILE" 2>&1
else
    log "  Dependencies up to date, skipping install"
fi

# 3. 构建 Next.js SSG（prebuild 会自动生成搜索索引）
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

# 4. 部署到站点目录
# - data/ 保留不覆盖（由爬虫直接写入）
# - daily/ 中已有的旧页面保留，只增量添加新页面
# - _next/static/ 只追加不删除，防止浏览器缓存的旧 HTML 引用 chunk 404
log "Step 4: Deploying..."

# 4a. 同步除 data/、daily/、_next/static/ 外的文件（首页、搜索页、归档页等）
rsync -a --delete --exclude="data/" --exclude="daily/" --exclude="_next/static/" out/ "$SITE_DIR/"

# 4b. _next/static/ 只追加不删除旧 chunk（保证旧 HTML 引用仍有效）
rsync -a out/_next/static/ "$SITE_DIR/_next/static/"

# 4c. 增量同步 daily/ 目录（只添加新的，不删除旧的）
if [ -d "out/daily" ]; then
    rsync -a out/daily/ "$SITE_DIR/daily/"
fi

# 4d. 搜索数据文件（search-data-{hash}.json + search-manifest.json）
# 这些文件由 Next.js 从 public/ 自动复制到 out/，由 4a 的 rsync 同步过去
# 旧的 search-data-{hash}.json 会被 --delete 自动清理
if ls "$SITE_DIR"/search-data-*.json &>/dev/null; then
    INDEX_FILE=$(ls "$SITE_DIR"/search-data-*.json | head -1)
    INDEX_SIZE=$(du -h "$INDEX_FILE" | cut -f1)
    log "  Search data deployed: $(basename $INDEX_FILE) ($INDEX_SIZE)"
fi

# 4e. 清理过期的旧 chunk 文件（保留 7 天，防止磁盘积累）
OLD_CHUNKS=$(find "$SITE_DIR/_next/static" -type f -mtime +7 2>/dev/null | wc -l)
if [ "$OLD_CHUNKS" -gt 0 ]; then
    find "$SITE_DIR/_next/static" -type f -mtime +7 -delete 2>/dev/null || true
    find "$SITE_DIR/_next/static" -type d -empty -delete 2>/dev/null || true
    log "  Cleaned $OLD_CHUNKS old chunk files (>7 days)"
fi

DEPLOYED_DATES=$(ls -d "$SITE_DIR/daily/"* 2>/dev/null | wc -l || echo 0)
log "  Deployed: $DEPLOYED_DATES date pages (incremental)"
log "=== Frontend build completed ==="
