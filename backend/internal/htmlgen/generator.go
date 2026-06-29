// Package htmlgen 负责将日报数据渲染为静态 HTML 文件。
package htmlgen

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"ai-news-crawler/internal/model"
)

//go:embed templates/*.html
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

// Generator 负责将数据渲染为静态 HTML 并写入磁盘。
type Generator struct {
	outputDir string
	funcMap   template.FuncMap
}

// New 创建一个新的 Generator 实例。
func New(outputDir string) (*Generator, error) {
	funcMap := template.FuncMap{
		"catColorClass":  catColorClass,
		"catEmoji":       catEmoji,
		"displayTitle":   displayTitle,
		"englishTitle":   englishTitle,
		"splitTags":      splitTags,
		"slice3":         slice3,
		"isCategoryTag":  isCategoryTag,
		"topTags":        topTags,
	}

	// 验证模板是否可以正确解析（快速失败）
	if _, err := template.New("").Funcs(funcMap).ParseFS(templateFS, "templates/layout.html"); err != nil {
		return nil, fmt.Errorf("parse layout template: %w", err)
	}

	return &Generator{
		outputDir: outputDir,
		funcMap:   funcMap,
	}, nil
}

// pageData 是传递给所有页面模板的公共数据。
type pageData struct {
	Title       string
	Description string
	Year        int
}

// homeData 首页模板数据。
type homeData struct {
	pageData
	Report *model.DailyReport
}

// dailyData 日报详情页模板数据。
type dailyData struct {
	pageData
	Report *model.DailyReport
}

// archiveMonth 归档页按月分组。
type archiveMonth struct {
	Label string
	Days  []model.DailyIndex
}

// archiveCategoryStat 归档页分类统计。
type archiveCategoryStat struct {
	Name       string
	Emoji      string
	ColorClass string
	Count      int
}

// archiveData 归档页模板数据。
type archiveData struct {
	pageData
	TotalDays      int
	Months         []archiveMonth
	CategoryStats  []archiveCategoryStat
}

// searchItem 搜索数据的压缩结构，减小内嵌 JSON 体积。
type searchItem struct {
	CT string `json:"ct"`          // chinese_title
	OT string `json:"ot"`          // original_title
	SM string `json:"sm"`          // summary
	TG string `json:"tg"`          // tags
	CA string `json:"ca"`          // category
	SC string `json:"sc"`          // source
	DT string `json:"dt"`          // date
}

// searchData 搜索页模板数据。
type searchData struct {
	pageData
	SearchDataJSON template.JS
}

// aboutData 关于页模板数据。
type aboutData struct {
	pageData
}

// GenerateAll 重新生成整个站点的所有页面。
func (g *Generator) GenerateAll(
	latestReport *model.DailyReport,
	allReports []*model.DailyReport,
	index *model.IndexFile,
) error {
	slog.Info("generating all HTML pages", "output_dir", g.outputDir)

	// 复制静态资源
	if err := g.copyStaticAssets(); err != nil {
		return fmt.Errorf("copy static assets: %w", err)
	}

	// 生成 robots.txt
	if err := g.generateRobotsTxt(); err != nil {
		return fmt.Errorf("generate robots.txt: %w", err)
	}

	// 生成首页（使用最新日报）
	if err := g.GenerateHome(latestReport); err != nil {
		return fmt.Errorf("generate home: %w", err)
	}

	// 生成所有日报详情页
	for _, report := range allReports {
		if err := g.GenerateDaily(report); err != nil {
			slog.Error("generate daily page failed", "date", report.Date, "error", err)
			continue
		}
	}

	// 生成归档页
	if err := g.GenerateArchive(index); err != nil {
		return fmt.Errorf("generate archive: %w", err)
	}

	// 生成搜索页
	if err := g.GenerateSearch(allReports); err != nil {
		return fmt.Errorf("generate search: %w", err)
	}

	// 生成关于页
	if err := g.GenerateAbout(); err != nil {
		return fmt.Errorf("generate about: %w", err)
	}

	slog.Info("all HTML pages generated successfully")
	return nil
}

