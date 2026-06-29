package crawler

import (
	"time"
)

// NewChipHuyenCrawler 创建 Chip Huyen 博客 RSS 采集器。
// Chip Huyen 是 MLOps/LLMOps 领域的思想领袖，著有《Designing Machine Learning Systems》。
func NewChipHuyenCrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"chip_huyen",
		"https://huyenchip.com/feed.xml",
		"Chip Huyen",
		timeout,
	)
}
