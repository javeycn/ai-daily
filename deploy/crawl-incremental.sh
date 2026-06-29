#!/bin/bash
# AI Daily 增量爬虫脚本
# 由 crontab 在 14:00 和 20:00 执行，只采集最近 6 小时的增量数据
# 增量模式下只更新首页和当天日报页，不做全站重建

set -euo pipefail

INSTALL_DIR="/data/system/aidaily"
LOG_DIR="${INSTALL_DIR}/logs"
CONFIG="${INSTALL_DIR}/configs/config-prod.yaml"
BINARY="${INSTALL_DIR}/bin/ai-news-crawler"

# 确保日志目录存在
mkdir -p "${LOG_DIR}"

# 日志文件按日期命名（增量模式用后缀区分）
LOG_FILE="${LOG_DIR}/crawl-$(date +%Y-%m-%d)-incr-$(date +%H%M).log"

echo "========================================" >> "${LOG_FILE}"
echo "AI Daily Incremental Crawl Start: $(date '+%Y-%m-%d %H:%M:%S')" >> "${LOG_FILE}"
echo "========================================" >> "${LOG_FILE}"

# 执行增量爬虫（--incremental 标志 + 默认回溯 6 小时）
cd "${INSTALL_DIR}"
export WEBSEARCH_API_KEY="${WEBSEARCH_API_KEY:-}"
"${BINARY}" --config "${CONFIG}" --incremental --since-hours 6 >> "${LOG_FILE}" 2>&1

EXIT_CODE=$?

echo "========================================" >> "${LOG_FILE}"
echo "AI Daily Incremental Crawl End: $(date '+%Y-%m-%d %H:%M:%S'), exit code: ${EXIT_CODE}" >> "${LOG_FILE}"
echo "========================================" >> "${LOG_FILE}"

# 清理 30 天前的日志
find "${LOG_DIR}" -name "crawl-*.log" -mtime +30 -delete

exit ${EXIT_CODE}
