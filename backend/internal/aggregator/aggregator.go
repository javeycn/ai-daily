// Package aggregator 负责将采集到的文章聚合成每日日报。
package aggregator

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"ai-news-crawler/internal/config"
	"ai-news-crawler/internal/model"
	"ai-news-crawler/internal/repository"
	"ai-news-crawler/internal/summarizer"
)

// Aggregator 聚合文章生成每日日报。
type Aggregator struct {
	repo       repository.ArticleRepository
	summarizer *summarizer.Summarizer
	cfg        *config.AggregatorConfig
}

// New 创建一个新的 Aggregator 实例。
func New(repo repository.ArticleRepository, summarizer *summarizer.Summarizer, cfg *config.AggregatorConfig) *Aggregator {
	return &Aggregator{
		repo:       repo,
		summarizer: summarizer,
		cfg:        cfg,
	}
}

// Aggregate 生成指定日期的日报。
// 采用分级处理策略：先预选候选文章 → 只对候选调用 LLM 摘要 → 最终采样，
// 避免对全库所有文章都做昂贵的 LLM 调用。
func (a *Aggregator) Aggregate(ctx context.Context, date time.Time) (*model.DailyReport, error) {
	end := time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 0, date.Location())
	start := end.AddDate(0, 0, -3)

	articles, err := a.repo.GetByDate(ctx, start, end)
	if err != nil {
		return nil, fmt.Errorf("get articles for %s: %w", start.Format("2006-01-02"), err)
	}

	slog.Info("fetched articles from db", "date", start.Format("2006-01-02"), "count", len(articles))

	if len(articles) == 0 {
		return nil, fmt.Errorf("no articles found for %s", start.Format("2006-01-02"))
	}

	// === 分级处理策略 ===
	// 第一级：从全量文章中预选候选（纯本地计算，按来源优先级+时间排序）
	// 候选数量 = 最终需要数 × 3 倍冗余（默认 90 篇），确保 LLM 摘要后有足够高质量文章
	candidateCount := a.cfg.MaxDailyArticles * 3
	if candidateCount < 60 {
		candidateCount = 60
	}
	candidates := a.preSelect(articles, candidateCount)
	slog.Info("pre-selected candidates for LLM summarization",
		"total_articles", len(articles),
		"candidates", len(candidates),
	)

	// 第二级：只对候选文章中未生成摘要的调用 LLM
	a.summarizeCandidates(ctx, candidates)

	// 重新查询完整数据（包含摘要）
	articles, err = a.repo.GetByDate(ctx, start, end)
	if err != nil {
		return nil, fmt.Errorf("re-fetch articles: %w", err)
	}

	// === 后置重分类：将含国内云厂商关键词但被 LLM 错误归类的文章纠正回"云服务与平台" ===
	reclassified := a.reclassifyDomesticCloudArticles(ctx, articles)
	if reclassified > 0 {
		slog.Info("domestic cloud reclassification completed", "reclassified", reclassified)
	}

	// 第三级：最终多源均衡采样，选出首页精选的 30 篇日报文章
	featured := a.diverseSample(articles, a.cfg.MaxDailyArticles)
	featuredCount := len(featured)

	// 收集所有已生成摘要的文章（用于归档全量展示）
	var allSummarized []*model.Article
	featuredIDs := make(map[string]bool)
	for _, fa := range featured {
		featuredIDs[fa.ID] = true
	}
	for _, art := range articles {
		if art.ChineseTitle != "" && !featuredIDs[art.ID] {
			allSummarized = append(allSummarized, art)
		}
	}

	// 合并：精选文章在前 + 其余已摘要文章在后
	// 精选文章保持 diverseSample 的排序（按发布时间），其余按发布时间排序
	sort.Slice(allSummarized, func(i, j int) bool {
		return allSummarized[i].PublishedAt.After(allSummarized[j].PublishedAt)
	})
	allArticles := make([]*model.Article, 0, len(featured)+len(allSummarized))
	allArticles = append(allArticles, featured...)
	allArticles = append(allArticles, allSummarized...)

	// 统计标签（基于精选文章，保持日报摘要的聚焦性）
	tagStats := a.computeTagStats(featured)

	// 生成日报概要（基于精选文章）
	reportSummary := a.generateReportSummary(featured, date)

	// 转换为值类型切片
	articleValues := make([]model.Article, len(allArticles))
	for i, ap := range allArticles {
		articleValues[i] = *ap
	}

	// 精选文章转换为值类型（用于首页精选分组）
	featuredValues := make([]model.Article, len(featured))
	for i, ap := range featured {
		featuredValues[i] = *ap
	}

	// 按分类分组：全量（日报详情页）+ 精选（首页）
	categoryGroups := a.buildCategoryGroups(articleValues)
	featuredCategoryGroups := a.buildCategoryGroups(featuredValues)

	report := &model.DailyReport{
		Date:                   date.Format("2006-01-02"),
		Title:                  fmt.Sprintf("AI 日报 - %s", date.Format("2006年01月02日")),
		Summary:                reportSummary,
		TotalCount:             len(allArticles),
		FeaturedCount:          featuredCount,
		TagStats:               tagStats,
		Articles:               articleValues,
		CategoryGroups:         categoryGroups,
		FeaturedCategoryGroups: featuredCategoryGroups,
		PublishedAt:            time.Now().Format("2006-01-02 15:04:05"),
	}

	slog.Info("daily report generated", "date", report.Date, "featured", featuredCount, "total", report.TotalCount)
	return report, nil
}

