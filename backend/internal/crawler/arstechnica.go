package crawler

import (
	"time"
)

// NewArsTechnicaCrawler 创建 Ars Technica AI 频道采集器。
func NewArsTechnicaCrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"ars_technica",
		"https://feeds.arstechnica.com/arstechnica/technology-lab",
		"Ars Technica",
		timeout,
	)
}
