#!/bin/bash
# AI Daily 定时爬虫脚本
# 由 crontab 每天执行，采集新闻 + LLM 摘要 + 导出 JSON

set -euo pipefail

INSTALL_DIR="/data/system/aidaily"
LOG_DIR="${INSTALL_DIR}/logs"
CONFIG="${INSTALL_DIR}/configs/config-prod.yaml"
BINARY="${INSTALL_DIR}/bin/ai-news-crawler"

# 确保日志目录存在
mkdir -p "${LOG_DIR}"

# 日志文件按日期命名
LOG_FILE="${LOG_DIR}/crawl-$(date +%Y-%m-%d).log"

echo "========================================" >> "${LOG_FILE}"
echo "AI Daily Crawl Start: $(date '+%Y-%m-%d %H:%M:%S')" >> "${LOG_FILE}"
echo "========================================" >> "${LOG_FILE}"

# 执行爬虫
cd "${INSTALL_DIR}"
export WEBSEARCH_API_KEY="${WEBSEARCH_API_KEY:-}"
"${BINARY}" --config "${CONFIG}" >> "${LOG_FILE}" 2>&1

EXIT_CODE=$?

echo "========================================" >> "${LOG_FILE}"
echo "AI Daily Crawl End: $(date '+%Y-%m-%d %H:%M:%S'), exit code: ${EXIT_CODE}" >> "${LOG_FILE}"
echo "========================================" >> "${LOG_FILE}"

# 清理 30 天前的日志
find "${LOG_DIR}" -name "crawl-*.log" -mtime +30 -delete

exit ${EXIT_CODE}
