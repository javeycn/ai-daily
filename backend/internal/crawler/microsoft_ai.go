package crawler

import (
	"time"
)

// NewMicrosoftAICrawler 创建 Microsoft AI Blog 采集器。
func NewMicrosoftAICrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"microsoft_ai",
		"https://blogs.microsoft.com/ai/feed/",
		"Microsoft AI",
		timeout,
	)
}
