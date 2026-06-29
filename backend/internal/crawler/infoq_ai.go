package crawler

import (
	"time"
)

// NewInfoQAICrawler 创建 InfoQ AI 频道 RSS 采集器。
// InfoQ 是架构师社区的 AI 板块，偏工程实践和技术深度文章。
func NewInfoQAICrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"infoq_ai",
		"https://feed.infoq.com/ai-ml-data-eng/",
		"InfoQ AI",
		timeout,
	)
}
