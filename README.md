# AI Daily - 每日 AI 资讯聚合平台

> **线上地址**：https://www.javey.pro/ai-daily/

每天自动从全球 **45+ 个新闻源**采集最新 AI 资讯，通过大语言模型生成中文摘要与智能分类，经过三级聚合流水线精选 **30 条**高质量日报，覆盖 **8 大分类**，内置国内/国际内容均衡策略和厂商去重机制，部署为静态站点。

---

## 功能特性

### 采集层
- **45+ 新闻源**：覆盖 8 大分组 — 国际科技媒体、AI 专业媒体、AI 大厂官方博客、云服务 AI 博客、AI 前沿思想领袖 & Newsletter、AI 学术 & 论文、创投 & 科技商业、国内科技媒体
- **Web 搜索补充**：集成 Serper.dev Google Search API，20 组搜索查询覆盖突发新闻和 RSS 难以触达的主题（按需启用）
- **7 策略图片提取**：media:content(带 medium) → media:content(无 medium) → media:thumbnail → enclosure → content:encoded 中 img → description 中 img → summary 中 img，配合 `looksLikeImageURL` + `isValidArticleImage` 过滤追踪像素和图标
- **177 个 AI 关键词过滤**：二次确认采集内容与 AI 相关
- **URL 指纹去重**：基于 SHA256 哈希 + SQLite UNIQUE 约束双重去重
- **增量模式**：`--incremental --since-hours 6`，只采集最近 N 小时新文章，局部更新 HTML

### 智能处理层
- **LLM 双 Provider 支持**：OpenAI（默认）和 Claude，支持自定义 BaseURL，环境变量覆盖
- **8 类分类体系**：按内容本质分类（非地域），含结构化 Prompt 边界判定规则和源感知分类
- **5 维标签系统**：技术主题、工程与工具、行业与场景、公司与平台、商业与治理
- **双语标题**：LLM 返回 `original_title` 时自动拼接 "英文原标题 / 中文标题" 格式
- **三级聚合流水线**：preSelect（本地预选 ~90 篇）→ summarizeCandidates（LLM 批量摘要）→ diverseSample（最终采样 30 篇）
- **智能采样算法（diverseSample）**：三阶段软性配额 — Phase 1 按 8 分类保底配额（共 20 条 Min，含国内/国际均衡策略），Phase 2 按源优先级+分类配额偏好填充，Phase 3 自由补位至 30 条
- **内容均衡策略**：分类级别的国内/国际内容比例控制 — 云服务国内≥60%、模型前沿 1:1、产品与应用 1:1 + AI Agent 优先，厂商去重（同分类同厂商最多 2 条）
- **5 级来源优先级（rankSources）**：从 Tier 1（16 个 AI 大厂+思想领袖）到 Tier 5（高产量国内源），国内头部媒体提权至 Tier 3，AWS/Azure/GCP 降权至 Tier 4

### 前端展示层
- **分类筛选**：首页按 8 大分类 Tab 筛选，每个分类带 Emoji 和彩色标签
- **分类折叠 + 渐进渲染**：每个分类组默认折叠显示（移动端 2 条 / 桌面端 3 条），"展开更多"按钮触发完整渲染；IntersectionObserver 懒渲染视口外分类（先显示 skeleton 占位，滚动到附近才渲染真实 DOM），大幅减少首屏 DOM 数量和内存占用
- **滚动触发入场动画**：卡片级别 IntersectionObserver 检测，进入视口时 fade-in 上滑动画，`transitionDelay` 上限 0.2s，避免长列表动画堆积
- **全文搜索**：客户端 Fuse.js 模糊搜索，6 字段加权（chinese_title ×2.0, original_title ×1.5, tags ×1.5, category ×1.5, summary ×1.0, source ×0.5）
- **深色/浅色主题**：支持主题切换，localStorage 持久化
- **精选 + 全量**：首页展示 30 条精选，日报详情页展示全部已摘要文章
- **双语标题展示**：自动拆分中英文标题，桌面端显示英文副标题
- **响应式设计**：移动端优化，分类栏横向滚动带渐变遮罩，展开按钮 44px 触摸目标

### 部署层
- **Cron 3 次/日调度**：08:00 全量 + 14:00/20:00 增量
- **前端自动构建**：爬虫完成后自动触发 Next.js SSG 重建，无需本地构建上传
- **Nginx 静态托管**：静态资源缓存 7 天，JSON 数据缓存 10 分钟
- **纯 Go 编译**：CGO_ENABLED=0 交叉编译，无外部依赖

