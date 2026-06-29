/**
 * 生成搜索索引 JSON 文件 + 预构建 Fuse 索引
 *
 * 优化策略：
 * 1. 只保留最近 30 天数据，大幅减小体积
 * 2. 预构建 Fuse.js 索引（构建时序列化），浏览器用 Fuse.parseIndex 直接加载
 *    避免在 19000+ 条数据上运行时建索引的主线程阻塞（3-8 秒）
 * 3. 输出文件包含版本号（基于 mtime hash），配合 Nginx 长缓存
 * 4. 同时输出 manifest 让客户端知道当前最新版本
 *
 * 输出文件：
 *   public/search-data-{hash}.json    — { items, fuseIndex }
 *   public/search-manifest.json        — { version: "{hash}" }
 */
import fs from "fs";
import path from "path";
import crypto from "crypto";
import { fileURLToPath } from "url";
import Fuse from "fuse.js";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const DATA_DIR = path.join(__dirname, "..", "data");
const DAILY_DIR = path.join(DATA_DIR, "daily");
const OUTPUT_DIR = path.join(__dirname, "..", "public");

const MAX_DAYS = 30;
const SUMMARY_MAX_LEN = 60;

const FUSE_KEYS = [
  { name: "t", weight: 2 },
  { name: "s", weight: 1 },
  { name: "tg", weight: 1.5 },
  { name: "c", weight: 1.5 },
  { name: "src", weight: 0.5 },
];

function main() {
  // 读取索引
  const indexPath = path.join(DATA_DIR, "index.json");
  if (!fs.existsSync(indexPath)) {
    console.error("index.json not found");
    process.exit(1);
  }

  const index = JSON.parse(fs.readFileSync(indexPath, "utf-8"));
  const dates = index.days.map((d) => d.date).slice(0, MAX_DAYS);

  console.log(`Generating search index for ${dates.length} days...`);

  const items = [];
  for (const date of dates) {
    const filePath = path.join(DAILY_DIR, `${date}.json`);
    if (!fs.existsSync(filePath)) continue;

    const report = JSON.parse(fs.readFileSync(filePath, "utf-8"));
    if (!report.articles) continue;

    for (const article of report.articles) {
      items.push({
        id: article.id,
        t: article.chinese_title || article.original_title,
        s: (article.summary || "").slice(0, SUMMARY_MAX_LEN),
        tg: article.tags || "",
        c: article.category || "",
        src: article.source || "",
        d: date,
      });
    }
  }

  // 预构建 Fuse 索引（关键优化：把建索引的耗时从浏览器搬到构建阶段）
  console.log(`Building Fuse index for ${items.length} items...`);
  const fuseIndex = Fuse.createIndex(FUSE_KEYS.map((k) => k.name), items);
  const serializedFuseIndex = fuseIndex.toJSON();

  // 组装输出
  const payload = {
    items,
    fuseIndex: serializedFuseIndex,
  };
  const json = JSON.stringify(payload);

  // 计算版本 hash（基于内容）
  const hash = crypto.createHash("md5").update(json).digest("hex").slice(0, 8);
  const filename = `search-data-${hash}.json`;
  const outputPath = path.join(OUTPUT_DIR, filename);

  // 确保输出目录存在
  fs.mkdirSync(OUTPUT_DIR, { recursive: true });

  // 清理旧的 search-data-*.json（保留最新的）
  for (const f of fs.readdirSync(OUTPUT_DIR)) {
    if (/^search-data-[a-f0-9]+\.json$/.test(f) && f !== filename) {
      fs.unlinkSync(path.join(OUTPUT_DIR, f));
    }
  }

  fs.writeFileSync(outputPath, json);

  // 写入 manifest
  const manifestPath = path.join(OUTPUT_DIR, "search-manifest.json");
  fs.writeFileSync(
    manifestPath,
    JSON.stringify({ version: hash, file: filename, count: items.length })
  );

  const sizeMB = (Buffer.byteLength(json) / 1024 / 1024).toFixed(2);
  console.log(`✓ Search data generated:`);
  console.log(`  Items: ${items.length}`);
  console.log(`  Size: ${sizeMB} MB`);
  console.log(`  File: ${filename}`);
  console.log(`  Manifest: search-manifest.json`);
}

main();
