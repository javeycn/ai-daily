package crawler

import (
	"time"
)

// NewAINewsCrawler 创建 Artificial Intelligence News 采集器。
func NewAINewsCrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"ai_news",
		"https://www.artificialintelligence-news.com/feed/",
		"AI News",
		timeout,
	)
}
