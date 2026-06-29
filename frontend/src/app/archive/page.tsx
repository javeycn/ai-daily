import { loadIndex } from "@/lib/data";
import ArchiveClient from "./ArchiveClient";

export function generateMetadata() {
  return {
    title: "历史归档 - AI Daily",
  };
}

export default function ArchivePage() {
  const index = loadIndex();
  return <ArchiveClient index={index} />;
}
