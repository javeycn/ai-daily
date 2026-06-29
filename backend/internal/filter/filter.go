// Package filter 提供 AI 相关关键词过滤功能。
package filter

import (
	"strings"

	"ai-news-crawler/internal/model"
)

// aiKeywords 包含用于判断文章是否与 AI 相关的关键词。
var aiKeywords = []string{
	// 通用 AI 关键词
	"artificial intelligence",
	"machine learning",
	"deep learning",
	"neural network",
	"large language model",
	"llm",
	"gpt",
	"generative ai",
	"chatgpt",
	"claude",
	"gemini",
	"transformer",
	"natural language processing",
	"nlp",
	"computer vision",
	"reinforcement learning",
	"autonomous driving",
	"self-driving",
	"robotics",
	"foundation model",
	"multimodal",
	"retrieval augmented generation",
	"rag",
	"fine-tuning",
	"ai regulation",
	"ai ethics",
	"ai safety",
	"agentic",
	"ai agent",
	"copilot",
	"speech recognition",
	"image generation",
	"code generation",
	"training data",
	"inference",
	"prompt engineering",
	"embeddings",
	"vector database",
	"knowledge graph",

	// 国际 AI 模型和厂商
	"openai",
	"anthropic",
	"google ai",
	"google deepmind",
	"meta ai",
	"stability ai",
	"midjourney",
	"dall-e",
	"stable diffusion",
	"microsoft ai",
	"microsoft copilot",
	"mistral",
	"cohere",
	"perplexity",
	"xai",
	"grok",
	"sora",
	"o1",
	"o3",

	// 国内 AI 厂商和产品
	"百度",
	"文心一言",
	"ernie",
	"阿里",
	"通义千问",
	"qwen",
	"腾讯",
	"混元",
	"hunyuan",
	"字节跳动",
	"豆包",
	"doubao",
	"讯飞",
	"星火",
	"智谱",
	"chatglm",
	"glm",
	"月之暗面",
	"kimi",
	"minimax",
	"零一万物",
	"yi",
	"deepseek",
	"商汤",
	"sensetime",
	"旷视",
	"megvii",
	// 国内 AI 新锐 & 应用生态补充
	"百川智能",
	"baichuan",
	"moonshot",
	"深度求索",
	"coze",
	"扣子",
	"千帆大模型",
	"阿里云百炼",
	"腾讯云智能",
	"华为昇思",
	"mindspore",
	"日日新",
	"海螺ai",
	"wps ai",
	"钉钉 ai",
	"飞书 ai",
	"国产大模型",
	"算力中心",

	// 云计算
	"aws",
	"amazon web services",
	"azure",
	"google cloud",
	"gcp",
	"阿里云",
	"alicloud",
	"腾讯云",
	"tencent cloud",
	"华为云",
	"huawei cloud",
	"cloud computing",
	"serverless",
	"kubernetes",
	"cloud native",
	"saas",
	"paas",
	"iaas",

	// 开源
	"open source",
	"opensource",
	"huggingface",
	"hugging face",
	"llama",
	"pytorch",
	"tensorflow",
	"jax",
	"langchain",
	"llamaindex",
	"ollama",
	"vllm",
	"ggml",
	"gguf",
	"transformers",
	"github",

	// 硬件
	"nvidia",
	"gpu",
	"tpu",
	"ai chip",
	"ai芯片",
	"cuda",
	"amd",
	"intel",
	"芯片",
	"semiconductor",
	"h100",
	"h200",
	"a100",
	"blackwell",
	"groq",
	"cerebras",
	"寒武纪",
	"昇腾",
	"ascend",

	// 通用中文关键词
	"人工智能",
	"机器学习",
	"深度学习",
	"大模型",
	"大语言模型",
	"生成式AI",
	"自动驾驶",
	"推荐算法",
	"AI助手",
	"智能体",
	"多模态",
	"具身智能",
	"ai startup",
	"ai funding",
	"ai research",
}

// IsAIRelated 判断文章是否与 AI 相关。
func IsAIRelated(article *model.Article) bool {
	text := strings.ToLower(article.OriginalTitle + " " + article.Summary)

	for _, kw := range aiKeywords {
		if strings.Contains(text, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

// FilterArticles 过滤出与 AI 相关的文章。
func FilterArticles(articles []*model.Article) []*model.Article {
	var result []*model.Article
	for _, a := range articles {
		if IsAIRelated(a) {
			result = append(result, a)
		}
	}
	return result
}