// GenerateHome 生成首页 index.html。
func (g *Generator) GenerateHome(report *model.DailyReport) error {
	data := homeData{
		pageData: pageData{
			Title:       "AI Daily — 每日 AI 资讯",
			Description: "自动化 AI 资讯聚合平台，每日精选全球 AI 新闻",
			Year:        time.Now().Year(),
		},
		Report: report,
	}

	outPath := filepath.Join(g.outputDir, "index.html")
	if err := g.renderToFile("home.html", data, outPath); err != nil {
		return fmt.Errorf("render home page: %w", err)
	}

	slog.Info("home page generated", "file", outPath)
	return nil
}

// GenerateDaily 生成单日日报详情页 daily/{date}/index.html。
func (g *Generator) GenerateDaily(report *model.DailyReport) error {
	if report == nil {
		return nil
	}

	data := dailyData{
		pageData: pageData{
			Title:       fmt.Sprintf("%s — %s", report.Date, report.Title),
			Description: report.Summary,
			Year:        time.Now().Year(),
		},
		Report: report,
	}

	dir := filepath.Join(g.outputDir, "daily", report.Date)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create daily directory %s: %w", report.Date, err)
	}

	outPath := filepath.Join(dir, "index.html")
	if err := g.renderToFile("daily.html", data, outPath); err != nil {
		return fmt.Errorf("render daily page %s: %w", report.Date, err)
	}

	slog.Info("daily page generated", "date", report.Date, "file", outPath)
	return nil
}

// GenerateArchive 生成归档页 archive/index.html。
func (g *Generator) GenerateArchive(index *model.IndexFile) error {
	months := groupByMonth(index.Days)

	// 从所有日报的 tags 中提取分类统计
	catStats := extractCategoryStats(index.Days)

	data := archiveData{
		pageData: pageData{
			Title:       "历史归档",
			Description: "AI Daily 历史日报归档",
			Year:        time.Now().Year(),
		},
		TotalDays:     len(index.Days),
		Months:        months,
		CategoryStats: catStats,
	}

	dir := filepath.Join(g.outputDir, "archive")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create archive directory: %w", err)
	}

	outPath := filepath.Join(dir, "index.html")
	if err := g.renderToFile("archive.html", data, outPath); err != nil {
		return fmt.Errorf("render archive page: %w", err)
	}

	slog.Info("archive page generated", "file", outPath, "total_days", len(index.Days))
	return nil
}

// GenerateSearch 生成搜索页 search/index.html。
func (g *Generator) GenerateSearch(allReports []*model.DailyReport) error {
	items := buildSearchData(allReports)

	jsonBytes, err := json.Marshal(items)
	if err != nil {
		return fmt.Errorf("marshal search data: %w", err)
	}

	data := searchData{
		pageData: pageData{
			Title:       "搜索资讯",
			Description: "搜索 AI Daily 全部资讯",
			Year:        time.Now().Year(),
		},
		SearchDataJSON: template.JS(jsonBytes),
	}

	dir := filepath.Join(g.outputDir, "search")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create search directory: %w", err)
	}

	outPath := filepath.Join(dir, "index.html")
	if err := g.renderToFile("search.html", data, outPath); err != nil {
		return fmt.Errorf("render search page: %w", err)
	}

	slog.Info("search page generated", "file", outPath, "articles", len(items))
	return nil
}

// GenerateAbout 生成关于页 about/index.html。
func (g *Generator) GenerateAbout() error {
	data := aboutData{
		pageData: pageData{
			Title:       "关于",
			Description: "关于 AI Daily 项目",
			Year:        time.Now().Year(),
		},
	}

	dir := filepath.Join(g.outputDir, "about")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create about directory: %w", err)
	}

	outPath := filepath.Join(dir, "index.html")
	if err := g.renderToFile("about.html", data, outPath); err != nil {
		return fmt.Errorf("render about page: %w", err)
	}

	slog.Info("about page generated", "file", outPath)
	return nil
}