// buildCategoryGroups 按分类对文章进行分组。
func (a *Aggregator) buildCategoryGroups(articles []model.Article) []model.CategoryGroup {
	grouped := make(map[string][]model.Article)
	for _, article := range articles {
		cat := article.Category
		if cat == "" {
			cat = model.CategoryModelFrontier
		}
		grouped[cat] = append(grouped[cat], article)
	}

	var groups []model.CategoryGroup
	for _, cat := range model.AllCategories {
		arts, exists := grouped[cat]
		if !exists || len(arts) == 0 {
			continue
		}

		// 云服务与平台分类：多级排序 — 头部国内大厂 > 其他国内 > 中性 > 国际
		// 全量视图中也严格执行"国内 ≥ 国际"的数量限制
		if cat == model.CategoryCloud {
			sort.SliceStable(arts, func(i, j int) bool {
				iRank := cloudSortRank(&arts[i])
				jRank := cloudSortRank(&arts[j])
				if iRank != jRank {
					return iRank < jRank // rank 越小越靠前
				}
				if arts[i].ImportanceScore != arts[j].ImportanceScore {
					return arts[i].ImportanceScore > arts[j].ImportanceScore // 高分优先
				}
				return arts[i].PublishedAt.After(arts[j].PublishedAt)
			})

			// 全量视图中也限制国际文章数量：国际文章 ≤ 国内文章
			arts = enforceCloudDomesticRatio(arts)
		}

		// 产品与应用分类：热门国内产品 > 普通国内 > 国际，同级按重要度+时间倒序
		if cat == model.CategoryProduct {
			sort.SliceStable(arts, func(i, j int) bool {
				iRank := productSortRank(&arts[i])
				jRank := productSortRank(&arts[j])
				if iRank != jRank {
					return iRank < jRank
				}
				if arts[i].ImportanceScore != arts[j].ImportanceScore {
					return arts[i].ImportanceScore > arts[j].ImportanceScore
				}
				return arts[i].PublishedAt.After(arts[j].PublishedAt)
			})
		}

		// 其他分类：按重要度+时间倒序
		if cat != model.CategoryCloud && cat != model.CategoryProduct {
			sort.SliceStable(arts, func(i, j int) bool {
				if arts[i].ImportanceScore != arts[j].ImportanceScore {
					return arts[i].ImportanceScore > arts[j].ImportanceScore
				}
				return arts[i].PublishedAt.After(arts[j].PublishedAt)
			})
		}

		emoji := model.CategoryEmoji[cat]
		groups = append(groups, model.CategoryGroup{
			Category: cat,
			Emoji:    emoji,
			Articles: arts,
		})
	}

	return groups
}

// cloudSortRank 返回云服务文章的排序等级（越小越靠前）。
// 0: 头部国内大厂（腾讯/阿里/火山引擎/百度/华为）
// 1: 其他国内云厂商
// 2: 中性文章（未明确匹配国内或国际关键词）
// 3: 国际云厂商
func cloudSortRank(a *model.Article) int {
	if a.IsTopDomesticCloud() {
		return 0
	}
	if a.IsDomesticCloud() {
		return 1
	}
	if a.IsInternationalCloud() {
		return 3
	}
	return 2 // 中性文章（未明确匹配关键词）排在国内后、国际前
}

// enforceCloudDomesticRatio 在全量视图中限制国际云服务文章数量。
// 规则：国际文章条数 ≤ 国内文章条数，如果超出则截断多余的国际文章（保留排序后靠前的）。
// 这确保了即使在全量（非精选）视图中，国内内容也不会被国际内容淹没。
func enforceCloudDomesticRatio(arts []model.Article) []model.Article {
	domesticCount := 0
	internationalCount := 0
	for _, art := range arts {
		if art.IsInternationalCloud() {
			internationalCount++
		} else {
			domesticCount++
		}
	}

	// 国际文章不超过国内文章数量
	if internationalCount <= domesticCount {
		return arts // 比例已经满足
	}

	// 需要截断国际文章
	maxIntl := domesticCount
	if maxIntl < 2 {
		maxIntl = 2 // 至少保留 2 条国际文章
	}

	var result []model.Article
	intlAdded := 0
	for _, art := range arts {
		if art.IsInternationalCloud() {
			if intlAdded >= maxIntl {
				continue // 跳过多余的国际文章
			}
			intlAdded++
		}
		result = append(result, art)
	}

	slog.Info("cloud domestic ratio enforced in full view",
		"original_total", len(arts),
		"domestic", domesticCount,
		"international_before", internationalCount,
		"international_after", intlAdded,
		"result_total", len(result),
	)

	return result
}

// productSortRank 返回产品与应用文章的排序等级（越小越靠前）。
// 0: 热门国内大厂产品 1: 普通国内产品 2: 热门国际产品 3: 其他国际产品
func productSortRank(a *model.Article) int {
	isDomestic := a.IsDomesticProduct()
	isHot := a.IsHotProduct()
	if isDomestic && a.IsTopDomesticProduct() {
		return 0
	}
	if isDomestic {
		return 1
	}
	if isHot {
		return 2
	}
	return 3
}

// knownCategorySet 已知分类名称集合，用于从标签中排除分类名称。
var knownCategorySet = map[string]bool{
	model.CategoryModelFrontier: true,
	model.CategoryProduct:       true,
	model.CategoryInsight:       true,
	model.CategoryCloud:         true,
	model.CategoryAIEng:         true,
	model.CategoryInfra:         true,
	model.CategoryBiz:           true,
	model.CategorySafety:        true,
}

// computeTagStats 统计文章标签分布（排除分类名称，只统计纯标签）。
func (a *Aggregator) computeTagStats(articles []*model.Article) []model.TagStat {
	tagCount := make(map[string]int)

	for _, article := range articles {
		if article.Tags == "" {
			continue
		}
		// 将中文逗号统一替换为英文逗号
		normalized := strings.ReplaceAll(article.Tags, "，", ",")
		tags := strings.Split(normalized, ",")
		for _, tag := range tags {
			tag = strings.TrimSpace(tag)
			if tag == "" {
				continue
			}
			// 排除分类名称，标签和分类是独立维度
			if knownCategorySet[tag] {
				continue
			}
			tagCount[tag]++
		}
	}

	var stats []model.TagStat
	for tag, count := range tagCount {
		stats = append(stats, model.TagStat{Tag: tag, Count: count})
	}

	sort.Slice(stats, func(i, j int) bool {
		if stats[i].Count != stats[j].Count {
			return stats[i].Count > stats[j].Count
		}
		return stats[i].Tag < stats[j].Tag
	})

	return stats
}

// generateReportSummary 生成日报概要描述。
func (a *Aggregator) generateReportSummary(articles []*model.Article, date time.Time) string {
	if len(articles) == 0 {
		return ""
	}

	// 统计来源
	sources := make(map[string]int)
	for _, article := range articles {
		sources[article.Source]++
	}
	var sourceNames []string
	for s := range sources {
		sourceNames = append(sourceNames, s)
	}

	return fmt.Sprintf(
		"今日共收录 %d 条 AI 资讯，来源包括 %s。",
		len(articles),
		strings.Join(sourceNames, "、"),
	)
}

