package crawler

import (
	"time"
)

// NewDeepMindCrawler 创建 Google DeepMind 官方博客采集器。
func NewDeepMindCrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"deepmind",
		"https://deepmind.google/blog/rss.xml",
		"DeepMind",
		timeout,
	)
}