---

## 系统架构

```
                    crontab (3次/天)
                    ├── 08:00 全量 crawl.sh
                    ├── 14:00 增量 crawl-incremental.sh
                    └── 20:00 增量 crawl-incremental.sh
                             │
                             ▼
                  cmd/crawler/main.go (入口)
                             │
         ┌───────────────────┼───────────────────┐
         ▼                   ▼                   ▼
   43+1 Crawlers         filter              dedup/URLHash
   (42 RSS +           (177 关键词)          (SHA256 去重)
    1 WebSearch)
         │
         ▼
   SQLite (WAL 模式, modernc.org/sqlite, 无 CGO)
         │
         ▼
   Aggregator (三级分级处理 + 内容均衡)
   ├── preSelect    (本地预选 ~90 篇, 按源优先级轮询)
   ├── summarize    (LLM 批量摘要, OpenAI/Claude, 并发控制)
   └── diverseSample (三阶段配额采样 + 国内/国际均衡 + 厂商去重 → 30 篇精选)
         │
         ▼
   Exporter → daily/{date}.json + index.json
         │
         ├──────────────────────────────────┐
         ▼                                  ▼
   htmlgen → 静态 HTML               build-and-deploy.sh
   (Go 生成简易 HTML,                (同步 JSON → Next.js SSG →
    Nginx 托管于 /ai-daily/)          rsync 部署到站点目录)
```

### 5 步处理流程

| 步骤 | 说明 |
|------|------|
| Step 1 | 并发采集（semaphore 控制并发数 = 5），保存到 SQLite，利用 URL 唯一约束去重 |
| AI Filter | `filter.FilterArticles()` 做二次 AI 相关性确认（177 个关键词匹配） |
| Step 2 | 三级聚合：preSelect → summarizeCandidates → diverseSample，生成日报 |
| Step 3 | 导出 JSON 日报 + 更新索引文件 |
| Step 4 | 生成静态 HTML（增量模式仅更新首页+当天日报页；全量模式重建整站） |

---

## 8 分类体系

按内容本质分类，不按地域。百度发模型和 OpenAI 发模型是同一类事。

| # | 分类名称 | Emoji | 内容范围 |
|---|---------|-------|---------|
| 1 | 模型前沿 | 🧠 | 模型能力突破、开源模型发布、Benchmark 评测、训练/推理新范式 |
| 2 | 产品与应用 | 🚀 | AI 产品发布、行业落地案例、AI 工具 |
| 3 | 深度洞察 | 📊 | 趋势分析、领袖观点、行业报告、深度技术博文 |
| 4 | 云服务与平台 | ☁️ | 云厂商 AI 服务动态、MaaS 平台、API 定价 |
| 5 | AI工程 | ⚙️ | Agent 框架、RAG 架构、MCP/A2A 协议、Prompt 工程、LLMOps、AI 编程工具 |
| 6 | AI基础设施 | 🔧 | AI 芯片、推理引擎、训练框架、向量数据库 |
| 7 | 商业与投资 | 💰 | 融资、并购、IPO、市场格局、企业战略 |
| 8 | AI安全 | 🛡️ | AI 监管法规、安全对齐、数据隐私、开源许可、伦理争议 |

### 软性配额

| 分类 | 最低保底 | 建议上限 | 均衡策略 |
|------|---------|---------|---------|
| 模型前沿 | 3 | 4 | 国内/国际 ≈ 1:1，头部厂商优先 |
| 产品与应用 | 3 | 5 | 国内/国际 ≈ 1:1，AI Agent 内容优先 |
| 深度洞察 | 1 | 4 | — |
| 云服务与平台 | 5 | 7 | 国内 ≥ 60%，同厂商 ≤ 2 条 |
| AI工程 | 2 | 4 | — |
| AI基础设施 | 2 | 3 | — |
| 商业与投资 | 2 | 3 | — |
| AI安全 | 2 | 3 | ≤ 云服务，与 AI工程/基础设施持平 |
| **合计 Min** | **20** | — | — |

---

## 45+ 新闻源完整列表

### 国际科技媒体（7 个）
The Verge, VentureBeat, TechCrunch, Ars Technica, Wired AI, MIT Technology Review, THE DECODER

### AI 专业媒体 & 聚合（8 个）
Unite.AI, AI News, MarkTechPost, Synced Review, IEEE Spectrum, AI Hub Today, DailyAI, InfoQ AI

