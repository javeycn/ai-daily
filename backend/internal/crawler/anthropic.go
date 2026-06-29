package crawler

import (
	"time"
)

// NewAnthropicCrawler 创建 Anthropic 官方博客采集器。
func NewAnthropicCrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"anthropic",
		"https://www.anthropic.com/feed.xml",
		"Anthropic",
		timeout,
	)
}
