package crawler

import (
	"time"
)

// NewMetaEngineeringCrawler 创建 Meta Engineering 博客采集器。
func NewMetaEngineeringCrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"meta_engineering",
		"https://engineering.fb.com/feed/",
		"Meta Engineering",
		timeout,
	)
}
