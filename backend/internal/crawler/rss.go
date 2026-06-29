package crawler

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"ai-news-crawler/internal/model"

	"github.com/antchfx/xmlquery"
)

// imgTagRegex 用于从 HTML 内容中提取 <img> 标签的 src 属性。
var imgTagRegex = regexp.MustCompile(`<img[^>]+src=["']([^"']+)["']`)

// RSSCrawler 是基于 RSS/Atom feed 的通用采集器基类。
type RSSCrawler struct {
	name        string
	feedURL     string
	source      string
	httpTimeout time.Duration
}

// NewRSSCrawler 创建一个新的 RSS 采集器。
func NewRSSCrawler(name, feedURL, source string, httpTimeout time.Duration) *RSSCrawler {
	return &RSSCrawler{
		name:        name,
		feedURL:     feedURL,
		source:      source,
		httpTimeout: httpTimeout,
	}
}

// Name 返回采集器名称。
func (c *RSSCrawler) Name() string {
	return c.name
}

// Crawl 从 RSS feed 采集文章。
func (c *RSSCrawler) Crawl(ctx context.Context, since time.Time) ([]*model.Article, error) {
	slog.Info("crawling rss feed", "source", c.source, "url", c.feedURL)

	doc, err := xmlquery.LoadURL(c.feedURL)
	if err != nil {
		return nil, fmt.Errorf("load rss feed %s: %w", c.feedURL, err)
	}

	var articles []*model.Article

	// 尝试解析 RSS 2.0 格式
	items := xmlquery.Find(doc, "//item")
	if len(items) == 0 {
		// 尝试 Atom 格式
		items = xmlquery.Find(doc, "//entry")
	}

	for _, item := range items {
		article := c.parseItem(item)
		if article == nil {
			continue
		}

		// 只采集 since 之后的文章
		if article.PublishedAt.Before(since) {
			continue
		}

		articles = append(articles, article)
	}

	slog.Info("rss crawl completed", "source", c.source, "count", len(articles))
	return articles, nil
}

// parseItem 从 RSS item/Atom entry 解析文章。
func (c *RSSCrawler) parseItem(item *xmlquery.Node) *model.Article {
	title := getText(item, "//title")
	link := getText(item, "//link")
	if link == "" {
		link = getAttr(item, "//link", "href")
	}
	if title == "" || link == "" {
		return nil
	}

	description := getText(item, "//description")
	if description == "" {
		description = getText(item, "//summary")
	}
	if description == "" {
		description = getText(item, "//content")
	}

	// 提取图片 URL（多种策略）
	imageURL := c.extractImageURL(item, description)

	pubDate := getText(item, "//pubDate")
	if pubDate == "" {
		pubDate = getText(item, "//published")
		if pubDate == "" {
			pubDate = getText(item, "//updated")
		}
	}

	publishedAt := parseTime(pubDate)
	now := time.Now()

	hash := sha256.Sum256([]byte(link))

	return &model.Article{
		ID:            fmt.Sprintf("%x", hash)[:16],
		URL:           link,
		OriginalTitle: title,
		Summary:       description,
		Source:        c.source,
		ImageURL:      imageURL,
		PublishedAt:   publishedAt,
		CrawledAt:     now,
		Hash:          fmt.Sprintf("%x", hash),
	}
}

