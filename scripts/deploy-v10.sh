#!/bin/bash
# AI Daily v10 部署脚本（本地化字体：Google Fonts → fontsource 本地包）
# 用法：
# 1. 本地: scp -P YOUR_PORT /tmp/aidaily-v10-assets.tar.gz YOUR_USER@YOUR_SERVER_IP:/tmp/
# 2. 本地: scp -P YOUR_PORT /tmp/deploy-v10.sh YOUR_USER@YOUR_SERVER_IP:/tmp/
# 3. 服务器: sudo bash /tmp/deploy-v10.sh

set -euo pipefail

SITE_DIR="/data/web/www/ai-daily"
BACKUP_DIR="/data/web/www/ai-daily-bak-$(date +%Y%m%d%H%M%S)"
TMP_TAR="/tmp/aidaily-v10-assets.tar.gz"

echo "=== AI Daily v10 部署 ==="
echo "变更内容："
echo "  1. 字体本地化：Google Fonts → fontsource 本地包"
echo "  2. 消除 fonts.googleapis.com / fonts.gstatic.com 外部依赖"
echo "  3. 字体按 unicode-range 分片按需加载，优化中国大陆访问体验"
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
echo "--- 字体文件 ---"
FONT_COUNT=$(find "$SITE_DIR/_next/static/media/" -name "*.woff2" 2>/dev/null | wc -l)
echo "  woff2 字体文件数: $FONT_COUNT"
echo "--- 页面文件 ---"
echo "  index.html: $(test -f $SITE_DIR/index.html && echo '✅' || echo '❌')"
echo "  archive/:   $(test -d $SITE_DIR/archive && echo '✅' || echo '❌')"
echo "  about/:     $(test -d $SITE_DIR/about && echo '✅' || echo '❌')"
echo "  daily/:     $(test -d $SITE_DIR/daily && echo '✅' || echo '❌')"
echo "  search/:    $(test -d $SITE_DIR/search && echo '✅' || echo '❌')"
echo ""

# 6. 检查是否还有 Google Fonts 引用
echo "--- Google Fonts 检查 ---"
if grep -r "fonts.googleapis" "$SITE_DIR/_next/static/css/" 2>/dev/null; then
  echo "  ⚠️  CSS 中仍有 Google Fonts 引用！"
else
  echo "  ✅ CSS 中无 Google Fonts 引用"
fi
echo ""
echo "=== 部署完成 ==="
echo "⚠️  data/ 和 data/daily/ 目录完整保留，未被修改"
echo "⚠️  备份目录: $BACKUP_DIR"
