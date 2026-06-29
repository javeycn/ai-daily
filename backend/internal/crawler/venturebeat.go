package crawler

import (
	"time"
)

// NewVentureBeatCrawler 创建 VentureBeat AI 频道采集器。
func NewVentureBeatCrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"venturebeat",
		"https://venturebeat.com/category/ai/feed/",
		"VentureBeat",
		timeout,
	)
}