// diverseSample 从全部文章中均衡采样，确保多源、多分类、国内/国际均衡的多样性。
// 策略（三阶段软性配额 + 分类级均衡策略 + 同厂商去重）：
// 1. 第一阶段（分类保底）：遍历 8 个分类，每个分类按特定均衡策略取最低保底数量
//   - 云服务与平台：国内优先（国内:国际 ≥ 3:2），同厂商最多 2 条
//   - 模型前沿：国内/国际 1:1 均衡，头部厂商优先，同厂商最多 2 条
//   - 产品与应用：国内/国际 1:1 均衡，同厂商最多 2 条
//   - 其他分类：按来源优先级+时间排序选取，同厂商最多 2 条
//
// 2. 第二阶段（优先补位）：剩余配额优先补齐未达 Preferred 的分类
// 3. 第三阶段（自由补位）：剩余名额按来源优先级填满
// 4. 最终按发布时间排序
func (a *Aggregator) diverseSample(articles []*model.Article, maxCount int) []*model.Article {
	if len(articles) <= maxCount {
		sort.Slice(articles, func(i, j int) bool {
			return articles[i].PublishedAt.After(articles[j].PublishedAt)
		})
		return articles
	}

	// 只保留已生成摘要的文章
	var summarized []*model.Article
	for _, a := range articles {
		if a.ChineseTitle != "" {
			summarized = append(summarized, a)
		}
	}
	if len(summarized) == 0 {
		summarized = articles
	}

	// 按分类分组
	byCategory := make(map[string][]*model.Article)
	for _, article := range summarized {
		cat := article.Category
		if cat == "" {
			cat = model.CategoryModelFrontier
		}
		byCategory[cat] = append(byCategory[cat], article)
	}

	// 每个分类内部按来源优先级+重要度+时间排序（高优先级来源 + 高分 + 更新的文章排在前面）
	bySource := make(map[string][]*model.Article)
	for _, article := range summarized {
		bySource[article.Source] = append(bySource[article.Source], article)
	}
	sourceOrder := a.rankSources(bySource)
	sourcePriority := make(map[string]int)
	for i, s := range sourceOrder {
		sourcePriority[s] = i
	}
	for cat := range byCategory {
		arts := byCategory[cat]
		sort.Slice(arts, func(i, j int) bool {
			pi := sourcePriority[arts[i].Source]
			pj := sourcePriority[arts[j].Source]
			if pi != pj {
				return pi < pj
			}
			if arts[i].ImportanceScore != arts[j].ImportanceScore {
				return arts[i].ImportanceScore > arts[j].ImportanceScore // 高分优先
			}
			return arts[i].PublishedAt.After(arts[j].PublishedAt)
		})
	}

	selected := make(map[string]bool)
	var result []*model.Article

	// 每个分类内的厂商计数器（用于同厂商去重限制）
	catVendorCount := make(map[string]map[string]int) // category -> vendor -> count

	// canPickByVendor 检查该文章的厂商是否已达上限
	canPickByVendor := func(art *model.Article, cat string) bool {
		vendor := art.ExtractVendor()
		if vendor == "" {
			return true // 未识别厂商，不限制
		}
		if catVendorCount[cat] == nil {
			return true
		}
		return catVendorCount[cat][vendor] < model.MaxPerVendorInCategory
	}

	// recordVendor 记录厂商选取
	recordVendor := func(art *model.Article, cat string) {
		vendor := art.ExtractVendor()
		if vendor == "" {
			return
		}
		if catVendorCount[cat] == nil {
			catVendorCount[cat] = make(map[string]int)
		}
		catVendorCount[cat][vendor]++
	}

	// pickArticle 选取一篇文章并记录
	pickArticle := func(art *model.Article, cat string) {
		result = append(result, art)
		selected[art.ID] = true
		recordVendor(art, cat)
	}

	// === 第一阶段：分类保底（带分类级均衡策略）===
	for _, cat := range model.AllCategories {
		quota, ok := model.CategoryQuotas[cat]
		if !ok {
			continue
		}
		arts := byCategory[cat]
		picked := 0

		switch cat {
		case model.CategoryCloud:
			// 云服务分类：头部国内优先（腾讯/阿里/火山引擎），国内:国际 ≥ 3:2，同厂商限 2 条
			// 策略：国内必须占 60%+，国际最多 2 条（硬上限）
			// 选取顺序：头部国内云 → 普通国内云 → 国际云
			internationalHardMax := 2 // 国际绝对硬上限

			// 第一步：优先选头部国内云厂商（腾讯云、阿里云、火山引擎系）
			domesticPicked := 0
			for _, art := range arts {
				if picked >= quota.Min {
					break
				}
				if !selected[art.ID] && art.IsTopDomesticCloud() && canPickByVendor(art, cat) {
					pickArticle(art, cat)
					picked++
					domesticPicked++
				}
			}
			// 第二步：选其他国内云服务文章
			for _, art := range arts {
				if picked >= quota.Min {
					break
				}
				if !selected[art.ID] && art.IsDomesticCloud() && canPickByVendor(art, cat) {
					pickArticle(art, cat)
					picked++
					domesticPicked++
				}
			}
			// 从国内源的云服务文章补充
			for _, art := range arts {
				if picked >= quota.Min {
					break
				}
				if !selected[art.ID] && model.DomesticCloudSources[art.Source] && canPickByVendor(art, cat) {
					pickArticle(art, cat)
					picked++
					domesticPicked++
				}
			}

			// 第三步：选国际云服务文章（严格限制硬上限 2 条）
			// 只有当国内已有足够数量保证 ≥ 60% 时才补国际
			internationalPicked := 0
			for _, art := range arts {
				if internationalPicked >= internationalHardMax || picked >= quota.Min {
					break
				}
				// 动态检查比例：加了这条国际后国内占比必须 ≥ 60%
				if domesticPicked*5 < (picked+1)*3 {
					break
				}
				if !selected[art.ID] && art.IsInternationalCloud() && canPickByVendor(art, cat) {
					pickArticle(art, cat)
					picked++
					internationalPicked++
				}
			}

			// 兜底：如果还不够保底配额
			// 先放宽厂商限制补头部国内
			if picked < quota.Min {
				for _, art := range arts {
					if picked >= quota.Min {
						break
					}
					if !selected[art.ID] && art.IsTopDomesticCloud() {
						pickArticle(art, cat)
						picked++
						domesticPicked++
					}
				}
			}
			// 放宽厂商限制补国内
			if picked < quota.Min {
				for _, art := range arts {
					if picked >= quota.Min {
						break
					}
					if !selected[art.ID] && art.IsDomesticCloud() {
						pickArticle(art, cat)
						picked++
						domesticPicked++
					}
				}
			}
			// 再从国内源补充（不限关键词匹配，只看来源）
			if picked < quota.Min {
				for _, art := range arts {
					if picked >= quota.Min {
						break
					}
					if !selected[art.ID] && model.DomesticCloudSources[art.Source] {
						pickArticle(art, cat)
						picked++
						domesticPicked++
					}
				}
			}
			// 将"非国际"的文章视为中性/国内，优先补充
			if picked < quota.Min {
				for _, art := range arts {
					if picked >= quota.Min {
						break
					}
					if !selected[art.ID] && !art.IsInternationalCloud() {
						pickArticle(art, cat)
						picked++
						domesticPicked++
					}
				}
			}
			// 最终兜底：只有在比例允许时才加国际
			if picked < quota.Min {
				for _, art := range arts {
					if picked >= quota.Min {
						break
					}
					if selected[art.ID] {
						continue
					}
					// 国际文章：检查比例 + 硬上限
					if domesticPicked*5 < (picked+1)*3 || internationalPicked >= internationalHardMax {
						continue
					}
					pickArticle(art, cat)
					picked++
					internationalPicked++
				}
			}

			slog.Info("cloud category balanced",
				"domestic", domesticPicked,
				"international", internationalPicked,
				"total", picked,
			)

		case model.CategoryModelFrontier:
			// 模型前沿：国内/国际 1:1 均衡，头部厂商优先，同厂商限 2 条
			halfQuota := quota.Min / 2
			if halfQuota < 1 {
				halfQuota = 1
			}

			// 先选国内头部厂商
			domesticPicked := 0
			for _, art := range arts {
				if domesticPicked >= halfQuota {
					break
				}
				if !selected[art.ID] && art.IsTopDomesticModel() && canPickByVendor(art, cat) {
					pickArticle(art, cat)
					picked++
					domesticPicked++
				}
			}
			// 不够则从所有国内模型文章补
			for _, art := range arts {
				if domesticPicked >= halfQuota {
					break
				}
				if !selected[art.ID] && art.IsDomesticModel() && canPickByVendor(art, cat) {
					pickArticle(art, cat)
					picked++
					domesticPicked++
				}
			}

			// 再选国际头部厂商
			internationalPicked := 0
			internationalTarget := quota.Min - picked
			for _, art := range arts {
				if internationalPicked >= internationalTarget {
					break
				}
				if !selected[art.ID] && art.IsTopInternationalModel() && canPickByVendor(art, cat) {
					pickArticle(art, cat)
					picked++
					internationalPicked++
				}
			}
			// 不够则从所有国际模型文章补
			for _, art := range arts {
				if internationalPicked >= internationalTarget || picked >= quota.Min {
					break
				}
				if !selected[art.ID] && art.IsInternationalModel() && canPickByVendor(art, cat) {
					pickArticle(art, cat)
					picked++
					internationalPicked++
				}
			}

			// 仍不够则不限地域补满
			for _, art := range arts {
				if picked >= quota.Min {
					break
				}
				if !selected[art.ID] && canPickByVendor(art, cat) {
					pickArticle(art, cat)
					picked++
				}
			}

			slog.Info("model frontier balanced",
				"domestic", domesticPicked,
				"international", internationalPicked,
				"total", picked,
			)

		case model.CategoryProduct:
			// 产品与应用：国内大厂热门产品优先 + Agent 优先 + 国内/国际严格 1:1 均衡
			// 选取顺序：国内大厂热门产品 → Agent/热门产品 → 按 1:1 补齐国内/国际
			halfQuota := quota.Min / 2
			if halfQuota < 1 {
				halfQuota = 1
			}

			// 第一步：优先选国内大厂热门产品（腾讯/阿里/字节/火山引擎系）
			domesticPicked := 0
			for _, art := range arts {
				if domesticPicked >= halfQuota {
					break
				}
				if !selected[art.ID] && art.IsTopDomesticProduct() && canPickByVendor(art, cat) {
					pickArticle(art, cat)
					picked++
					domesticPicked++
				}
			}
			// 补充其他国内产品到半数配额
			for _, art := range arts {
				if domesticPicked >= halfQuota {
					break
				}
				if !selected[art.ID] && art.IsDomesticProduct() && canPickByVendor(art, cat) {
					pickArticle(art, cat)
					picked++
					domesticPicked++
				}
			}

			// 第二步：选国际产品（优先热门产品，保证 1:1）
			internationalPicked := 0
			internationalTarget := quota.Min - picked
			// 先选热门国际产品
			for _, art := range arts {
				if internationalPicked >= internationalTarget || picked >= quota.Min {
					break
				}
				if !selected[art.ID] && !art.IsDomesticProduct() && art.IsHotProduct() && canPickByVendor(art, cat) {
					pickArticle(art, cat)
					picked++
					internationalPicked++
				}
			}
			// 再选其他国际产品
			for _, art := range arts {
				if internationalPicked >= internationalTarget || picked >= quota.Min {
					break
				}
				if !selected[art.ID] && !art.IsDomesticProduct() && canPickByVendor(art, cat) {
					pickArticle(art, cat)
					picked++
					internationalPicked++
				}
			}

			// 仍不够补满
			for _, art := range arts {
				if picked >= quota.Min {
					break
				}
				if !selected[art.ID] && canPickByVendor(art, cat) {
					pickArticle(art, cat)
					picked++
				}
			}

			slog.Info("product category balanced",
				"top_domestic", domesticPicked,
				"international", internationalPicked,
				"total", picked,
			)

		default:
			// 其他分类：按来源优先级+时间排序选取，同厂商限 2 条
			for _, art := range arts {
				if picked >= quota.Min {
					break
				}
				if !selected[art.ID] && canPickByVendor(art, cat) {
					pickArticle(art, cat)
					picked++
				}
			}
			// 如果因厂商限制选不够，放宽限制
			for _, art := range arts {
				if picked >= quota.Min {
					break
				}
				if !selected[art.ID] {
					pickArticle(art, cat)
					picked++
				}
			}
		}

		slog.Info("category quota guaranteed",
			"category", cat,
			"guaranteed", picked,
			"available", len(arts),
			"min_quota", quota.Min,
		)
	}

	// 统计第一阶段每个分类已选数量
	catCount := make(map[string]int)
	for _, art := range result {
		cat := art.Category
		if cat == "" {
			cat = model.CategoryModelFrontier
		}
		catCount[cat]++
	}

	// === 第二阶段：优先补位（未达 Preferred 的分类优先补齐）===
	// 遍历各分类，对尚未达到 Preferred 上限的分类按来源优先级补充
	// 对模型前沿/产品与应用/云服务继续执行均衡策略
	for _, cat := range model.AllCategories {
		if len(result) >= maxCount {
			break
		}
		quota, ok := model.CategoryQuotas[cat]
		if !ok {
			continue
		}
		if catCount[cat] >= quota.Preferred {
			continue
		}
		arts := byCategory[cat]
		remaining := quota.Preferred - catCount[cat]

		switch cat {
		case model.CategoryCloud:
			// 继续头部国内优先策略补位，确保最终国内:国际 ≥ 3:2
			// 先统计当前已选的国内/国际数
			currentDomestic := 0
			currentIntl := 0
			for _, art := range result {
				if art.Category == model.CategoryCloud {
					if art.IsInternationalCloud() {
						currentIntl++
					} else {
						currentDomestic++
					}
				}
			}

			// 优先补头部国内云厂商
			domesticAdded := 0
			for _, art := range arts {
				if len(result) >= maxCount || catCount[cat] >= quota.Preferred {
					break
				}
				if !selected[art.ID] && art.IsTopDomesticCloud() && canPickByVendor(art, cat) {
					pickArticle(art, cat)
					catCount[cat]++
					domesticAdded++
					currentDomestic++
				}
			}
			// 补普通国内云文章（IsDomesticCloud 匹配）
			for _, art := range arts {
				if len(result) >= maxCount || catCount[cat] >= quota.Preferred {
					break
				}
				if !selected[art.ID] && art.IsDomesticCloud() && canPickByVendor(art, cat) {
					pickArticle(art, cat)
					catCount[cat]++
					domesticAdded++
					currentDomestic++
				}
			}
			// 从国内源补充
			for _, art := range arts {
				if len(result) >= maxCount || catCount[cat] >= quota.Preferred {
					break
				}
				if !selected[art.ID] && model.DomesticCloudSources[art.Source] && canPickByVendor(art, cat) {
					pickArticle(art, cat)
					catCount[cat]++
					domesticAdded++
					currentDomestic++
				}
			}
			// "非国际"文章视为中性/国内
			for _, art := range arts {
				if len(result) >= maxCount || catCount[cat] >= quota.Preferred {
					break
				}
				if !selected[art.ID] && !art.IsInternationalCloud() && canPickByVendor(art, cat) {
					pickArticle(art, cat)
					catCount[cat]++
					domesticAdded++
					currentDomestic++
				}
			}

			// 国际补位：总国际硬上限 2 条 + 比例检查
			intlAdded := 0
			for _, art := range arts {
				if len(result) >= maxCount || catCount[cat] >= quota.Preferred {
					break
				}
				// 国际总数硬上限 2 条
				if currentIntl+intlAdded >= 2 {
					break
				}
				// 动态检查：补了这一条国际后，国内占比是否仍 ≥ 60%
				if currentDomestic*5 < (currentDomestic+currentIntl+intlAdded+1)*3 {
					break
				}
				if !selected[art.ID] && art.IsInternationalCloud() && canPickByVendor(art, cat) {
					pickArticle(art, cat)
					catCount[cat]++
					intlAdded++
				}
			}
			slog.Info("cloud phase2 fillup",
				"domestic_added", domesticAdded,
				"intl_added", intlAdded,
				"total_domestic", currentDomestic,
				"total_intl", currentIntl+intlAdded,
			)

		case model.CategoryModelFrontier:
			// 继续 1:1 均衡补位
			halfRemaining := remaining / 2
			if halfRemaining < 1 {
				halfRemaining = 1
			}
			// 补国内
			dAdded := 0
			for _, art := range arts {
				if len(result) >= maxCount || dAdded >= halfRemaining {
					break
				}
				if !selected[art.ID] && art.IsDomesticModel() && canPickByVendor(art, cat) {
					pickArticle(art, cat)
					catCount[cat]++
					dAdded++
				}
			}
			// 补国际
			iAdded := 0
			iTarget := remaining - dAdded
			for _, art := range arts {
				if len(result) >= maxCount || catCount[cat] >= quota.Preferred || iAdded >= iTarget {
					break
				}
				if !selected[art.ID] && art.IsInternationalModel() && canPickByVendor(art, cat) {
					pickArticle(art, cat)
					catCount[cat]++
					iAdded++
				}
			}

		case model.CategoryProduct:
			// 优先补国内大厂热门产品
			dAdded := 0
			for _, art := range arts {
				if len(result) >= maxCount || catCount[cat] >= quota.Preferred || dAdded >= remaining/2 {
					break
				}
				if !selected[art.ID] && art.IsTopDomesticProduct() && canPickByVendor(art, cat) {
					pickArticle(art, cat)
					catCount[cat]++
					dAdded++
				}
			}
			// 补其他国内产品
			for _, art := range arts {
				if len(result) >= maxCount || catCount[cat] >= quota.Preferred || dAdded >= remaining/2 {
					break
				}
				if !selected[art.ID] && art.IsDomesticProduct() && canPickByVendor(art, cat) {
					pickArticle(art, cat)
					catCount[cat]++
					dAdded++
				}
			}
			// 补国际产品（保持 1:1）
			iAdded := 0
			iTarget := remaining - dAdded
			for _, art := range arts {
				if len(result) >= maxCount || catCount[cat] >= quota.Preferred || iAdded >= iTarget {
					break
				}
				if !selected[art.ID] && !art.IsDomesticProduct() && canPickByVendor(art, cat) {
					pickArticle(art, cat)
					catCount[cat]++
					iAdded++
				}
			}

		default:
			for _, art := range arts {
				if len(result) >= maxCount || catCount[cat] >= quota.Preferred {
					break
				}
				if !selected[art.ID] && canPickByVendor(art, cat) {
					pickArticle(art, cat)
					catCount[cat]++
				}
			}
		}
	}

	// === 第三阶段：自由补位（剩余名额按来源优先级+时间，但尊重 Preferred 上限）===
	var remainingArts []*model.Article
	for _, art := range summarized {
		if !selected[art.ID] {
			remainingArts = append(remainingArts, art)
		}
	}
	sort.Slice(remainingArts, func(i, j int) bool {
		pi := sourcePriority[remainingArts[i].Source]
		pj := sourcePriority[remainingArts[j].Source]
		if pi != pj {
			return pi < pj
		}
		return remainingArts[i].PublishedAt.After(remainingArts[j].PublishedAt)
	})

	for _, art := range remainingArts {
		if len(result) >= maxCount {
			break
		}
		cat := art.Category
		if cat == "" {
			cat = model.CategoryModelFrontier
		}
		// 尊重 Preferred 上限，已达上限的分类跳过
		if quota, ok := model.CategoryQuotas[cat]; ok && catCount[cat] >= quota.Preferred {
			continue
		}
		// 云服务分类：国际文章需检查比例 + 硬上限 2 条
		if cat == model.CategoryCloud && art.IsInternationalCloud() {
			cloudDomestic := 0
			cloudIntl := 0
			for _, r := range result {
				if r.Category == model.CategoryCloud {
					if r.IsInternationalCloud() {
						cloudIntl++
					} else {
						cloudDomestic++
					}
				}
			}
			// 国际硬上限 2 条 或 补了这条国际后国内占比 < 60% 则跳过
			if cloudIntl >= 2 || cloudDomestic*5 < (cloudDomestic+cloudIntl+1)*3 {
				continue
			}
		}
		if canPickByVendor(art, cat) {
			pickArticle(art, cat)
			catCount[cat]++
		}
	}

	// 如果所有分类都达 Preferred 上限但还有剩余名额，不限分类补满
	for _, art := range remainingArts {
		if len(result) >= maxCount {
			break
		}
		if !selected[art.ID] {
			cat := art.Category
			if cat == "" {
				cat = model.CategoryModelFrontier
			}
			// 云服务分类国际文章仍然检查比例 + 硬上限
			if cat == model.CategoryCloud && art.IsInternationalCloud() {
				cloudDomestic := 0
				cloudIntl := 0
				for _, r := range result {
					if r.Category == model.CategoryCloud {
						if r.IsInternationalCloud() {
							cloudIntl++
						} else {
							cloudDomestic++
						}
					}
				}
				if cloudIntl >= 2 || cloudDomestic*5 < (cloudDomestic+cloudIntl+1)*3 {
					continue
				}
			}
			pickArticle(art, cat)
			catCount[cat]++
		}
	}

	// 按发布时间排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].PublishedAt.After(result[j].PublishedAt)
	})

	// 输出各分类最终统计
	for _, cat := range model.AllCategories {
		slog.Info("final category count",
			"category", cat,
			"count", catCount[cat],
			"preferred", model.CategoryQuotas[cat].Preferred,
		)
	}

	slog.Info("diverse sample completed",
		"total_candidates", len(summarized),
		"categories", len(byCategory),
		"sources", len(bySource),
		"selected", len(result),
	)

	return result
}

