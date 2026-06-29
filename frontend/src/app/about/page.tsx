export default function AboutPage() {
  return (
    <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-16">
      <div className="animate-fade-in">
        <h1 className="text-3xl font-bold gradient-text mb-8">关于 AI Daily</h1>

        <div className="glass-card p-6 sm:p-8 space-y-6">
          <section>
            <h2 className="text-xl font-semibold text-[var(--text-primary)] mb-3">项目介绍</h2>
            <p className="text-[var(--text-secondary)] leading-relaxed">
              AI Daily 是一个全自动化的 AI 资讯聚合平台。系统每天从全球 <strong className="text-[var(--text-primary)]">45+ 个新闻源</strong>自动采集最新 AI 资讯，
              通过大语言模型（LLM）智能生成中文摘要与分类标签，经三级聚合流水线精选 <strong className="text-[var(--text-primary)]">30 条</strong>高质量日报，
              覆盖 <strong className="text-[var(--text-primary)]">8 大分类</strong>，部署为静态站点。
            </p>
          </section>

          <section>
            <h2 className="text-xl font-semibold text-[var(--text-primary)] mb-3">数据来源（45+ 新闻源）</h2>
            <div className="space-y-3 text-[var(--text-secondary)]">
              {[
                { group: "国际科技媒体（7 个）", sources: "The Verge, VentureBeat, TechCrunch, Ars Technica, Wired AI, MIT Technology Review, THE DECODER" },
                { group: "AI 专业媒体 & 聚合（8 个）", sources: "Unite.AI, AI News, MarkTechPost, Synced Review, IEEE Spectrum, AI Hub Today, DailyAI, InfoQ AI" },
                { group: "AI 大厂官方博客（8 个）", sources: "OpenAI, DeepMind, Google Research, Meta Engineering, HuggingFace, Anthropic, Microsoft AI, NVIDIA" },
                { group: "云服务 AI 博客（3 个）", sources: "AWS ML Blog, Google Cloud AI Blog, Azure AI Services Blog" },
                { group: "AI 前沿思想领袖（10 个）", sources: "Import AI, Simon Willison, Lilian Weng, Latent Space, The Gradient, The Batch, BAIR Blog, Chip Huyen, Eugene Yan, Karpathy" },
                { group: "AI 学术 & 论文（2 个）", sources: "arXiv cs.AI, Papers With Code" },
                { group: "创投 & 科技商业（2 个）", sources: "Hacker News, PanDaily" },
                { group: "国内科技媒体（4 个）", sources: "36氪, 机器之心, IT之家, 智东西" },
                { group: "Web 搜索补充（1 个）", sources: "Serper.dev — 20 组搜索查询覆盖突发新闻" },
              ].map((item) => (
                <div key={item.group}>
                  <p className="text-sm font-medium text-[var(--text-primary)] mb-0.5">{item.group}</p>
                  <p className="text-xs text-[var(--text-secondary)] leading-relaxed">{item.sources}</p>
                </div>
              ))}
            </div>
          </section>

          <section>
            <h2 className="text-xl font-semibold text-[var(--text-primary)] mb-3">技术栈</h2>
            <div className="grid sm:grid-cols-2 gap-3">
              {[
                { label: "后端采集", value: "Go 1.22 + RSS/WebSearch" },
                { label: "数据存储", value: "SQLite WAL (modernc.org/sqlite, 无 CGO)" },
                { label: "AI 摘要", value: "OpenAI / Claude 双 Provider" },
                { label: "聚合策略", value: "三级流水线 + 两阶段配额采样" },
                { label: "前端站点", value: "Next.js 14 SSG + Tailwind CSS" },
                { label: "搜索", value: "Fuse.js 客户端模糊搜索" },
                { label: "部署", value: "静态站点 + Nginx, Cron 3 次/日" },
                { label: "交叉编译", value: "CGO_ENABLED=0, 纯 Go 无外部依赖" },
              ].map((item) => (
                <div key={item.label} className="flex items-start gap-3 p-3 rounded-lg bg-white/5">
                  <span className="text-sm font-medium text-[var(--text-primary)] w-20 flex-shrink-0">{item.label}</span>
                  <span className="text-sm text-[var(--text-secondary)]">{item.value}</span>
                </div>
              ))}
            </div>
          </section>

          <section>
            <h2 className="text-xl font-semibold text-[var(--text-primary)] mb-3">智能处理</h2>
            <div className="space-y-2 text-sm text-[var(--text-secondary)]">
              <p>• <strong className="text-[var(--text-primary)]">8 类分类体系</strong> — 按内容本质分类（非地域），含结构化 Prompt 边界判定规则和源感知分类</p>
              <p>• <strong className="text-[var(--text-primary)]">5 维标签系统</strong> — 技术主题、工程与工具、行业与场景、公司与平台、商业与治理</p>
              <p>• <strong className="text-[var(--text-primary)]">6 级来源优先级</strong> — 从 AI 大厂官方 + 思想领袖（Tier 1）到高产量国内源（Tier 6，限额 15%）</p>
              <p>• <strong className="text-[var(--text-primary)]">177 个 AI 关键词过滤</strong> — 二次确认采集内容与 AI 相关</p>
              <p>• <strong className="text-[var(--text-primary)]">双语标题</strong> — 自动拼接"英文原标题 / 中文标题"格式</p>
            </div>
          </section>

          <section>
            <h2 className="text-xl font-semibold text-[var(--text-primary)] mb-3">自动化流程</h2>
            <div className="flex items-center gap-2 flex-wrap text-sm text-[var(--text-secondary)]">
              {["45+ 源采集", "→", "AI 关键词过滤", "→", "URL 去重", "→", "三级聚合", "→", "LLM 摘要", "→", "配额采样 30 篇", "→", "JSON 导出", "→", "静态发布"].map((step, i) => (
                <span
                  key={i}
                  className={`${
                    step === "→" ? "text-blue-500" : "px-3 py-1 rounded-full bg-blue-500/10 text-blue-400 border border-blue-500/20"
                  }`}
                >
                  {step}
                </span>
              ))}
            </div>
          </section>

          <section className="pt-4 border-t border-[var(--border-color)]">
            <p className="text-sm text-[var(--text-secondary)]">
              本站内容为 AI 自动生成摘要，版权归原作者所有。如有任何问题，请通过 javeyim#gmail.com 反馈。
            </p>
          </section>
        </div>
      </div>
    </div>
  );
}
