package crawler

import (
	"time"
)

// NewWiredAICrawler 创建 Wired AI 频道采集器。
func NewWiredAICrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"wired_ai",
		"https://www.wired.com/feed/tag/ai/latest/rss",
		"Wired",
		timeout,
	)
}