### AI 大厂官方博客（8 个）
OpenAI Blog, DeepMind, Google Research, Meta Engineering, HuggingFace, Anthropic, Microsoft AI, NVIDIA Blog

### 云服务 AI 博客（3 个）
AWS ML Blog, Google Cloud AI Blog, Azure AI Services Blog

### AI 前沿思想领袖 & Newsletter（10 个）
Jack Clark (Import AI), Simon Willison, Lilian Weng, Latent Space, The Gradient, The Batch (吴恩达), BAIR Blog (伯克利), Chip Huyen, Eugene Yan, Karpathy

### AI 学术 & 论文（2 个）
arXiv cs.AI, Papers With Code

### 创投 & 科技商业（2 个）
Hacker News (YC), PanDaily

### 国内科技媒体（4 个）
36氪, 机器之心, IT之家, 智东西

### Web 搜索补充（1 个，按需启用）
WebSearchCrawler (Serper.dev) — 20 组搜索查询，覆盖通用 AI 突发新闻、大厂动态、产品发布、融资商业、AI 大佬动态、AI 安全与伦理、国内外云厂商

---

## 5 级来源优先级

| 层级 | 来源 | 说明 |
|------|------|------|
| Tier 1 | OpenAI Blog, DeepMind, Google Research, Anthropic, Microsoft AI, Import AI, Simon Willison, Lilian Weng, Latent Space, The Gradient, HuggingFace, The Batch, BAIR Blog, Chip Huyen, Eugene Yan, Karpathy (共 16 个) | AI 大厂官方 + 前沿思想领袖 |
| Tier 2 | TechCrunch, The Verge, Wired, MIT Technology Review, Ars Technica, IEEE Spectrum, NVIDIA Blog, InfoQ AI (共 8 个) | 国际顶级科技媒体 |
| Tier 3 | 机器之心, 智东西, 量子位, InfoQ中文, THE DECODER, MarkTechPost, Unite.AI, AI News, Synced Review, Meta Engineering, arXiv cs.AI, Papers With Code, AI Hub Today, DailyAI (共 14 个) | **国内头部媒体**（从 Tier 5/6 提权）+ 国际 AI 专业媒体 & 学术 |
| Tier 4 | Hacker News, PanDaily, VentureBeat, Web Search, **AWS ML Blog, GCP AI Blog, Azure AI Blog** (共 7 个) | 创投 & 科技商业 + 搜索补充 + **国际云服务博客**（从 Tier 2 降权） |
| Tier 5 | 36氪, IT之家 (共 2 个) | 高产量国内源（适当限制配额） |

> **V9 调整要点**：国内头部媒体（机器之心/智东西/量子位/InfoQ中文）从原 Tier 5/6 提升到 Tier 3，与国际专业媒体同级；AWS/Azure/GCP 云服务博客从原 Tier 2 降至 Tier 4，避免国际云服务内容挤占配额。

---

## 内容均衡策略（V9 新增）

### 国内/国际内容比例控制

针对不同分类，采用差异化的国内/国际内容均衡策略：

| 分类 | 均衡目标 | 实现机制 |
|------|---------|---------|
| 云服务与平台 | 国内 ≥ 60%（≥3:2） | Phase 1 优先选国内云厂商内容，60% 配额保留给国内 |
| 模型前沿 | 国内/国际 ≈ 1:1 | 交替选取国内/国际头部模型厂商内容 |
| 产品与应用 | 国内/国际 ≈ 1:1 | AI Agent 内容优先，再按国内/国际交替填充 |
| 其他分类 | 自然分布 | 仅做厂商去重，不强制比例 |

### 厂商去重机制

- **同分类同厂商最多 2 条**（`MaxPerVendorInCategory = 2`），避免单一厂商霸占某个分类
- **厂商名称归一化**：通过 `VendorAliases` 映射表（约 40 个条目），将不同写法统一为标准厂商名
  - 例如：`openai` / `gpt` / `chatgpt` → `OpenAI`，`kimi` / `月之暗面` / `moonshot` → `月之暗面`
- **三阶段均适用**：Phase 1 保底、Phase 2 偏好填充、Phase 3 自由补位均检查厂商限额

### AI Agent 内容优先

在「产品与应用」分类中，AI Agent/智能体相关内容享有优先选取权：

