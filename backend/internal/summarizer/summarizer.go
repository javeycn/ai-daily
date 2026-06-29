// Package summarizer 提供基于 LLM API 的文章摘要生成功能。
package summarizer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"

	"ai-news-crawler/internal/config"
	"ai-news-crawler/internal/model"
)

// Summarizer 使用 LLM API 生成文章摘要。
type Summarizer struct {
	cfg    *config.LLMConfig
	client *http.Client
}

// SummarizeResult 是单篇文章的摘要结果。
type SummarizeResult struct {
	Article         *model.Article
	ChineseTitle    string
	Summary         string
	Tags            string
	Category        string
	ImportanceScore int
	Recommendation  string
	Error           error
}

// New 创建一个新的 Summarizer 实例。
func New(cfg *config.LLMConfig) *Summarizer {
	return &Summarizer{
		cfg: cfg,
		client: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		},
	}
}

// SummarizeBatch 批量生成文章摘要，控制并发数。
func (s *Summarizer) SummarizeBatch(ctx context.Context, articles []*model.Article) []*SummarizeResult {
	results := make([]*SummarizeResult, len(articles))

	sem := make(chan struct{}, s.cfg.MaxConcurrent)
	var wg sync.WaitGroup

	for i, article := range articles {
		wg.Add(1)
		go func(idx int, a *model.Article) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			chineseTitle, summary, tags, category, importanceScore, recommendation, err := s.summarizeOne(ctx, a)
			results[idx] = &SummarizeResult{
				Article:         a,
				ChineseTitle:    chineseTitle,
				Summary:         summary,
				Tags:            tags,
				Category:        category,
				ImportanceScore: importanceScore,
				Recommendation:  recommendation,
				Error:           err,
			}
		}(i, article)
	}

	wg.Wait()
	return results
}

// summarizeOne 为单篇文章生成摘要。
func (s *Summarizer) summarizeOne(ctx context.Context, article *model.Article) (chineseTitle, summary, tags, category string, importanceScore int, recommendation string, err error) {
	prompt := s.buildPrompt(article)

	switch s.cfg.Provider {
	case "claude":
		chineseTitle, summary, tags, category, importanceScore, recommendation, err = s.callClaude(ctx, prompt)
	default:
		chineseTitle, summary, tags, category, importanceScore, recommendation, err = s.callOpenAI(ctx, prompt)
	}

	if err != nil {
		slog.Error("summarize failed, using fallback", "article_id", article.ID, "error", err)
		chineseTitle = article.OriginalTitle
		summary = truncateText(article.Summary, 300)
		tags = "其他"
		category = model.CategoryModelFrontier
		importanceScore = 5
		recommendation = ""
		err = nil
	}

	return
}

