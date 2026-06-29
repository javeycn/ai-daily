// Package main 是 AI 新闻采集服务的入口。
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"sync"
	"time"

	"ai-news-crawler/internal/aggregator"
	"ai-news-crawler/internal/config"
	"ai-news-crawler/internal/crawler"
	"ai-news-crawler/internal/dedup"
	"ai-news-crawler/internal/exporter"
	"ai-news-crawler/internal/filter"
	"ai-news-crawler/internal/htmlgen"
	"ai-news-crawler/internal/model"
	"ai-news-crawler/internal/repository"
	"ai-news-crawler/internal/summarizer"
)

func main() {
	configPath := flag.String("config", "configs/config.yaml", "配置文件路径")
	dateStr := flag.String("date", "", "指定采集日期（格式：2006-01-02），默认为今天")
	summarizeOnly := flag.Bool("summarize-only", false, "跳过采集步骤，仅对已有文章生成摘要并导出")
	incremental := flag.Bool("incremental", false, "增量模式：只采集最近6小时新文章，局部更新HTML")
	sinceHours := flag.Int("since-hours", 6, "增量模式下的回溯小时数，默认6小时")
	fixBroken := flag.Bool("fix-broken", false, "修复数据库中 LLM 解析失败导致标题为 JSON 残片的文章，重新调用 LLM 生成摘要")
	flag.Parse()

	// 初始化日志
	initLogger()

	// 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		slog.Error("load config failed", "error", err)
		os.Exit(1)
	}

	// 确定采集日期
	targetDate := time.Now()
	if *dateStr != "" {
		parsed, err := time.Parse("2006-01-02", *dateStr)
		if err != nil {
			slog.Error("invalid date format", "date", *dateStr, "error", err)
			os.Exit(1)
		}
		targetDate = parsed
	}

	// 确定采集时间范围
	crawlDate := targetDate
	var since time.Time
	if *incremental {
		since = time.Now().Add(-time.Duration(*sinceHours) * time.Hour)
		slog.Info("incremental mode enabled",
			"since_hours", *sinceHours,
			"since", since.Format("2006-01-02 15:04:05"),
		)
	} else {
		// 全量模式：采集最近3天的数据，确保有足够的文章
		since = crawlDate.AddDate(0, 0, -3)
	}

	slog.Info("starting ai news crawler",
		"target_date", targetDate.Format("2006-01-02"),
		"crawl_since", since.Format("2006-01-02 15:04:05"),
		"mode", modeLabel(*incremental),
	)

	ctx := context.Background()

	// 初始化数据库
	repo, err := repository.NewSQLiteRepo(cfg.Database.Path)
	if err != nil {
		slog.Error("init database failed", "error", err)
		os.Exit(1)
	}
	defer repo.Close()

	// --fix-broken 模式：修复数据库中 LLM 解析失败的文章
	if *fixBroken {
		fixBrokenSummaries(ctx, repo)
		return
	}

	// 初始化采集器
	crawlers := initCrawlers(cfg)

	if !*summarizeOnly {
		// 第一步：并发采集数据（semaphore 控制并发数为 5）
		slog.Info("step 1: crawling articles from all sources (concurrent)")
		const maxConcurrentCrawlers = 5
		sem := make(chan struct{}, maxConcurrentCrawlers)
		var wg sync.WaitGroup
		var mu sync.Mutex
		var allArticles []*model.Article

		for _, c := range crawlers {
			wg.Add(1)
			go func(cr crawler.Crawler) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				articles, err := cr.Crawl(ctx, since)
				if err != nil {
					slog.Error("crawl failed", "source", cr.Name(), "error", err)
					return
				}

				// 保存到数据库（利用唯一 URL 约束去重，SQLite WAL 模式支持并发写入）
				savedCount := 0
				imgUpdated := 0
				for _, a := range articles {
					a.Hash = dedup.URLHash(a.URL)
					if err := repo.Save(ctx, a); err != nil {
						slog.Error("save article failed", "url", a.URL, "error", err)
						continue
					}
					savedCount++
					if a.ImageURL != "" {
						if err := repo.UpdateImageURL(ctx, a.URL, a.ImageURL); err == nil {
							imgUpdated++
						}
					}
				}

				slog.Info("source crawl completed",
					"source", cr.Name(),
					"fetched", len(articles),
					"saved", savedCount,
				)

				mu.Lock()
				allArticles = append(allArticles, articles...)
				mu.Unlock()
			}(c)
		}
		wg.Wait()

		slog.Info("total articles crawled", "count", len(allArticles))

		// AI 关键词过滤（采集阶段已由 RSS URL 过滤，这里做二次确认）
		filtered := filter.FilterArticles(allArticles)
		slog.Info("articles after AI filter", "count", len(filtered))
	} else {
		slog.Info("summarize-only mode: skipping crawl and filter steps")
	}

	// 第三步：生成摘要并聚合日报
	slog.Info("step 2: generating summaries and aggregating daily report")
	summ := summarizer.New(&cfg.LLM)
	agg := aggregator.New(repo, summ, &cfg.Aggregator)

	report, err := agg.Aggregate(ctx, crawlDate)
	if err != nil {
		slog.Error("aggregate failed", "error", err)
		os.Exit(1)
	}

	// 第四步：导出数据
	slog.Info("step 3: exporting daily report")
	exp := exporter.New(&cfg.Output)
	if err := exp.ExportDaily(report); err != nil {
		slog.Error("export daily failed", "error", err)
		os.Exit(1)
	}

	// 更新索引
	allReports, err := exp.LoadAllDailyReports()
	if err != nil {
		slog.Error("load reports for index failed", "error", err)
	} else {
		if err := exp.UpdateIndex(allReports); err != nil {
			slog.Error("update index failed", "error", err)
		}
	}

	// 第五步：生成静态 HTML 页面
	slog.Info("step 4: generating static HTML pages")
	if cfg.Output.HTMLDir != "" {
		gen, err := htmlgen.New(cfg.Output.HTMLDir)
		if err != nil {
			slog.Error("init html generator failed", "error", err)
			os.Exit(1)
		}

		if *incremental {
			// 增量模式：只更新首页和当天日报页，跳过归档、搜索等页面的全量重建
			slog.Info("incremental HTML update: home + daily page only")
			if err := gen.GenerateHome(report); err != nil {
				slog.Error("generate home page failed", "error", err)
			}
			if err := gen.GenerateDaily(report); err != nil {
				slog.Error("generate daily page failed", "error", err)
			}
		} else {
			// 全量模式：重新生成整个站点
			index := buildIndex(allReports)
			if err := gen.GenerateAll(report, allReports, index); err != nil {
				slog.Error("generate HTML pages failed", "error", err)
				os.Exit(1)
			}
		}
		slog.Info("static HTML pages generated", "output_dir", cfg.Output.HTMLDir, "mode", modeLabel(*incremental))
	} else {
		slog.Info("html_dir not configured, skipping HTML generation")
	}

	slog.Info("ai news crawler completed successfully",
		"date", report.Date,
		"articles", report.TotalCount,
	)
}

