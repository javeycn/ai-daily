package crawler

import (
	"time"
)

// NewGCPAIBlogCrawler 创建 Google Cloud AI Blog 采集器。
// GCP AI Blog 涵盖 Vertex AI、Gemini API、Cloud TPU 等 Google 云 AI 服务动态。
func NewGCPAIBlogCrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"gcp_ai_blog",
		"https://cloud.google.com/blog/products/ai-machine-learning/rss",
		"Google Cloud AI",
		timeout,
	)
}
