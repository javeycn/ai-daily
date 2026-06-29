package crawler

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"ai-news-crawler/internal/config"
	"ai-news-crawler/internal/model"
)

// defaultSearchQueries 是默认的搜索查询模板，覆盖突发新闻和 RSS 不易覆盖的主题。
// 借鉴 daily-ai-news-skill 的搜索查询策略，按类别分组。
var defaultSearchQueries = []string{
	// 通用 AI 突发新闻
	`"AI news today" OR "artificial intelligence breakthrough"`,
	// 大厂动态
	`"OpenAI announcement" OR "GPT update" OR "ChatGPT news"`,
	`"Google AI announcement" OR "Gemini update" OR "DeepMind"`,
	`"Anthropic news" OR "Claude update"`,
	`"Meta AI" OR "LLaMA update"`,
	`"Microsoft AI" OR "Copilot update" OR "Azure AI news"`,
	`"xAI" OR "Grok update" OR "Apple AI" OR "Apple Intelligence"`,
	// 产品发布 & 开源
	`"AI product release" OR "new AI tool" OR "LLM release"`,
	`"open source AI" OR "AI model release"`,
	// 融资 & 商业
	`"AI startup funding" OR "artificial intelligence investment"`,
	// AI 大佬动态（补充 RSS 覆盖不到的推文和演讲）
	`"Sam Altman" OR "Andrej Karpathy" OR "Demis Hassabis" AI`,
	`"Yann LeCun" OR "Andrew Ng" OR "Dario Amodei" AI opinion`,
	// AI 安全与伦理
	`"AI safety" OR "AI regulation" OR "AI policy" OR "AI ethics"`,

	// === 国内云厂商 & AI 平台动态 ===
	`"腾讯云 AI" OR "混元大模型" OR "腾讯混元" OR "腾讯云智能"`,
	`"阿里云 AI" OR "通义千问" OR "阿里达摩院" OR "阿里云百炼"`,
	`"华为云 AI" OR "盘古大模型" OR "昇腾" OR "华为昇思"`,
	`"百度智能云" OR "文心一言" OR "文心大模型" OR "千帆大模型平台"`,
	`"火山引擎" OR "豆包大模型" OR "字节跳动 AI" OR "扣子 Coze"`,
	// 国内 AI 新锐公司
	`"DeepSeek" OR "深度求索" OR "DeepSeek-V3" OR "DeepSeek-R1"`,
	`"Kimi" OR "月之暗面" OR "Moonshot AI"`,
	`"智谱AI" OR "ChatGLM" OR "智谱清言" OR "GLM-4"`,
	`"MiniMax" OR "海螺AI" OR "abab大模型"`,
	`"零一万物" OR "Yi大模型" OR "李开复 AI"`,
	`"百川智能" OR "Baichuan" OR "百川大模型"`,
	`"商汤科技" OR "SenseTime" OR "日日新大模型"`,
	// 国内 AI 应用生态
	`"讯飞星火" OR "科大讯飞 AI" OR "星火大模型"`,
	`"金山办公 AI" OR "WPS AI" OR "小米大模型" OR "小爱同学 AI"`,
	`"钉钉 AI" OR "飞书 AI" OR "企业微信 AI" OR "AI办公"`,
	// 国内 AI 政策与产业
	`"中国 AI 政策" OR "国产大模型" OR "AI 芯片国产化" OR "算力中心"`,

	// === 国内云厂商补充查询（加强国内候选池）===
	`"腾讯云" AI 发布 OR "腾讯混元" 更新 OR "腾讯云智能" 新功能`,
	`"阿里云" AI 平台 OR "通义千问" 升级 OR "百炼" 大模型`,
	`"华为云" AI 方案 OR "盘古大模型" 应用 OR "昇腾" 算力`,
	`"百度智能云" 发布 OR "千帆" 平台 OR "文心" 企业版`,
	`"火山引擎" 大模型 OR "豆包" 企业版 OR "Coze" 插件`,
	`"天翼云" AI OR "移动云" AI OR "联通云" AI OR "浪潮云" AI`,

	// === 国际云厂商 AI 服务动态（合并为单条，限制数量）===
	`"AWS Bedrock" OR "Azure OpenAI" OR "Vertex AI" cloud update`,
}

