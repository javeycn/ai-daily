package crawler

import (
	"time"
)

// NewPapersWithCodeCrawler 创建 arXiv cs.LG (Machine Learning) 论文采集器。
// 注意: paperswithcode.com 无公开 RSS feed，使用 arXiv cs.LG RSS 替代，
// 与 arXiv cs.AI 互补覆盖 AI 学术前沿。
func NewPapersWithCodeCrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"papers_with_code",
		"https://rss.arxiv.org/rss/cs.LG",
		"Papers With Code",
		timeout,
	)
}
