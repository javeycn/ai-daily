package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"ai-news-crawler/internal/model"

	_ "modernc.org/sqlite"
)

// SQLiteRepo 是 ArticleRepository 的 SQLite 实现。
type SQLiteRepo struct {
	db *sql.DB
}

// NewSQLiteRepo 创建 SQLite 数据库连接并初始化表结构。
func NewSQLiteRepo(dbPath string) (*SQLiteRepo, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// 启用 WAL 模式，提升并发读写性能
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("set WAL mode: %w", err)
	}
	if _, err := db.Exec("PRAGMA busy_timeout=5000"); err != nil {
		return nil, fmt.Errorf("set busy timeout: %w", err)
	}

	if err := initSchema(db); err != nil {
		return nil, fmt.Errorf("init schema: %w", err)
	}

	return &SQLiteRepo{db: db}, nil
}

// initSchema 初始化数据库表结构。
func initSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS articles (
		id            TEXT PRIMARY KEY,
		url           TEXT UNIQUE NOT NULL,
		original_title TEXT NOT NULL,
		chinese_title  TEXT DEFAULT '',
		summary       TEXT DEFAULT '',
		source        TEXT NOT NULL,
		image_url     TEXT DEFAULT '',
		tags          TEXT DEFAULT '',
		category      TEXT DEFAULT '',
		importance_score INTEGER DEFAULT 0,
		published_at  DATETIME NOT NULL,
		crawled_at    DATETIME NOT NULL,
		hash          TEXT NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_articles_url ON articles(url);
	CREATE INDEX IF NOT EXISTS idx_articles_published_at ON articles(published_at);
	CREATE INDEX IF NOT EXISTS idx_articles_source ON articles(source);
	CREATE INDEX IF NOT EXISTS idx_articles_hash ON articles(hash);
	CREATE INDEX IF NOT EXISTS idx_articles_category ON articles(category);
	`
	if _, err := db.Exec(schema); err != nil {
		return err
	}

	// 增量迁移：为已有数据库添加 importance_score 列
	_, _ = db.Exec("ALTER TABLE articles ADD COLUMN importance_score INTEGER DEFAULT 0")

	// 增量迁移：为已有数据库添加 recommendation 列（推荐理由）
	_, _ = db.Exec("ALTER TABLE articles ADD COLUMN recommendation TEXT DEFAULT ''")

	return nil
}

// Save 保存一篇文章，如果 URL 已存在则跳过。
func (r *SQLiteRepo) Save(ctx context.Context, article *model.Article) error {
	query := `
		INSERT OR IGNORE INTO articles (id, url, original_title, chinese_title, summary, source, image_url, tags, category, published_at, crawled_at, hash)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		article.ID, article.URL, article.OriginalTitle,
		article.ChineseTitle, article.Summary, article.Source,
		article.ImageURL, article.Tags, article.Category,
		article.PublishedAt.Format(time.RFC3339),
		article.CrawledAt.Format(time.RFC3339),
		article.Hash,
	)
	if err != nil {
		return fmt.Errorf("save article %s: %w", article.ID, err)
	}
	return nil
}

// UpdateImageURL 回填文章的图片 URL（仅当新图片非空且原图片为空时更新）。
func (r *SQLiteRepo) UpdateImageURL(ctx context.Context, url, imageURL string) error {
	query := `UPDATE articles SET image_url = ? WHERE url = ? AND (image_url = '' OR image_url IS NULL)`
	_, err := r.db.ExecContext(ctx, query, imageURL, url)
	if err != nil {
		return fmt.Errorf("update image_url for %s: %w", url, err)
	}
	return nil
}

// GetByURL 根据 URL 查询文章。
func (r *SQLiteRepo) GetByURL(ctx context.Context, url string) (*model.Article, error) {
	query := `SELECT id, url, original_title, chinese_title, summary, recommendation, source, image_url, tags, category, importance_score, published_at, crawled_at, hash FROM articles WHERE url = ?`
	row := r.db.QueryRowContext(ctx, query, url)

	var a model.Article
	var publishedStr, crawledStr string
	err := row.Scan(&a.ID, &a.URL, &a.OriginalTitle, &a.ChineseTitle,
		&a.Summary, &a.Recommendation, &a.Source, &a.ImageURL, &a.Tags, &a.Category,
		&a.ImportanceScore, &publishedStr, &crawledStr, &a.Hash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get article by url: %w", err)
	}
	a.PublishedAt = parseTimeStr(publishedStr)
	a.CrawledAt = parseTimeStr(crawledStr)
	return &a, nil
}