- **27 个 Agent 关键词**：Agent、Manus、OpenClaw、WorkBuddy、AutoGPT、CrewAI、MetaGPT、AutoGen、Devin、Copilot、Cursor、Bolt、Windsurf、CodeBuddy、Lovable、Coze 等
- Phase 1 优先选取匹配 Agent 关键词的文章，再按国内/国际均衡策略填充剩余配额

### 三阶段采样算法

```
Phase 1 — 保底阶段（≈20 条）
  ├── 按 8 分类的 Min 配额逐个填充
  ├── 每个分类内部应用各自的均衡策略：
  │   ├── 云服务：国内 60% 优先 + 厂商限额
  │   ├── 模型前沿：1:1 国内/国际 + 头部厂商优先
  │   ├── 产品与应用：Agent 优先 + 1:1 均衡
  │   └── 其他：按源优先级 + 厂商限额
  └── 输出 Min 配额条数

Phase 2 — 偏好填充阶段
  ├── 按 Preferred 上限继续填充（不超过 30）
  ├── 延续 Phase 1 的分类均衡策略
  └── 已满 Preferred 的分类跳过

Phase 3 — 自由补位阶段
  ├── 不再区分分类偏好，按整体质量排序
  ├── 仍然检查厂商限额
  └── 填充至总数 30 条
```

---

## 快速开始

### 环境要求

- Go 1.22+
- Node.js 18+
- OpenAI / Claude API Key

### 安装

```bash
# 1. 初始化环境
bash scripts/setup.sh

# 2. 配置 API Key（二选一）
#    方式一：编辑配置文件
vim backend/configs/config.yaml

#    方式二：环境变量
export OPENAI_API_KEY=your_api_key_here

# 3. 一键部署（采集 + 前端构建）
bash scripts/deploy.sh
```

### 手动运行

```bash
# 全量采集（默认采集最近 3 天数据）
cd backend
go run ./cmd/crawler/ -config configs/config.yaml

# 指定日期
go run ./cmd/crawler/ -config configs/config.yaml -date 2026-03-27

# 增量采集（最近 6 小时）
go run ./cmd/crawler/ -config configs/config.yaml -incremental -since-hours 6

# 仅重建摘要（跳过采集）
go run ./cmd/crawler/ -config configs/config.yaml -summarize-only -date 2026-03-27

# 开发前端
cd frontend
npm run dev
```

### CLI 参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `-config` | string | `configs/config.yaml` | 配置文件路径 |
| `-date` | string | 今天 | 指定采集日期（格式：`2006-01-02`） |
| `-summarize-only` | bool | false | 跳过采集，仅对已有文章生成摘要并导出 |
| `-incremental` | bool | false | 增量模式：只采集最近 N 小时新文章，局部更新 HTML |
| `-since-hours` | int | 6 | 增量模式下的回溯小时数 |

### 定时任务

```bash
# 编辑 crontab
crontab -e

# 粘贴以下内容（根据实际路径修改 INSTALL_DIR）
INSTALL_DIR=/data/system/aidaily

# 全量采集（每天 8:00）—— 采集最近 3 天 + LLM 摘要 + 全站 HTML 重建
0 8 * * * ${INSTALL_DIR}/deploy/crawl.sh

# 增量采集（每天 14:00）—— 采集最近 6 小时 + 局部 HTML 更新
0 14 * * * ${INSTALL_DIR}/deploy/crawl-incremental.sh

# 增量采集（每天 20:00）—— 采集最近 6 小时 + 局部 HTML 更新
0 20 * * * ${INSTALL_DIR}/deploy/crawl-incremental.sh
```

---

## 项目结构

