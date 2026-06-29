"use client";

import { IndexFile, DailyIndex, CATEGORY_COLORS, CATEGORY_EMOJI } from "@/lib/types";
import { useState, useEffect, useMemo } from "react";

const BASE_PATH = "/ai-daily";

// 卡片内最多展示的标签数量
const MAX_VISIBLE_TAGS = 3;

// 固定 8 分类顺序（与首页一致）
const CATEGORY_ORDER = [
  "模型前沿", "产品与应用", "深度洞察", "云服务与平台",
  "AI工程", "AI基础设施", "商业与投资", "AI安全",
];

// 新 8 分类 + 旧 5 分类兼容
const KNOWN_CATEGORIES = new Set([
  ...CATEGORY_ORDER,
  "国际AI模型", "国内AI厂商", "产品落地", "开源", "商业硬件",
]);

interface ArchiveClientProps {
  index: IndexFile;
}

// TagMore 组件：折叠的标签，hover 展开
function TagMore({ tags }: { tags: string[] }) {
  const [open, setOpen] = useState(false);

  return (
    <span
      className="relative text-[0.6875rem] px-1.5 py-0.5 rounded-full font-medium bg-[var(--bg-primary)] text-[var(--text-secondary)] border border-[var(--border-color)] cursor-default flex-shrink-0"
      onMouseEnter={() => setOpen(true)}
      onMouseLeave={() => setOpen(false)}
    >
      +{tags.length}
      {open && (
        <span className="absolute bottom-full right-0 mb-2 z-20 flex flex-col gap-1 p-2.5 rounded-xl bg-[var(--bg-card)] border border-[var(--border-color)] shadow-lg min-w-[120px]" style={{ boxShadow: '0 8px 24px rgba(0,0,0,0.2)' }}>
          {tags.map((cat) => {
            const colors = CATEGORY_COLORS[cat] || { bg: "bg-gray-500/10", text: "text-gray-400", border: "border-gray-500/20" };
            return (
              <span
                key={cat}
                className={`text-[0.6875rem] px-2 py-0.5 rounded-full font-medium whitespace-nowrap ${colors.bg} ${colors.text} border ${colors.border}`}
              >
                {CATEGORY_EMOJI[cat] || ""} {cat}
              </span>
            );
          })}
        </span>
      )}
    </span>
  );
}