// GetByDate 查询指定时间范围内的文章。
func (r *SQLiteRepo) GetByDate(ctx context.Context, start, end time.Time) ([]*model.Article, error) {
	query := `
		SELECT id, url, original_title, chinese_title, summary, recommendation, source, image_url, tags, category, importance_score, published_at, crawled_at, hash
		FROM articles
		WHERE published_at >= ? AND published_at < ?
		ORDER BY published_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, start.Format(time.RFC3339), end.Format(time.RFC3339))
	if err != nil {
		return nil, fmt.Errorf("query articles by date: %w", err)
	}
	defer rows.Close()

	var articles []*model.Article
	for rows.Next() {
		var a model.Article
		var publishedStr, crawledStr string
		err := rows.Scan(&a.ID, &a.URL, &a.OriginalTitle, &a.ChineseTitle,
			&a.Summary, &a.Recommendation, &a.Source, &a.ImageURL, &a.Tags, &a.Category,
			&a.ImportanceScore, &publishedStr, &crawledStr, &a.Hash)
		if err != nil {
			return nil, fmt.Errorf("scan article: %w", err)
		}
		a.PublishedAt = parseTimeStr(publishedStr)
		a.CrawledAt = parseTimeStr(crawledStr)
		articles = append(articles, &a)
	}
	return articles, rows.Err()
}

// UpdateSummary 更新文章的中文标题、摘要、标签、分类、重要度评分和推荐理由。
func (r *SQLiteRepo) UpdateSummary(ctx context.Context, id, chineseTitle, summary, tags, category string, importanceScore int, recommendation string) error {
	query := `UPDATE articles SET chinese_title = ?, summary = ?, tags = ?, category = ?, importance_score = ?, recommendation = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, chineseTitle, summary, tags, category, importanceScore, recommendation, id)
	if err != nil {
		return fmt.Errorf("update summary for %s: %w", id, err)
	}
	return nil
}

// GetUnsummarized 获取尚未生成中文摘要的文章。
func (r *SQLiteRepo) GetUnsummarized(ctx context.Context, limit int) ([]*model.Article, error) {
	query := `
		SELECT id, url, original_title, chinese_title, summary, recommendation, source, image_url, tags, category, importance_score, published_at, crawled_at, hash
		FROM articles
		WHERE chinese_title = ''
		ORDER BY published_at DESC
		LIMIT ?
	`
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("query unsummarized articles: %w", err)
	}
	defer rows.Close()

	var articles []*model.Article
	for rows.Next() {
		var a model.Article
		var publishedStr, crawledStr string
		err := rows.Scan(&a.ID, &a.URL, &a.OriginalTitle, &a.ChineseTitle,
			&a.Summary, &a.Recommendation, &a.Source, &a.ImageURL, &a.Tags, &a.Category,
			&a.ImportanceScore, &publishedStr, &crawledStr, &a.Hash)
		if err != nil {
			return nil, fmt.Errorf("scan article: %w", err)
		}
		a.PublishedAt = parseTimeStr(publishedStr)
		a.CrawledAt = parseTimeStr(crawledStr)
		articles = append(articles, &a)
	}
	return articles, rows.Err()
}

// Close 关闭数据库连接。
func (r *SQLiteRepo) Close() error {
	return r.db.Close()
}

// GetBrokenSummaries 获取 chinese_title 像 JSON 残片的文章。
// 匹配条件：标题为 "{" 或以 "{" 开头，或 summary 中包含 "chinese_title" 键名。
func (r *SQLiteRepo) GetBrokenSummaries(ctx context.Context) ([]*model.Article, error) {
	query := `
		SELECT id, url, original_title, chinese_title, summary, recommendation, source, image_url, tags, category, importance_score, published_at, crawled_at, hash
		FROM articles
		WHERE chinese_title IN ('{', '}')
		   OR (chinese_title != '' AND summary LIKE '%"chinese_title"%')
		   OR (chinese_title != '' AND summary LIKE '%"summary"%' AND summary LIKE '%"category"%')
		ORDER BY published_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query broken summaries: %w", err)
	}
	defer rows.Close()

	var articles []*model.Article
	for rows.Next() {
		var a model.Article
		var publishedStr, crawledStr string
		err := rows.Scan(&a.ID, &a.URL, &a.OriginalTitle, &a.ChineseTitle,
			&a.Summary, &a.Recommendation, &a.Source, &a.ImageURL, &a.Tags, &a.Category,
			&a.ImportanceScore, &publishedStr, &crawledStr, &a.Hash)
		if err != nil {
			return nil, fmt.Errorf("scan article: %w", err)
		}
		a.PublishedAt = parseTimeStr(publishedStr)
		a.CrawledAt = parseTimeStr(crawledStr)
		articles = append(articles, &a)
	}
	return articles, rows.Err()
}

// parseTimeStr 将字符串时间解析为 time.Time。
func parseTimeStr(s string) time.Time {
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t
		}
	}
	return time.Now()
}
