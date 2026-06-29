// Package config 负责加载和管理应用配置。
package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 是应用的全局配置结构。
type Config struct {
	Crawler    CrawlerConfig    `yaml:"crawler"`
	Database   DatabaseConfig   `yaml:"database"`
	LLM        LLMConfig        `yaml:"llm"`
	Output     OutputConfig     `yaml:"output"`
	Aggregator AggregatorConfig `yaml:"aggregator"`
	WebSearch  WebSearchConfig  `yaml:"web_search"`
	Log        LogConfig        `yaml:"log"`
}

// CrawlerConfig 爬虫相关配置。
type CrawlerConfig struct {
	Delay              int `yaml:"delay"`
	Timeout            int `yaml:"timeout"`
	MaxRetries         int `yaml:"max_retries"`
	MaxArticlesPerSource int `yaml:"max_articles_per_source"`
}

// DatabaseConfig 数据库配置。
type DatabaseConfig struct {
	Path string `yaml:"path"`
}

// LLMConfig LLM API 配置。
type LLMConfig struct {
	Provider         string `yaml:"provider"`
	BaseURL          string `yaml:"base_url"`
	APIKey           string `yaml:"api_key"`
	Model            string `yaml:"model"`
	MaxConcurrent    int    `yaml:"max_concurrent"`
	Timeout          int    `yaml:"timeout"`
	MaxSummaryTokens int    `yaml:"max_summary_tokens"`
}

// OutputConfig 数据输出配置。
type OutputConfig struct {
	Dir         string `yaml:"dir"`
	DailyFormat string `yaml:"daily_format"`
	HTMLDir     string `yaml:"html_dir"`
}

// AggregatorConfig 聚合配置。
type AggregatorConfig struct {
	MaxDailyArticles int `yaml:"max_daily_articles"`
	MinDailyArticles int `yaml:"min_daily_articles"`
}

// WebSearchConfig Web 搜索 API 配置（使用 Serper.dev Google Search API）。
type WebSearchConfig struct {
	Enabled        bool     `yaml:"enabled"`
	APIKey         string   `yaml:"api_key"`
	MaxResults     int      `yaml:"max_results"`
	Timeout        int      `yaml:"timeout"`
	Queries        []string `yaml:"queries"`
	ExcludeDomains []string `yaml:"exclude_domains"`
}

// LogConfig 日志配置。
type LogConfig struct {
	Level string `yaml:"level"`
	File  string `yaml:"file"`
}

// Load 从指定路径加载配置文件，环境变量优先级更高。
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	cfg.applyEnvOverrides()
	cfg.applyDefaults()

	return cfg, nil
}

// applyEnvOverrides 用环境变量覆盖配置值。
func (c *Config) applyEnvOverrides() {
	if v := os.Getenv("OPENAI_API_KEY"); v != "" {
		c.LLM.APIKey = v
	}
	if v := os.Getenv("CLAUDE_API_KEY"); v != "" {
		c.LLM.APIKey = v
	}
	if v := os.Getenv("LLM_PROVIDER"); v != "" {
		c.LLM.Provider = v
	}
	if v := os.Getenv("LLM_MODEL"); v != "" {
		c.LLM.Model = v
	}
	if v := os.Getenv("LLM_BASE_URL"); v != "" {
		c.LLM.BaseURL = v
	}
	if v := os.Getenv("DB_PATH"); v != "" {
		c.Database.Path = v
	}
	if v := os.Getenv("OUTPUT_DIR"); v != "" {
		c.Output.Dir = v
	}
	if v := os.Getenv("HTML_DIR"); v != "" {
		c.Output.HTMLDir = v
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		c.Log.Level = v
	}
	if v := os.Getenv("WEBSEARCH_API_KEY"); v != "" {
		c.WebSearch.APIKey = v
	}
}

// applyDefaults 设置配置默认值。
func (c *Config) applyDefaults() {
	if c.Crawler.Delay <= 0 {
		c.Crawler.Delay = 2
	}
	if c.Crawler.Timeout <= 0 {
		c.Crawler.Timeout = 30
	}
	if c.Crawler.MaxRetries <= 0 {
		c.Crawler.MaxRetries = 3
	}
	if c.Crawler.MaxArticlesPerSource <= 0 {
		c.Crawler.MaxArticlesPerSource = 50
	}
	if c.Database.Path == "" {
		c.Database.Path = "./data/ai_news.db"
	}
	if c.LLM.MaxConcurrent <= 0 {
		c.LLM.MaxConcurrent = 5
	}
	if c.LLM.Timeout <= 0 {
		c.LLM.Timeout = 60
	}
	if c.LLM.MaxSummaryTokens <= 0 {
		c.LLM.MaxSummaryTokens = 500
	}
	if c.Output.Dir == "" {
		c.Output.Dir = "../frontend/data"
	}
	if c.Output.DailyFormat == "" {
		c.Output.DailyFormat = "2006-01-02"
	}
	// html_dir 不设默认值：为空时 main.go 会跳过 HTML 生成。
	// 现在前端由 Next.js 渲染，Go 模板 HTML 已不再需要。
	if c.Aggregator.MaxDailyArticles <= 0 {
		c.Aggregator.MaxDailyArticles = 30
	}
	if c.Aggregator.MinDailyArticles <= 0 {
		c.Aggregator.MinDailyArticles = 20
	}
	if c.Log.Level == "" {
		c.Log.Level = "info"
	}
	if c.WebSearch.MaxResults <= 0 {
		c.WebSearch.MaxResults = 10
	}
	if c.WebSearch.Timeout <= 0 {
		c.WebSearch.Timeout = 15
	}
}

// CrawlerTimeout 返回爬虫超时时间。
func (c *Config) CrawlerTimeout() time.Duration {
	return time.Duration(c.Crawler.Timeout) * time.Second
}

// LLMTimeout 返回 LLM 请求超时时间。
func (c *Config) LLMTimeout() time.Duration {
	return time.Duration(c.LLM.Timeout) * time.Second
}

// WebSearchTimeout 返回 Web 搜索请求超时时间。
func (c *Config) WebSearchTimeout() time.Duration {
	return time.Duration(c.WebSearch.Timeout) * time.Second
}
