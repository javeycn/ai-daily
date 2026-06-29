package crawler

import (
	"time"
)

// NewNvidiaBlogCrawler 创建 NVIDIA AI Blog 采集器。
func NewNvidiaBlogCrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"nvidia_blog",
		"https://blogs.nvidia.com/feed/",
		"NVIDIA Blog",
		timeout,
	)
}