// renderToFile 将指定页面模板（与 layout.html 组合）渲染到文件。
// 每次独立解析 layout.html + 目标页面模板，避免多个 "content" block 冲突。
func (g *Generator) renderToFile(pageTmplName string, data any, outPath string) error {
	tmpl, err := template.New("").Funcs(g.funcMap).ParseFS(
		templateFS,
		"templates/layout.html",
		"templates/"+pageTmplName,
	)
	if err != nil {
		return fmt.Errorf("parse templates for %s: %w", pageTmplName, err)
	}

	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer f.Close()

	if err := tmpl.ExecuteTemplate(f, "layout.html", data); err != nil {
		return fmt.Errorf("execute template %s: %w", pageTmplName, err)
	}

	return nil
}

// copyStaticAssets 将嵌入的静态资源复制到输出目录。
func (g *Generator) copyStaticAssets() error {
	staticDir := filepath.Join(g.outputDir, "static")
	if err := os.MkdirAll(staticDir, 0o755); err != nil {
		return fmt.Errorf("create static directory: %w", err)
	}

	return fs.WalkDir(staticFS, "static", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 计算相对于 "static" 的路径
		relPath, _ := filepath.Rel("static", path)
		outPath := filepath.Join(staticDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(outPath, 0o755)
		}

		content, err := staticFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read embedded file %s: %w", path, err)
		}

		if err := os.WriteFile(outPath, content, 0o644); err != nil {
			return fmt.Errorf("write static file %s: %w", outPath, err)
		}

		return nil
	})
}

// generateRobotsTxt 生成 robots.txt 文件，允许搜索引擎抓取。
func (g *Generator) generateRobotsTxt() error {
	content := `User-agent: *
Allow: /
Sitemap: https://www.javey.org/ai-daily/sitemap.xml
`
	outPath := filepath.Join(g.outputDir, "robots.txt")
	if err := os.WriteFile(outPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write robots.txt: %w", err)
	}
	slog.Info("generated robots.txt", "path", outPath)
	return nil
}

// ============================
// 模板函数
// ============================

// categoryColorMap 分类名到 CSS class 的映射。
var categoryColorMap = map[string]string{
	model.CategoryModelFrontier: "cat-blue",   // 🧠 模型前沿 → 蓝色（科技感）
	model.CategoryProduct:       "cat-green",  // 🚀 产品与应用 → 绿色（活力）
	model.CategoryInsight:       "cat-pink",   // 📊 深度洞察 → 粉色（温度）
	model.CategoryCloud:         "cat-teal",   // ☁️ 云服务与平台 → 青绿色（云端）
	model.CategoryAIEng:         "cat-purple", // ⚙️ AI工程 → 紫色（工程）
	model.CategoryInfra:         "cat-cyan",   // 🔧 AI基础设施 → 青色（底层）
	model.CategoryBiz:           "cat-orange", // 💰 商业与投资 → 橙色（商务）
	model.CategorySafety:        "cat-red",    // 🛡️ AI安全 → 红色（警示）
}

// catColorClass 返回分类对应的 CSS 颜色 class。
func catColorClass(category string) string {
	if class, ok := categoryColorMap[category]; ok {
		return class
	}
	return "cat-blue"
}

// catEmoji 返回分类对应的 emoji。
func catEmoji(category string) string {
	if emoji, ok := model.CategoryEmoji[category]; ok {
		return emoji
	}
	return "📰"
}

// displayTitle 返回文章的展示标题，优先中文标题。
func displayTitle(a model.Article) string {
	if a.ChineseTitle != "" {
		return a.ChineseTitle
	}
	return a.OriginalTitle
}

// englishTitle 返回文章的英文标题（当中文标题存在时返回原始标题）。
func englishTitle(a model.Article) string {
	if a.ChineseTitle != "" && a.OriginalTitle != "" && a.ChineseTitle != a.OriginalTitle {
		return a.OriginalTitle
	}
	return ""
}

