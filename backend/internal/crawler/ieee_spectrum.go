package crawler

import (
	"time"
)

// NewIEEESpectrumCrawler 创建 IEEE Spectrum AI 频道采集器。
func NewIEEESpectrumCrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"ieee_spectrum",
		"https://spectrum.ieee.org/feeds/topic/artificial-intelligence.rss",
		"IEEE Spectrum",
		timeout,
	)
}