// searchResult 表示 Serper API 返回的单条结果。
type searchResult struct {
	Title   string
	Link    string
	Snippet string
	Date    string
}

// WebSearchCrawler 通过 Serper.dev Google Search API 补充 RSS 覆盖不到的 AI 新闻。
type WebSearchCrawler struct {
	apiKey         string
	client         *http.Client
	queries        []string
	maxResults     int
	excludeDomains map[string]bool
}

// NewWebSearchCrawler 创建一个新的 Web 搜索采集器。
// 当配置未启用或 API Key 为空时返回 nil。
func NewWebSearchCrawler(cfg *config.WebSearchConfig) *WebSearchCrawler {
	if cfg == nil || !cfg.Enabled || cfg.APIKey == "" {
		return nil
	}

	queries := cfg.Queries
	if len(queries) == 0 {
		queries = defaultSearchQueries
	}

	maxResults := cfg.MaxResults
	if maxResults <= 0 {
		maxResults = 10
	}

	timeout := time.Duration(cfg.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 15 * time.Second
	}

	excludeDomains := make(map[string]bool, len(cfg.ExcludeDomains))
	for _, d := range cfg.ExcludeDomains {
		excludeDomains[strings.ToLower(d)] = true
	}

	return &WebSearchCrawler{
		apiKey:         cfg.APIKey,
		client:         &http.Client{Timeout: timeout},
		queries:        queries,
		maxResults:     maxResults,
		excludeDomains: excludeDomains,
	}
}

// Name 返回采集器名称。
func (c *WebSearchCrawler) Name() string {
	return "web_search"
}

// Crawl 执行多个搜索查询，汇聚结果并去重。
func (c *WebSearchCrawler) Crawl(ctx context.Context, since time.Time) ([]*model.Article, error) {
	slog.Info("web search crawling started", "queries", len(c.queries), "max_results_per_query", c.maxResults)

	seen := make(map[string]bool)
	var allArticles []*model.Article

	for _, query := range c.queries {
		// 国际云厂商查询限制返回数量，避免国际云内容挤占候选池
		queryMaxResults := c.maxResults
		if isInternationalCloudQuery(query) {
			queryMaxResults = 3
			slog.Debug("international cloud query limited", "query", query, "max_results", queryMaxResults)
		}
		results, err := c.searchSerper(ctx, query, queryMaxResults)
		if err != nil {
			slog.Warn("web search query failed", "query", query, "error", err)
			continue
		}

		for _, r := range results {
			if r.Link == "" || seen[r.Link] {
				continue
			}
			// 排除配置中指定的域名
			if c.isExcludedDomain(r.Link) {
				continue
			}
			seen[r.Link] = true

			article := c.resultToArticle(r, since)
			if article == nil {
				continue
			}
			allArticles = append(allArticles, article)
		}

		slog.Debug("search query completed", "query", query, "results", len(results))
	}

	slog.Info("web search crawling completed", "total_articles", len(allArticles))
	return allArticles, nil
}

// resultToArticle 将搜索结果转换为 Article 模型。
func (c *WebSearchCrawler) resultToArticle(r searchResult, since time.Time) *model.Article {
	// 解析发布时间
	publishedAt := parseSearchDate(r.Date)

	// 如果搜索结果有日期信息且在 since 之前，跳过
	if !publishedAt.IsZero() && publishedAt.Before(since) {
		return nil
	}

	// 如果没有日期信息，使用当前时间（搜索 API 通常返回近期结果）
	if publishedAt.IsZero() {
		publishedAt = time.Now()
	}

	hash := sha256.Sum256([]byte(r.Link))
	hashStr := fmt.Sprintf("%x", hash)

	return &model.Article{
		ID:            hashStr[:16],
		URL:           r.Link,
		OriginalTitle: r.Title,
		Summary:       r.Snippet,
		Source:        "Web Search",
		PublishedAt:   publishedAt,
		CrawledAt:     time.Now(),
		Hash:          hashStr,
	}
}

