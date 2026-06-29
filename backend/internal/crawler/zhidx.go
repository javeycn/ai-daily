package crawler

import (
	"time"
)

// NewZhidxCrawler 创建智东西采集器（国内 AI 硬件与产业报道）。
func NewZhidxCrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"zhidx",
		"https://zhidx.com/rss",
		"智东西",
		timeout,
	)
}
