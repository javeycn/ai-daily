package crawler

import (
	"time"
)

// NewKarpathyCrawler 创建 Andrej Karpathy 博客 RSS 采集器。
// Karpathy 是前 OpenAI/Tesla AI 负责人，深度学习领域重要的思想领袖。
func NewKarpathyCrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"karpathy",
		"https://karpathy.github.io/feed.xml",
		"Karpathy",
		timeout,
	)
}
