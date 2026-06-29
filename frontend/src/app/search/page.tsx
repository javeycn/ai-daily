import SearchClient from "./SearchClient";

export function generateMetadata() {
  return {
    title: "搜索资讯 - AI Daily",
  };
}

export default function SearchPage() {
  return <SearchClient />;
}
