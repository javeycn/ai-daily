"use client";

import { DailyReport, Article, CategoryGroup, IndexFile, CATEGORY_COLORS, CATEGORY_EMOJI } from "@/lib/types";
import { useState, useEffect } from "react";
import { useLazyRender, useIsMobile } from "@/lib/hooks";

const BASE_PATH = "/ai-daily";

interface HomeClientProps {
  report: DailyReport | null;
}

export default function HomeClient({ report: initialReport }: HomeClientProps) {
  const [report, setReport] = useState<DailyReport | null>(initialReport);
  const [activeCategory, setActiveCategory] = useState("全部");
  const [mounted, setMounted] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setMounted(true);
    const theme = localStorage.getItem("theme");
    if (theme === "light") {
      document.documentElement.classList.add("light");
    } else {
      document.documentElement.classList.remove("light");
    }

    // 动态加载最新日报
    async function loadLatest() {
      try {
        const idxRes = await fetch(`${BASE_PATH}/data/index.json`);
        if (!idxRes.ok) throw new Error("fetch index failed");
        const idx: IndexFile = await idxRes.json();
        if (idx.days.length === 0) {
          setLoading(false);
          return;
        }
        const latestDate = idx.days[0].date;
        const reportRes = await fetch(`${BASE_PATH}/data/daily/${latestDate}.json`);
        if (!reportRes.ok) throw new Error("fetch report failed");
        const data: DailyReport = await reportRes.json();
        setReport(data);
      } catch (err) {
        console.error("load latest report error:", err);
      } finally {
        setLoading(false);
      }
    }
    loadLatest();
  }, []);

  if (loading && !report) {
    return (
      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-20">
        <div className="text-center animate-fade-in">
          <div className="w-12 h-12 mx-auto mb-4 rounded-full border-4 border-blue-500/30 border-t-blue-500 animate-spin"></div>
          <p className="text-[var(--text-secondary)]">加载中...</p>
        </div>
      </div>
    );
  }

  if (!report) {
    return (
      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-20">
        <div className="text-center animate-fade-in">
          <div className="w-16 h-16 mx-auto mb-6 rounded-2xl bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center shadow-xl shadow-blue-500/20">
            <svg className="w-8 h-8 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <h2 className="text-2xl font-bold gradient-text mb-3">数据尚未就绪</h2>
          <p className="text-[var(--text-secondary)] max-w-md mx-auto">
            AI 采集服务正在运行中，每日日报将在首次采集完成后自动发布。
          </p>
        </div>
      </div>
    );
  }

  // 首页只展示精选文章（featured_count 篇）
  const featuredCount = report.featured_count || report.total_count;

  // 优先使用后端提供的 featured_category_groups（精选分组），没有则 fallback 过滤
  const categoryGroups = (() => {
    if (report.featured_category_groups && report.featured_category_groups.length > 0) {
      return report.featured_category_groups;
    }
    // fallback: 从 articles 前 N 篇 ID 过滤 category_groups
    const allCategoryGroups = report.category_groups || [];
    const featuredArticleIds = new Set(
      (report.articles || []).slice(0, featuredCount).map((a) => a.id)
    );
    return allCategoryGroups
      .map((g) => ({
        ...g,
        articles: g.articles.filter((a) => featuredArticleIds.has(a.id)),
      }))
      .filter((g) => g.articles.length > 0);
  })();

  const categoryNames = ["全部", ...categoryGroups.map((g) => g.category)];

  const displayGroups =
    activeCategory === "全部"
      ? categoryGroups
      : categoryGroups.filter((g) => g.category === activeCategory);

  return (
    <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      {/* Daily Header */}
      <div className={`mb-8 animate-fade-in ${mounted ? "" : "opacity-0"}`}>
        <div className="glass-card p-4 sm:p-8">
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-4">
            <div>
              <p className="text-sm text-[var(--text-secondary)] mb-1">{report.date}</p>
              <h1 className="text-2xl sm:text-3xl font-bold text-[var(--text-primary)]">
                {report.title}
              </h1>
            </div>
            <div className="flex items-center gap-2">
              <span className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full bg-blue-500/10 text-blue-400 text-sm font-medium border border-blue-500/20">
                <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z" />
                </svg>
                {featuredCount} 条精选
              </span>
            </div>
          </div>

          <p className="text-[var(--text-secondary)] text-sm leading-relaxed">{report.summary.replace(/共收录 \d+ 条/, `共收录 ${report.total_count} 条`)}</p>
        </div>
      </div>

      {/* Category Filter */}
      <div className={`mb-6 animate-slide-up ${mounted ? "" : "opacity-0"}`} style={{ animationDelay: "0.1s" }}>
        <div className="relative overflow-hidden">
          {/* 右侧渐变遮罩提示可滚动 */}
          <div className="absolute right-0 top-0 bottom-2 w-8 bg-gradient-to-l from-[var(--bg-primary)] to-transparent z-10 pointer-events-none sm:hidden" />
          <div className="flex gap-2 overflow-x-auto pb-2 scrollbar-thin">
            {categoryNames.map((cat) => {
              const emoji = CATEGORY_EMOJI[cat] || "";
              return (
                <button
                  key={cat}
                  onClick={() => setActiveCategory(cat)}
                  className={`tag-badge whitespace-nowrap cursor-pointer ${
                    activeCategory === cat ? "tag-badge-active" : ""
                  }`}
                >
                  {emoji && <span className="mr-1">{emoji}</span>}
                  {cat}
                  {cat !== "全部" && (
                    <span className="ml-1.5 text-[10px] opacity-60">
                      {categoryGroups.find((g) => g.category === cat)?.articles.length || 0}
                    </span>
                  )}
                </button>
              );
            })}
          </div>
        </div>
      </div>

      {/* Articles by Category */}
      <div className="space-y-8">
        {displayGroups.map((group) => (
          <CategorySection key={group.category} group={group} mounted={mounted} />
        ))}
      </div>

      {displayGroups.length === 0 && (
        <div className="text-center py-16 text-[var(--text-secondary)]">
          <p>该分类下暂无资讯</p>
        </div>
      )}
    </div>
  );
}

