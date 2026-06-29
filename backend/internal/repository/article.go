// Package repository 定义文章数据存储的接口和实现。
package repository

import (
	"context"
	"time"

	"ai-news-crawler/internal/model"
)

// ArticleRepository 定义文章数据的存储接口。
type ArticleRepository interface {
	// Save 保存一篇文章，如果已存在（URL 去重）则跳过。
	Save(ctx context.Context, article *model.Article) error

	// GetByURL 根据 URL 查询文章是否已存在。
	GetByURL(ctx context.Context, url string) (*model.Article, error)

	// GetByDate 查询指定日期范围内的所有文章。
	GetByDate(ctx context.Context, start, end time.Time) ([]*model.Article, error)

	// UpdateSummary 更新文章的中文标题、摘要、标签、分类、重要度评分和推荐理由。
	UpdateSummary(ctx context.Context, id, chineseTitle, summary, tags, category string, importanceScore int, recommendation string) error

	// UpdateImageURL 回填文章的图片 URL。
	UpdateImageURL(ctx context.Context, url, imageURL string) error

	// GetUnsummarized 获取尚未生成摘要的文章列表。
	GetUnsummarized(ctx context.Context, limit int) ([]*model.Article, error)

	// GetBrokenSummaries 获取 chinese_title 像 JSON 残片的文章（标题为 "{" 等非正常内容）。
	GetBrokenSummaries(ctx context.Context) ([]*model.Article, error)

	// Close 关闭数据库连接。
	Close() error
}