// rankSources 对来源按优先级排序，前沿 AI 源排前面。
// 优先级原则：
//   - 国内头部媒体提权（与国际专业媒体同级），确保国内内容有足够曝光
//   - AWS/Azure/GCP 官方博客适当降权（避免国际云服务内容霸占配额）
//   - 国际 AI 大厂官方博客和思想领袖保持最高优先级
func (a *Aggregator) rankSources(bySource map[string][]*model.Article) []string {
	// 定义来源优先级（数字越小优先级越高）
	priority := map[string]int{
		// Tier 1: AI 大厂官方 & 前沿思想领袖（国际头部模型厂商）
		"OpenAI Blog":       1,
		"DeepMind":          1,
		"Google Research":   1,
		"Anthropic":         1,
		"Microsoft AI":      3, // 降权：微软 AI 包含大量 Azure 云服务内容
		"Import AI":         1,
		"Simon Willison":    1,
		"Lilian Weng":       1,
		"Latent Space":      1,
		"The Gradient":      1,
		"HuggingFace":       1,
		"The Batch":         1,
		"BAIR Blog":         1,
		"Chip Huyen":        1,
		"Eugene Yan":        1,
		"Karpathy":          1,
		// Tier 2: 国际顶级科技媒体（不含云服务博客）
		"TechCrunch":            2,
		"The Verge":             2,
		"Wired":                 2,
		"MIT Technology Review": 2,
		"Ars Technica":          2,
		"IEEE Spectrum":         2,
		"NVIDIA Blog":           2,
		"InfoQ AI":              2,
		// Tier 3: 国内头部科技媒体 & 国际专业媒体 & 云服务博客
		// 国内头部媒体提权到与国际专业媒体同级，确保国内内容充足
		"机器之心":            3,
		"智东西":             3,
		"量子位":             3,
		"InfoQ中文":          3,
		"THE DECODER":       3,
		"MarkTechPost":      3,
		"Unite.AI":          3,
		"AI News":           3,
		"Synced Review":     3,
		"Meta Engineering":  3,
		"arXiv cs.AI":       3,
		"Papers With Code":  3,
		"AI Hub Today":      3,
		"DailyAI":           3,
		// Tier 4: 创投 & 科技商业 & 搜索补充 & 云服务博客（降权）
		"Hacker News":  4,
		"PanDaily":     4,
		"VentureBeat":  4,
		"Web Search":   4,
		// AWS/Azure/GCP 博客降到 Tier 4（降权，避免国际云服务内容挤占配额）
		"AWS ML Blog":   4,
		"GCP AI Blog":   4,
		"Azure AI Blog": 4,
		// Tier 5: 国内科技媒体（产量大，适当限制配额）
		"36氪":  5,
		"IT之家": 5,
	}

	var sources []string
	for source := range bySource {
		sources = append(sources, source)
	}

	sort.Slice(sources, func(i, j int) bool {
		pi := priority[sources[i]]
		pj := priority[sources[j]]
		if pi == 0 {
			pi = 5
		}
		if pj == 0 {
			pj = 5
		}
		if pi != pj {
			return pi < pj
		}
		return sources[i] < sources[j]
	})

	return sources
}

