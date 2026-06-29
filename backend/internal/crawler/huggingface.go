package crawler

import (
	"time"
)

// NewHuggingFaceCrawler 创建 HuggingFace Blog 采集器。
func NewHuggingFaceCrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"huggingface",
		"https://huggingface.co/blog/feed.xml",
		"HuggingFace",
		timeout,
	)
}
