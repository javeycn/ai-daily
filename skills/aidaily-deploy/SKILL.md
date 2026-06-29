---
name: aidaily-deploy
description: AI Daily 项目构建与部署技能。当用户需要构建、部署、上传或更新线上 AI Daily 站点时使用。涵盖 Go 后端编译、Next.js 前端构建、服务器文件上传、线上部署验证等完整流程。
metadata:
  openclaw:
    requires:
      tools: [execute_command, read_file, write_to_file]
    optional:
      tools: [web_fetch]
---

# AI Daily 部署技能

## 何时使用

当用户提到以下意图时激活此技能：
- 部署 / 发布 / 上线
- 构建前端 / 编译后端
- 上传文件到服务器
- 更新线上站点
- 回滚到旧版本

## 环境信息

- **服务器**：YOUR_SERVER_IP，端口 YOUR_PORT，账号 YOUR_USER
- **连接方式**：`ssh -p YOUR_PORT YOUR_USER@YOUR_SERVER_IP`
- **安装目录**：`/data/system/aidaily`
- **站点目录**：`/data/web/www/ai-daily`
- **切换 root**：`sudo su -`

## 部署流程

### Phase 1: 本地构建

#### Go 后端
```bash
cd backend
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /tmp/ai-news-crawler ./cmd/crawler/
```

#### Next.js 前端
```bash
# 1. 从服务器同步最新数据
scp -P YOUR_PORT YOUR_USER@YOUR_SERVER_IP:/data/web/www/ai-daily/data/index.json frontend/data/
scp -P YOUR_PORT 'YOUR_USER@YOUR_SERVER_IP:/data/web/www/ai-daily/data/daily/*.json' frontend/data/daily/

# 2. 构建
cd frontend && rm -rf .next out && npm run build

# 3. 打包（排除 data）
cd out && tar czf /tmp/aidaily-frontend.tar.gz --exclude='data' .
```

### Phase 2: 上传到服务器

```bash
scp -P YOUR_PORT /tmp/aidaily-frontend.tar.gz YOUR_USER@YOUR_SERVER_IP:/tmp/
scp -P YOUR_PORT /tmp/ai-news-crawler YOUR_USER@YOUR_SERVER_IP:/tmp/
```

### Phase 3: 服务器部署

```bash
# SSH 到服务器执行
SITE_DIR="/data/web/www/ai-daily"

# 1. 备份
cp -r "$SITE_DIR" "${SITE_DIR}-bak-$(date +%Y%m%d%H%M%S)"

# 2. 清理旧前端资源（保留 data/）
rm -rf "$SITE_DIR/_next" "$SITE_DIR/daily" "$SITE_DIR/archive" "$SITE_DIR/about" "$SITE_DIR/search"
rm -f "$SITE_DIR/index.html" "$SITE_DIR/404.html" "$SITE_DIR/index.txt"

# 3. 解压新资源
cd "$SITE_DIR" && tar xzf /tmp/aidaily-frontend.tar.gz

# 4. 替换后端二进制（如需要）
cp /tmp/ai-news-crawler /data/system/aidaily/bin/ai-news-crawler
chmod +x /data/system/aidaily/bin/ai-news-crawler
```

### Phase 4: 验证

```bash
# 检查文件完整性
test -f "$SITE_DIR/index.html" && echo "✅ index.html"
test -d "$SITE_DIR/_next" && echo "✅ _next/"
test -d "$SITE_DIR/daily" && echo "✅ daily/"
ls "$SITE_DIR/data/daily/"*.json | wc -l  # JSON 文件数

# 检查外部依赖
grep -r "fonts.googleapis" "$SITE_DIR/_next/static/css/" || echo "✅ 无 Google Fonts"
grep "requestIdleCallback" "$SITE_DIR/index.html" && echo "✅ 统计脚本延迟加载"
```

## 回滚流程

```bash
# 查看备份列表
ls -la /data/web/www/ai-daily-bak-*

# 回滚到指定备份
sudo rm -rf /data/web/www/ai-daily
sudo cp -r /data/web/www/ai-daily-bak-{timestamp} /data/web/www/ai-daily
```

## 注意事项

- 部署前端时**绝不删除 data/ 目录**
- 每次部署必须先**备份**
- Go 编译必须使用 `CGO_ENABLED=0`
- 前端构建前先从服务器**同步最新 JSON 数据**