// maxArticlesPerSource 根据来源和总配额计算每个来源的最大贡献量。
func (a *Aggregator) maxArticlesPerSource(source string, totalSources, maxTotal int) int {
	// 高产量国内源限制配额，避免挤占其他来源
	highVolumeSources := map[string]bool{
		"IT之家": true,
		"36氪":  true,
	}

	if highVolumeSources[source] {
		// 高产量源最多占总配额的 15%
		limit := maxTotal * 15 / 100
		if limit < 2 {
			limit = 2
		}
		return limit
	}

	// 其他来源根据总来源数动态分配
	perSource := maxTotal / totalSources
	if perSource < 2 {
		perSource = 2
	}
	if perSource > 5 {
		perSource = 5
	}
	return perSource
}

// preSelect 从全量文章中预选候选文章，用于后续 LLM 摘要。
// 这是一步纯本地计算（无 LLM 调用），目标是将文章数从 1000+ 缩减到 ~90 篇。
// 策略：复用 diverseSample 的来源优先级 + 轮询逻辑，但取更大的候选池。
// 已有摘要的文章直接入选（不浪费之前的 LLM 投入）。
// 关键改进：
//  1. 即使已摘要文章充足，也保证从未摘要文章中选取一批高优先级新文章
//  2. 保底选取国内云服务相关文章，确保 diverseSample 有足够国内云内容可选
func (a *Aggregator) preSelect(articles []*model.Article, maxCandidates int) []*model.Article {
	if len(articles) <= maxCandidates {
		return articles
	}

	// 已有摘要的文章直接入选（它们已经消耗了 LLM 调用，不该浪费）
	var alreadySummarized []*model.Article
	var unsummarized []*model.Article
	for _, art := range articles {
		if art.ChineseTitle != "" {
			alreadySummarized = append(alreadySummarized, art)
		} else {
			unsummarized = append(unsummarized, art)
		}
	}

	slog.Info("pre-select breakdown",
		"already_summarized", len(alreadySummarized),
		"unsummarized", len(unsummarized),
	)

	// 即使已摘要文章充足，也从未摘要中按来源优先级选取一批新文章
	// 保证新采集的高质量内容能被 LLM 处理（至少选取 maxDailyArticles 篇新文章）
	minNewArticles := a.cfg.MaxDailyArticles
	if minNewArticles < 30 {
		minNewArticles = 30
	}

	var newPicked []*model.Article
	if len(unsummarized) > 0 {
		pickCount := maxCandidates - len(alreadySummarized)
		if pickCount < minNewArticles {
			pickCount = minNewArticles
		}
		if pickCount > len(unsummarized) {
			pickCount = len(unsummarized)
		}
		newPicked = a.pickBySourcePriority(unsummarized, pickCount)
		slog.Info("pre-select new articles for LLM",
			"unsummarized_total", len(unsummarized),
			"picked_for_llm", len(newPicked),
		)
	}

	result := make([]*model.Article, 0, len(alreadySummarized)+len(newPicked))
	result = append(result, alreadySummarized...)
	result = append(result, newPicked...)

	// 保底：确保国内云服务文章在候选池中有足够数量
	// 跨分类搜索：即使文章被 LLM 归入了其他分类，只要匹配国内云关键词也要加入候选池
	// 这样后置重分类步骤才能将它们纠正到云服务分类
	selectedIDs := make(map[string]bool)
	for _, art := range result {
		selectedIDs[art.ID] = true
	}
	domesticCloudCount := 0
	for _, art := range result {
		// 不再限定 art.Category == model.CategoryCloud，跨分类检测
		if art.IsDomesticCloud() || art.IsTopDomesticCloud() {
			domesticCloudCount++
		}
	}
	minDomesticCloud := 12 // 提高保底数量到 12 条（供后置重分类 + diverseSample 选取足够国内云文章）
	if domesticCloudCount < minDomesticCloud {
		for _, art := range articles {
			if domesticCloudCount >= minDomesticCloud {
				break
			}
			if !selectedIDs[art.ID] && (art.IsDomesticCloud() || art.IsTopDomesticCloud()) {
				result = append(result, art)
				selectedIDs[art.ID] = true
				domesticCloudCount++
			}
		}
		slog.Info("pre-select domestic cloud boost (cross-category)",
			"domestic_cloud_total", domesticCloudCount,
		)
	}

	return result
}

