#!/bin/bash
# AI Daily 全量采集（宿主机 cron 调用）
# 用法：/data/aidaily/crawl.sh [--incremental]
set -euo pipefail

AIDAILY_DIR="/data/aidaily"
LOG_DIR="${AIDAILY_DIR}/logs"
mkdir -p "${LOG_DIR}"

MODE="full"
EXTRA_ARGS=""
if [ "${1:-}" = "--incremental" ]; then
    MODE="incremental"
    EXTRA_ARGS="--incremental --since-hours 6"
fi

LOG_FILE="${LOG_DIR}/crawl-$(date +%Y-%m-%d)-${MODE}-$(date +%H%M).log"

echo "========================================" >> "${LOG_FILE}"
echo "AI Daily ${MODE} Crawl Start: $(date "+%Y-%m-%d %H:%M:%S")" >> "${LOG_FILE}"
echo "========================================" >> "${LOG_FILE}"

cd "${AIDAILY_DIR}"

# 1. 运行爬虫容器
docker compose run --rm \
    aidaily-crawler \
    --config /app/config-prod.yaml ${EXTRA_ARGS} >> "${LOG_FILE}" 2>&1
CRAWL_EXIT=$?

echo "Crawl exit code: ${CRAWL_EXIT}" >> "${LOG_FILE}"

# 2. 爬虫成功后触发前端构建
if [ ${CRAWL_EXIT} -eq 0 ]; then
    echo "Triggering frontend build..." >> "${LOG_FILE}"
    docker compose run --rm aidaily-builder >> "${LOG_FILE}" 2>&1 || true
fi

echo "========================================" >> "${LOG_FILE}"
echo "AI Daily ${MODE} Crawl End: $(date "+%Y-%m-%d %H:%M:%S")" >> "${LOG_FILE}"
echo "========================================" >> "${LOG_FILE}"

# 清理 30 天前日志
find "${LOG_DIR}" -name "crawl-*.log" -mtime +30 -delete 2>/dev/null || true

exit ${CRAWL_EXIT}
