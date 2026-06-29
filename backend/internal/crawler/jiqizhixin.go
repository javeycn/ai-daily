package crawler

import (
	"time"
)

// NewJiqizhixinCrawler 创建机器之心采集器。
func NewJiqizhixinCrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"jiqizhixin",
		"https://www.jiqizhixin.com/rss",
		"机器之心",
		timeout,
	)
}
