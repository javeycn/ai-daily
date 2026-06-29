// Package exporter 负责将日报数据导出为 JSON 文件。
package exporter

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"ai-news-crawler/internal/config"
	"ai-news-crawler/internal/model"
)

// Exporter 导出日报数据为 JSON 文件。
type Exporter struct {
	cfg *config.OutputConfig
}

// New 创建一个新的 Exporter 实例。
func New(cfg *config.OutputConfig) *Exporter {
	return &Exporter{cfg: cfg}
}

// ExportDaily 导出单日日报为 JSON 文件。
func (e *Exporter) ExportDaily(report *model.DailyReport) error {
	dailyDir := filepath.Join(e.cfg.Dir, "daily")
	if err := os.MkdirAll(dailyDir, 0o755); err != nil {
		return fmt.Errorf("create daily directory: %w", err)
	}

	filename := filepath.Join(dailyDir, report.Date+".json")
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal daily report: %w", err)
	}

	if err := os.WriteFile(filename, data, 0o644); err != nil {
		return fmt.Errorf("write daily report: %w", err)
	}

	slog.Info("daily report exported", "file", filename, "articles", report.TotalCount)
	return nil
}

// UpdateIndex 更新全量索引文件。
func (e *Exporter) UpdateIndex(reports []*model.DailyReport) error {
	if err := os.MkdirAll(e.cfg.Dir, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	// 按日期降序排序
	sort.Slice(reports, func(i, j int) bool {
		return reports[i].Date > reports[j].Date
	})

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

	index := model.IndexFile{
		Days:    days,
		Updated: time.Now().Format("2006-01-02 15:04:05"),
	}

	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal index: %w", err)
	}

	indexPath := filepath.Join(e.cfg.Dir, "index.json")
	if err := os.WriteFile(indexPath, data, 0o644); err != nil {
		return fmt.Errorf("write index file: %w", err)
	}

	slog.Info("index file updated", "file", indexPath, "days", len(days))
	return nil
}

// LoadAllDailyReports 加载所有已导出的日报文件，用于重建索引。
func (e *Exporter) LoadAllDailyReports() ([]*model.DailyReport, error) {
	dailyDir := filepath.Join(e.cfg.Dir, "daily")
	entries, err := os.ReadDir(dailyDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read daily directory: %w", err)
	}

	var reports []*model.DailyReport
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(dailyDir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			slog.Warn("skip invalid daily file", "file", filePath, "error", err)
			continue
		}

		var report model.DailyReport
		if err := json.Unmarshal(data, &report); err != nil {
			slog.Warn("skip malformed daily file", "file", filePath, "error", err)
			continue
		}

		reports = append(reports, &report)
	}

	return reports, nil
}
