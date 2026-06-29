package crawler

import (
	"time"
)

// NewArxivAICrawler 创建 arXiv cs.AI 最新论文采集器。
func NewArxivAICrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"arxiv_ai",
		"https://rss.arxiv.org/rss/cs.AI",
		"arXiv cs.AI",
		timeout,
	)
}
