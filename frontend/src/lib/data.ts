// 数据加载工具函数

import { DailyReport, DailyIndex, IndexFile } from "./types";
import fs from "fs";
import path from "path";

const DATA_DIR = path.join(process.cwd(), "data");
const DAILY_DIR = path.join(DATA_DIR, "daily");

/**
 * 加载全量索引文件
 */
export function loadIndex(): IndexFile {
  const indexPath = path.join(DATA_DIR, "index.json");
  try {
    const raw = fs.readFileSync(indexPath, "utf-8");
    return JSON.parse(raw) as IndexFile;
  } catch {
    return { days: [], updated: "" };
  }
}

/**
 * 加载指定日期的日报
 */
export function loadDailyReport(date: string): DailyReport | null {
  const filePath = path.join(DAILY_DIR, `${date}.json`);
  try {
    const raw = fs.readFileSync(filePath, "utf-8");
    return JSON.parse(raw) as DailyReport;
  } catch {
    return null;
  }
}

/**
 * 加载最新一期日报
 */
export function loadLatestReport(): DailyReport | null {
  const index = loadIndex();
  if (index.days.length === 0) return null;
  return loadDailyReport(index.days[0].date);
}

/**
 * 获取所有日报用于静态参数生成
 */
export function getAllDailyDates(): string[] {
  const index = loadIndex();
  return index.days.map((d) => d.date);
}

/**
 * 获取最近 N 天的文章精简数据用于搜索索引
 * 只保留搜索必要字段，减小索引体积
 */
export function buildSearchIndex(maxDays: number = 90): SearchIndexItem[] {
  const dates = getAllDailyDates().slice(0, maxDays);
  const items: SearchIndexItem[] = [];

  for (const date of dates) {
    const report = loadDailyReport(date);
    if (report) {
      for (const article of report.articles) {
        items.push({
          id: article.id,
          t: article.chinese_title || article.original_title,
          ot: article.original_title,
          s: article.summary,
          tg: article.tags,
          c: article.category,
          src: article.source,
          d: date,
        });
      }
    }
  }

  return items;
}

/**
 * 搜索索引精简结构（字段名缩短以减小 JSON 体积）
 */
export interface SearchIndexItem {
  id: string;   // article id
  t: string;    // chinese_title 或 original_title
  ot: string;   // original_title
  s: string;    // summary
  tg: string;   // tags
  c: string;    // category
  src: string;  // source
  d: string;    // date
}
