package crawler

import (
	"time"
)

// NewPanDailyCrawler 创建 PanDaily 采集器（中国科技创投出海资讯）。
func NewPanDailyCrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"pandaily",
		"https://pandaily.com/feed/",
		"PanDaily",
		timeout,
	)
}
