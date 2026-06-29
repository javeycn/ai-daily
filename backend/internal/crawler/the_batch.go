package crawler

import (
	"time"
)

// NewTheBatchCrawler 创建 The Batch (DeepLearning.AI) 采集器（吴恩达的 AI 周报）。
func NewTheBatchCrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"the_batch",
		"https://www.deeplearning.ai/the-batch/feed/",
		"The Batch",
		timeout,
	)
}