// splitTags 将逗号分隔的标签字符串拆分为 slice，最多返回 max 个。
// 同时支持中文逗号（，）和英文逗号（,）作为分隔符。
func splitTags(tags string, max int) []string {
	if tags == "" {
		return nil
	}
	// 将中文逗号统一替换为英文逗号
	tags = strings.ReplaceAll(tags, "，", ",")
	parts := strings.Split(tags, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	if len(result) > max {
		result = result[:max]
	}
	return result
}

// slice3 返回字符串切片的前 3 个元素。
func slice3(tags []string) []string {
	if len(tags) <= 3 {
		return tags
	}
	return tags[:3]
}

// isCategoryTag 判断一个 tag 是否为已知分类名称。
func isCategoryTag(tag string) bool {
	return knownCategories[tag]
}

// topTags 返回 TagStats 的前 max 个（已按 count 降序排列）。
func topTags(tags []model.TagStat, max int) []model.TagStat {
	if len(tags) <= max {
		return tags
	}
	return tags[:max]
}

// ============================
// 数据处理辅助函数
// ============================

// knownCategories 已知的分类名称集合，用于从 tags 中识别分类标签。
var knownCategories = map[string]bool{
	model.CategoryModelFrontier: true,
	model.CategoryProduct:       true,
	model.CategoryInsight:       true,
	model.CategoryCloud:         true,
	model.CategoryAIEng:         true,
	model.CategoryInfra:         true,
	model.CategoryBiz:           true,
	model.CategorySafety:        true,
}

// extractCategoryStats 从索引的 tags 中提取分类出现次数统计。
func extractCategoryStats(days []model.DailyIndex) []archiveCategoryStat {
	countMap := make(map[string]int)
	for _, day := range days {
		for _, tag := range day.Tags {
			if knownCategories[tag] {
				countMap[tag]++
			}
		}
	}

	// 按 AllCategories 顺序输出，保证展示一致
	var stats []archiveCategoryStat
	for _, cat := range model.AllCategories {
		if cnt, ok := countMap[cat]; ok && cnt > 0 {
			stats = append(stats, archiveCategoryStat{
				Name:       cat,
				Emoji:      catEmoji(cat),
				ColorClass: catColorClass(cat),
				Count:      cnt,
			})
		}
	}
	return stats
}

// groupByMonth 将日报索引按月份分组。
func groupByMonth(days []model.DailyIndex) []archiveMonth {
	if len(days) == 0 {
		return nil
	}

	// 按日期降序排列
	sort.Slice(days, func(i, j int) bool {
		return days[i].Date > days[j].Date
	})

	monthMap := make(map[string]*archiveMonth)
	var monthOrder []string

	for _, day := range days {
		// 从 "2006-01-02" 中提取月份 "2006-01"
		if len(day.Date) < 7 {
			continue
		}
		monthKey := day.Date[:7]

		if _, ok := monthMap[monthKey]; !ok {
			// 解析月份标签，如 "2026年03月"
			t, err := time.Parse("2006-01", monthKey)
			if err != nil {
				continue
			}
			label := fmt.Sprintf("%d年%02d月", t.Year(), t.Month())
			monthMap[monthKey] = &archiveMonth{Label: label}
			monthOrder = append(monthOrder, monthKey)
		}

		monthMap[monthKey].Days = append(monthMap[monthKey].Days, day)
	}

	var result []archiveMonth
	for _, key := range monthOrder {
		result = append(result, *monthMap[key])
	}

	return result
}

// buildSearchData 从所有日报中提取搜索数据。
func buildSearchData(reports []*model.DailyReport) []searchItem {
	var items []searchItem

	for _, report := range reports {
		for _, article := range report.Articles {
			items = append(items, searchItem{
				CT: article.ChineseTitle,
				OT: article.OriginalTitle,
				SM: article.Summary,
				TG: article.Tags,
				CA: article.Category,
				SC: article.Source,
				DT: report.Date,
			})
		}
	}

	return items
}