// buildPrompt 构建摘要生成的 prompt，支持 8 分类体系（按内容本质分类）、来源感知、双语输出和重要度评分。
func (s *Summarizer) buildPrompt(article *model.Article) string {
	return fmt.Sprintf(`你是一位专业的 AI 行业分析师和科技新闻编辑。请对以下新闻文章进行处理：

原标题：%s
原文摘要：%s
来源：%s

请按以下 JSON 格式输出（不要包含 markdown 代码块标记）：
{
  "chinese_title": "中文标题",
  "original_title": "英文原标题（如原文是英文则保留原文标题，如原文是中文则留空）",
  "summary": "3到5行的中文内容摘要",
  "recommendation": "1句话推荐理由，回答为什么值得读",
  "tags": "标签1,标签2,标签3",
  "category": "从以下8个分类中选择最匹配的一个",
  "importance_score": 7
}

## 分类说明（按内容本质分类，不按地域。必须从以下 8 个分类中选择一个）：

1. 模型前沿 — 模型能力突破、开源模型发布、Benchmark 评测、训练/推理新范式。包括 GPT/Claude/Gemini/LLaMA/Qwen 等模型的能力进展、新架构（MoE、SSM 等）、Scaling Laws 研究、模型评测排行榜、RLHF/DPO 等对齐训练技术
2. 产品与应用 — AI 产品发布、行业落地案例、AI 工具。包括 AI 在医疗/教育/金融/法律等行业的落地应用、AI 助手/搜索/创作工具、ToB/ToC 产品上线、用户增长数据
3. 深度洞察 — 趋势分析、领袖观点、行业报告、深度技术博文。包括行业领袖对 AI 趋势的判断、深度技术文章、播客精华摘要、研究报告解读、AI 行业年度回顾/展望
4. 云服务与平台 — AWS/Google Cloud/Azure/腾讯云/阿里云/华为云/火山引擎等云厂商的 AI 服务动态、MaaS 平台更新、API 定价变化、GPU 实例上新、模型托管服务
5. AI工程 — Agent 框架（LangChain/CrewAI/AutoGen）、RAG 架构、MCP/A2A 协议、Prompt 工程、LLMOps（评估/监控/部署）、AI 编程工具（Cursor/Copilot/Devin）、向量搜索优化
6. AI基础设施 — AI 芯片（NVIDIA/AMD/Intel/自研芯片）、推理引擎（vLLM/TensorRT）、训练框架（PyTorch/JAX）、向量数据库、AI 集群网络、高性能计算
7. 商业与投资 — 融资、并购、IPO、市场格局分析、企业战略调整。包括 AI 公司融资事件、并购交易、IPO 进展、市场规模预测、竞争态势变化
8. AI安全 — AI 监管法规（EU AI Act 等）、安全对齐研究、数据隐私、开源许可争议、AI 伦理问题、深度伪造治理、模型安全漏洞

## 重要度评分说明（importance_score，1-10 整数）：

综合评估文章的行业影响力、内容质量和时效性，给出 1-10 的整数评分：
- 10 分：行业重大事件（如头部模型重大升级、大型并购、重要法规出台），每个 AI 从业者都应该知道
- 7-9 分：高价值内容（如重要产品发布、深度技术分析、有影响力的行业报告），对多数 AI 从业者有价值
- 4-6 分：中等价值（如特定领域的技术更新、常规产品迭代、一般性行业动态）
- 1-3 分：低价值（如纯营销软文、信息量少的简讯、过时内容）

评分时特别注意：
- 深度原创分析 > 简单新闻报道 > 转载摘要
- 有具体数据/基准测试/案例支撑的 > 纯观点输出
- 首次披露的独家信息 > 多家媒体已报道的旧闻

## 标签说明（与分类是不同维度，标签描述文章涉及的具体技术/产品/主题）：

从以下候选标签中选择 1~3 个最匹配的标签，用逗号分隔。如果候选列表中没有合适的，可以新建一个简短的中文标签（2~6 个字）。

### 技术主题标签：
LLM, GPT, Claude, Gemini, LLaMA, Qwen, Mistral, DeepSeek, 多模态, 语音AI, 视觉AI, 视频生成, 图像生成, 代码生成, 文本生成, NLP, 知识图谱, 强化学习, 微调, 量化, 蒸馏, 对齐, Scaling Law, MoE, Transformer, Diffusion, 机器人

### 工程与工具标签：
Agent, RAG, MCP, A2A, LangChain, Prompt工程, LLMOps, Cursor, Copilot, Devin, 向量数据库, 推理优化, vLLM, TensorRT, PyTorch, 模型部署, API, SDK, 开源

### 行业与场景标签：
医疗AI, 教育AI, 金融AI, 法律AI, 自动驾驶, 搜索, 推荐系统, 客服, 翻译, 创作工具, 办公AI, 编程助手, 科学研究, 药物发现, 气候AI

### 公司与平台标签：
OpenAI, Google, Meta, Anthropic, Microsoft, NVIDIA, Apple, 百度, 阿里, 腾讯, 字节跳动, AWS, Azure, Hugging Face, Stability AI, Midjourney, xAI

### 商业与治理标签：
融资, 并购, IPO, 估值, 开源许可, AI监管, 数据隐私, AI伦理, 版权, 深度伪造, AI安全, EU AI Act

### 标签选择原则：
- 标签用于帮助读者快速识别文章涉及的具体技术/产品/公司，与 category 互补而非重复
- 不要将 category 的名称（如"模型前沿""AI工程"）作为标签
- 优先选择候选列表中的标签，保持一致性
- 如需新建标签，使用 2~6 个中文字，简洁明了

## 关键边界判定（遇到模糊情况时参考）：
- 云厂商上架新模型（如 Azure 上架 GPT-5）→ 云服务与平台（侧重平台动态）
- 模型本身的能力突破（如 GPT-5 的技术细节）→ 模型前沿
- 云厂商 GPU 实例/AI 芯片定价 → 云服务与平台
- AI 芯片本身的技术规格和性能 → AI基础设施
- Agent 框架设计/MCP 协议/RAG 架构 → AI工程
- 某公司上线 AI 客服产品 → 产品与应用
- 开源模型发布（如 LLaMA 4）→ 模型前沿（开源模型属于模型能力维度）
- 有工程价值的学术论文 → 模型前沿或AI工程（按论文核心贡献分类）
- 纯学术 Benchmark/评测 → 模型前沿

## 【重要】国内云厂商优先归类为"云服务与平台"：
当文章主要涉及以下国内云厂商时，应优先归类为"云服务与平台"（除非文章核心内容是纯模型能力突破或纯融资事件）：
- 腾讯云、阿里云、华为云、火山引擎、百度智能云、金山云、京东云、商汤科技
- 以上厂商的平台级动态（如开放平台更新、AI 服务发布、算力调价、智能体平台上线、大模型 API 开放、生态合作等）都应归入"云服务与平台"
- 例如：华为盘古大模型团队动态 → 云服务与平台（华为云 AI 平台相关）
- 例如：腾讯混元模型开放 API → 云服务与平台（腾讯云 MaaS 平台动态）
- 例如：火山引擎发布豆包新功能 → 云服务与平台（火山引擎平台更新）
- 例如：阿里通义千问降价/开放新能力 → 云服务与平台（阿里云百炼平台动态）
- 仅当文章完全聚焦于模型本身的技术细节（架构、性能指标、训练方法论）且不涉及平台/服务层面时，才归入"模型前沿"
- 仅当文章核心是融资/估值/IPO 等纯商业事件时，才归入"商业与投资"

## 来源分类提示：
以下来源的文章通常应优先考虑归入"深度洞察"分类：
Import AI, Simon Willison, Lilian Weng, Latent Space, The Gradient, The Batch, BAIR Blog, DailyAI, Chip Huyen, Eugene Yan, Karpathy
（但如果文章内容明确是纯模型发布公告、纯工程实践教程或纯产品发布，则按实际内容分类）

## 分类决策优先级：
- 先看文章的核心内容是什么（模型能力？产品落地？深度分析？云服务？工程实践？基础设施？融资？安全政策？）
- 不要按来源公司所在国家分类，而是按内容本质分类
- 例如：百度发布新模型 → "模型前沿"；百度 AI 产品上线 → "产品与应用"；百度云 AI 服务更新 → "云服务与平台"；百度融资消息 → "商业与投资"

## 输出要求：
1. chinese_title：简洁准确的中文标题，体现新闻核心信息
2. original_title：如果原文是英文，保留完整的英文原标题；如果原文是中文，此字段留空
3. summary：3到5行的中文摘要，需遵循以下规范：
   - 第 1 句直接点明核心事件或话题（禁止用"本文讨论了..."、"本文介绍了..."等模板化开头）
   - 第 2-3 句阐述关键论据、技术方案或数据支撑
   - 最后 1 句给出结论或行业影响
   - 必须保留关键数字（性能提升百分比、用户量、融资金额、版本号等）
   - 包含具体技术术语和产品名称，不要泛泛概述
   - 目标：读者花 30 秒看摘要，就能决定是否花 10 分钟阅读原文
4. recommendation：1 句话推荐理由（15~40 个中文字），回答"为什么这篇文章值得读"。与 summary 互补，不要重复摘要内容。侧重揭示文章的独特价值点，例如首发数据、独家观点、对行业的影响等。示例："首次公开 GPT-5 的多模态架构细节，含完整 Benchmark 对比"
5. tags：从候选标签中选择 1~3 个，用逗号分隔，不要使用分类名称作为标签
6. category：必须从上面8个分类中选择恰好一个，输出分类的中文名称
7. importance_score：1-10 的整数，按上述评分标准评估文章重要度`, article.OriginalTitle, truncateText(article.Summary, 800), article.Source)
}

