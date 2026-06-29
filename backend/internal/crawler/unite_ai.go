package crawler

import (
	"time"
)

// NewUniteAICrawler 创建 Unite.AI 采集器（综合性 AI 新闻平台）。
func NewUniteAICrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"unite_ai",
		"https://www.unite.ai/feed/",
		"Unite.AI",
		timeout,
	)
}
