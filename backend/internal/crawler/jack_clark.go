package crawler

import (
	"time"
)

// NewJackClarkCrawler 创建 Import AI (Jack Clark) 采集器（Anthropic 联创的 AI Newsletter）。
func NewJackClarkCrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"jack_clark",
		"https://jack-clark.net/feed/",
		"Import AI",
		timeout,
	)
}
