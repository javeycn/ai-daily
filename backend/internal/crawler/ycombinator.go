package crawler

import (
	"time"
)

// NewYCombinatorCrawler 创建 Y Combinator Hacker News 采集器。
func NewYCombinatorCrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"ycombinator",
		"https://news.ycombinator.com/rss",
		"Hacker News",
		timeout,
	)
}
