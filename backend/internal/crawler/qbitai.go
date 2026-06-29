package crawler

import (
	"time"
)

// NewQbitAICrawler 创建量子位采集器。
// 量子位（QbitAI）是国内头部 AI 科技媒体，报道覆盖 AI 大模型、云计算、自动驾驶等领域。
func NewQbitAICrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"qbitai",
		"https://www.qbitai.com/feed",
		"量子位",
		timeout,
	)
}