```
ai_website/
├── backend/                    # Go 后端采集服务
│   ├── cmd/crawler/            # 程序入口 (main.go)
│   ├── internal/               # 核心模块
│   │   ├── config/             # 配置管理 (YAML + 环境变量覆盖)
│   │   ├── crawler/            # 新闻采集器 (42 个 RSS + 1 个 WebSearch)
│   │   │   ├── crawler.go      # Crawler 接口定义
│   │   │   ├── rss.go          # RSS/Atom 通用基类 + 7 策略图片提取
│   │   │   ├── websearch.go    # Serper.dev Web 搜索采集器
│   │   │   └── *.go            # 各新闻源采集器实现
│   │   ├── dedup/              # URL SHA256 去重
│   │   ├── filter/             # AI 关键词过滤 (177 个关键词)
│   │   ├── summarizer/         # LLM 摘要生成 (OpenAI/Claude 双 Provider)
│   │   ├── aggregator/         # 三级聚合 + 智能采样
│   │   ├── exporter/           # JSON 导出 + 索引更新
│   │   ├── htmlgen/            # 静态 HTML 页面生成
│   │   ├── repository/         # SQLite WAL 数据存储 (modernc.org/sqlite)
│   │   └── model/              # 数据模型 + 8 分类 + 配额表
│   ├── configs/                # 配置文件 (config.yaml / config-prod.yaml)
│   └── docs/                   # 设计方案文档
├── frontend/                   # Next.js 14 前端 (SSG)
│   ├── src/app/                # 页面路由 (首页/日报/归档/搜索/关于)
│   ├── src/lib/                # 类型定义 + 分类颜色/Emoji 映射 + 自定义 hooks
│   └── data/                   # 采集数据 (JSON)
├── deploy/                     # 部署配置
│   ├── crawl.sh                # 全量采集脚本（完成后触发前端构建）
│   ├── crawl-incremental.sh    # 增量采集脚本（完成后触发前端构建）
│   ├── crontab.example         # Cron 调度配置示例
│   └── nginx-ai-daily.conf    # Nginx location 配置
├── scripts/                    # 初始化和部署脚本
│   └── build-and-deploy.sh     # 前端自动构建部署脚本（服务端 Next.js SSG）
└── skills/                     # CodeBuddy Skill 定义文件
```

---

## 配置说明

主要配置项（`backend/configs/config.yaml`）：

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `crawler.delay` | 采集请求间隔（秒） | 2 |
| `crawler.timeout` | HTTP 超时（秒） | 30 |
| `crawler.max_retries` | 最大重试次数 | 3 |
| `crawler.max_articles_per_source` | 每个源最大采集数 | 50 |
| `database.path` | SQLite 数据库路径 | `./data/ai_news.db` |
| `llm.provider` | LLM 提供商（`openai` / `claude`） | openai |
| `llm.base_url` | API 基础 URL（支持自定义） | — |
| `llm.api_key` | API Key | — |
| `llm.model` | 模型名称 | gpt-4o-mini |
| `llm.max_concurrent` | LLM 并发数 | 5 |
| `llm.timeout` | LLM 请求超时（秒） | 60 |
| `llm.max_summary_tokens` | 最大摘要 token 数 | 500 |
| `aggregator.max_daily_articles` | 每日精选文章数 | 30 |
| `aggregator.min_daily_articles` | 每日最小文章数 | 20 |
| `output.dir` | JSON 输出目录 | `../frontend/data` |
| `output.html_dir` | HTML 输出目录 | `../frontend/out` |
| `web_search.enabled` | 是否启用 Web 搜索 | false |
| `web_search.api_key` | Serper.dev API Key | — |
| `web_search.max_results` | 每次查询最大结果数 | 10 |
| `web_search.timeout` | 搜索超时（秒） | 15 |
| `web_search.queries` | 自定义查询列表（空则使用默认 20 组） | — |
| `web_search.exclude_domains` | 排除域名列表 | — |
| `log.level` | 日志级别 | info |
| `log.file` | 日志文件路径 | — |

### 环境变量覆盖

| 环境变量 | 覆盖字段 |
|----------|----------|
| `OPENAI_API_KEY` | `llm.api_key` |
| `CLAUDE_API_KEY` | `llm.api_key` |
| `LLM_PROVIDER` | `llm.provider` |
| `LLM_MODEL` | `llm.model` |
| `LLM_BASE_URL` | `llm.base_url` |
| `DB_PATH` | `database.path` |
| `OUTPUT_DIR` | `output.dir` |
| `HTML_DIR` | `output.html_dir` |
| `LOG_LEVEL` | `log.level` |
| `WEBSEARCH_API_KEY` | `web_search.api_key` |

---

## 技术亮点

### 三级聚合流水线

避免对全库所有文章都做昂贵的 LLM 调用，采用分级处理策略：

1. **preSelect（本地预选）**：纯本地计算，按来源优先级轮询从 1000+ 篇缩减到 ~90 篇候选。已有摘要的文章直接入选（不浪费 LLM 投入），同时保证新采集的高优先级文章有机会被 LLM 处理
2. **summarizeCandidates（LLM 摘要）**：仅对候选中未生成摘要的调用 LLM，分批处理（每批 20 篇），信号量控制并发
3. **diverseSample（智能采样）**：三阶段软性配额算法 — Phase 1 按 8 分类保底 + 国内/国际均衡策略（共 20 条），Phase 2 按分类偏好填充，Phase 3 自由补位至 30 条。全程检查厂商去重（同分类同厂商 ≤ 2 条）

