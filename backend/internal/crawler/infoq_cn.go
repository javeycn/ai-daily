package crawler

import (
	"time"
)

// NewInfoQCNCrawler 创建 InfoQ 中文站采集器。
// InfoQ 中文站是面向中文开发者的技术社区，覆盖 AI、云原生、架构等领域的深度内容。
func NewInfoQCNCrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"infoq_cn",
		"https://www.infoq.cn/feed",
		"InfoQ中文",
		timeout,
	)
}
