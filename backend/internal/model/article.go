// Package model 定义 AI 新闻采集系统的核心数据模型。
package model

import (
	"strings"
	"time"
)

// Category 常量定义八大分类（按内容本质分类，不按地域）。
const (
	CategoryModelFrontier = "模型前沿"   // 模型能力、开源模型、Benchmark、训练/推理新范式
	CategoryProduct       = "产品与应用"  // AI 产品发布、行业落地案例、AI 工具
	CategoryInsight       = "深度洞察"   // 趋势分析、领袖观点、行业报告、深度技术博文
	CategoryCloud         = "云服务与平台" // 云厂商 AI 服务动态、MaaS 平台、API 定价
	CategoryAIEng         = "AI工程"    // Agent 框架、RAG 架构、MCP/A2A 协议、Prompt 工程、LLMOps、AI 编程工具
	CategoryInfra         = "AI基础设施"  // AI 芯片、推理引擎、训练框架、向量数据库
	CategoryBiz           = "商业与投资"  // 融资、并购、IPO、市场格局、企业战略
	CategorySafety        = "AI安全"    // AI 监管法规、安全对齐、数据隐私、开源许可、伦理争议
)

// CategoryQuota 定义每个分类的软性配额。
type CategoryQuota struct {
	Min       int // 最低保底条数
	Preferred int // 建议条数上限
}

// CategoryQuotas 8 分类的软性配额表。
// Min 合计 = 21，总精选 30 篇，留 9 个名额由第二阶段按来源优先级补位。
// 设计原则：
//   - 模型前沿控制占比（Preferred=4），避免挤占其他分类
//   - 云服务与平台重点加强（Min=5），国内 ≥ 3 条
//   - AI安全 ≤ 云服务，与 AI工程/AI基础设施持平
var CategoryQuotas = map[string]CategoryQuota{
	CategoryModelFrontier: {Min: 3, Preferred: 4}, // 模型能力突破（适当控制，让位给其他分类）
	CategoryProduct:       {Min: 3, Preferred: 5}, // 产品落地案例（国内/国际均衡）
	CategoryInsight:       {Min: 1, Preferred: 4}, // 深度分析（精选少而精）
	CategoryCloud:         {Min: 5, Preferred: 6}, // 云服务（头部国内优先+国际≤2，控制首页条数≤6）
	CategoryAIEng:         {Min: 2, Preferred: 4}, // AI 工程实践
	CategoryInfra:         {Min: 2, Preferred: 3}, // 基础设施
	CategoryBiz:           {Min: 2, Preferred: 3}, // 商业与投资
	CategorySafety:        {Min: 2, Preferred: 3}, // AI 安全（≤ 云服务，与 AI工程/基础设施持平）
}

// CategoryEmoji 分类对应的 emoji 图标。
var CategoryEmoji = map[string]string{
	CategoryModelFrontier: "🧠",
	CategoryProduct:       "🚀",
	CategoryInsight:       "📊",
	CategoryCloud:         "☁️",
	CategoryAIEng:         "⚙️",
	CategoryInfra:         "🔧",
	CategoryBiz:           "💰",
	CategorySafety:        "🛡️",
}

// DomesticCloudKeywords 国内云厂商关键词，用于识别国内云服务相关内容。
// 注意：匹配时会做大小写归一化（ToLower）。
// 不包含通用术语（如 MaaS）避免误匹配国际云文章。
var DomesticCloudKeywords = []string{
	"腾讯", "腾讯云", "阿里", "阿里云", "百度", "百度智能云",
	"华为", "华为云", "金山", "金山云", "火山", "火山引擎", "字节",
	"商汤", "智谱", "讯飞", "京东", "京东云", "美团",
	"通义", "千问", "文心", "盘古", "混元",
	"tokenhub", "qwen", "deepseek",
	"天翼云", "移动云", "联通云", "浪潮云", "紫光云", "青云",
	"七牛", "又拍云", "ucloud",
	"昇腾", "昇思", "千帆", "百炼",
}