function CategorySection({ group, mounted }: { group: CategoryGroup; mounted: boolean }) {
  const colors = CATEGORY_COLORS[group.category] || { bg: "bg-gray-500/10", text: "text-gray-400", border: "border-gray-500/20" };
  const isMobile = useIsMobile();
  const defaultVisible = isMobile ? 2 : group.articles.length;
  const [expanded, setExpanded] = useState(!isMobile);
  const [sectionRef, isInView] = useLazyRender<HTMLDivElement>(isMobile ? "100px" : "200px");

  const visibleArticles = expanded
    ? group.articles
    : group.articles.slice(0, defaultVisible);
  const hiddenCount = group.articles.length - defaultVisible;
  const showExpandBtn = !expanded && hiddenCount > 0;

  return (
    <div ref={sectionRef} className={`scroll-fade-in ${isInView ? "is-visible" : ""}`}>
      {isInView ? (
        <>
          <div className="flex items-center gap-3 mb-4">
            <span className="text-2xl">{group.emoji}</span>
            <h2 className="text-xl font-bold text-[var(--text-primary)]">{group.category}</h2>
            <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${colors.bg} ${colors.text} border ${colors.border}`}>
              {group.articles.length} 条
            </span>
            <div className="flex-1 h-px bg-gradient-to-r from-white/10 to-transparent" />
          </div>

          <div className={`grid gap-4 ${showExpandBtn ? "collapse-fade-mask" : ""}`}>
            {visibleArticles.map((article, index) => (
              <ArticleCard key={article.id} article={article} index={index} mounted={mounted} />
            ))}
          </div>

          {showExpandBtn && (
            <button
              onClick={() => setExpanded(true)}
              className="expand-btn mt-2"
              aria-label={`展开${group.category}分类下剩余 ${hiddenCount} 条`}
            >
              <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 8.25l-7.5 7.5-7.5-7.5" />
              </svg>
              展开更多 {hiddenCount} 条
            </button>
          )}

          {expanded && hiddenCount > 0 && (
            <button
              onClick={() => setExpanded(false)}
              className="expand-btn mt-2"
              aria-label="收起"
            >
              <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M4.5 15.75l7.5-7.5 7.5 7.5" />
              </svg>
              收起
            </button>
          )}
        </>
      ) : (
        /* Skeleton 占位 — 视口外分类组不渲染真实 DOM */
        <div className="space-y-4">
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 rounded-lg bg-white/5 animate-pulse" />
            <div className="h-5 w-24 rounded bg-white/5 animate-pulse" />
          </div>
          {Array.from({ length: isMobile ? 2 : 4 }).map((_, i) => (
            <div key={i} className="glass-card p-4 sm:p-5">
              <div className="h-4 w-3/4 rounded bg-white/5 animate-pulse mb-3" />
              <div className="h-3 w-full rounded bg-white/5 animate-pulse mb-2" />
              <div className="h-3 w-2/3 rounded bg-white/5 animate-pulse" />
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

function ArticleCard({
  article,
  index,
  mounted,
}: {
  article: Article;
  index: number;
  mounted: boolean;
}) {
  const [cardRef, isCardVisible] = useLazyRender<HTMLAnchorElement>("50px");
  const titleParts = article.chinese_title ? article.chinese_title.split(" / ") : [];
  const hasEnglish = titleParts.length >= 2;
  const displayTitle = hasEnglish ? titleParts[titleParts.length - 1] : (article.chinese_title || article.original_title);
  const englishTitle = hasEnglish ? titleParts[0] : (article.original_title !== article.chinese_title ? article.original_title : "");

  const colors = CATEGORY_COLORS[article.category] || { bg: "bg-gray-500/10", text: "text-gray-400", border: "border-gray-500/20" };

  return (
    <a
      ref={cardRef}
      href={article.url}
      target="_blank"
      rel="noopener noreferrer"
      className={`glass-card p-4 sm:p-5 flex gap-4 sm:gap-5 transition-all duration-300 scroll-fade-in block group ${
        isCardVisible ? "is-visible" : ""
      }`}
      style={{ transitionDelay: `${Math.min(index * 0.05, 0.2)}s` }}
    >
      {/* Image */}
      <div className="hidden sm:block flex-shrink-0 w-28 h-24 sm:w-32 sm:h-28 rounded-xl overflow-hidden bg-dark-800 relative">
        {article.image_url ? (
          <>
            <img
              src={article.image_url}
              alt={displayTitle}
              className="w-full h-full object-cover"
              loading="lazy"
              referrerPolicy="no-referrer"
              onError={(e) => {
                const target = e.currentTarget;
                target.style.display = "none";
                const fallback = target.nextElementSibling as HTMLElement;
                if (fallback) fallback.style.display = "flex";
              }}
            />
            <div className="w-full h-full items-center justify-center bg-gradient-to-br from-blue-500/10 to-purple-500/10 absolute inset-0" style={{ display: "none" }}>
              <svg className="w-8 h-8 text-blue-500/30" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M2.25 15.75l5.159-5.159a2.25 2.25 0 013.182 0l5.159 5.159m-1.5-1.5l1.409-1.41a2.25 2.25 0 013.182 0l2.909 2.91m-18 3.75h16.5a1.5 1.5 0 001.5-1.5V6a1.5 1.5 0 00-1.5-1.5H3.75A1.5 1.5 0 002.25 6v12a1.5 1.5 0 001.5 1.5zm10.5-11.25h.008v.008h-.008V8.25zm.375 0a.375.375 0 11-.75 0 .375.375 0 01.75 0z" />
              </svg>
            </div>
          </>
        ) : (
          <div className="w-full h-full flex items-center justify-center bg-gradient-to-br from-blue-500/10 to-purple-500/10">
            <svg className="w-8 h-8 text-blue-500/30" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M2.25 15.75l5.159-5.159a2.25 2.25 0 013.182 0l5.159 5.159m-1.5-1.5l1.409-1.41a2.25 2.25 0 013.182 0l2.909 2.91m-18 3.75h16.5a1.5 1.5 0 001.5-1.5V6a1.5 1.5 0 00-1.5-1.5H3.75A1.5 1.5 0 002.25 6v12a1.5 1.5 0 001.5 1.5zm10.5-11.25h.008v.008h-.008V8.25zm.375 0a.375.375 0 11-.75 0 .375.375 0 01.75 0z" />
            </svg>
          </div>
        )}
      </div>

      {/* Content */}
      <div className="flex-1 min-w-0">
        <div className="flex items-start justify-between gap-2 sm:gap-3 mb-1">
          <div className="min-w-0">
            <h3 className="text-sm sm:text-base font-semibold text-[var(--text-primary)] leading-snug line-clamp-2 sm:line-clamp-none">
              {displayTitle}
            </h3>
            {englishTitle && (
              <p className="text-xs sm:text-sm text-[var(--text-secondary)] opacity-60 mt-0.5 line-clamp-1 hidden sm:block">
                {englishTitle}
              </p>
            )}
          </div>
          <svg className="w-4 h-4 flex-shrink-0 text-[var(--text-secondary)] mt-0.5 opacity-40 sm:opacity-0 sm:group-hover:opacity-100 transition-opacity" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M13.5 6H5.25A2.25 2.25 0 003 8.25v10.5A2.25 2.25 0 005.25 21h10.5A2.25 2.25 0 0018 18.75V10.5m-10.5 6L21 3m0 0h-5.25M21 3v5.25" />
          </svg>
        </div>

        {article.summary && (
          <p className="text-xs sm:text-sm text-[var(--text-secondary)] leading-relaxed mb-2 sm:mb-3 line-clamp-3 sm:line-clamp-5">
            {article.summary}
          </p>
        )}

        <div className="flex items-center gap-2 sm:gap-3 flex-wrap">
          <span className="text-[11px] sm:text-xs text-[var(--text-secondary)] flex items-center gap-1">
            <span className="inline-block w-1.5 h-1.5 rounded-full bg-green-400"></span>
            {article.source}
          </span>
          {article.category && (
            <span className={`inline-block px-2 py-0.5 rounded-full text-[10px] sm:text-[11px] font-medium ${colors.bg} ${colors.text} border ${colors.border}`}>
              {CATEGORY_EMOJI[article.category] || ""} {article.category}
            </span>
          )}
          {article.tags && (
            <span className="text-xs text-[var(--text-secondary)] hidden sm:inline">
              {article.tags.split(",").map((tag) => tag.trim()).filter(Boolean).slice(0, 2).map((tag) => (
                <span key={tag} className="inline-block px-2 py-0.5 rounded-full bg-white/5 text-[11px] mr-1">
                  {tag}
                </span>
              ))}
            </span>
          )}
        </div>
      </div>
    </a>
  );
}
