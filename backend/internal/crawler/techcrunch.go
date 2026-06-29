package crawler

import (
	"time"
)

// NewTechCrunchCrawler 创建 TechCrunch AI 频道采集器。
func NewTechCrunchCrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"techcrunch",
		"https://techcrunch.com/category/artificial-intelligence/feed/",
		"TechCrunch",
		timeout,
	)
}