// isInternationalCloudQuery 判断搜索查询是否专门针对国际云厂商。
// 用于限制此类查询的返回数量，避免国际云内容挤占候选池。
func isInternationalCloudQuery(query string) bool {
	q := strings.ToLower(query)
	intlCloudKeywords := []string{
		"aws", "amazon bedrock", "sagemaker",
		"azure openai", "azure ai", "copilot studio",
		"google cloud ai", "vertex ai", "gcp",
	}
	for _, kw := range intlCloudKeywords {
		if strings.Contains(q, kw) {
			return true
		}
	}
	return false
}

// isExcludedDomain 检查 URL 是否属于配置中排除的域名。
func (c *WebSearchCrawler) isExcludedDomain(link string) bool {
	if len(c.excludeDomains) == 0 {
		return false
	}
	u, err := url.Parse(link)
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Hostname())
	for domain := range c.excludeDomains {
		if host == domain || strings.HasSuffix(host, "."+domain) {
			return true
		}
	}
	return false
}

// parseSearchDate 尝试解析搜索 API 返回的日期字符串。
func parseSearchDate(s string) time.Time {
	if s == "" {
		return time.Time{}
	}

	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05",
		"2006-01-02",
		"Jan 2, 2006",
		"January 2, 2006",
		"02 Jan 2006",
		"2 Jan 2006",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t
		}
	}

	// 处理 "X hours ago"、"X days ago" 等相对时间
	now := time.Now()
	s = strings.ToLower(strings.TrimSpace(s))
	if strings.Contains(s, "hour") {
		return now.Add(-2 * time.Hour) // 粗略估计
	}
	if strings.Contains(s, "day") {
		return now.AddDate(0, 0, -1) // 粗略估计
	}
	if strings.Contains(s, "minute") {
		return now.Add(-30 * time.Minute)
	}

	return time.Time{}
}

// --- Serper.dev Google Search API ---

const serperBaseURL = "https://google.serper.dev/search"

// serperRequest Serper API 请求体。
type serperRequest struct {
	Q   string `json:"q"`
	Num int    `json:"num"`
	TBS string `json:"tbs,omitempty"` // 时间过滤: qdr:d (24h), qdr:w (week)
}

// serperResponse Serper API 响应体。
type serperResponse struct {
	Organic []serperOrganicResult `json:"organic"`
	News    []serperNewsResult    `json:"news"`
}

// serperOrganicResult Serper 有机搜索结果。
type serperOrganicResult struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	Snippet string `json:"snippet"`
	Date    string `json:"date"`
}

// serperNewsResult Serper 新闻搜索结果。
type serperNewsResult struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	Snippet string `json:"snippet"`
	Date    string `json:"date"`
	Source  string `json:"source"`
}

// searchSerper 执行 Serper.dev 搜索查询。
func (c *WebSearchCrawler) searchSerper(ctx context.Context, query string, maxResults int) ([]searchResult, error) {
	reqBody := serperRequest{
		Q:   query,
		Num: maxResults,
		TBS: "qdr:d", // 仅搜索最近 24 小时的结果
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal serper request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, serperBaseURL, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, fmt.Errorf("create serper request: %w", err)
	}

	req.Header.Set("X-API-KEY", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("serper api call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("serper api returned status %d: %s", resp.StatusCode, string(body))
	}

	var serperResp serperResponse
	if err := json.NewDecoder(resp.Body).Decode(&serperResp); err != nil {
		return nil, fmt.Errorf("decode serper response: %w", err)
	}

	var results []searchResult

	// 合并新闻结果（优先）和有机搜索结果
	for _, n := range serperResp.News {
		results = append(results, searchResult{
			Title:   n.Title,
			Link:    n.Link,
			Snippet: n.Snippet,
			Date:    n.Date,
		})
	}
	for _, o := range serperResp.Organic {
		results = append(results, searchResult{
			Title:   o.Title,
			Link:    o.Link,
			Snippet: o.Snippet,
			Date:    o.Date,
		})
	}

	return results, nil
}
