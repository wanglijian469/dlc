import { FormEvent, useEffect, useState } from "react";
import { Plus, Search } from "lucide-react";
import { Link, useSearchParams } from "react-router-dom";
import { listMarketPosts } from "../api/market";
import { MarketPostCard } from "../components/public/MarketPostCard";
import { PageFrame } from "../components/public/PageFrame";
import { Pagination } from "../components/public/Pagination";
import type { MarketPost, PageResult } from "../types/api";

export function MarketPostsPage() {
  const [params, setParams] = useSearchParams();
  const [keyword, setKeyword] = useState(params.get("keyword") || "");
  const [result, setResult] = useState<PageResult<MarketPost> | null>(null);
  const [error, setError] = useState("");
  const type = params.get("type") as "supply" | "demand" | null;
  const page = Number(params.get("page") || 1);
  useEffect(() => { setError(""); listMarketPosts({ type: type || undefined, keyword: params.get("keyword") || undefined, page, pageSize: 18 }).then(setResult).catch(() => setError("供求信息加载失败")); }, [params]);
  const submit = (event: FormEvent) => { event.preventDefault(); const next = new URLSearchParams(params); keyword ? next.set("keyword", keyword) : next.delete("keyword"); next.delete("page"); setParams(next); };
  const selectType = (nextType?: string) => { const next = new URLSearchParams(params); nextType ? next.set("type", nextType) : next.delete("type"); next.delete("page"); setParams(next); };
  return <PageFrame title="供求信息" subtitle="采购商发布真实求购，源头厂家发布供应信息">
    <div className="market-toolbar"><div className="market-type-tabs"><button className={!type ? "active" : ""} onClick={() => selectType()}>全部</button><button className={type === "demand" ? "active" : ""} onClick={() => selectType("demand")}>求购</button><button className={type === "supply" ? "active" : ""} onClick={() => selectType("supply")}>供应</button></div><Link className="primary-btn market-publish-button" to="/publish"><Plus size={17} />发布信息</Link></div>
    <form className="market-search" onSubmit={submit}><Search size={18} /><input placeholder="搜索配件、机型或需求" value={keyword} onChange={(event) => setKeyword(event.target.value)} /><button>搜索</button></form>
    {error && <div className="state-page"><p>{error}</p></div>}
    {!result ? <div className="market-post-grid market-loading">正在加载…</div> : result.items.length ? <><div className="market-post-grid">{result.items.map((post) => <MarketPostCard key={post.id} post={post} />)}</div><Pagination page={result.page} pageSize={result.pageSize} total={result.total} onChange={(nextPage) => { const next = new URLSearchParams(params); next.set("page", String(nextPage)); setParams(next); }} /></> : <div className="state-page"><p>暂无符合条件的供求信息</p><Link className="primary-btn" to="/publish">发布第一条信息</Link></div>}
  </PageFrame>;
}
