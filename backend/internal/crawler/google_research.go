package crawler

import (
	"time"
)

// NewGoogleResearchCrawler 创建 Google Research 博客采集器。
func NewGoogleResearchCrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"google_research",
		"https://research.google/blog/rss/",
		"Google Research",
		timeout,
	)
}
