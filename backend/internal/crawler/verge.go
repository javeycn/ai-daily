package crawler

import (
	"time"
)

// NewVergeCrawler 创建 The Verge AI 频道采集器。
func NewVergeCrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"the_verge",
		"https://www.theverge.com/rss/ai-artificial-intelligence/index.xml",
		"The Verge",
		timeout,
	)
}
