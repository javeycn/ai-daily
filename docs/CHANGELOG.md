# Changelog

AI Daily 项目的完整版本变更记录。

---

## [V13] — 2026-06-12 — 域名切换 javey.pro → javey.me

### 变更
- **替换域名**：停用 `javey.pro`，切换为 `javey.me` 作为站点备用域名
- **跳转策略**（与 javey.org 完全一致）：
  - `http://javey.me/*` / `http://www.javey.me/*` → `https://www.javey.me/*`（HTTPS 强制）
  - `https://javey.me/*` → `https://www.javey.me/*`（裸域跳 www）
  - `https://www.javey.me/` = Javey Talk 首页（MD5 与 `www.javey.org` 一致）
  - `https://www.javey.me/ai-daily/` = AI Daily 静态站点
- **TLS 配置**：TLSv1.2/1.3、HTTP/2、OCSP Stapling
- **javey.pro 已停用**：配置文件重命名为 `.disabled`，证书保留不删

### 变更文件
| 文件 | 操作 | 说明 |
|------|------|------|
| `deploy/ngx_javey.me.conf` | NEW | 替代 ngx_javey.pro.conf |
| `deploy/ngx_javey.pro.conf` | DELETE | 已停用 |
| 服务器 `ngx_javey.pro.conf` | DISABLED | 重命名为 .disabled |
| 服务器 `ngx_javey.me.conf` | MODIFY | 完整配置（裸域跳转 + 首页 + AI Daily） |

---

## [V12] — 2026-05-26 — 新增 javey.pro 域名绑定

### 新增
- **绑定新域名 `javey.pro`**：与 `www.javey.org` 共用同一前端站点（包括 Javey Talk 首页与 AI Daily 子站），通过 Nginx 复用 `/www/javey-root/` 与 `/www/ai-daily/` 物理目录
- **域名跳转策略**：
  - `http://javey.pro/*` 与 `http://www.javey.pro/*` → `https://www.javey.pro/*`（强制 HTTPS）
  - `https://javey.pro/*` → `https://www.javey.pro/*`（裸域跳 www，与 javey.org 风格一致）
  - `https://www.javey.pro/` 默认首页与 `https://www.javey.org/` 行为一致（同一 Javey Talk 首页）
  - `https://www.javey.pro/ai-daily/` 提供 AI Daily 静态站点
- **TLS 配置**：TLSv1.2/1.3、HTTP/2、OCSP Stapling，与 javey.org 保持一致
- **证书**：TrustAsia DV TLS RSA CA 2025 颁发，覆盖 `javey.pro` + `www.javey.pro`，有效期至 2026-08-23

### 变更文件
| 文件 | 操作 | 说明 |
|------|------|------|
| `deploy/ngx_javey.pro.conf` | NEW | 独立的 Nginx server block 配置 |
| 服务器 `/data/web/services/nginx/ssl/javey.pro/` | NEW | 证书目录（`javey.pro_bundle.crt` + `javey.pro.key`） |
| 服务器 `/data/web/services/nginx/conf.d/ngx_javey.pro.conf` | NEW | 部署后的实际配置文件 |

### 验证
- `http://javey.pro/` → 301 `https://www.javey.pro/`
- `https://javey.pro/` → 301 `https://www.javey.pro/`
- `https://www.javey.pro/` 返回 HTTP/2 200，首页内容 MD5 与 `https://www.javey.org/` **完全一致**
- `https://www.javey.pro/ai-daily/` 返回 HTTP/2 200

---

## [V11] — 2026-04-11 — Inter 字体迁移至 jsDelivr CDN

### 优化
- **移除本地 Inter 字体包**：删除 `@fontsource-variable/inter` npm 依赖，清除构建产物中 7 个 woff2 文件（218KB）
- **改用 jsDelivr CDN 加载**：仅引用 `latin` 子集可变字体（1 个请求 48KB），通过 `@font-face` + `font-display: swap` 声明
  - CDN 地址：`https://cdn.jsdelivr.net/fontsource/fonts/inter:vf@latest/latin-wght-normal.woff2`
  - jsDelivr 国内有 CDN 节点，TTFB 稳定在 ~170ms
- **移除不需要的语言子集**：cyrillic、cyrillic-ext、greek、greek-ext、vietnamese、latin-ext 对中文站点无用
- **预连接优化**：`layout.tsx` 添加 `<link rel="preconnect" href="https://cdn.jsdelivr.net">`，提前完成 DNS + TLS 握手

