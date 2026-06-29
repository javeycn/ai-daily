package crawler

import (
	"time"
)

// NewSimonWillisonCrawler 创建 Simon Willison 博客采集器（LLM 应用开发领域意见领袖）。
func NewSimonWillisonCrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"simon_willison",
		"https://simonwillison.net/atom/everything/",
		"Simon Willison",
		timeout,
	)
}
