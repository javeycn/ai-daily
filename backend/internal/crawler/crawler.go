// Package crawler 定义新闻采集器的统一接口。
package crawler

import (
	"context"
	"time"

	"ai-news-crawler/internal/model"
)

// Result 是单个采集器返回的采集结果。
type Result struct {
	Articles []*model.Article
	Source   string
	Error    error
}

// Crawler 定义新闻采集器的接口，每个新闻源实现此接口。
type Crawler interface {
	// Name 返回采集器的名称标识（如 "the_verge"）。
	Name() string

	// Crawl 从新闻源采集文章，返回结果通道。
	Crawl(ctx context.Context, since time.Time) ([]*model.Article, error)
}
