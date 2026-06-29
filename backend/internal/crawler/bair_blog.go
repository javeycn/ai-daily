package crawler

import (
	"time"
)

// NewBAIRBlogCrawler 创建 BAIR Blog 采集器（伯克利人工智能研究实验室博客）。
func NewBAIRBlogCrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"bair_blog",
		"https://bair.berkeley.edu/blog/feed.xml",
		"BAIR Blog",
		timeout,
	)
}