### 效果
| 指标 | 优化前 | 优化后 |
|------|--------|--------|
| 自有服务器字体文件 | 7 个 woff2（218KB） | 0 个 |
| 字体请求数 | 7 | 1（CDN） |
| 字体传输量 | ~218KB | 48KB |
| JS 总传输量（Gzip） | 208KB | 131KB |
| CDN 字体 TTFB | — | ~170ms（稳定） |

### 变更文件
| 文件 | 操作 | 说明 |
|------|------|------|
| `frontend/src/styles/globals.css` | MODIFY | 移除 @fontsource-variable/inter 导入，改用 CDN @font-face |
| `frontend/package.json` | MODIFY | 移除 @fontsource-variable/inter 依赖 |
| `frontend/src/app/layout.tsx` | MODIFY | 添加 jsDelivr preconnect |
| `frontend/next.config.js` | MODIFY | 移除 optimizePackageImports |

---

## [V10] — 2026-04-11 — 性能深度优化 + 服务器源码同步机制修复

### 修复
- **增量构建覆盖问题**：服务器 `/data/system/aidaily/frontend/` 源码未同步本地改动，每次增量采集后 `build-and-deploy.sh` 用旧代码重建，导致 V9 的前端优化被覆盖。现已将完整最新源码同步到服务器，后续增量构建自动使用新代码

### 优化
- **Nginx HTTP/2 启用**：`javey.org` server block 的 `listen 443 ssl;` 改为 `listen 443 ssl http2;`，多路复用减少连接开销
- **OCSP Stapling**：添加 `ssl_stapling on; ssl_stapling_verify on;`，使用阿里/腾讯 DNS（223.5.5.5 / 119.29.29.29），加速 TLS 握手
- **Next.js 构建优化**：启用 `experimental.optimizePackageImports`、`swcMinify`
- **DNS Prefetch**：layout.tsx `<head>` 中添加 `<link rel="dns-prefetch">` 预解析外部域名

### 分析结论
- **无国内不可访问的外部依赖**：所有 JS/CSS/字体/数据均从 `www.javey.org` 自有服务器加载
- 爱站统计（`node96.aizhantj.com`）和 GitHub 链接是仅有的两个外部引用，国内均可访问
- 加载慢的根因是**源站到用户的网络延迟**（TLS 握手 200ms~1.3s 波动），建议后续接入国内 CDN 进一步加速

### 首页资源传输量（Gzip 后）
| 资源 | 大小 |
|------|------|
| HTML | 4KB |
| CSS | 6KB |
| JS 合计（9 个 chunk） | 208KB |
| **总计** | **~218KB** |

### 缓存策略
| 资源类型 | 策略 |
|----------|------|
| JS/CSS（`_next/static/`） | 365 天 + immutable |
| JSON 数据 | 10 分钟 + must-revalidate |
| HTML 页面 | 无缓存，始终最新 |

### 变更文件
| 文件 | 操作 | 说明 |
|------|------|------|
| `frontend/next.config.js` | MODIFY | 添加 optimizePackageImports + swcMinify |
| `frontend/src/app/layout.tsx` | MODIFY | 添加 dns-prefetch meta |
| `deploy/nginx-ai-daily.conf` | — | Nginx 配置在服务器侧直接修改 |
| 服务器 `ngx_javey.org.conf` | MODIFY | 启用 HTTP/2 + OCSP Stapling |

---

## [V9] — 2026-04-11 — 前端体验优化（5 项）+ 字体瘦身 + Gzip 增强

### 优化 1：移除 Noto Sans SC 中文字体
- 移除 `@fontsource-variable/noto-sans-sc` 依赖和 CSS 导入
- woff2 字体文件从 **216 个** 减少到 **7 个**（仅 Inter 英文字体子集）
- 改用系统字体栈：`Inter Variable` → `-apple-system` → `PingFang SC` → `Hiragino Sans GB` → `Microsoft YaHei` → `Noto Sans SC`（系统自带时才用）→ `sans-serif`
- macOS/iOS 用苹方，Windows 用微软雅黑，**完全无需下载中文字体**

### 优化 2：Nginx Gzip 增强
- 压缩级别 2 → 6
- 新增 `application/json` 和 `image/svg+xml` 类型压缩
- JSON 数据传输量：35KB → 6KB（压缩比 82%）

### 优化 3：首页 PC 端全部展示
- PC 端 `defaultVisible` 从固定 3 条改为 `group.articles.length`（全部展示）
- PC 端 `expanded` 初始值改为 `true`
- 移动端保持原有 2 条折叠逻辑不变

### 优化 4：搜索页热门关键词扩充
- 从 8 个扩充到 16 个，铺满 2 排
- 新增：Agent、多模态、Claude、推理优化、AI安全、视频生成、Anthropic、AI编程

