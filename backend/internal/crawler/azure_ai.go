package crawler

import (
	"time"
)

// NewAzureAIBlogCrawler 创建 Azure AI Services Blog 采集器。
// Azure AI Blog 涵盖 Azure OpenAI Service、Cognitive Services、Copilot Studio 等微软云 AI 服务动态。
func NewAzureAIBlogCrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"azure_ai_blog",
		"https://techcommunity.microsoft.com/plugins/custom/microsoft/o365/custom-blog-rss?board=Azure-AI-Services-blog",
		"Azure AI",
		timeout,
	)
}
