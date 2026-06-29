import { loadDailyReport, getAllDailyDates } from "@/lib/data";
import DailyClient from "./DailyClient";

export function generateStaticParams() {
  // 增量构建：只预渲染最近 7 天页面
  // 历史日报已存在于站点目录中，部署时不会被覆盖（rsync --exclude="daily/"）
  const dates = getAllDailyDates().slice(0, 7);
  return dates.map((date) => ({ date }));
}

export function generateMetadata({ params }: { params: { date: string } }) {
  return {
    title: `AI 日报 - ${params.date}`,
  };
}

export default function DailyPage({ params }: { params: { date: string } }) {
  // SSG 预渲染时尝试加载本地数据，找不到也不 404，交给客户端动态加载
  const report = loadDailyReport(params.date);
  return <DailyClient report={report} date={params.date} />;
}