// extractImageURL 从 RSS item 中尽可能提取文章配图 URL。
// 按优先级依次尝试：media:content > media:thumbnail > enclosure > content 中的 img 标签 > description 中的 img 标签。
func (c *RSSCrawler) extractImageURL(item *xmlquery.Node, description string) string {
	// 策略 1: media:content（带 medium="image" 属性）
	if n := xmlquery.FindOne(item, "//media:content[@medium='image']"); n != nil {
		if u := getNodeAttr(n, "url"); u != "" {
			return u
		}
	}

	// 策略 2: media:content（不带 medium 过滤，但 URL 像图片）
	if n := xmlquery.FindOne(item, "//media:content"); n != nil {
		if u := getNodeAttr(n, "url"); u != "" && looksLikeImageURL(u) {
			return u
		}
	}

	// 策略 3: media:thumbnail
	if n := xmlquery.FindOne(item, "//media:thumbnail"); n != nil {
		if u := getNodeAttr(n, "url"); u != "" {
			return u
		}
	}

	// 策略 4: enclosure（图片类型）
	if n := xmlquery.FindOne(item, "//enclosure"); n != nil {
		encType := getNodeAttr(n, "type")
		if strings.HasPrefix(encType, "image/") {
			if u := getNodeAttr(n, "url"); u != "" {
				return u
			}
		}
		// 即使没有 type，如果 URL 是图片
		if u := getNodeAttr(n, "url"); u != "" && looksLikeImageURL(u) {
			return u
		}
	}

	// 策略 5: 从 content:encoded 中提取第一个 <img> 标签
	contentEncoded := getInnerXML(item, "//content:encoded")
	if contentEncoded == "" {
		contentEncoded = getInnerXML(item, "//content")
	}
	if u := extractFirstImgSrc(contentEncoded); u != "" {
		return u
	}

	// 策略 6: 从 description 的 HTML 中提取第一个 <img> 标签
	descRaw := getInnerXML(item, "//description")
	if u := extractFirstImgSrc(descRaw); u != "" {
		return u
	}

	// 策略 7: 从 summary 的 HTML 中提取
	summaryRaw := getInnerXML(item, "//summary")
	if u := extractFirstImgSrc(summaryRaw); u != "" {
		return u
	}

	return ""
}

// extractFirstImgSrc 从 HTML 内容中提取第一个 <img> 标签的 src URL。
func extractFirstImgSrc(html string) string {
	if html == "" {
		return ""
	}
	matches := imgTagRegex.FindStringSubmatch(html)
	if len(matches) >= 2 {
		u := strings.TrimSpace(matches[1])
		// 过滤掉太小的图标/追踪像素
		if isValidArticleImage(u) {
			return u
		}
	}
	return ""
}

// looksLikeImageURL 判断 URL 是否像图片链接。
func looksLikeImageURL(u string) bool {
	lower := strings.ToLower(u)
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg", ".avif"}
	for _, ext := range imageExts {
		if strings.Contains(lower, ext) {
			return true
		}
	}
	// 常见图片 CDN 路径
	imageCDNs := []string{"/image", "/photo", "/media", "/wp-content/uploads", "/cdn-cgi/image"}
	for _, cdn := range imageCDNs {
		if strings.Contains(lower, cdn) {
			return true
		}
	}
	return false
}

// isValidArticleImage 过滤掉追踪像素、图标等无效图片。
func isValidArticleImage(u string) bool {
	if u == "" {
		return false
	}
	lower := strings.ToLower(u)
	// 排除常见追踪像素和图标
	excludePatterns := []string{
		"pixel", "tracking", "beacon", "1x1",
		"favicon", "icon", "logo",
		"spacer", "blank", "transparent",
		"feeds.feedburner.com",
		"stats.wordpress.com",
	}
	for _, p := range excludePatterns {
		if strings.Contains(lower, p) {
			return false
		}
	}
	// 必须以 http 开头
	if !strings.HasPrefix(lower, "http") {
		return false
	}
	return true
}

// getNodeAttr 直接从节点获取属性值（无需 XPath 查找）。
func getNodeAttr(node *xmlquery.Node, attrName string) string {
	if node == nil {
		return ""
	}
	for _, a := range node.Attr {
		if a.Name.Local == attrName {
			return a.Value
		}
	}
	return ""
}

// getInnerXML 获取 XML 节点的内部 XML（包含 HTML 标签）。
func getInnerXML(node *xmlquery.Node, xpath string) string {
	n := xmlquery.FindOne(node, xpath)
	if n == nil {
		return ""
	}
	return n.OutputXML(false)
}

// getText 获取 XML 节点的文本内容。
func getText(node *xmlquery.Node, xpath string) string {
	n := xmlquery.FindOne(node, xpath)
	if n == nil {
		return ""
	}
	return n.InnerText()
}

// getAttr 获取 XML 节点的属性值。
func getAttr(node *xmlquery.Node, xpath, attr string) string {
	n := xmlquery.FindOne(node, xpath)
	if n == nil {
		return ""
	}
	for _, a := range n.Attr {
		if a.Name.Local == attr {
			return a.Value
		}
	}
	return ""
}

// parseTime 尝试多种常见时间格式解析时间字符串。
func parseTime(s string) time.Time {
	formats := []string{
		time.RFC1123,
		time.RFC1123Z,
		time.RFC3339,
		time.RFC822,
		time.RFC822Z,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t
		}
	}
	// 解析失败则返回当前时间
	return time.Now()
}
