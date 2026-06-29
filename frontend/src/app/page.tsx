import HomeClient from "./HomeClient";

export const metadata = {
  title: "AI Daily - 每日 AI 资讯",
};

export default function HomePage() {
  return <HomeClient report={null} />;
}
