#!/bin/bash
# AI Daily 采集脚本（宿主机直接运行版本）
# 用法：/data/aidaily/crawl.sh [--incremental]
set -euo pipefail

AIDAILY_DIR="/data/aidaily"
LOG_DIR="${AIDAILY_DIR}/logs"
CONFIG="${AIDAILY_DIR}/configs/config-prod.yaml"
BINARY="${AIDAILY_DIR}/bin/ai-news-crawler"

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

# 加载环境变量
[ -f "${AIDAILY_DIR}/.env" ] && export $(grep -v '^#' "${AIDAILY_DIR}/.env" | xargs) 2>/dev/null || true

# 1. 运行爬虫
"${BINARY}" --config "${CONFIG}" ${EXTRA_ARGS} >> "${LOG_FILE}" 2>&1
CRAWL_EXIT=$?

echo "Crawl exit code: ${CRAWL_EXIT}" >> "${LOG_FILE}"

# 2. 爬虫成功后触发前端构建
if [ ${CRAWL_EXIT} -eq 0 ]; then
    echo "Triggering frontend build..." >> "${LOG_FILE}"
    "${AIDAILY_DIR}/scripts/build-and-deploy.sh" >> "${LOG_FILE}" 2>&1 || true
fi

echo "========================================" >> "${LOG_FILE}"
echo "AI Daily ${MODE} Crawl End: $(date "+%Y-%m-%d %H:%M:%S")" >> "${LOG_FILE}"
echo "========================================" >> "${LOG_FILE}"

# 清理 30 天前日志
find "${LOG_DIR}" -name "crawl-*.log" -mtime +30 -delete 2>/dev/null || true

exit ${CRAWL_EXIT}
