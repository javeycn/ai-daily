package crawler

import (
	"time"
)

// NewOpenAIBlogCrawler 创建 OpenAI Blog 采集器。
func NewOpenAIBlogCrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"openai_blog",
		"https://openai.com/blog/rss.xml",
		"OpenAI Blog",
		timeout,
	)
}