// callOpenAI 调用 OpenAI API 生成摘要。
func (s *Summarizer) callOpenAI(ctx context.Context, prompt string) (chineseTitle, summary, tags, category string, importanceScore int, recommendation string, err error) {
	reqBody := map[string]interface{}{
		"model":       s.cfg.Model,
		"messages":    []map[string]string{{"role": "user", "content": prompt}},
		"temperature": 0.3,
		"max_tokens":  s.cfg.MaxSummaryTokens,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", "", "", "", 0, "", fmt.Errorf("marshal request: %w", err)
	}

	baseURL := s.cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com"
	}
	apiURL := strings.TrimRight(baseURL, "/") + "/v1/chat/completions"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return "", "", "", "", 0, "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.cfg.APIKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", "", "", "", 0, "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", "", "", 0, "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", "", "", "", 0, "", fmt.Errorf("api error status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var chatResp openAIResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", "", "", "", 0, "", fmt.Errorf("unmarshal response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", "", "", "", 0, "", fmt.Errorf("empty choices in response")
	}

	return parseLLMOutput(chatResp.Choices[0].Message.Content)
}

// callClaude 调用 Claude API 生成摘要。
func (s *Summarizer) callClaude(ctx context.Context, prompt string) (chineseTitle, summary, tags, category string, importanceScore int, recommendation string, err error) {
	reqBody := map[string]interface{}{
		"model":      s.cfg.Model,
		"max_tokens": s.cfg.MaxSummaryTokens,
		"messages":   []map[string]string{{"role": "user", "content": prompt}},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", "", "", "", 0, "", fmt.Errorf("marshal request: %w", err)
	}

	baseURL := s.cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}
	apiURL := strings.TrimRight(baseURL, "/") + "/v1/messages"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return "", "", "", "", 0, "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", s.cfg.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", "", "", "", 0, "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", "", "", 0, "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", "", "", "", 0, "", fmt.Errorf("api error status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var claudeResp claudeResponse
	if err := json.Unmarshal(respBody, &claudeResp); err != nil {
		return "", "", "", "", 0, "", fmt.Errorf("unmarshal response: %w", err)
	}

	if len(claudeResp.Content) == 0 {
		return "", "", "", "", 0, "", fmt.Errorf("empty content in response")
	}

	textContent := getTextContent(&claudeResp)
	if textContent == "" {
		return "", "", "", "", 0, "", fmt.Errorf("no text content in response")
	}

	return parseLLMOutput(textContent)
}

