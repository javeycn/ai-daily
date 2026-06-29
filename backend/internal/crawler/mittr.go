package crawler

import (
	"time"
)

// NewMITTechReviewCrawler 创建 MIT Technology Review 采集器。
func NewMITTechReviewCrawler(timeout time.Duration) *RSSCrawler {
	return NewRSSCrawler(
		"mit_tech_review",
		"https://www.technologyreview.com/feed/",
		"MIT Technology Review",
		timeout,
	)
}
