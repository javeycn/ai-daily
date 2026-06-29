"use client";

import { useEffect, useState } from "react";

export default function LayoutClient({ children }: { children: React.ReactNode }) {
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
    const theme = localStorage.getItem("theme");
    if (theme === "light") {
      document.documentElement.classList.remove("dark");
      document.documentElement.classList.add("light");
    }
  }, []);

  return (
    <>
      <Header />
      <main className="flex-1">{children}</main>
      <Footer />
    </>
  );
}

function Header() {
  return (
    <header className="sticky top-0 z-50 border-b border-[var(--border-color)] bg-[var(--bg-primary)]/80 backdrop-blur-xl">
      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">
          {/* Logo */}
          <a href="/ai-daily/" className="flex items-center gap-3 group">
            <div className="w-9 h-9 rounded-xl bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center shadow-lg shadow-blue-500/20 group-hover:shadow-blue-500/40 transition-shadow">
              <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
            </div>
            <div>
              <h1 className="text-lg font-bold gradient-text">AI Daily</h1>
              <p className="text-[10px] text-[var(--text-secondary)] -mt-0.5">每日 AI 资讯</p>
            </div>
          </a>

          {/* Navigation */}
          <nav className="flex items-center gap-1 sm:gap-2">
            <a
              href="/ai-daily/archive/"
              className="flex items-center justify-center px-3 py-2 sm:px-4 sm:py-2 text-sm text-[var(--text-secondary)] hover:text-[var(--text-primary)] rounded-lg hover:bg-white/5 transition-all min-w-[44px] min-h-[44px] sm:min-w-0 sm:min-h-0"
            >
              <span className="hidden sm:inline">历史归档</span>
              <svg className="w-5 h-5 sm:hidden" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </a>
            <a
              href="/ai-daily/search/"
              className="flex items-center justify-center px-3 py-2 sm:px-4 sm:py-2 text-sm text-[var(--text-secondary)] hover:text-[var(--text-primary)] rounded-lg hover:bg-white/5 transition-all min-w-[44px] min-h-[44px] sm:min-w-0 sm:min-h-0"
            >
              <span className="hidden sm:inline">搜索</span>
              <svg className="w-5 h-5 sm:hidden" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.607 10.607z" />
              </svg>
            </a>
            <ThemeToggle />
          </nav>
        </div>
      </div>
    </header>
  );
}

function ThemeToggle() {
  const handleToggle = () => {
    document.documentElement.classList.toggle("light");
    document.documentElement.classList.toggle("dark");
    const isLight = document.documentElement.classList.contains("light");
    localStorage.setItem("theme", isLight ? "light" : "dark");
  };

  return (
    <button
      className="flex items-center justify-center p-2 rounded-lg text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-white/5 transition-all min-w-[44px] min-h-[44px] sm:min-w-0 sm:min-h-0"
      aria-label="切换主题"
      onClick={handleToggle}
    >
      <svg className="w-5 h-5 dark:hidden" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M21.752 15.002A9.718 9.718 0 0118 15.75c-5.385 0-9.75-4.365-9.75-9.75 0-1.33.266-2.597.748-3.752A9.753 9.753 0 003 11.25C3 16.635 7.365 21 12.75 21a9.753 9.753 0 009.002-5.998z" />
      </svg>
      <svg className="w-5 h-5 hidden dark:block" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M12 3v2.25m6.364.386l-1.591 1.591M21 12h-2.25m-.386 6.364l-1.591-1.591M12 18.75V21m-4.773-4.227l-1.591 1.591M5.25 12H3m4.227-4.773L5.636 5.636M15.75 12a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0z" />
      </svg>
    </button>
  );
}

function Footer() {
  return (
    <footer className="border-t border-[var(--border-color)] bg-[var(--bg-primary)]/50 mt-8 sm:mt-16">
      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-6 sm:py-8">
        <div className="flex flex-col sm:flex-row items-center justify-between gap-4">
          <div className="flex items-center gap-2 text-sm text-[var(--text-secondary)]">
            <div className="w-6 h-6 rounded-lg bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center">
              <svg className="w-3.5 h-3.5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
            </div>
            <span>&copy; {new Date().getFullYear()} AI Daily. Powered by AI&amp;LLM, JAVEY.org</span>
          </div>
          <div className="flex items-center gap-4 text-sm text-[var(--text-secondary)]">
            <a href="/ai-daily/about/" className="hover:text-[var(--text-primary)] transition-colors">关于</a>
            <span className="text-[var(--border-color)]">|</span>
            <span className="flex items-center gap-1">
              <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
                <path d="M6.18 15.64a2.18 2.18 0 0 1 2.18 2.18C8.36 19 7.38 20 6.18 20C5 20 4 19 4 17.82a2.18 2.18 0 0 1 2.18-2.18M4 4.44A15.56 15.56 0 0 1 19.56 20h-2.83A12.73 12.73 0 0 0 4 7.27V4.44m0 5.66a9.9 9.9 0 0 1 9.9 9.9h-2.83A7.07 7.07 0 0 0 4 12.93V10.1z" />
              </svg>
              RSS
            </span>
            <span className="text-[var(--border-color)]">|</span>
            <a
              href="https://github.com"
              target="_blank"
              rel="noopener noreferrer"
              className="hover:text-[var(--text-primary)] transition-colors"
            >
              <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
                <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z" />
              </svg>
            </a>
          </div>
        </div>
      </div>
    </footer>
  );
}
