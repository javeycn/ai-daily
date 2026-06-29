#!/bin/bash
# AI Daily 迁移脚本 — 从 tc106 迁移到 tc197
# 本脚本在本地 Mac 执行，通过 SSH 操作两台服务器
#
# 前置条件：
#   - SSH 别名 tc106 和 tc197 可用
#   - tc197 已安装 Docker + Docker Compose
#
# 用法：bash deploy/docker/migrate.sh

set -euo pipefail

TC106="tc106"
TC197="tc197"
LOCAL_DOCKER_DIR="$(cd "$(dirname "$0")" && pwd)"
REMOTE_DIR="/data/aidaily"

echo "========================================"
echo "AI Daily Migration: tc106 -> tc197"
echo "========================================"
echo ""

# ===== Step 1: 在 tc197 创建目录结构 =====
echo "[Step 1] Creating directory structure on tc197..."
ssh ${TC197} "sudo mkdir -p ${REMOTE_DIR}/{bin,configs,data,frontend,www/ai-daily,logs} && sudo chown -R \$(whoami):\$(whoami) ${REMOTE_DIR}"
echo "  ✓ Directory structure created"

# ===== Step 2: 从 tc106 打包数据传输到 tc197 =====
echo "[Step 2] Packaging data from tc106..."

# 2a. Go 二进制
echo "  Transferring Go binary..."
ssh ${TC106} "sudo cat /data/system/aidaily/bin/ai-news-crawler" | ssh ${TC197} "cat > ${REMOTE_DIR}/bin/ai-news-crawler && chmod +x ${REMOTE_DIR}/bin/ai-news-crawler"
echo "  ✓ Go binary transferred (19MB)"

# 2b. SQLite 数据库
echo "  Transferring SQLite database..."
ssh ${TC106} "sudo cat /data/system/aidaily/data/ai_news.db" | ssh ${TC197} "cat > ${REMOTE_DIR}/data/ai_news.db"
echo "  ✓ Database transferred (151MB)"

# 2c. 静态产物（www/ai-daily）
echo "  Transferring static site (this may take a while)..."
ssh ${TC106} "sudo tar czf - -C /data/web/www ai-daily" | ssh ${TC197} "tar xzf - -C ${REMOTE_DIR}/www/"
echo "  ✓ Static site transferred (186MB)"

# 2d. 前端源码（排除 node_modules、.next、out）
echo "  Transferring frontend source..."
ssh ${TC106} "sudo tar czf - -C /data/system/aidaily/frontend --exclude=node_modules --exclude=.next --exclude=out ." | ssh ${TC197} "tar xzf - -C ${REMOTE_DIR}/frontend/"
echo "  ✓ Frontend source transferred"

echo ""

# ===== Step 3: 上传 Docker 部署文件 =====
echo "[Step 3] Uploading Docker deployment files..."
scp ${LOCAL_DOCKER_DIR}/docker-compose.yml ${TC197}:${REMOTE_DIR}/
scp ${LOCAL_DOCKER_DIR}/config-prod.yaml ${TC197}:${REMOTE_DIR}/configs/
scp ${LOCAL_DOCKER_DIR}/crawl.sh ${TC197}:${REMOTE_DIR}/
scp -r ${LOCAL_DOCKER_DIR}/crawler ${TC197}:${REMOTE_DIR}/
scp -r ${LOCAL_DOCKER_DIR}/builder ${TC197}:${REMOTE_DIR}/
scp ${LOCAL_DOCKER_DIR}/.env.example ${TC197}:${REMOTE_DIR}/.env
ssh ${TC197} "chmod +x ${REMOTE_DIR}/crawl.sh ${REMOTE_DIR}/builder/build.sh"
echo "  ✓ Docker files uploaded"

# ===== Step 4: 上传 Nginx location 片段 =====
echo "[Step 4] Deploying Nginx AI Daily location config..."
scp ${LOCAL_DOCKER_DIR}/ngx_ai-daily.location.conf ${TC197}:/tmp/
ssh ${TC197} "sudo cp /tmp/ngx_ai-daily.location.conf /etc/nginx/locations.d/ngx_ai-daily.location.conf && sudo nginx -t && sudo nginx -s reload"
echo "  ✓ Nginx config deployed and reloaded"

# ===== Step 5: 构建 Docker 镜像 =====
echo "[Step 5] Building Docker images on tc197..."
ssh ${TC197} "cd ${REMOTE_DIR} && sudo docker compose build"
echo "  ✓ Docker images built"

# ===== Step 6: 测试运行 =====
echo "[Step 6] Testing frontend builder..."
ssh ${TC197} "cd ${REMOTE_DIR} && sudo docker compose run --rm aidaily-builder"
echo "  ✓ Frontend build test passed"

echo ""
echo "========================================"
echo "Migration completed!"
echo ""
echo "Next steps:"
echo "  1. Edit ${REMOTE_DIR}/.env on tc197 (fill in WEBSEARCH_API_KEY)"
echo "  2. Test crawler: ssh tc197 'cd ${REMOTE_DIR} && sudo docker compose run --rm aidaily-crawler'"
echo "  3. Verify: curl https://www.javey.pro/ai-daily/"
echo "  4. Set up cron on tc197 (see below)"
echo "  5. Stop cron on tc106"
echo ""
echo "Crontab for tc197 (sudo crontab -e):"
echo "  0 8 * * *  ${REMOTE_DIR}/crawl.sh >> ${REMOTE_DIR}/logs/cron.log 2>&1"
echo "  0 14 * * * ${REMOTE_DIR}/crawl.sh --incremental >> ${REMOTE_DIR}/logs/cron.log 2>&1"
echo "  0 20 * * * ${REMOTE_DIR}/crawl.sh --incremental >> ${REMOTE_DIR}/logs/cron.log 2>&1"
echo "========================================"