### LLM Prompt 工程

结构化 Prompt 包含：
- **角色设定**：AI 行业分析师和科技新闻编辑
- **8 类分类指南**：每个分类含具体内容范围和典型案例
- **关键边界判定**：明确云服务 vs AI 基础设施、AI 工程 vs 产品与应用等边界
- **源感知分类**：指定 Import AI、Simon Willison 等来源优先考虑"深度洞察"
- **5 维标签体系**：80+ 候选标签，按技术主题/工程与工具/行业与场景/公司与平台/商业与治理五维度组织
- **双语输出**：自动返回 `original_title` + `chinese_title`

### 7 策略图片提取

按优先级依次尝试 7 种策略从 RSS 中提取文章配图：

| 优先级 | 策略 | 说明 |
|--------|------|------|
| 1 | `media:content[@medium='image']` | 带 medium="image" 属性的 media:content |
| 2 | `media:content`（无 medium 过滤） | URL 看起来像图片的 media:content |
| 3 | `media:thumbnail` | 媒体缩略图 |
| 4 | `enclosure` | type 为 `image/*` 的 enclosure，或 URL 看起来像图片 |
| 5 | `content:encoded` / `content` 中的 `<img>` | 从正文 HTML 提取第一个 img src |
| 6 | `description` 的 HTML 中的 `<img>` | 从描述 HTML 提取 |
| 7 | `summary` 的 HTML 中的 `<img>` | 从摘要 HTML 提取 |

配合 `looksLikeImageURL`（检查 7 种图片扩展名 + CDN 路径）和 `isValidArticleImage`（排除追踪像素/favicon/logo/spacer 等）双重过滤。

### SQLite WAL 模式

- **纯 Go 驱动**：`modernc.org/sqlite`，无 CGO 依赖，支持 `CGO_ENABLED=0` 交叉编译
- **WAL 模式**：`PRAGMA journal_mode=WAL`，提升并发读写性能
- **Busy Timeout**：`PRAGMA busy_timeout=5000`，避免并发写入时 "database is locked"
- **5 个索引**：url（UNIQUE）、published_at、source、hash、category

---

## 扩展新采集源

实现 `crawler.Crawler` 接口即可添加新采集源：

```go
// internal/crawler/newsource.go
package crawler

import "time"

// NewSourceCrawler 创建一个新的 XXX 采集器。
func NewSourceCrawler(timeout time.Duration) *RSSCrawler {
    return NewRSSCrawler(
        "source_name",      // 采集器标识
        "https://example.com/feed.xml",  // RSS URL
        "Source Name",       // 来源显示名称
        timeout,
    )
}
```

然后在 `cmd/crawler/main.go` 的 `initCrawlers` 中注册，并在 `aggregator.go` 的 `rankSources` 中添加优先级即可。

---

## 迭代历程

### V9 — 内容均衡策略 + 服务端前端自动构建（2026-03-29）

**内容均衡策略（4 大分类级别的国内/国际内容比例控制）：**
- **云服务与平台**：国内云厂商优先级高于国际厂商，国内:国际 ≥ 3:2（60%），同厂商最多 2 条。AWS/Azure/GCP 来源降权至 Tier 4
- **模型前沿**：降低整体占比（Preferred 5→4），国内/国际 ≈ 1:1 均衡。头部厂商（智谱/minimax/kimi/qwen/通义/豆包/元宝/混元/DeepSeek + Anthropic/OpenAI/Google/Grok/Meta）优先
- **产品与应用**：国内/国际 ≈ 1:1，AI Agent/智能体内容优先（27 个 Agent 关键词：Manus、OpenClaw、WorkBuddy、AutoGPT、CrewAI、Devin、Cursor 等）
- **AI 安全**：每日总数控制在与 AI 工程和 AI 基础设施持平，不超过云服务与平台

**厂商去重机制：**
- `MaxPerVendorInCategory = 2`，同分类同厂商最多 2 条
- `VendorAliases` 映射表（约 40 条目）统一厂商名称（如 kimi/月之暗面/moonshot → 月之暗面）
- `ExtractVendor()` 方法从文章标题和摘要中提取归一化厂商名

**来源优先级调整（6 级 → 5 级）：**
- 国内头部媒体（机器之心/智东西/量子位/InfoQ中文）从 Tier 5/6 提升到 Tier 3
- AWS ML Blog / GCP AI Blog / Azure AI Blog 从 Tier 2 降到 Tier 4

