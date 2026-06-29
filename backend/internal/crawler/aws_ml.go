package crawler

import (
	"time"
)

// NewAWSMLBlogCrawler 创建 AWS Machine Learning Blog 采集器。
// AWS ML Blog 是 Amazon 官方的机器学习博客，涵盖 SageMaker、Bedrock 等云 AI 服务动态。
func NewAWSMLBlogCrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"aws_ml_blog",
		"https://aws.amazon.com/blogs/machine-learning/feed/",
		"AWS ML Blog",
		timeout,
	)
}
