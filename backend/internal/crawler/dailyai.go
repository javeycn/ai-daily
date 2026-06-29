package crawler

import (
	"time"
)

// NewDailyAICrawler 创建 DailyAI 采集器（AI 新闻与深度报道聚合网站）。
func NewDailyAICrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"dailyai",
		"https://dailyai.com/feed/",
		"DailyAI",
		timeout,
	)
}