// ParseLLMOutput 是 parseLLMOutput 的导出版本，供外部包调用。
// 用于修复已损坏的数据：从存储的内容中重新提取 JSON 字段。
func ParseLLMOutput(content string) (chineseTitle, summary, tags, category string, importanceScore int, recommendation string, err error) {
	return parseLLMOutput(content)
}

// jsonBlockRe 用于从 LLM 输出中提取第一个完整的 JSON 对象 {...}。
var jsonBlockRe = regexp.MustCompile(`(?s)\{[^{}]*("chinese_title"|"summary"|"category")[^{}]*\}`)

// llmResult 是 LLM 返回的 JSON 结构。
type llmResult struct {
	ChineseTitle    string `json:"chinese_title"`
	OriginalTitle   string `json:"original_title"`
	Summary         string `json:"summary"`
	Recommendation  string `json:"recommendation"`
	Tags            string `json:"tags"`
	Category        string `json:"category"`
	ImportanceScore int    `json:"importance_score"`
}

// parseLLMOutput 解析 LLM 返回的 JSON 格式输出。
// 采用多级策略增强鲁棒性：
// 1. 直接解析清理后的全文
// 2. 清理不可见控制字符后重试
// 3. 正则提取 JSON 块后解析
// 4. 最终降级：返回错误让上层使用原标题 fallback
func parseLLMOutput(content string) (chineseTitle, summary, tags, category string, importanceScore int, recommendation string, err error) {
	content = strings.TrimSpace(content)

	// 去掉 markdown 代码块标记（LLM 可能用 ```json...``` 包裹）
	content = stripMarkdownCodeBlock(content)

	var result llmResult

	// 策略 1：直接 JSON 解析
	if tryUnmarshal(content, &result) {
		return formatResult(&result)
	}

	// 策略 2：清理 JSON 中的控制字符（LLM 有时在字符串值中输出未转义的换行/制表符）
	cleaned := sanitizeJSONString(content)
	if tryUnmarshal(cleaned, &result) {
		slog.Debug("parseLLMOutput: succeeded after sanitizing control characters")
		return formatResult(&result)
	}

	// 策略 3：用正则提取 JSON 块（LLM 在 JSON 前后可能添加了说明文字）
	if jsonStr := extractJSONBlock(content); jsonStr != "" {
		cleanedJSON := sanitizeJSONString(jsonStr)
		if tryUnmarshal(cleanedJSON, &result) {
			slog.Debug("parseLLMOutput: succeeded after regex extraction")
			return formatResult(&result)
		}
	}

	// 策略 4：尝试逐行查找 JSON 开始和结束位置（处理 JSON 前后有多余文本的情况）
	if jsonStr := extractJSONByBraces(content); jsonStr != "" {
		cleanedJSON := sanitizeJSONString(jsonStr)
		if tryUnmarshal(cleanedJSON, &result) {
			slog.Debug("parseLLMOutput: succeeded after brace-based extraction")
			return formatResult(&result)
		}
	}

	// 所有策略均失败，返回错误让上层走 fallback（使用原标题+截断摘要）
	slog.Warn("parseLLMOutput: all JSON parsing strategies failed",
		"content_length", len(content),
		"content_preview", truncateText(content, 200),
	)
	return "", "", "", "", 0, "", fmt.Errorf("failed to parse LLM output as JSON: content_length=%d", len(content))
}

