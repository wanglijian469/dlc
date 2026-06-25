import { PageFrame } from "../components/public/PageFrame";

export function PlaceholderPage({ title, description }: { title: string; description?: string }) {
  return (
    <PageFrame title={title}>
      <div className="state-panel">
        <p>{description || "该栏目已预留入口，内容可在后续版本继续扩展。"}</p>
      </div>
    </PageFrame>
  );
}