### 优化 5：归档页分页
- 默认展示最近 2 个月数据
- 底部"加载更多"按钮，每次加载 2 个月
- 全部展示完毕显示"已展示全部 N 个月的归档"

### 其他
- **外部 URL 检查**：全面扫描确认无国内不可访问的外部引用
- **静态资源缓存**：JS/CSS 从 7 天延长到 365 天（`public, immutable`）
- **数据同步修复**：从正确的线上目录 `/data/web/www/ai-daily/data/` 同步数据

### 变更文件
| 文件 | 操作 | 说明 |
|------|------|------|
| `frontend/src/styles/globals.css` | MODIFY | 移除 noto-sans-sc 导入，改用系统字体栈 |
| `frontend/package.json` | MODIFY | 移除 @fontsource-variable/noto-sans-sc 依赖 |
| `frontend/src/app/HomeClient.tsx` | MODIFY | PC 端全展开 + 骨架屏数量调整 |
| `frontend/src/app/search/SearchClient.tsx` | MODIFY | 热词扩充到 16 个 |
| `frontend/src/app/archive/ArchiveClient.tsx` | MODIFY | 添加分页逻辑（默认 2 个月 + 加载更多） |
| `deploy/nginx-ai-daily.conf` | MODIFY | 添加 Gzip 配置 + 缓存延长 |

---

## [V8] — 2026-03-27 — 前端页面加载优化（折叠 + 懒渲染 + 动画修复）

### 新增
- **`frontend/src/lib/hooks.ts`** — 两个共享自定义 hooks：
  - `useLazyRender<T>()` — IntersectionObserver 懒渲染，元素进入视口时触发一次性渲染，`rootMargin` 可配置
  - `useIsMobile()` — 响应式断点判断（对应 Tailwind sm 640px），使用 `matchMedia` 监听
- **分类折叠**：每个分类组默认只显示前 N 条（移动端 2 条 / 桌面端 3 条），超出部分折叠
  - "展开更多 X 条"按钮 + "收起"按钮
  - `min-height: 44px` 满足 Apple HIG 触摸目标
- **IntersectionObserver 渐进渲染**：视口外分类组先渲染 skeleton 占位，滚动到附近时才渲染真实 DOM
- **CSS 组件类**：
  - `.scroll-fade-in` / `.is-visible` — 滚动触发淡入动画
  - `.expand-btn` — 展开/收起按钮样式
  - `.collapse-fade-mask` — 折叠底部渐变遮罩（`::after` 伪元素）

### 变更
- **`HomeClient.tsx`**：
  - `CategorySection` 重写：添加折叠逻辑 + `useLazyRender` 懒渲染 + skeleton 占位
  - `ArticleCard` 动画修复：从 `animate-slide-up` + 固定 `animationDelay` 改为 `scroll-fade-in` + IntersectionObserver 触发，`transitionDelay` 上限 0.2s
- **`DailyClient.tsx`**：同 HomeClient 同构改造
- **`globals.css`**：新增 3 个 CSS 组件类（scroll-fade-in、expand-btn、collapse-fade-mask）

### 变更文件
| 文件 | 操作 | 说明 |
|------|------|------|
| `frontend/src/lib/hooks.ts` | NEW | useLazyRender + useIsMobile 共享 hooks |
| `frontend/src/app/HomeClient.tsx` | MODIFY | CategorySection 折叠 + 懒渲染 + 动画修复 |
| `frontend/src/app/daily/[date]/DailyClient.tsx` | MODIFY | CategorySection 折叠 + 懒渲染 + 动画修复 |
| `frontend/src/styles/globals.css` | MODIFY | 新增 scroll-fade-in / expand-btn / collapse-fade-mask |

### 效果
- 首屏 DOM 节点：全量渲染 → 仅视口内分类 + 每分类 2-3 条
- 动画延迟：100 条 × 0.03s = 3s → 上限 0.2s
- 零新增第三方依赖（IntersectionObserver 原生 API）

---

## [V7] — 2026-03-27 — 前端兼容修复 + 云服务源扩充

### 新增
- **3 个云服务 AI 博客 RSS 源**
  - `internal/crawler/aws_ml.go` — AWS Machine Learning Blog (Tier 3)
  - `internal/crawler/gcp_ai.go` — Google Cloud AI Blog (Tier 3)
  - `internal/crawler/azure_ai.go` — Azure AI Services Blog (Tier 3)
