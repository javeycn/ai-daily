// TypeScript 类型定义

export interface Article {
  id: string;
  url: string;
  original_title: string;
  chinese_title: string;
  summary: string;
  recommendation?: string;
  source: string;
  image_url: string;
  tags: string;
  category: string;
  published_at: string;
  crawled_at: string;
}

export interface TagStat {
  tag: string;
  count: number;
}

export interface CategoryGroup {
  category: string;
  emoji: string;
  articles: Article[];
}

export interface DailyReport {
  date: string;
  title: string;
  summary: string;
  total_count: number;
  featured_count: number;
  tag_stats: TagStat[];
  articles: Article[];
  category_groups: CategoryGroup[];
  featured_category_groups?: CategoryGroup[];
  published_at: string;
}

export interface DailyIndex {
  date: string;
  title: string;
  summary: string;
  total_count: number;
  tags: string[];
}

export interface IndexFile {
  days: DailyIndex[];
  updated: string;
}

// 分类颜色映射（8 分类体系 + 旧分类兼容）
export const CATEGORY_COLORS: Record<string, { bg: string; text: string; border: string }> = {
  // 新 8 分类体系
  "模型前沿": { bg: "bg-blue-500/10", text: "text-blue-400", border: "border-blue-500/20" },
  "产品与应用": { bg: "bg-green-500/10", text: "text-green-400", border: "border-green-500/20" },
  "深度洞察": { bg: "bg-pink-500/10", text: "text-pink-400", border: "border-pink-500/20" },
  "云服务与平台": { bg: "bg-teal-500/10", text: "text-teal-400", border: "border-teal-500/20" },
  "AI工程": { bg: "bg-purple-500/10", text: "text-purple-400", border: "border-purple-500/20" },
  "AI基础设施": { bg: "bg-cyan-500/10", text: "text-cyan-400", border: "border-cyan-500/20" },
  "商业与投资": { bg: "bg-orange-500/10", text: "text-orange-400", border: "border-orange-500/20" },
  "AI安全": { bg: "bg-red-500/10", text: "text-red-400", border: "border-red-500/20" },
  // 旧分类兼容（数据迁移前过渡使用）
  "国际AI模型": { bg: "bg-blue-500/10", text: "text-blue-400", border: "border-blue-500/20" },
  "国内AI厂商": { bg: "bg-indigo-500/10", text: "text-indigo-400", border: "border-indigo-500/20" },
  "产品落地": { bg: "bg-green-500/10", text: "text-green-400", border: "border-green-500/20" },
  "开源": { bg: "bg-purple-500/10", text: "text-purple-400", border: "border-purple-500/20" },
  "商业硬件": { bg: "bg-orange-500/10", text: "text-orange-400", border: "border-orange-500/20" },
};

// 分类 emoji 映射（8 分类体系 + 旧分类兼容）
export const CATEGORY_EMOJI: Record<string, string> = {
  // 新 8 分类体系
  "模型前沿": "🧠",
  "产品与应用": "🚀",
  "深度洞察": "📊",
  "云服务与平台": "☁️",
  "AI工程": "⚙️",
  "AI基础设施": "🔧",
  "商业与投资": "💰",
  "AI安全": "🛡️",
  // 旧分类兼容（数据迁移前过渡使用）
  "国际AI模型": "🌐",
  "国内AI厂商": "🇨🇳",
  "产品落地": "📱",
  "开源": "💻",
  "商业硬件": "🔩",
};