**采样算法升级（两阶段 → 三阶段）：**
- Phase 1：保底阶段（按 Min 配额 + 分类级均衡策略）
- Phase 2：偏好填充（按 Preferred 上限 + 延续均衡策略）
- Phase 3：自由补位（按整体质量排序 + 厂商限额检查）

**配额调整：**
- 模型前沿 Preferred 5→4，云服务 Min 2→5 / Preferred 4→7，AI工程 Min 3→2，深度洞察 Min 2→1

**服务端前端自动构建：**
- 前端源码部署到服务器，爬虫完成后自动触发 `build-and-deploy.sh`
- 脚本流程：同步 JSON 数据 → Next.js SSG 重建 → rsync 部署到站点目录
- 集成到 `crawl.sh` 和 `crawl-incremental.sh`，无需本地构建上传
- 解决新日期归档页 404 问题（SSG 页面随数据自动生成）

**新增代码结构：**
- `model/article.go`：新增 `DomesticModelKeywords`、`TopDomesticModelKeywords`、`InternationalModelKeywords`、`TopInternationalModelKeywords`、`DomesticProductKeywords`、`AgentProductKeywords`、`VendorAliases` 等数据表和 `IsDomesticModel()`、`IsAgentProduct()`、`ExtractVendor()` 等方法
- `aggregator/aggregator.go`：`diverseSample()` 完全重写为三阶段算法，`rankSources()` 重构为 5 级优先级
- `scripts/build-and-deploy.sh`：新增前端自动构建部署脚本

### V8 — 前端页面加载优化（2026-03-27）

- **分类折叠**：每个分类组默认只显示前 N 条（移动端 2 条 / 桌面端 3 条），点击"展开更多"显示全部，点击"收起"折叠回去
- **IntersectionObserver 渐进渲染**：视口外的分类组先渲染 skeleton 占位，滚动到附近时才渲染真实卡片 DOM，`rootMargin` 移动端 100px / 桌面端 200px
- **滚动触发入场动画**：取消原来的固定 `animationDelay: 0.03 * index`（100 条 = 3s 延迟），改为 IntersectionObserver 触发的 `scroll-fade-in` CSS transition，`transitionDelay` 上限 0.2s
- **折叠底部渐变遮罩**：`.collapse-fade-mask::after` 伪元素，暗示下方有更多内容
- **展开按钮移动端优化**：`min-height: 44px`（Apple HIG 触摸目标）、`-webkit-tap-highlight-color: transparent`
- **共享 hooks**：新增 `useLazyRender`（IntersectionObserver 懒渲染）和 `useIsMobile`（响应式断点判断，matchMedia 监听）
- 零新增第三方依赖，新增代码约 80 行

### V7 — 前端兼容修复 + 云服务源扩充（2026-03-27）

- **新增 3 个云服务 AI 博客 RSS 源**：AWS ML Blog (Tier 3)、Google Cloud AI Blog (Tier 3)、Azure AI Services Blog (Tier 3)
- **WebSearch 搜索关键词扩充**（13 → 21 条）：新增 8 条覆盖腾讯云/阿里云/华为云/百度云/火山引擎/AWS/GCP/Azure
- **首页分类样式兼容**：新增旧 5 分类（国际AI模型、国内AI厂商、产品落地、开源、商业硬件）的颜色和 emoji 映射，解决数据迁移前分类标签显示灰色
- **归档页分类筛选**：重构归档页，新增可点击分类筛选标签栏，支持按分类过滤日报列表
- **移动端宽度一致性**：统一 glass-card padding 为 `p-4 sm:p-8`，消除分类筛选栏移动端溢出
- 爬虫总数：41 → 44，搜索关键词：13 → 21

### V6 — 8 分类体系重构（2026-03-26）

- **拆分"大模型与基础设施"**为 4 个独立维度：模型前沿、AI 工程、AI 基础设施、云服务与平台
- **取消低效分类**："研究与论文"和"开源生态"不再独立，内容分散到对应分类
- **新 8 分类体系**设计与实现，含软性配额采样策略
- **LLM Prompt 全面重写**：8 分类指南 + 边界判定规则 + 源感知分类
- `normalizeCategory` 模糊匹配重写，覆盖 8 分类中英文关键词
- **新增 4 个 RSS 源**：Chip Huyen (Tier 1)、Eugene Yan (Tier 1)、Karpathy (Tier 1)、InfoQ AI (Tier 2)
- HTML 颜色映射更新 + `cat-teal` CSS 类
- 搜索热词和关于页更新