// stripMarkdownCodeBlock 去掉 markdown 代码块标记。
func stripMarkdownCodeBlock(s string) string {
	s = strings.TrimSpace(s)
	// 处理 ```json\n...\n``` 和 ```\n...\n```
	if strings.HasPrefix(s, "```") {
		// 去掉首行 ```json 或 ```
		if idx := strings.Index(s, "\n"); idx != -1 {
			s = s[idx+1:]
		} else {
			s = strings.TrimPrefix(s, "```json")
			s = strings.TrimPrefix(s, "```")
		}
	}
	s = strings.TrimSuffix(strings.TrimSpace(s), "```")
	return strings.TrimSpace(s)
}

// tryUnmarshal 尝试将字符串解析为 llmResult，返回是否成功。
func tryUnmarshal(s string, result *llmResult) bool {
	// 重置 result 以免上次尝试的残留数据干扰
	*result = llmResult{}
	if err := json.Unmarshal([]byte(s), result); err != nil {
		return false
	}
	// 验证至少有一个关键字段非空
	return result.ChineseTitle != "" || result.Summary != ""
}

// sanitizeJSONString 清理 JSON 字符串中的非法控制字符。
// JSON 规范要求字符串值中的控制字符必须用 \uXXXX 转义，
// 但 LLM 有时直接输出原始换行、制表符等。
func sanitizeJSONString(s string) string {
	var buf strings.Builder
	buf.Grow(len(s))
	inString := false
	escaped := false

	for _, r := range s {
		if escaped {
			buf.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' && inString {
			buf.WriteRune(r)
			escaped = true
			continue
		}
		if r == '"' {
			inString = !inString
			buf.WriteRune(r)
			continue
		}
		if inString && unicode.IsControl(r) {
			// 替换控制字符为合适的转义序列
			switch r {
			case '\n':
				buf.WriteString("\\n")
			case '\r':
				buf.WriteString("\\r")
			case '\t':
				buf.WriteString("\\t")
			default:
				buf.WriteString(fmt.Sprintf("\\u%04x", r))
			}
			continue
		}
		buf.WriteRune(r)
	}
	return buf.String()
}