export default function ArchiveClient({ index: ssrIndex }: ArchiveClientProps) {
  const [mounted, setMounted] = useState(false);
  const [index, setIndex] = useState<IndexFile>(ssrIndex);
  const [loading, setLoading] = useState(true);
  // 分页：每次展示的月份数量
  const MONTHS_PER_PAGE = 2;
  const [visibleMonthCount, setVisibleMonthCount] = useState(MONTHS_PER_PAGE);

  useEffect(() => {
    setMounted(true);
    const theme = localStorage.getItem("theme");
    if (theme === "light") {
      document.documentElement.classList.add("light");
    } else {
      document.documentElement.classList.remove("light");
    }

    // 客户端动态加载最新 index.json（覆盖 SSG 数据，确保数据完整）
    async function loadIndex() {
      try {
        const res = await fetch(`${BASE_PATH}/data/index.json`);
        if (!res.ok) throw new Error("fetch index failed");
        const data: IndexFile = await res.json();
        setIndex(data);
      } catch (err) {
        console.error("load index error:", err);
        // 失败时保留 SSG 数据
      } finally {
        setLoading(false);
      }
    }
    loadIndex();
  }, []);

  // 全局最大条数，用于计算进度条比例
  const maxCount = useMemo(
    () => Math.max(...index.days.map((d) => d.total_count), 1),
    [index.days],
  );

  // 从所有日报的 tags 中提取分类名，并按固定顺序排列
  const categoryStats = useMemo(() => {
    const stats: Record<string, number> = {};
    for (const day of index.days) {
      for (const tag of day.tags) {
        if (KNOWN_CATEGORIES.has(tag)) {
          stats[tag] = (stats[tag] || 0) + 1;
        }
      }
    }
    // 按固定 8 分类顺序排列，新分类在前，旧分类在后
    const ordered: [string, number][] = [];
    for (const cat of CATEGORY_ORDER) {
      if (stats[cat]) ordered.push([cat, stats[cat]]);
    }
    // 旧分类（如果有的话）追加在最后
    for (const [cat, count] of Object.entries(stats)) {
      if (!CATEGORY_ORDER.includes(cat)) {
        ordered.push([cat, count]);
      }
    }
    return ordered;
  }, [index.days]);

  // 按月份分组
  const grouped = index.days.reduce<Record<string, DailyIndex[]>>((acc, day) => {
    const month = day.date.substring(0, 7);
    if (!acc[month]) acc[month] = [];
    acc[month].push(day);
    return acc;
  }, {});

  const months = Object.keys(grouped).sort().reverse();

  if (loading && index.days.length === 0) {
    return (
      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-20">
        <div className="text-center animate-fade-in">
          <div className="w-12 h-12 mx-auto mb-4 rounded-full border-4 border-blue-500/30 border-t-blue-500 animate-spin"></div>
          <p className="text-[var(--text-secondary)]">加载中...</p>
        </div>
      </div>
    );
  }

  if (index.days.length === 0) {
    return (
      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-20">
        <div className="text-center animate-fade-in">
          <h2 className="text-2xl font-bold gradient-text mb-3">暂无归档</h2>
          <p className="text-[var(--text-secondary)]">日报归档将在首次采集完成后出现。</p>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-6 sm:py-8">
      <div className={`mb-6 sm:mb-8 animate-fade-in ${mounted ? "" : "opacity-0"}`}>
        <h1 className="text-2xl sm:text-3xl font-bold text-[var(--text-primary)] mb-1 sm:mb-2">历史归档</h1>
        <p className="text-sm text-[var(--text-secondary)]">
          共 {index.days.length} 期日报
        </p>
      </div>

      {/* Category Stats — 纯展示，不可点击 */}
      {categoryStats.length > 0 && (
        <div className={`mb-6 animate-slide-up ${mounted ? "" : "opacity-0"}`} style={{ animationDelay: "0.05s" }}>
          <div className="relative">
            <div className="absolute right-0 top-0 bottom-2 w-8 bg-gradient-to-l from-[var(--bg-primary)] to-transparent z-10 pointer-events-none sm:hidden" />
            <div className="flex gap-2 overflow-x-auto pb-2 scrollbar-thin -mx-4 px-4 sm:mx-0 sm:px-0">
              {categoryStats.map(([cat, count]) => {
                const emoji = CATEGORY_EMOJI[cat] || "";
                const colors = CATEGORY_COLORS[cat] || { bg: "bg-gray-500/10", text: "text-gray-400", border: "border-gray-500/20" };
                return (
                  <span
                    key={cat}
                    className={`inline-flex items-center gap-1 whitespace-nowrap px-2.5 py-1 rounded-full text-xs font-medium ${colors.bg} ${colors.text} border ${colors.border}`}
                  >
                    {emoji && <span>{emoji}</span>}
                    {cat}
                  </span>
                );
              })}
            </div>
          </div>
        </div>
      )}

      <div className="space-y-8 sm:space-y-10">
        {months.slice(0, visibleMonthCount).map((month, mIdx) => (
          <div
            key={month}
            className={`animate-slide-up ${mounted ? "" : "opacity-0"}`}
            style={{ animationDelay: `${0.05 * mIdx}s` }}
          >
            <h2 className="text-base sm:text-lg font-semibold text-[var(--text-primary)] mb-3 sm:mb-4 flex items-center gap-2">
              <span className="w-2 h-2 rounded-full bg-blue-500"></span>
              {month.replace("-", " 年 ") + " 月"}
            </h2>
            <div className="grid gap-3 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3">
              {grouped[month].map((day) => {
                const dayCategories = day.tags.filter((tag) => KNOWN_CATEGORIES.has(tag));
                const visibleTags = dayCategories.slice(0, MAX_VISIBLE_TAGS);
                const hiddenTags = dayCategories.slice(MAX_VISIBLE_TAGS);
                const countPct = Math.round((day.total_count / maxCount) * 100);

                return (
                  <a
                    key={day.date}
                    href={`/ai-daily/daily/${day.date}/`}
                    className="glass-card p-4 transition-all duration-300 group active:scale-[0.98] flex flex-col"
                  >
                    {/* 卡片头部：日期（加粗加大）+ 条数进度条 */}
                    <div className="flex items-center justify-between mb-2.5">
                      <span className="text-[0.9375rem] font-semibold tracking-tight text-[var(--text-primary)]">{day.date}</span>
                      <div className="flex items-center gap-2">
                        <div className="w-9 h-1 rounded-sm bg-[var(--border-color)] overflow-hidden hidden sm:block">
                          <div
                            className="h-full rounded-sm bg-blue-500/80 transition-all duration-500"
                            style={{ width: `${countPct}%` }}
                          />
                        </div>
                        <span className="text-xs font-semibold text-blue-400 bg-blue-500/10 px-2 py-0.5 rounded-full tabular-nums whitespace-nowrap">
                          {day.total_count} 条
                        </span>
                      </div>
                    </div>

                    {/* 摘要：展示 3 行，hover 变色；将摘要中"共收录 N 条"替换为实际 total_count */}
                    <p className="text-sm leading-relaxed text-[var(--text-secondary)] line-clamp-3 group-hover:text-[var(--text-primary)] transition-colors flex-1">
                      {day.summary.replace(/共收录 \d+ 条/, `共收录 ${day.total_count} 条`)}
                    </p>

                    {/* 标签区域：Top3 + 折叠 */}
                    {dayCategories.length > 0 && (
                      <div className="flex items-center gap-1.5 mt-3 flex-nowrap overflow-hidden">
                        {visibleTags.map((cat) => {
                          const colors = CATEGORY_COLORS[cat] || { bg: "bg-gray-500/10", text: "text-gray-400", border: "border-gray-500/20" };
                          return (
                            <span
                              key={cat}
                              className={`text-[0.6875rem] px-2 py-0.5 rounded-full font-medium whitespace-nowrap flex-shrink-0 ${colors.bg} ${colors.text} border ${colors.border}`}
                            >
                              {CATEGORY_EMOJI[cat] || ""} {cat}
                            </span>
                          );
                        })}
                        {hiddenTags.length > 0 && (
                          <TagMore tags={hiddenTags} />
                        )}
                      </div>
                    )}
                  </a>
                );
              })}
            </div>
          </div>
        ))}
      </div>

      {/* 加载更多按钮 */}
      {visibleMonthCount < months.length && (
        <div className="flex justify-center mt-8 sm:mt-10">
          <button
            onClick={() => setVisibleMonthCount((prev) => Math.min(prev + MONTHS_PER_PAGE, months.length))}
            className="inline-flex items-center gap-2 px-6 py-3 rounded-xl text-sm font-medium text-[var(--text-secondary)] hover:text-blue-400 bg-[var(--bg-card)] border border-[var(--border-color)] hover:border-blue-500/30 transition-all duration-200 cursor-pointer"
          >
            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 8.25l-7.5 7.5-7.5-7.5" />
            </svg>
            加载更多（还有 {months.length - visibleMonthCount} 个月）
          </button>
        </div>
      )}

      {visibleMonthCount >= months.length && months.length > MONTHS_PER_PAGE && (
        <div className="text-center mt-8 text-sm text-[var(--text-secondary)] opacity-60">
          已展示全部 {months.length} 个月的归档
        </div>
      )}
    </div>
  );
}