// InternationalCloudKeywords 国际云厂商关键词，用于识别国际云服务内容。
// 同时包含英文和中文名称，确保中文翻译的文章也能正确识别。
var InternationalCloudKeywords = []string{
	"aws", "amazon bedrock", "sagemaker", "amazon web services",
	"azure", "microsoft azure", "copilot studio", "azure openai",
	"google cloud", "vertex ai", "google kubernetes", "gke",
	"gcp", "bigquery", "google data cloud",
	"oracle cloud", "ibm cloud", "ibm watsonx",
	"salesforce einstein", "snowflake",
	// 中文名称（确保中文翻译文章也能识别）
	"谷歌云", "亚马逊云", "微软云", "甲骨文云",
}

// TopDomesticCloudKeywords 头部国内云厂商关键词（腾讯云、阿里云、火山引擎、百度云、华为云），优先级最高。
// 涵盖 CodeBuddy、WorkBuddy、OpenClaw、Qwen、豆包等核心产品。
var TopDomesticCloudKeywords = []string{
	"腾讯", "腾讯云", "混元", "tokenhub",
	"阿里", "阿里云", "通义", "千问", "qwen", "百炼",
	"火山引擎", "火山", "字节", "豆包", "火山方舟",
	"百度", "百度云", "百度智能云", "文心", "千帆",
	"华为", "华为云", "盘古", "昇腾", "昇思",
	"codebuddy", "workbuddy", "openclaw",
}

// HotProductKeywords 当前热门 AI 产品关键词，用于在产品与应用分类中提权。
var HotProductKeywords = []string{
	// 国内热门产品
	"CodeBuddy", "codebuddy", "WorkBuddy", "workbuddy",
	"OpenClaw", "openclaw",
	"豆包", "Doubao",
	"Kimi", "kimi", "月之暗面",
	"元宝", "混元",
	"通义", "千问",
	"文心一言",
	"扣子", "Coze", "coze",
	"飞书", "钉钉",
	"Manus", "manus",
	// 国际热门产品
	"ChatGPT", "chatgpt",
	"Claude", "claude",
	"Cursor", "cursor",
	"Copilot", "copilot",
	"Gemini", "gemini",
	"Perplexity", "perplexity",
	"Midjourney", "midjourney",
	"Sora", "sora",
	"Devin", "devin",
	"Windsurf", "windsurf",
	"Lovable", "lovable",
	"Bolt", "bolt",
}

// TopDomesticProductKeywords 国内大厂热门产品关键词，在产品与应用分类中权重最高。
var TopDomesticProductKeywords = []string{
	"腾讯", "元宝", "混元", "企业微信", "微信",
	"阿里", "通义", "千问", "钉钉", "淘宝", "天猫",
	"字节", "火山", "豆包", "飞书", "抖音",
	"CodeBuddy", "codebuddy", "WorkBuddy", "workbuddy",
	"OpenClaw", "openclaw",
	"Kimi", "kimi", "月之暗面",
	"百度", "文心一言",
	"科大讯飞", "讯飞",
	"扣子", "Coze", "coze",
	"Manus", "manus",
}

// DomesticCloudSources 国内来源名称，辅助国内内容识别。
var DomesticCloudSources = map[string]bool{
	"智东西":    true,
	"36氪":    true,
	"IT之家":   true,
	"机器之心":   true,
	"量子位":    true,
	"InfoQ中文": true,
}

// DomesticModelKeywords 国内模型厂商关键词，用于识别国内模型相关内容。
var DomesticModelKeywords = []string{
	"智谱", "GLM", "ChatGLM",
	"minimax", "MiniMax",
	"kimi", "Kimi", "月之暗面", "Moonshot",
	"qwen", "Qwen", "通义", "千问", "阿里云",
	"豆包", "字节", "火山", "云雀",
	"元宝", "腾讯", "混元", "Hunyuan",
	"百度", "文心", "ERNIE",
	"讯飞", "星火",
	"商汤", "SenseChat",
	"零一万物", "Yi",
	"百川", "Baichuan",
	"昆仑万维", "天工",
	"深度求索", "DeepSeek",
	"阶跃星辰", "Step",
}

