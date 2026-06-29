package crawler

import (
	"time"
)

// NewTheDecoderCrawler 创建 THE DECODER 采集器（AI 新闻与深度分析）。
func NewTheDecoderCrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"the_decoder",
		"https://the-decoder.com/feed/",
		"THE DECODER",
		timeout,
	)
}
