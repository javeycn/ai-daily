package crawler

import (
	"time"
)

// NewITHomeCrawler 创建 IT之家采集器。
func NewITHomeCrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"ithome",
		"https://www.ithome.com/rss/",
		"IT之家",
		timeout,
	)
}
