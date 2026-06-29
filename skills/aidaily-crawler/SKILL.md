---
name: aidaily-crawler
description: AI Daily 采集器开发与维护技能。当用户需要新增 RSS 源、修改聚合算法、调整分类配额、优化 LLM Prompt 或排查采集问题时使用。
metadata:
  openclaw:
    requires:
      tools: [read_file, write_to_file, execute_command]
---

# AI Daily 采集器开发技能

## 何时使用

当用户提到以下意图时激活此技能：
- 新增 / 删除 / 修改新闻源
- 调整分类配额或采样算法
- 修改 LLM Prompt 或摘要生成逻辑
- 排查采集失败或数据异常
- 调整来源优先级

## 新增 RSS 源（完整流程）

### Step 1: 创建采集器文件

在 `backend/internal/crawler/` 下新建文件：

```go
// internal/crawler/{source_name}.go
package crawler

import "time"

// New{SourceName}Crawler 创建 {Source Name} 采集器。
func New{SourceName}Crawler(timeout time.Duration) *RSSCrawler {
    return NewRSSCrawler(
        "{source_id}",                   // 唯一标识（小写，用于日志和去重）
        "https://example.com/rss.xml",   // RSS Feed URL
        "{Source Display Name}",          // 来源显示名称
        timeout,
    )
}
```

### Step 2: 注册采集器

在 `cmd/crawler/main.go` 的 `initCrawlers()` 中添加：
```go
crawler.New{SourceName}Crawler(timeout),  // 注释说明
```

### Step 3: 添加来源优先级

在 `internal/aggregator/aggregator.go` 的 `rankSources()` 中添加，按重要性选择 Tier 1-5：

| Tier | 说明 |
|------|------|
| 1 | AI 大厂官方 + 顶级思想领袖（16 个） |
| 2 | 国际顶级科技媒体（8 个） |
| 3 | 国内头部媒体 + 国际 AI 专业媒体（14 个） |
| 4 | 创投商业 + 搜索 + 国际云服务博客（7 个） |
| 5 | 高产量国内源（2 个） |

### Step 4: 更新文档

- `README.md` 新闻源列表
- `CHANGELOG.md` 变更记录

## 核心模块说明

### 三级聚合算法（`aggregator/aggregator.go`）

```
preSelect (~90 篇)        → 按来源优先级轮询，纯本地计算
    ↓
summarizeCandidates      → LLM 批量摘要（每批 20 篇，信号量并发控制）
    ↓
diverseSample (30 篇)    → 三阶段软性配额采样
  Phase 1: 保底（按 Min 配额 + 国内/国际均衡）
  Phase 2: 偏好填充（按 Preferred 上限）
  Phase 3: 自由补位（按质量排序 + 厂商去重）
```

### 内容均衡策略（`model/article.go`）

- **云服务**：国内 ≥ 60%
- **模型前沿**：国内/国际 ≈ 1:1
- **产品与应用**：1:1 + AI Agent 优先（27 个 Agent 关键词）
- **厂商去重**：同分类同厂商 ≤ 2 条（`VendorAliases` 约 40 条映射）

### LLM Prompt（`summarizer/summarizer.go`）

- 结构化 Prompt：角色设定 + 8 分类指南 + 边界判定 + 源感知分类
- 输出 JSON 格式：`chinese_title` + `summary` + `tags` + `category` + `importance_score` + `recommendation`
- `normalizeCategory()` 做中英文模糊匹配
- Fallback 机制：LLM 失败 → 原标题 + 截断摘要 + 默认分类

## 常用排查命令

```bash
# 查看最新采集日志
ssh -p YOUR_PORT YOUR_USER@YOUR_SERVER_IP "tail -100 /data/system/aidaily/logs/crawl-$(date +%Y-%m-%d).log"

# 查看数据库文章数
ssh -p YOUR_PORT YOUR_USER@YOUR_SERVER_IP "sqlite3 /data/system/aidaily/data/ai_news.db 'SELECT COUNT(*) FROM articles'"

# 查看今日各源采集数
ssh -p YOUR_PORT YOUR_USER@YOUR_SERVER_IP "sqlite3 /data/system/aidaily/data/ai_news.db \"SELECT source, COUNT(*) FROM articles WHERE date(crawled_at) = date('now') GROUP BY source ORDER BY COUNT(*) DESC\""

# 手动触发全量采集
ssh -p YOUR_PORT YOUR_USER@YOUR_SERVER_IP "/data/system/aidaily/deploy/crawl.sh"

# 手动触发增量采集
ssh -p YOUR_PORT YOUR_USER@YOUR_SERVER_IP "/data/system/aidaily/deploy/crawl-incremental.sh"

# 仅重建摘要（不采集）
ssh -p YOUR_PORT YOUR_USER@YOUR_SERVER_IP "/data/system/aidaily/bin/ai-news-crawler --config /data/system/aidaily/configs/config-prod.yaml --summarize-only --date $(date +%Y-%m-%d)"
```

## 关键文件清单

| 文件 | 说明 |
|------|------|
| `model/article.go` | 数据模型、8 分类常量、配额表、厂商映射 |
| `aggregator/aggregator.go` | 三级聚合 + 采样算法 + 来源优先级 |
| `summarizer/summarizer.go` | LLM Prompt + 解析 + normalizeCategory |
| `crawler/rss.go` | RSS 基类 + 7 策略图片提取 |
| `crawler/websearch.go` | Web 搜索采集器（Serper.dev） |
| `filter/filter.go` | 177 个 AI 关键词过滤 |
| `config/config.go` | 配置加载 + 环境变量覆盖 |