- **8 条 WebSearch 搜索关键词**（13 → 21 条），覆盖国内外云厂商 AI 平台：
  - 腾讯云/混元、阿里云/通义千问、华为云/盘古/昇腾、百度/文心一言、火山引擎/豆包
  - AWS/Bedrock/SageMaker、Google Cloud/Vertex AI、Azure OpenAI/Copilot Studio
- **归档页分类筛选功能**：从 `index.json` 提取已知分类名，新增可点击分类筛选标签栏

### 修复
- **首页分类样式兼容**：`frontend/src/lib/types.ts` 新增旧 5 分类（国际AI模型、国内AI厂商、产品落地、开源、商业硬件）的 `CATEGORY_COLORS` 和 `CATEGORY_EMOJI` 映射，解决数据迁移前分类标签显示灰色
- **移动端宽度一致性**：统一 `HomeClient.tsx` 和 `DailyClient.tsx` 的 glass-card padding 为 `p-4 sm:p-8`，移除分类筛选栏的 `-mx-4 px-4` 负 margin，改用 `overflow-hidden` 容器消除溢出

### 变更文件
| 文件 | 操作 | 说明 |
|------|------|------|
| `frontend/src/lib/types.ts` | MODIFY | 新增旧 5 分类颜色/emoji 兼容映射 |
| `frontend/src/app/HomeClient.tsx` | MODIFY | glass-card padding 统一 |
| `frontend/src/app/daily/[date]/DailyClient.tsx` | MODIFY | glass-card padding 统一 |
| `frontend/src/app/archive/ArchiveClient.tsx` | MODIFY | 重构添加分类筛选功能 |
| `internal/crawler/aws_ml.go` | NEW | AWS ML Blog RSS 采集器 |
| `internal/crawler/gcp_ai.go` | NEW | Google Cloud AI Blog RSS 采集器 |
| `internal/crawler/azure_ai.go` | NEW | Azure AI Services Blog RSS 采集器 |
| `internal/crawler/websearch.go` | MODIFY | 新增 8 条云服务搜索词 |
| `cmd/crawler/main.go` | MODIFY | 注册 3 个云服务采集器 |

### 统计
- 爬虫总数：41 → 44
- 搜索关键词：13 → 21

---

## [V6] — 2026-03-26 — 8 分类体系重构

### 新增
- **8 分类体系**：模型前沿、产品与应用、深度洞察、云服务与平台、AI工程、AI基础设施、商业与投资、AI安全
- **软性配额采样策略**：两阶段（分类保底 19 条 + 质量补位 11 条 = 30 条精选）
- **`CategoryQuota` 结构体**：定义每个分类的 Min/Preferred 配额
- **4 个 RSS 源**：
  - Chip Huyen (Tier 1) — MLOps/LLMOps 思想领袖
  - Eugene Yan (Tier 1) — Amazon 首席 AI 工程师
  - Karpathy (Tier 1) — 前 OpenAI/Tesla AI 负责人
  - InfoQ AI (Tier 2) — 架构师社区 AI 板块
- **`cat-teal` CSS 类**：对应"云服务与平台"的青绿色调

### 变更
- **`model/article.go`**：旧 7 分类常量替换为 8 分类常量，新增 `CategoryQuotas` 配额表、`CategoryEmoji` 映射、`AllCategories` 有序列表
- **`summarizer/summarizer.go`**：
  - `buildPrompt` 完全重写，新增 8 分类指南、关键边界判定段、源感知分类提示
  - `normalizeCategory` 完全重写，覆盖 8 分类中英文模糊匹配
  - Fallback 默认分类更新为 `CategoryModelFrontier`
- **`aggregator/aggregator.go`**：
  - `diverseSample` 重写为两阶段软性配额采样
  - `rankSources` 新增 4 个来源优先级
  - `buildCategoryGroups` / `generateReportSummary` 默认分类更新
- **`htmlgen/generator.go`**：`categoryColorMap` 替换为 8 分类颜色映射
- **搜索热词更新**：GPT、开源模型、Agent、云服务、RAG、AI芯片、融资、AI安全

### 移除
- 旧 7 分类常量：大模型与基础设施、Agent与应用、研究与论文、开源生态、融资与商业、安全与治理、行业洞察

---

## [V5] — 2026-03-26 — 7 分类体系（从地域分类改为内容本质分类）

### 变更
- 从地域分类（国际AI模型/国内AI厂商/产品落地/开源/商业硬件）改为内容本质分类的 7 分类体系
- DB 全量清洗 + 5 天日报逐天重建

### 新增
- **3 个 RSS 源**：AI Hub Today、BAIR Blog (伯克利)、DailyAI
- **Web Search 查询增强**：8 → 13 条
- **`rankSources` 6 级来源优先级体系**

