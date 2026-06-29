package crawler

import (
	"time"
)

// NewTheGradientCrawler 创建 The Gradient 采集器（斯坦福 AI 深度分析）。
func NewTheGradientCrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"the_gradient",
		"https://thegradient.pub/rss/",
		"The Gradient",
		timeout,
	)
}