// extractJSONBlock 用正则从文本中提取包含关键字段的 JSON 对象。
func extractJSONBlock(s string) string {
	match := jsonBlockRe.FindString(s)
	return match
}

// extractJSONByBraces 通过匹配花括号定位 JSON 块。
// 处理嵌套花括号和字符串内的花括号。
func extractJSONByBraces(s string) string {
	start := strings.IndexByte(s, '{')
	if start == -1 {
		return ""
	}

	depth := 0
	inString := false
	escaped := false
	runes := []rune(s[start:])

	for i, r := range runes {
		if escaped {
			escaped = false
			continue
		}
		if r == '\\' && inString {
			escaped = true
			continue
		}
		if r == '"' {
			inString = !inString
			continue
		}
		if !inString {
			if r == '{' {
				depth++
			} else if r == '}' {
				depth--
				if depth == 0 {
					return string(runes[:i+1])
				}
			}
		}
	}
	return ""
}

// formatResult 将解析好的 llmResult 格式化为返回值。
func formatResult(result *llmResult) (chineseTitle, summary, tags, category string, importanceScore int, recommendation string, err error) {
	// 如果 LLM 返回了原始英文标题，拼接为双语标题格式
	if result.OriginalTitle != "" && result.ChineseTitle != "" {
		result.ChineseTitle = result.OriginalTitle + " / " + result.ChineseTitle
	}

	// 验证 category 是否在允许范围内
	category = normalizeCategory(result.Category)

	// 验证并钳住 importance_score 到 1-10 范围
	importanceScore = result.ImportanceScore
	if importanceScore < 1 {
		importanceScore = 5 // 未返回时默认 5 分
	}
	if importanceScore > 10 {
		importanceScore = 10
	}

	return result.ChineseTitle, result.Summary, result.Tags, category, importanceScore, result.Recommendation, nil
}

