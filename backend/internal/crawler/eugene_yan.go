package crawler

import (
	"time"
)

// NewEugeneYanCrawler 创建 Eugene Yan 博客 RSS 采集器。
// Eugene Yan 是 Amazon 首席 AI 工程师，专注 AI 系统设计和推荐系统。
func NewEugeneYanCrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"eugene_yan",
		"https://eugeneyan.com/rss/",
		"Eugene Yan",
		timeout,
	)
}
