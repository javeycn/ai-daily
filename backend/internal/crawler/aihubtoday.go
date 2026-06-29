package crawler

import (
	"time"
)

// NewAIHubTodayCrawler 创建 AI Hub Today 采集器（AI 新闻聚合平台，daily-ai-news-skill Tier 1 推荐）。
func NewAIHubTodayCrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"aihubtoday",
		"https://ai.hubtoday.app/feed",
		"AI Hub Today",
		timeout,
	)
}
