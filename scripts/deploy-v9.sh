#!/bin/bash
# AI Daily v9 部署脚本（方案 A 归档页优化 + 关于页/Footer 文字更新）
# 用法：
# 1. 本地: scp -P YOUR_PORT /tmp/aidaily-v9-assets.tar.gz YOUR_USER@YOUR_SERVER_IP:/tmp/
# 2. 本地: scp -P YOUR_PORT /tmp/deploy-v9.sh YOUR_USER@YOUR_SERVER_IP:/tmp/
# 3. 服务器: sudo bash /tmp/deploy-v9.sh

set -euo pipefail

SITE_DIR="/data/web/www/ai-daily"
BACKUP_DIR="/data/web/www/ai-daily-bak-$(date +%Y%m%d%H%M%S)"
TMP_TAR="/tmp/aidaily-v9-assets.tar.gz"

echo "=== AI Daily v9 部署 ==="
echo "变更内容："
echo "  1. 归档页方案 A 优化：标签 Top3 折叠 + 微型进度条 + 摘要 3 行"
echo "  2. 关于页声明改为 javey#qq.com"
echo "  3. Footer 改为 Powered by AI&LLM, JAVEY.org"
echo ""

# 1. 检查文件
if [ ! -f "$TMP_TAR" ]; then
  echo "❌ 找不到 $TMP_TAR，请先上传"
  exit 1
fi

# 2. 备份
echo "--- 备份 $SITE_DIR → $BACKUP_DIR ---"
cp -r "$SITE_DIR" "$BACKUP_DIR"

# 3. 只替换前端资源，保留 data/ 目录
echo "--- 删除旧的 _next/ 和 HTML 页面 ---"
rm -rf "$SITE_DIR/_next"
rm -f "$SITE_DIR/index.html" "$SITE_DIR/404.html" "$SITE_DIR/index.txt"
rm -rf "$SITE_DIR/404" "$SITE_DIR/about" "$SITE_DIR/archive" "$SITE_DIR/search"
rm -rf "$SITE_DIR/daily"

# 4. 解压新资源
echo "--- 解压新的前端资源 ---"
cd "$SITE_DIR"
tar xzf "$TMP_TAR"

# 5. 验证
echo ""
echo "=== 验证 ==="
echo "--- data 完整性 ---"
ls "$SITE_DIR/data/daily/" 2>/dev/null | wc -l | xargs -I{} echo "  data/daily/ 文件数: {}"
echo "--- _next build ---"
ls "$SITE_DIR/_next/static/" 2>/dev/null | head -3
echo "--- 页面文件 ---"
echo "  index.html: $(test -f $SITE_DIR/index.html && echo '✅' || echo '❌')"
echo "  archive/:   $(test -d $SITE_DIR/archive && echo '✅' || echo '❌')"
echo "  about/:     $(test -d $SITE_DIR/about && echo '✅' || echo '❌')"
echo "  daily/:     $(test -d $SITE_DIR/daily && echo '✅' || echo '❌')"
echo "  search/:    $(test -d $SITE_DIR/search && echo '✅' || echo '❌')"
echo ""
echo "=== 部署完成 ==="
echo "⚠️  data/ 和 data/daily/ 目录完整保留，未被修改"