// pickBySourcePriority 从文章列表中按来源优先级+时间轮询选取指定数量。
func (a *Aggregator) pickBySourcePriority(articles []*model.Article, maxCount int) []*model.Article {
	if len(articles) <= maxCount {
		return articles
	}

	// 按来源分组
	bySource := make(map[string][]*model.Article)
	for _, art := range articles {
		bySource[art.Source] = append(bySource[art.Source], art)
	}

	// 每个来源内部按时间排序（最新在前）
	for source := range bySource {
		s := bySource[source]
		sort.Slice(s, func(i, j int) bool {
			return s[i].PublishedAt.After(s[j].PublishedAt)
		})
	}

	// 按优先级排序来源
	sourceOrder := a.rankSources(bySource)

	// 轮询选取
	selected := make(map[string]bool)
	var result []*model.Article
	sourceIdx := make(map[string]int)

	// 第一轮：每源取 1 篇
	for _, source := range sourceOrder {
		arts := bySource[source]
		if len(arts) > 0 && len(result) < maxCount {
			result = append(result, arts[0])
			selected[arts[0].ID] = true
			sourceIdx[source] = 1
		}
	}

	// 后续轮：继续轮询补充
	for len(result) < maxCount {
		added := false
		for _, source := range sourceOrder {
			if len(result) >= maxCount {
				break
			}
			arts := bySource[source]
			idx := sourceIdx[source]
			if idx < len(arts) && !selected[arts[idx].ID] {
				result = append(result, arts[idx])
				selected[arts[idx].ID] = true
				added = true
			}
			sourceIdx[source] = idx + 1
		}
		if !added {
			break
		}
	}

	return result
}

