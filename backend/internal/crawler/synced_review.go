package crawler

import (
	"time"
)

// NewSyncedReviewCrawler 创建 Synced Review 采集器（全球 AI 研究进展）。
func NewSyncedReviewCrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"synced_review",
		"https://syncedreview.com/feed/",
		"Synced Review",
		timeout,
	)
}
