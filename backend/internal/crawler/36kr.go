package crawler

import (
	"time"
)

// New36KrCrawler 创建 36氪 AI 频道采集器。
func New36KrCrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"36kr",
		"https://36kr.com/feed",
		"36氪",
		timeout,
	)
}
