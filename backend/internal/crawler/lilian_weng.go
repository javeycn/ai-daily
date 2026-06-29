package crawler

import (
	"time"
)

// NewLilianWengCrawler 创建 Lilian Weng 博客采集器（OpenAI 研究员的 AI 技术博客）。
func NewLilianWengCrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"lilian_weng",
		"https://lilianweng.github.io/index.xml",
		"Lilian Weng",
		timeout,
	)
}