// initCrawlers 初始化所有新闻源采集器。
func initCrawlers(cfg *config.Config) []crawler.Crawler {
	timeout := cfg.CrawlerTimeout()

	crawlers := []crawler.Crawler{
		// === 国际科技媒体 ===
		crawler.NewVergeCrawler(timeout),
		crawler.NewVentureBeatCrawler(timeout),
		crawler.NewTechCrunchCrawler(timeout),
		crawler.NewArsTechnicaCrawler(timeout),
		crawler.NewWiredAICrawler(timeout),
		crawler.NewMITTechReviewCrawler(timeout),
		crawler.NewTheDecoderCrawler(timeout),

		// === AI 专业媒体 & 聚合 ===
		crawler.NewUniteAICrawler(timeout),
		crawler.NewAINewsCrawler(timeout),
		crawler.NewMarkTechPostCrawler(timeout),
		crawler.NewSyncedReviewCrawler(timeout),
		crawler.NewIEEESpectrumCrawler(timeout),
		crawler.NewAIHubTodayCrawler(timeout),       // AI Hub Today — 聚合型 AI 新闻（Skills Tier 1 推荐）
		crawler.NewDailyAICrawler(timeout),           // DailyAI — AI 新闻深度报道
		crawler.NewInfoQAICrawler(timeout),           // InfoQ AI — 架构师社区 AI 板块

		// === AI 大厂官方博客 ===
		crawler.NewOpenAIBlogCrawler(timeout),
		crawler.NewDeepMindCrawler(timeout),
		crawler.NewGoogleResearchCrawler(timeout),
		crawler.NewMetaEngineeringCrawler(timeout),
		crawler.NewHuggingFaceCrawler(timeout),
		crawler.NewAnthropicCrawler(timeout),         // Anthropic 官方博客
		crawler.NewMicrosoftAICrawler(timeout),       // Microsoft AI Blog
		crawler.NewNvidiaBlogCrawler(timeout),        // NVIDIA AI Blog

		// === 云服务 AI 博客 ===
		crawler.NewAWSMLBlogCrawler(timeout),         // AWS Machine Learning Blog
		crawler.NewGCPAIBlogCrawler(timeout),         // Google Cloud AI Blog
		crawler.NewAzureAIBlogCrawler(timeout),       // Azure AI Services Blog

		// === AI 前沿思想领袖 & Newsletter ===
		crawler.NewJackClarkCrawler(timeout),        // Import AI — Anthropic 联创
		crawler.NewSimonWillisonCrawler(timeout),     // LLM 应用开发意见领袖
		crawler.NewLilianWengCrawler(timeout),        // OpenAI 研究员博客
		crawler.NewLatentSpaceCrawler(timeout),       // AI 工程师社区
		crawler.NewTheGradientCrawler(timeout),       // 斯坦福 AI 深度分析
		crawler.NewTheBatchCrawler(timeout),          // The Batch — 吴恩达 AI 周报
		crawler.NewBAIRBlogCrawler(timeout),          // BAIR Blog — 伯克利 AI 研究实验室
		crawler.NewChipHuyenCrawler(timeout),         // Chip Huyen — MLOps/LLMOps 思想领袖
		crawler.NewEugeneYanCrawler(timeout),         // Eugene Yan — Amazon 首席 AI 工程师
		crawler.NewKarpathyCrawler(timeout),          // Karpathy — 前 OpenAI/Tesla AI 负责人

		// === AI 学术 & 论文 ===
		crawler.NewArxivAICrawler(timeout),           // arXiv cs.AI 最新论文
		crawler.NewPapersWithCodeCrawler(timeout),    // Papers With Code 热门论文

		// === 创投 & 科技商业 ===
		crawler.NewYCombinatorCrawler(timeout),       // Hacker News (YC)
		crawler.NewPanDailyCrawler(timeout),          // 中国科技出海资讯

		// === 国内科技媒体 ===
		crawler.New36KrCrawler(timeout),
		crawler.NewJiqizhixinCrawler(timeout),
		crawler.NewITHomeCrawler(timeout),
		crawler.NewZhidxCrawler(timeout),             // 智东西
		crawler.NewQbitAICrawler(timeout),             // 量子位 — 国内头部 AI 科技媒体
		crawler.NewInfoQCNCrawler(timeout),            // InfoQ 中文 — 开发者社区 AI 深度内容
	}

	// === Web 搜索补充采集器（按需启用）===
	if ws := crawler.NewWebSearchCrawler(&cfg.WebSearch); ws != nil {
		crawlers = append(crawlers, ws)
		slog.Info("web search crawler enabled", "provider", "serper", "queries", len(cfg.WebSearch.Queries))
	}

	return crawlers
}

