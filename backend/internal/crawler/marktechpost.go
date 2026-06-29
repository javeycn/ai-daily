package crawler

import (
	"time"
)

// NewMarkTechPostCrawler 创建 MarkTechPost 采集器（AI/ML 论文解读和技术新闻）。
func NewMarkTechPostCrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"marktechpost",
		"https://www.marktechpost.com/feed/",
		"MarkTechPost",
		timeout,
	)
}
