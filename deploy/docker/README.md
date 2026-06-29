# AI Daily 容器化部署方案 — tc197 (javey.pro)

## 架构概览

```
┌─────────────────────────────────────────────────────────────┐
│ tc197 宿主机                                                 │
│                                                             │
│  Nginx (systemd)                                            │
│  ├── /etc/nginx/conf.d/ngx_javey.pro.conf  (已有，主配置)    │
│  └── /etc/nginx/locations.d/                                │
│       └── ngx_ai-daily.location.conf  ← 新增               │
│                  ↓                                          │
│          alias /data/aidaily/www/ai-daily/                   │
│                                                             │
│  Docker Containers (按需启动，非常驻)                         │
│  ├── aidaily-crawler   # Go 爬虫（采集+摘要+导出JSON）       │
│  └── aidaily-builder   # Node.js 前端构建（SSG→静态HTML）    │
│                                                             │
│  Cron (宿主机)                                              │
│  ├── 08:00 全量采集  → crawler → builder                    │
│  ├── 14:00 增量采集  → crawler → builder                    │
│  └── 20:00 增量采集  → crawler → builder                    │
│                                                             │
│  /data/aidaily/         # 数据持久化目录                     │
│  ├── bin/               # Go 二进制                         │
│  ├── configs/           # config-prod.yaml                  │
│  ├── data/              # ai_news.db (151MB)                │
│  ├── frontend/          # Next.js 源码                      │
│  ├── www/ai-daily/      # 静态产物（Nginx 直接 serve）       │
│  └── logs/              # 日志                              │
└─────────────────────────────────────────────────────────────┘
```

## 文件清单

| 文件 | 用途 |
|------|------|
| `docker-compose.yml` | 容器编排定义 |
| `crawler/Dockerfile` | 爬虫容器镜像（Alpine + 时区） |
| `builder/Dockerfile` | 前端构建容器镜像（Node 22 + rsync） |
| `builder/build.sh` | 容器内构建脚本 |
| `config-prod.yaml` | 生产配置（路径已适配容器） |
| `crawl.sh` | 宿主机 Cron 入口脚本 |
| `.env.example` | 环境变量模板 |
| `ngx_ai-daily.location.conf` | Nginx location 片段 |
| `migrate.sh` | 一键迁移脚本 |

## 一键迁移

```bash
# 在本地 Mac 执行
cd /Users/javey/CodeBuddy/ai_website
bash deploy/docker/migrate.sh
```

脚本自动完成：
1. tc197 创建目录结构
2. 从 tc106 流式传输数据（二进制、DB、静态产物、前端源码）
3. 上传 Docker 部署文件
4. 部署 Nginx 配置
5. 构建 Docker 镜像
6. 测试前端构建

## 迁移后手动步骤

### 1. 配置环境变量
```bash
ssh tc197
vi /data/aidaily/.env
# 填入 WEBSEARCH_API_KEY=xxx
```

### 2. 测试爬虫
```bash
ssh tc197 "cd /data/aidaily && sudo docker compose run --rm aidaily-crawler"
```

### 3. 验证访问
```bash
curl -I https://www.javey.pro/ai-daily/
```

### 4. 配置定时任务
```bash
ssh tc197 "sudo crontab -e"
# 添加：
# 0 8 * * *  /data/aidaily/crawl.sh >> /data/aidaily/logs/cron.log 2>&1
# 0 14 * * * /data/aidaily/crawl.sh --incremental >> /data/aidaily/logs/cron.log 2>&1
# 0 20 * * * /data/aidaily/crawl.sh --incremental >> /data/aidaily/logs/cron.log 2>&1
```

### 5. 停用 tc106
```bash
ssh tc106 "sudo crontab -l | grep -v aidaily | sudo crontab -"
```

## 日常运维

### 手动触发全量采集
```bash
ssh tc197 "cd /data/aidaily && sudo ./crawl.sh"
```

### 手动触发增量采集
```bash
ssh tc197 "cd /data/aidaily && sudo ./crawl.sh --incremental"
```

### 仅重建前端（不采集）
```bash
ssh tc197 "cd /data/aidaily && sudo docker compose run --rm aidaily-builder"
```

### 查看日志
```bash
ssh tc197 "tail -50 /data/aidaily/logs/crawl-$(date +%Y-%m-%d)-*.log"
```

### 更新前端代码
```bash
# 本地同步源码到服务器
rsync -avz --exclude=node_modules --exclude=.next --exclude=out \
  frontend/ tc197:/data/aidaily/frontend/
# 触发重建
ssh tc197 "cd /data/aidaily && sudo docker compose run --rm aidaily-builder"
```

### 更新爬虫二进制
```bash
# 本地交叉编译
cd backend && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ai-news-crawler ./cmd/crawler/
scp ai-news-crawler tc197:/data/aidaily/bin/
```