// initLogger 初始化日志。
func initLogger() {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slog.SetDefault(slog.New(handler))
}

// buildIndex 从日报列表构建索引数据。
func buildIndex(reports []*model.DailyReport) *model.IndexFile {
	var days []model.DailyIndex
	for _, r := range reports {
		var tags []string
		for _, ts := range r.TagStats {
			tags = append(tags, ts.Tag)
		}
		days = append(days, model.DailyIndex{
			Date:       r.Date,
			Title:      r.Title,
			Summary:    r.Summary,
			TotalCount: r.TotalCount,
			Tags:       tags,
		})
	}
	return &model.IndexFile{
		Days:    days,
		Updated: time.Now().Format("2006-01-02 15:04:05"),
	}
}

// modeLabel 返回运行模式的标签文本。
func modeLabel(incremental bool) string {
	if incremental {
		return "incremental"
	}
	return "full"
}

// fixBrokenSummaries 修复数据库中 LLM 解析失败导致标题为 JSON 残片的文章。
// 修复策略：
// 1. 先尝试从已存储的 summary 字段（可能包含完整 JSON）中提取正确数据
// 2. 如果无法从 summary 中修复，则清空 chinese_title 让下次运行时重新调用 LLM
func fixBrokenSummaries(ctx context.Context, repo *repository.SQLiteRepo) {
	slog.Info("fix-broken mode: scanning for broken summaries...")

	brokenArticles, err := repo.GetBrokenSummaries(ctx)
	if err != nil {
		slog.Error("get broken summaries failed", "error", err)
		os.Exit(1)
	}

	slog.Info("found broken articles", "count", len(brokenArticles))

	if len(brokenArticles) == 0 {
		slog.Info("no broken articles found, nothing to fix")
		return
	}

	fixedFromSummary := 0
	resetForResummarize := 0

	for _, art := range brokenArticles {
		slog.Info("processing broken article",
			"id", art.ID,
			"chinese_title", truncateForLog(art.ChineseTitle, 50),
			"summary_preview", truncateForLog(art.Summary, 100),
			"source", art.Source,
		)

		// 策略 1：尝试从现有的 chinese_title + summary 拼接还原完整 JSON 并解析
		// 因为降级逻辑将 "{" 作为标题，其余内容作为 summary，
		// 所以拼接 chinese_title + "\n" + summary 可能还原出完整 JSON
		fullContent := art.ChineseTitle + "\n" + art.Summary
		chineseTitle, summary, tags, category, importanceScore, recommendation, parseErr := summarizer.ParseLLMOutput(fullContent)

		if parseErr == nil && chineseTitle != "" && chineseTitle != "{" {
			// 从存储的数据中成功恢复
			if err := repo.UpdateSummary(ctx, art.ID, chineseTitle, summary, tags, category, importanceScore, recommendation); err != nil {
				slog.Error("update fixed summary failed", "id", art.ID, "error", err)
				continue
			}
			fixedFromSummary++
			slog.Info("fixed from stored data",
				"id", art.ID,
				"new_title", truncateForLog(chineseTitle, 80),
				"category", category,
			)
			continue
		}

		// 策略 2：无法从存储数据中恢复，清空 chinese_title 让下次运行重新调用 LLM
		if err := repo.UpdateSummary(ctx, art.ID, "", "", "", "", 0, ""); err != nil {
			slog.Error("reset summary failed", "id", art.ID, "error", err)
			continue
		}
		resetForResummarize++
		slog.Info("reset for re-summarization",
			"id", art.ID,
			"original_title", truncateForLog(art.OriginalTitle, 80),
		)
	}

	slog.Info("fix-broken completed",
		"total_broken", len(brokenArticles),
		"fixed_from_summary", fixedFromSummary,
		"reset_for_resummarize", resetForResummarize,
	)

	// 如果有重置的文章，提示用户运行 --summarize-only 重新生成
	if resetForResummarize > 0 {
		slog.Info("tip: run with --summarize-only to regenerate summaries for reset articles")
	}
}

// truncateForLog 截断字符串用于日志输出。
func truncateForLog(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}