// normalizeCategory 确保分类在允许范围内。
func normalizeCategory(cat string) string {
	cat = strings.TrimSpace(cat)
	for _, valid := range model.AllCategories {
		if cat == valid {
			return cat
		}
	}
	// 模糊匹配（按 8 分类体系）
	lower := strings.ToLower(cat)

	// 模型前沿
	if strings.Contains(lower, "模型前沿") || strings.Contains(lower, "模型") ||
		strings.Contains(lower, "model") || strings.Contains(lower, "frontier") ||
		strings.Contains(lower, "benchmark") || strings.Contains(lower, "评测") ||
		strings.Contains(lower, "scaling") || strings.Contains(lower, "训练范式") ||
		strings.Contains(lower, "开源模型") || strings.Contains(lower, "llama") ||
		strings.Contains(lower, "gpt") || strings.Contains(lower, "claude") {
		return model.CategoryModelFrontier
	}

	// 产品与应用
	if strings.Contains(lower, "产品") || strings.Contains(lower, "应用") ||
		strings.Contains(lower, "product") || strings.Contains(lower, "application") ||
		strings.Contains(lower, "落地") || strings.Contains(lower, "工具") ||
		strings.Contains(lower, "tool") || strings.Contains(lower, "assistant") ||
		strings.Contains(lower, "上线") || strings.Contains(lower, "发布产品") {
		return model.CategoryProduct
	}

	// 深度洞察
	if strings.Contains(lower, "深度") || strings.Contains(lower, "洞察") ||
		strings.Contains(lower, "insight") || strings.Contains(lower, "观点") ||
		strings.Contains(lower, "播客") || strings.Contains(lower, "podcast") ||
		strings.Contains(lower, "分析") || strings.Contains(lower, "趋势") ||
		strings.Contains(lower, "newsletter") || strings.Contains(lower, "博客") ||
		strings.Contains(lower, "blog") || strings.Contains(lower, "报告") ||
		strings.Contains(lower, "访谈") || strings.Contains(lower, "leader") ||
		strings.Contains(lower, "opinion") || strings.Contains(lower, "行业洞察") {
		return model.CategoryInsight
	}

	// 云服务与平台
	if strings.Contains(lower, "云服务") || strings.Contains(lower, "云平台") ||
		strings.Contains(lower, "cloud") || strings.Contains(lower, "maas") ||
		strings.Contains(lower, "aws") || strings.Contains(lower, "azure") ||
		strings.Contains(lower, "google cloud") || strings.Contains(lower, "gcp") ||
		strings.Contains(lower, "腾讯云") || strings.Contains(lower, "阿里云") ||
		strings.Contains(lower, "华为云") || strings.Contains(lower, "火山引擎") ||
		strings.Contains(lower, "api定价") || strings.Contains(lower, "api 定价") ||
		strings.Contains(lower, "云计算") || strings.Contains(lower, "平台服务") {
		return model.CategoryCloud
	}

	// AI工程
	if strings.Contains(lower, "ai工程") || strings.Contains(lower, "ai 工程") ||
		strings.Contains(lower, "engineering") || strings.Contains(lower, "agent") ||
		strings.Contains(lower, "rag") || strings.Contains(lower, "mcp") ||
		strings.Contains(lower, "a2a") || strings.Contains(lower, "prompt") ||
		strings.Contains(lower, "llmops") || strings.Contains(lower, "langchain") ||
		strings.Contains(lower, "框架") || strings.Contains(lower, "framework") ||
		strings.Contains(lower, "编程工具") || strings.Contains(lower, "cursor") ||
		strings.Contains(lower, "copilot") || strings.Contains(lower, "devin") {
		return model.CategoryAIEng
	}

	// AI基础设施
	if strings.Contains(lower, "基础设施") || strings.Contains(lower, "infrastructure") ||
		strings.Contains(lower, "芯片") || strings.Contains(lower, "chip") ||
		strings.Contains(lower, "gpu") || strings.Contains(lower, "nvidia") ||
		strings.Contains(lower, "推理引擎") || strings.Contains(lower, "inference") ||
		strings.Contains(lower, "训练框架") || strings.Contains(lower, "硬件") ||
		strings.Contains(lower, "hardware") || strings.Contains(lower, "vllm") ||
		strings.Contains(lower, "向量数据库") || strings.Contains(lower, "hpc") {
		return model.CategoryInfra
	}

	// 商业与投资
	if strings.Contains(lower, "商业") || strings.Contains(lower, "投资") ||
		strings.Contains(lower, "融资") || strings.Contains(lower, "business") ||
		strings.Contains(lower, "funding") || strings.Contains(lower, "investment") ||
		strings.Contains(lower, "并购") || strings.Contains(lower, "上市") ||
		strings.Contains(lower, "ipo") || strings.Contains(lower, "market") ||
		strings.Contains(lower, "竞争") || strings.Contains(lower, "战略") {
		return model.CategoryBiz
	}

	// AI安全
	if strings.Contains(lower, "安全") || strings.Contains(lower, "治理") ||
		strings.Contains(lower, "safety") || strings.Contains(lower, "governance") ||
		strings.Contains(lower, "监管") || strings.Contains(lower, "政策") ||
		strings.Contains(lower, "伦理") || strings.Contains(lower, "对齐") ||
		strings.Contains(lower, "regulation") || strings.Contains(lower, "alignment") ||
		strings.Contains(lower, "隐私") || strings.Contains(lower, "privacy") ||
		strings.Contains(lower, "许可") || strings.Contains(lower, "license") {
		return model.CategorySafety
	}

	return model.CategoryModelFrontier
}

// truncateText 截断文本到指定长度。
func truncateText(text string, maxLen int) string {
	text = strings.TrimSpace(text)
	runes := []rune(text)
	if len(runes) <= maxLen {
		return text
	}
	return string(runes[:maxLen]) + "..."
}

// openAIResponse OpenAI API 响应结构。
type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// claudeResponse Claude API 响应结构。
type claudeResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

// getTextContent 从 Claude 响应中提取 text 类型内容（跳过 thinking 块）。
func getTextContent(resp *claudeResponse) string {
	for _, c := range resp.Content {
		if c.Type == "text" && c.Text != "" {
			return c.Text
		}
	}
	for _, c := range resp.Content {
		if c.Text != "" {
			return c.Text
		}
	}
	return ""
}
