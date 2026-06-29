// Package main 只执行爬取步骤（不做 LLM 摘要），用于快速测试新源。
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"ai-news-crawler/internal/config"
	"ai-news-crawler/internal/crawler"
	"ai-news-crawler/internal/dedup"
	"ai-news-crawler/internal/repository"
)

func main() {
	configPath := flag.String("config", "configs/config.yaml", "配置文件路径")
	flag.Parse()

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(handler))

	cfg, err := config.Load(*configPath)
	if err != nil {
		slog.Error("load config failed", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	since := time.Now().AddDate(0, 0, -3)

	repo, err := repository.NewSQLiteRepo(cfg.Database.Path)
	if err != nil {
		slog.Error("init database failed", "error", err)
		os.Exit(1)
	}
	defer repo.Close()

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
		// === AI 专业媒体 ===
		crawler.NewUniteAICrawler(timeout),
		crawler.NewAINewsCrawler(timeout),
		crawler.NewMarkTechPostCrawler(timeout),
		crawler.NewSyncedReviewCrawler(timeout),
		crawler.NewIEEESpectrumCrawler(timeout),
		// === AI 大厂官方博客 ===
		crawler.NewOpenAIBlogCrawler(timeout),
		crawler.NewDeepMindCrawler(timeout),
		crawler.NewGoogleResearchCrawler(timeout),
		crawler.NewMetaEngineeringCrawler(timeout),
		crawler.NewHuggingFaceCrawler(timeout),
		crawler.NewAnthropicCrawler(timeout),
		crawler.NewMicrosoftAICrawler(timeout),
		crawler.NewNvidiaBlogCrawler(timeout),
		// === AI 前沿思想领袖 ===
		crawler.NewJackClarkCrawler(timeout),
		crawler.NewSimonWillisonCrawler(timeout),
		crawler.NewLilianWengCrawler(timeout),
		crawler.NewLatentSpaceCrawler(timeout),
		crawler.NewTheGradientCrawler(timeout),
		crawler.NewTheBatchCrawler(timeout),
		// === AI 学术 & 论文 ===
		crawler.NewArxivAICrawler(timeout),
		crawler.NewPapersWithCodeCrawler(timeout),
		// === 创投 ===
		crawler.NewYCombinatorCrawler(timeout),
		crawler.NewPanDailyCrawler(timeout),
		// === 国内科技媒体 ===
		crawler.New36KrCrawler(timeout),
		crawler.NewJiqizhixinCrawler(timeout),
		crawler.NewITHomeCrawler(timeout),
		crawler.NewZhidxCrawler(timeout),
	}

	// === Web 搜索补充采集器（按需启用）===
	if ws := crawler.NewWebSearchCrawler(&cfg.WebSearch); ws != nil {
		crawlers = append(crawlers, ws)
		slog.Info("web search crawler enabled", "provider", "serper")
	}

	totalNew := 0
	for _, c := range crawlers {
		articles, err := c.Crawl(ctx, since)
		if err != nil {
			slog.Error("crawl failed", "source", c.Name(), "error", err)
			continue
		}

		savedCount := 0
		imgCount := 0
		for _, a := range articles {
			a.Hash = dedup.URLHash(a.URL)
			if err := repo.Save(ctx, a); err != nil {
				continue
			}
			savedCount++
			if a.ImageURL != "" {
				if err := repo.UpdateImageURL(ctx, a.URL, a.ImageURL); err == nil {
					imgCount++
				}
			}
		}

		fmt.Printf("✅ %-20s  fetched: %3d  new: %3d  images: %3d\n", c.Name(), len(articles), savedCount, imgCount)
		totalNew += savedCount
	}

	fmt.Printf("\n📊 Total new articles saved: %d\n", totalNew)
}
