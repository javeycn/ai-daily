"use client";

import { useEffect, useState, useMemo, useRef, useCallback } from "react";
import Fuse from "fuse.js";

interface SearchIndexItem {
  id: string;
  t: string;    // chinese_title
  s: string;    // summary (truncated)
  tg: string;   // tags
  c: string;    // category
  src: string;  // source
  d: string;    // date
}

interface SearchManifest {
  version: string;
  file: string;
  count: number;
}

interface SearchPayload {
  items: SearchIndexItem[];
  fuseIndex: unknown; // Fuse 序列化的索引
}

const FUSE_KEYS = [
  { name: "t", weight: 2 },
  { name: "s", weight: 1 },
  { name: "tg", weight: 1.5 },
  { name: "c", weight: 1.5 },
  { name: "src", weight: 0.5 },
];

export default function SearchClient() {
  const [mounted, setMounted] = useState(false);
  const [query, setQuery] = useState("");
  const [fuse, setFuse] = useState<Fuse<SearchIndexItem> | null>(null);
  const [itemCount, setItemCount] = useState(0);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  // 异步加载搜索索引（manifest -> 数据文件）
  useEffect(() => {
    setMounted(true);
    let cancelled = false;

    const load = async () => {
      try {
        // Step 1: 加载 manifest，获取最新数据文件名
        const manifestRes = await fetch("/ai-daily/search-manifest.json", {
          cache: "no-cache",
        });
        if (!manifestRes.ok) throw new Error(`manifest HTTP ${manifestRes.status}`);
        const manifest: SearchManifest = await manifestRes.json();

        // Step 2: 加载实际数据文件（带 hash 的文件名可以长缓存）
        const dataRes = await fetch(`/ai-daily/${manifest.file}`);
        if (!dataRes.ok) throw new Error(`data HTTP ${dataRes.status}`);
        const payload: SearchPayload = await dataRes.json();

        if (cancelled) return;

        // Step 3: 用预构建的索引初始化 Fuse（避开运行时建索引的开销）
        // 让出主线程后再初始化，UI 先响应
        await new Promise((resolve) => setTimeout(resolve, 0));

        const parsedIndex = Fuse.parseIndex<SearchIndexItem>(payload.fuseIndex as never);
        const f = new Fuse(payload.items, {
          keys: FUSE_KEYS,
          threshold: 0.4,
          includeScore: true,
        }, parsedIndex);

        if (cancelled) return;
        setFuse(f);
        setItemCount(payload.items.length);
        setLoading(false);
      } catch (err) {
        if (cancelled) return;
        console.error("Search index load failed:", err);
        setLoading(false);
        setLoadError(true);
      }
    };

    load();
    return () => {
      cancelled = true;
    };
  }, []);

  const results = useMemo(() => {
    if (!query.trim() || !fuse) return [];
    return fuse.search(query).slice(0, 50);
  }, [query, fuse]);

  const handleKeywordClick = useCallback((kw: string) => {
    setQuery(kw);
    inputRef.current?.focus();
  }, []);

  const hotKeywords = [
    "GPT", "大模型", "开源", "OpenAI", "Agent", "多模态", "Claude", "推理优化",
    "云计算", "AI安全", "视频生成", "Anthropic", "融资", "AI编程", "硬件", "产品落地",
  ];

  const ready = !loading && !loadError && fuse !== null;

  return (
    <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      {/* Search Header */}
      <div className={`mb-8 animate-fade-in ${mounted ? "" : "opacity-0"}`}>
        <h1 className="text-3xl font-bold text-[var(--text-primary)] mb-6">搜索资讯</h1>

        <div className="relative">
          <svg
            className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-[var(--text-secondary)]"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            strokeWidth={2}
          >
            <path strokeLinecap="round" strokeLinejoin="round" d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.607 10.607z" />
          </svg>
          <input
            ref={inputRef}
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder={loading ? "加载搜索索引中..." : "搜索 AI 资讯标题、摘要、标签..."}
            disabled={!ready}
            className="w-full pl-12 pr-4 py-4 rounded-2xl glass-card border-[var(--border-color)] bg-[var(--bg-card)] text-[var(--text-primary)] placeholder:text-[var(--text-secondary)] focus:outline-none focus:border-blue-500/50 focus:ring-2 focus:ring-blue-500/20 transition-all text-lg disabled:opacity-60"
          />
        </div>

        {/* Loading indicator */}
        {loading && (
          <div className="flex items-center gap-2 mt-4 text-sm text-[var(--text-secondary)]">
            <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
            </svg>
            加载搜索索引...
          </div>
        )}

        {/* Error indicator */}
        {loadError && (
          <div className="flex items-center gap-2 mt-4 text-sm text-red-400">
            搜索索引加载失败，请刷新页面重试
          </div>
        )}

        {/* Hot Keywords */}
        {!query && ready && (
          <div className="flex gap-2 mt-4 flex-wrap">
            <span className="text-sm text-[var(--text-secondary)]">热门：</span>
            {hotKeywords.map((kw) => (
              <button
                key={kw}
                onClick={() => handleKeywordClick(kw)}
                className="tag-badge cursor-pointer"
              >
                {kw}
              </button>
            ))}
          </div>
        )}

        {/* Index info */}
        {ready && itemCount > 0 && !query && (
          <p className="mt-3 text-xs text-[var(--text-secondary)] opacity-60">
            索引包含最近 30 天共 {itemCount} 条资讯
          </p>
        )}
      </div>

      {/* Results */}
      {query.trim() && ready && (
        <div className="mb-4 text-sm text-[var(--text-secondary)]">
          {results.length > 0
            ? `找到 ${results.length} 条结果`
            : "未找到匹配结果"}
        </div>
      )}

      <div className="grid gap-4">
        {results.map((result, index) => {
          const item = result.item;
          return (
            <a
              key={`${item.id}-${item.d}`}
              href={`/ai-daily/daily/${item.d}/`}
              className={`glass-card p-5 flex gap-4 transition-all duration-300 animate-slide-up block ${
                mounted ? "" : "opacity-0"
              }`}
              style={{ animationDelay: `${0.03 * index}s` }}
            >
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 mb-1.5">
                  <h3 className="text-base font-semibold text-[var(--text-primary)] line-clamp-1">
                    {item.t}
                  </h3>
                </div>
                <p className="text-sm text-[var(--text-secondary)] line-clamp-2 mb-2">{item.s}</p>
                <div className="flex items-center gap-3">
                  <span className="text-xs text-blue-400">{item.d}</span>
                  <span className="text-xs text-[var(--text-secondary)]">{item.src}</span>
                  {item.tg && (
                    <span>
                      {item.tg.split(",").slice(0, 2).map((tag) => (
                        <span
                          key={tag}
                          className="inline-block text-[10px] px-2 py-0.5 rounded-full bg-white/5 text-[var(--text-secondary)] mr-1"
                        >
                          {tag.trim()}
                        </span>
                      ))}
                    </span>
                  )}
                </div>
              </div>
            </a>
          );
        })}
      </div>

      {/* Empty State */}
      {query.trim() && results.length === 0 && ready && (
        <div className="text-center py-20 animate-fade-in">
          <svg className="w-16 h-16 mx-auto mb-4 text-[var(--text-secondary)] opacity-30" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.607 10.607z" />
          </svg>
          <p className="text-[var(--text-secondary)]">没有找到与 &ldquo;{query}&rdquo; 相关的资讯</p>
        </div>
      )}
    </div>
  );
}
