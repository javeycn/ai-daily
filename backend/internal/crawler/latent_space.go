package crawler

import (
	"time"
)

// NewLatentSpaceCrawler 创建 Latent Space 采集器（AI 工程师社区播客/Newsletter）。
func NewLatentSpaceCrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"latent_space",
		"https://www.latent.space/feed",
		"Latent Space",
		timeout,
	)
}