// summarizeCandidates 只对候选文章中尚未生成摘要的进行 LLM 摘要。
func (a *Aggregator) summarizeCandidates(ctx context.Context, candidates []*model.Article) {
	// 筛出需要摘要的候选
	var needSummary []*model.Article
	for _, art := range candidates {
		if art.ChineseTitle == "" {
			needSummary = append(needSummary, art)
		}
	}

	if len(needSummary) == 0 {
		slog.Info("all candidates already summarized, skipping LLM calls")
		return
	}

	slog.Info("summarizing candidates",
		"total_candidates", len(candidates),
		"need_summary", len(needSummary),
	)

	// 分批处理
	batchSize := 20
	totalSummarized := 0
	for i := 0; i < len(needSummary); i += batchSize {
		end := i + batchSize
		if end > len(needSummary) {
			end = len(needSummary)
		}
		batch := needSummary[i:end]

		slog.Info("generating summaries batch",
			"batch_size", len(batch),
			"total_done", totalSummarized,
			"remaining", len(needSummary)-i,
		)

		results := a.summarizer.SummarizeBatch(ctx, batch)
		successCount := 0
		for _, r := range results {
			if r.Error == nil {
				if err := a.repo.UpdateSummary(ctx, r.Article.ID, r.ChineseTitle, r.Summary, r.Tags, r.Category, r.ImportanceScore, r.Recommendation); err != nil {
					slog.Error("update summary failed", "article_id", r.Article.ID, "error", err)
				} else {
					successCount++
				}
			}
		}
		totalSummarized += successCount
		slog.Info("batch completed", "success", successCount, "total_done", totalSummarized)
	}

	slog.Info("candidate summarization completed", "total_summarized", totalSummarized)
}

