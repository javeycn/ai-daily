import type { Metadata } from "next";
import "../styles/globals.css";
import LayoutClient from "./LayoutClient";

export const metadata: Metadata = {
  title: "AI Daily - 每日 AI 资讯",
  description: "每天自动采集全球 AI 新闻，通过大模型智能摘要，汇聚成每日精选 AI 日报。",
  keywords: ["AI", "人工智能", "新闻", "资讯", "大模型", "LLM", "日报"],
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="zh-CN" className="dark" suppressHydrationWarning>
      <head>
        {/* 预连接字体 CDN — 提前完成 DNS + TLS，加速 Inter 字体加载 */}
        <link rel="preconnect" href="https://cdn.jsdelivr.net" crossOrigin="anonymous" />
        <link rel="dns-prefetch" href="https://cdn.jsdelivr.net" />

        <meta httpEquiv="x-dns-prefetch-control" content="on" />
      </head>
      <body className="min-h-screen flex flex-col" suppressHydrationWarning>
        <LayoutClient>{children}</LayoutClient>
      </body>
    </html>
  );
}