### V5 — 7 分类体系（从地域分类改为内容本质分类）（2026-03-26）

- 从地域分类（国际/国内）改为内容本质分类的 7 分类体系
- DB 全量清洗 + 5 天日报逐天重建
- **新增 3 个 RSS 源**：AI Hub Today、BAIR Blog、DailyAI
- Web Search 查询增强（8 → 13 条）
- `rankSources` 6 级来源优先级体系建立
- `preSelect` 逻辑修复（保证新文章有机会被 LLM 处理）

### V4 — 采集优化（RSS 并发化 + 新增源 + 增量模式）（2026-03-25 ~ 03-26）

- **RSS 并发采集**：semaphore 控制并发数 = 5，大幅提升采集速度
- **LLM 并发提升**：信号量控制 `MaxConcurrent`，批量处理每批 20 篇
- **增量模式**：`--incremental --since-hours 6`，只采集最近 N 小时，局部更新 HTML
- **多频次 Cron 调度**：从单次 08:00 改为 3 次/日（08:00 全量 + 14:00/20:00 增量）
- **WebSearch 采集器**：集成 Serper.dev API，20 组搜索查询，突发新闻补充
- **新增大量 RSS 源**：从初始 ~15 个扩展到 40+ 个，覆盖 AI 大厂博客、思想领袖、学术论文等
- 部署脚本体系建立：`deploy/crawl.sh`、`deploy/crawl-incremental.sh`、`deploy/crontab.example`
- Nginx 配置优化：静态资源 7 天缓存 + JSON 10 分钟缓存

### V3 — 三级聚合流水线（2026-03-25）

- 从全量 LLM 摘要改为**三级分级处理**：preSelect → summarizeCandidates → diverseSample
- **preSelect**：纯本地计算预选候选，避免对 1000+ 篇文章全部调用 LLM
- **diverseSample**：两阶段均衡采样，按来源优先级+时间排序
- **高产量源限制**：IT之家、36氪等最多占总配额 15%
- 日报输出拆分为精选（`FeaturedCategoryGroups`）+ 全量（`CategoryGroups`）

### V2 — LLM 双 Provider + 结构化 Prompt（2026-03-25）

- **Claude Provider**：新增 Claude API 支持（`x-api-key` + `anthropic-version`）
- **结构化 Prompt**：含角色设定、分类指南、标签候选池、边界判定规则
- **5 维标签系统**：技术主题、工程与工具、行业与场景、公司与平台、商业与治理
- **双语标题**：自动返回 `original_title`，拼接 "英文原标题 / 中文标题"
- `normalizeCategory` 模糊匹配：对 LLM 返回的不精确分类名做中英文关键词匹配
- Fallback 机制：LLM 调用失败时使用原标题 + 截断摘要 + 默认分类

### V1 — 初始构建（2026-03-24 ~ 03-25）

- Go 1.22 后端 + Next.js 14 SSG 前端 + SQLite 数据存储
- 基础 RSS 采集器（~15 个来源）
- OpenAI API 摘要生成
- 简单聚合（按时间排序取 top N）
- JSON 导出 + 静态 HTML 生成
- Fuse.js 客户端搜索
- 深色/浅色主题切换
- Nginx 静态托管

---

## 编译与部署

### 交叉编译

```bash
# macOS → Linux amd64
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
go build -ldflags="-s -w" -o ai-news-crawler-linux-amd64 ./cmd/crawler/
```

### 部署流程

1. 编译新后端二进制
2. 上传二进制到服务器
3. 部署前端源码到服务器 `/data/system/aidaily/frontend/`
4. 安装 Node.js 依赖：`cd frontend && npm install`
5. 配置 Nginx：`include /data/system/aidaily/nginx-ai-daily.conf;`
6. 配置 Cron：参考 `deploy/crontab.example`
7. 首次运行全量采集验证（采集完成后自动触发前端构建）

> **前端自动构建**：爬虫脚本（`crawl.sh` / `crawl-incremental.sh`）完成数据采集后，自动调用 `scripts/build-and-deploy.sh` 进行 Next.js SSG 重建。脚本流程：同步 JSON 数据 → `npm run build` → `rsync` 部署到站点目录（保留 `data/` 目录不覆盖）。无需本地构建上传。

---

## License

MIT