// reclassifyDomesticCloudArticles 后置重分类：将含国内云厂商关键词但被 LLM 错误归类的文章纠正回"云服务与平台"。
// 仅针对以下情况进行纠正：
//  1. 文章标题/摘要包含头部国内云厂商关键词（腾讯云/阿里云/华为云/火山引擎/百度云等）
//  2. 文章当前分类不是"云服务与平台"
//  3. 文章当前分类也不是"模型前沿"（纯模型能力文章不强制纠正）
//  4. 文章不是纯融资/IPO 等商业事件
//
// 策略：优先纠正被归入"商业与投资""产品与应用""深度洞察"等分类的国内云厂商平台级内容。
// 对于"模型前沿"分类的文章，只有同时匹配云平台关键词（如"API""平台""服务"）时才纠正。
func (a *Aggregator) reclassifyDomesticCloudArticles(ctx context.Context, articles []*model.Article) int {
	// 云平台相关辅助关键词，用于区分"纯模型技术"和"平台/服务动态"
	cloudPlatformHints := []string{
		"云", "平台", "服务", "api", "开放", "上线", "发布", "更新",
		"降价", "免费", "调价", "算力", "实例", "gpu", "托管",
		"maas", "paas", "saas", "智能体平台", "模型商店",
		"生态", "合作", "开发者", "百炼", "千帆", "盘古",
		"混元", "tokenhub", "火山方舟",
	}

	// 纯商业关键词（如果文章核心是融资/并购，不应纠正到云服务）
	pureBizKeywords := []string{
		"融资", "估值", "ipo", "上市", "并购", "收购",
		"投资", "亿美元", "亿元", "series", "a轮", "b轮", "c轮",
	}

	// 非云服务场景排除关键词（汽车/硬件/手机/消费电子等）
	// 匹配到这些关键词的文章不应被纠正到云服务分类
	nonCloudExcludeKeywords := []string{
		"智行", "问界", "尚界", "享界", "智界", "鸿蒙座舱", "鸿蒙智行",
		"乾崑", "交付", "试驾", "上市售价", "预售", "车型",
		"激光雷达", "自动驾驶", "辅助驾驶", "ads ", "座椅",
		"手机", "平板", "笔记本", "电视", "穿戴", "耳机",
		"广汽", "丰田", "比亚迪", "蔚来", "小鹏", "理想", "极氪",
	}

	reclassified := 0

	for _, art := range articles {
		if art.ChineseTitle == "" {
			continue // 未摘要的文章跳过
		}
		if art.Category == model.CategoryCloud {
			continue // 已经是云服务分类的跳过
		}

		// 检查是否包含头部国内云厂商关键词
		if !art.IsTopDomesticCloud() && !art.IsDomesticCloud() {
			continue
		}

		text := strings.ToLower(art.ChineseTitle + " " + art.OriginalTitle + " " + art.Summary + " " + art.Tags)

		// 排除非云服务场景（汽车/硬件/消费电子等）
		nonCloudHits := 0
		for _, kw := range nonCloudExcludeKeywords {
			if strings.Contains(text, kw) {
				nonCloudHits++
			}
		}
		if nonCloudHits >= 1 {
			continue // 汽车/硬件类文章，不纠正到云服务
		}

		// 排除纯商业/融资文章
		bizHits := 0
		for _, kw := range pureBizKeywords {
			if strings.Contains(text, kw) {
				bizHits++
			}
		}
		if bizHits >= 2 {
			continue // 多个商业关键词命中，大概率是纯融资新闻
		}

		// 对于"模型前沿"分类的文章，需要额外的云平台关键词辅助判断
		if art.Category == model.CategoryModelFrontier {
			hasCloudHint := false
			for _, hint := range cloudPlatformHints {
				if strings.Contains(text, hint) {
					hasCloudHint = true
					break
				}
			}
			if !hasCloudHint {
				continue // 纯模型技术文章，不纠正
			}
		}

		// 执行重分类
		oldCat := art.Category
		art.Category = model.CategoryCloud

		// 同步更新数据库
		if err := a.repo.UpdateSummary(ctx, art.ID, art.ChineseTitle, art.Summary, art.Tags, art.Category, art.ImportanceScore, art.Recommendation); err != nil {
			slog.Error("reclassify update failed, reverting", "article_id", art.ID, "error", err)
			art.Category = oldCat // 回滚
			continue
		}

		reclassified++
		slog.Info("reclassified to cloud category",
			"article_id", art.ID,
			"title", truncateForLog(art.ChineseTitle, 50),
			"old_category", oldCat,
			"is_top_domestic", art.IsTopDomesticCloud(),
		)
	}

	return reclassified
}

// truncateForLog 截断字符串用于日志输出。
func truncateForLog(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
