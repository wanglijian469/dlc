import { Link } from "react-router-dom";

export function PlaceholderPage({ title }: { title: string }) {
  return (
    <main className="plain-page">
      <Link to="/">返回首页</Link>
      <h1>{title}</h1>
      <p>该栏目已预留入口，内容可在后续版本继续扩展。</p>
    </main>
  );
}