// TopDomesticModelKeywords 国内头部模型厂商关键词（优先级高于其他国内厂商）。
var TopDomesticModelKeywords = []string{
	"智谱", "GLM", "ChatGLM",
	"minimax", "MiniMax",
	"kimi", "Kimi", "月之暗面",
	"qwen", "Qwen", "通义", "千问",
	"豆包", "元宝", "混元",
	"DeepSeek", "深度求索",
}

// InternationalModelKeywords 国际模型厂商关键词。
var InternationalModelKeywords = []string{
	"OpenAI", "GPT", "ChatGPT", "o1", "o3",
	"Anthropic", "Claude",
	"Google", "Gemini", "Bard", "DeepMind",
	"Grok", "xAI",
	"Meta", "LLaMA", "Llama",
	"Mistral",
	"Cohere",
	"Stability", "Stable Diffusion",
	"Midjourney",
	"Perplexity",
	"Inflection",
}

// TopInternationalModelKeywords 国际头部模型厂商关键词（优先级高于其他国际厂商）。
var TopInternationalModelKeywords = []string{
	"OpenAI", "GPT", "ChatGPT", "o1", "o3",
	"Anthropic", "Claude",
	"Google", "Gemini", "DeepMind",
	"Grok", "xAI",
	"Meta", "LLaMA", "Llama",
}

// DomesticProductKeywords 国内产品/应用关键词，用于识别国内产品内容。
var DomesticProductKeywords = []string{
	"腾讯", "阿里", "百度", "华为", "字节", "美团", "京东",
	"小米", "OPPO", "vivo", "荣耀", "联想",
	"飞书", "钉钉", "企业微信", "WPS",
	"抖音", "快手", "微信", "支付宝", "淘宝", "天猫",
	"哔哩哔哩", "B站", "知乎", "小红书",
	"网易", "搜狐", "新浪",
	"科大讯飞", "商汤", "旷视",
	"国内", "中国",
}

// AgentProductKeywords AI Agent/智能体相关产品关键词，用于在产品与应用分类中提权。
var AgentProductKeywords = []string{
	"Agent", "agent", "智能体",
	"Manus", "manus",
	"OpenClaw", "openclaw",
	"WorkBuddy", "workbuddy", "Workbuddy",
	"AutoGPT", "autogpt", "Auto-GPT",
	"BabyAGI", "babyagi",
	"CrewAI", "crewai",
	"MetaGPT", "metagpt",
	"AutoGen", "autogen",
	"Devin", "devin",
	"Copilot", "copilot",
	"Cursor", "cursor",
	"Bolt", "bolt",
	"Replit Agent", "replit agent",
	"Windsurf", "windsurf",
	"CodeBuddy", "codebuddy",
	"Lovable", "lovable",
	"AI助手", "AI 助手", "AI工具", "AI 工具",
	"多智能体", "Multi-Agent", "multi-agent",
	"Agentic", "agentic",
	"AI编程", "AI 编程",
	"扣子", "Coze", "coze",
}

// MaxPerVendorInCategory 同一分类下同一厂商的最大条数（防止重复过多）。
const MaxPerVendorInCategory = 2

// NonCloudExcludeKeywords 非云服务场景排除关键词（汽车/硬件/消费电子等）。
// 包含这些关键词的文章即使匹配了云厂商名称，也不应被视为国内云服务内容。
var NonCloudExcludeKeywords = []string{
	"智行", "问界", "尚界", "享界", "智界", "鸿蒙座舱", "鸿蒙智行",
	"乾崑", "交付", "试驾", "上市售价", "预售", "车型",
	"激光雷达", "自动驾驶", "辅助驾驶", "座椅",
	"手机", "平板", "笔记本", "电视", "穿戴", "耳机",
	"广汽", "丰田", "比亚迪", "蔚来", "小鹏", "理想", "极氪",
}