### 修复
- `preSelect` 逻辑修复：保证新采集的高优先级内容有机会被 LLM 处理

---

## [V4] — 2026-03-25 ~ 03-26 — 采集优化（RSS 并发化 + 新增源 + 增量模式）

### 新增
- **增量模式**：`--incremental --since-hours 6` 参数，只采集最近 N 小时新文章，局部更新 HTML
- **多频次 Cron 调度**：从单次 08:00 改为 3 次/日（08:00 全量 + 14:00/20:00 增量）
- **WebSearch 采集器**：集成 Serper.dev Google Search API，20 组默认搜索查询
- **部署脚本体系**：
  - `deploy/crawl.sh` — 全量采集脚本
  - `deploy/crawl-incremental.sh` — 增量采集脚本
  - `deploy/crontab.example` — Cron 调度配置
  - `deploy/nginx-ai-daily.conf` — Nginx location 配置
- **大量新 RSS 源**：从初始 ~15 个扩展到 40+ 个
  - AI 大厂博客：Anthropic, Microsoft AI, NVIDIA Blog
  - 思想领袖：Jack Clark (Import AI), Simon Willison, Lilian Weng, Latent Space, The Gradient, The Batch
  - 学术：arXiv cs.AI, Papers With Code
  - 更多科技媒体：AI News, IEEE Spectrum, Synced Review 等

### 优化
- **RSS 并发采集**：semaphore 控制并发数 = 5，大幅提升采集速度
- **LLM 并发提升**：`SummarizeBatch` 信号量控制 `MaxConcurrent`，批量处理每批 20 篇
- **Nginx 缓存策略**：静态资源 7 天 + JSON 数据 10 分钟
- 日志自动清理 30 天前

---

## [V3] — 2026-03-25 — 三级聚合流水线

### 新增
- **三级分级处理策略**：preSelect → summarizeCandidates → diverseSample
- **preSelect**：纯本地计算，按来源优先级轮询预选 ~90 篇候选，避免全量 LLM 调用
- **diverseSample**：两阶段均衡采样（来源优先级 + 时间排序）
- **高产量源限制**：`maxArticlesPerSource` — IT之家/36氪最多占总配额 15%
- 日报输出拆分：精选（`FeaturedCategoryGroups`）+ 全量（`CategoryGroups`）
- `FeaturedCount` 字段：首页只展示精选文章

### 优化
- LLM 调用量从全库文章降低到 ~90 篇候选，成本大幅下降

---

## [V2] — 2026-03-25 — LLM 双 Provider + 结构化 Prompt

### 新增
- **Claude Provider**：`callClaude` 方法，支持 `x-api-key` + `anthropic-version: 2023-06-01`
- **结构化 Prompt**：角色设定、分类指南、标签候选池、边界判定规则
- **5 维标签系统**：80+ 候选标签（技术主题/工程与工具/行业与场景/公司与平台/商业与治理）
- **双语标题**：LLM 返回 `original_title` 时拼接 "英文 / 中文" 格式
- **normalizeCategory 模糊匹配**：中英文关键词匹配，确保 LLM 返回值落入有效分类
- **Fallback 机制**：LLM 失败时使用原标题 + 截断摘要 + 默认分类

### 变更
- `summarizeOne` 重构为 Provider 分发（`switch cfg.Provider`）
- LLM 输出解析增加 JSON 容错 + markdown 代码块清理

---

## [V1] — 2026-03-24 ~ 03-25 — 初始构建

### 新增
- **Go 1.22 后端**：基于标准库 net/http，Clean Architecture 分层
- **Next.js 14 SSG 前端**：静态站点生成，部署简单
- **SQLite 数据存储**：`modernc.org/sqlite` 纯 Go 驱动（无 CGO），WAL 模式
- **基础 RSS 采集器**：~15 个来源（The Verge, TechCrunch, VentureBeat, 36氪, 机器之心等）
- **OpenAI API 摘要生成**：基础 Prompt，JSON 格式输出
- **JSON 导出 + 索引更新**：`daily/{date}.json` + `index.json`
- **静态 HTML 生成**：`htmlgen` 模块，Go template 渲染
- **客户端搜索**：Fuse.js 模糊搜索，6 字段加权
- **主题切换**：深色/浅色主题，localStorage 持久化
- **URL 去重**：SHA256 哈希 + SQLite UNIQUE 约束
- **AI 关键词过滤**：177 个关键词二次确认
- **Nginx 静态托管**：`location /ai-daily/` 配置