// isNonCloudContent 检查文章是否为非云服务内容（汽车/硬件/消费电子等）。
func (a *Article) isNonCloudContent() bool {
	text := strings.ToLower(a.ChineseTitle + " " + a.OriginalTitle + " " + a.Summary)
	for _, kw := range NonCloudExcludeKeywords {
		if strings.Contains(text, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

// VendorAliases 厂商名称归一化映射，将不同写法映射到统一的厂商标识。
var VendorAliases = map[string]string{
	// 国际厂商
	"openai": "OpenAI", "gpt": "OpenAI", "chatgpt": "OpenAI", "o1": "OpenAI", "o3": "OpenAI",
	"anthropic": "Anthropic", "claude": "Anthropic",
	"google": "Google", "gemini": "Google", "deepmind": "Google", "bard": "Google",
	"grok": "xAI", "xai": "xAI",
	"meta": "Meta", "llama": "Meta",
	"aws": "AWS", "amazon": "AWS",
	"azure": "Azure", "microsoft": "Azure",
	"gcp": "GCP",
	"nvidia": "NVIDIA",
	"mistral": "Mistral",
	// 国内厂商
	"智谱": "智谱", "glm": "智谱", "chatglm": "智谱",
	"minimax": "MiniMax",
	"kimi": "月之暗面", "月之暗面": "月之暗面", "moonshot": "月之暗面",
	"qwen": "通义千问", "通义": "通义千问", "千问": "通义千问", "阿里云": "通义千问", "阿里": "通义千问",
	"豆包": "字节跳动", "字节": "字节跳动", "火山": "字节跳动", "云雀": "字节跳动",
	"元宝": "腾讯", "腾讯": "腾讯", "混元": "腾讯", "hunyuan": "腾讯",
	"百度": "百度", "文心": "百度", "ernie": "百度",
	"讯飞": "讯飞", "星火": "讯飞",
	"商汤": "商汤", "sensechat": "商汤",
	"deepseek": "DeepSeek", "深度求索": "DeepSeek",
	"零一万物": "零一万物", "yi": "零一万物",
	"百川": "百川", "baichuan": "百川",
	"华为": "华为", "盘古": "华为",
	"金山": "金山",
	"京东": "京东",
	"美团": "美团",
}

// IsDomesticCloud 判断文章是否与国内云服务相关。
// 使用大小写不敏感匹配。增加反向排除：若同时匹配国际云关键词，优先判定为国际。
// 排除非云服务场景（汽车/硬件/消费电子等）。
func (a *Article) IsDomesticCloud() bool {
	// 排除非云服务内容（汽车/硬件等）
	if a.isNonCloudContent() {
		return false
	}

	// 国内源 + 云服务分类 → 一定是国内云服务
	if DomesticCloudSources[a.Source] && a.Category == CategoryCloud {
		return true
	}

	text := strings.ToLower(a.ChineseTitle + " " + a.OriginalTitle + " " + a.Summary + " " + a.Tags)

	// 先检查是否匹配国内关键词
	matchedDomestic := false
	for _, kw := range DomesticCloudKeywords {
		if len(kw) > 0 && strings.Contains(text, strings.ToLower(kw)) {
			matchedDomestic = true
			break
		}
	}
	if !matchedDomestic {
		return false
	}

	// 反向排除：如果同时匹配国际云关键词，认为是国际文章（避免误判）
	// 例如 "Azure OpenAI" 的摘要中可能提到 "MaaS" 或 "tokenhub"
	srcLower := strings.ToLower(a.Source)
	fullText := text + " " + srcLower
	for _, kw := range InternationalCloudKeywords {
		if len(kw) > 0 && strings.Contains(fullText, strings.ToLower(kw)) {
			return false // 同时匹配国际 → 不算国内
		}
	}
	return true
}

// IsInternationalCloud 判断文章是否与国际云服务相关。
// 用于云服务分类中严格区分国际内容并控制其数量。
func (a *Article) IsInternationalCloud() bool {
	text := strings.ToLower(a.ChineseTitle + " " + a.OriginalTitle + " " + a.Summary + " " + a.Tags + " " + a.Source)
	for _, kw := range InternationalCloudKeywords {
		if len(kw) > 0 && strings.Contains(text, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

// IsTopDomesticCloud 判断文章是否与头部国内云厂商相关（腾讯云、阿里云、火山引擎系）。
// 这些厂商的文章在云服务分类中权重最高，排序最靠前。
// 排除非云服务场景（汽车/硬件/消费电子等）。
func (a *Article) IsTopDomesticCloud() bool {
	if a.isNonCloudContent() {
		return false
	}
	if a.IsInternationalCloud() {
		return false // 国际文章不算头部国内
	}
	text := strings.ToLower(a.ChineseTitle + " " + a.OriginalTitle + " " + a.Summary + " " + a.Tags)
	for _, kw := range TopDomesticCloudKeywords {
		if len(kw) > 0 && strings.Contains(text, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

// IsHotProduct 判断文章是否涉及当前热门 AI 产品。
func (a *Article) IsHotProduct() bool {
	text := strings.ToLower(a.ChineseTitle + " " + a.OriginalTitle + " " + a.Summary + " " + a.Tags)
	for _, kw := range HotProductKeywords {
		if strings.Contains(text, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

// IsTopDomesticProduct 判断文章是否与国内大厂热门产品相关。
func (a *Article) IsTopDomesticProduct() bool {
	text := strings.ToLower(a.ChineseTitle + " " + a.OriginalTitle + " " + a.Summary + " " + a.Tags)
	for _, kw := range TopDomesticProductKeywords {
		if strings.Contains(text, strings.ToLower(kw)) {
			return true
		}
	}
	// 国内源的产品文章
	if DomesticCloudSources[a.Source] && a.Category == CategoryProduct {
		return true
	}
	return false
}

// IsDomesticModel 判断文章是否与国内模型厂商相关。
func (a *Article) IsDomesticModel() bool {
	text := a.ChineseTitle + " " + a.OriginalTitle + " " + a.Summary + " " + a.Tags
	lower := strings.ToLower(text)
	for _, kw := range DomesticModelKeywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			return true
		}
	}
	// 国内源发布的模型前沿文章大概率是国内模型相关
	if DomesticCloudSources[a.Source] && a.Category == CategoryModelFrontier {
		return true
	}
	return false
}

// IsTopDomesticModel 判断文章是否与国内头部模型厂商相关。
func (a *Article) IsTopDomesticModel() bool {
	text := a.ChineseTitle + " " + a.OriginalTitle + " " + a.Summary + " " + a.Tags
	lower := strings.ToLower(text)
	for _, kw := range TopDomesticModelKeywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

// IsInternationalModel 判断文章是否与国际模型厂商相关。
func (a *Article) IsInternationalModel() bool {
	text := a.ChineseTitle + " " + a.OriginalTitle + " " + a.Summary + " " + a.Tags
	lower := strings.ToLower(text)
	for _, kw := range InternationalModelKeywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

// IsTopInternationalModel 判断文章是否与国际头部模型厂商相关。
func (a *Article) IsTopInternationalModel() bool {
	text := a.ChineseTitle + " " + a.OriginalTitle + " " + a.Summary + " " + a.Tags
	lower := strings.ToLower(text)
	for _, kw := range TopInternationalModelKeywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

// IsAgentProduct 判断文章是否与 AI Agent/智能体相关产品有关。
func (a *Article) IsAgentProduct() bool {
	text := a.ChineseTitle + " " + a.OriginalTitle + " " + a.Summary + " " + a.Tags
	lower := strings.ToLower(text)
	for _, kw := range AgentProductKeywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

// IsDomesticProduct 判断文章是否与国内产品/应用相关。
func (a *Article) IsDomesticProduct() bool {
	text := a.ChineseTitle + " " + a.OriginalTitle + " " + a.Summary + " " + a.Tags
	for _, kw := range DomesticProductKeywords {
		if len(kw) > 0 && containsStr(text, kw) {
			return true
		}
	}
	// 国内源的产品文章
	if DomesticCloudSources[a.Source] && a.Category == CategoryProduct {
		return true
	}
	return false
}

// ExtractVendor 从文章中提取厂商标识（归一化后的名称）。
// 用于同厂商去重。返回空字符串表示未识别到特定厂商。
func (a *Article) ExtractVendor() string {
	text := strings.ToLower(a.ChineseTitle + " " + a.OriginalTitle + " " + a.Summary)
	// 按关键词长度从长到短匹配，避免短关键词误匹配
	// 先尝试精确匹配较长的关键词
	bestVendor := ""
	bestLen := 0
	for keyword, vendor := range VendorAliases {
		if strings.Contains(text, keyword) && len(keyword) > bestLen {
			bestVendor = vendor
			bestLen = len(keyword)
		}
	}
	return bestVendor
}

// containsStr 简单子串匹配。
func containsStr(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) && strings.Contains(s, substr)
}

// AllCategories 按展示顺序排列的全部分类列表。
var AllCategories = []string{
	CategoryModelFrontier,
	CategoryProduct,
	CategoryInsight,
	CategoryCloud,
	CategoryAIEng,
	CategoryInfra,
	CategoryBiz,
	CategorySafety,
}

// Article 表示一篇采集到的 AI 新闻文章。
type Article struct {
	ID              string    `json:"id" db:"id"`
	URL             string    `json:"url" db:"url"`
	OriginalTitle   string    `json:"original_title" db:"original_title"`
	ChineseTitle    string    `json:"chinese_title" db:"chinese_title"`
	Summary         string    `json:"summary" db:"summary"`
	Recommendation  string    `json:"recommendation" db:"recommendation"` // LLM 生成的 1 句话推荐理由，回答"为什么值得读"
	Source          string    `json:"source" db:"source"`
	ImageURL        string    `json:"image_url" db:"image_url"`
	Tags            string    `json:"tags" db:"tags"`
	Category        string    `json:"category" db:"category"`
	ImportanceScore int       `json:"importance_score" db:"importance_score"` // LLM 评估的重要度评分（1-10），用于同分类内排序
	PublishedAt     time.Time `json:"published_at" db:"published_at"`
	CrawledAt       time.Time `json:"crawled_at" db:"crawled_at"`
	Hash            string    `json:"-" db:"hash"`
}

// CategoryGroup 按分类分组的文章集合。
type CategoryGroup struct {
	Category string    `json:"category"`
	Emoji    string    `json:"emoji"`
	Articles []Article `json:"articles"`
}

// DailyReport 表示一期每日 AI 资讯日报。
type DailyReport struct {
	Date                   string          `json:"date"`
	Title                  string          `json:"title"`
	Summary                string          `json:"summary"`
	TotalCount             int             `json:"total_count"`
	FeaturedCount          int             `json:"featured_count"`
	TagStats               []TagStat       `json:"tag_stats"`
	Articles               []Article       `json:"articles"`
	CategoryGroups         []CategoryGroup `json:"category_groups"`          // 全量分组（日报详情页使用）
	FeaturedCategoryGroups []CategoryGroup `json:"featured_category_groups"` // 精选分组（首页使用，仅含精选 30 条）
	PublishedAt            string          `json:"published_at"`
}

// TagStat 表示某个标签的文章统计。
type TagStat struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

// DailyIndex 全量索引条目，用于归档页和搜索。
type DailyIndex struct {
	Date       string   `json:"date"`
	Title      string   `json:"title"`
	Summary    string   `json:"summary"`
	TotalCount int      `json:"total_count"`
	Tags       []string `json:"tags"`
}

// IndexFile 全量索引文件结构。
type IndexFile struct {
	Days    []DailyIndex `json:"days"`
	Updated string       `json:"updated"`
}
