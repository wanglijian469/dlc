import { useEffect, useState } from "react";
import { Factory, Package, Plus, UserRound } from "lucide-react";
import { Link, useNavigate } from "react-router-dom";
import { listOwnMarketPosts, withdrawMarketPost } from "../api/market";
import { MarketPostCard } from "../components/public/MarketPostCard";
import { PageFrame } from "../components/public/PageFrame";
import type { MarketPost, PageResult } from "../types/api";

export function MyMarketPostsPage() {
  const navigate = useNavigate();
  const role = localStorage.getItem("cms_role");
  const username = localStorage.getItem("cms_username") || "";
  const [result, setResult] = useState<PageResult<MarketPost> | null>(null);
  const load = () => listOwnMarketPosts().then(setResult).catch(() => navigate(`/account/login?returnTo=${encodeURIComponent("/account/posts")}`, { replace: true }));
  useEffect(() => { void load(); }, []);
  return <PageFrame title="我的" subtitle={username ? `${username} · ${role === "vendor" ? "厂商账号" : "采购商账号"}` : "管理账号资料和供求信息"}>
    <div className="account-shortcuts"><Link to="/publish"><Plus size={20} /><span><strong>发布信息</strong><small>{role === "vendor" ? "发布供应" : "发布求购"}</small></span></Link>{role === "buyer" && <Link to="/account/profile"><UserRound size={20} /><span><strong>采购商资料</strong><small>维护联系人和地区</small></span></Link>}{role === "vendor" && <><Link to="/admin/vendor-profile"><Factory size={20} /><span><strong>厂家资料</strong><small>维护企业信息</small></span></Link><Link to="/admin/vendor-products"><Package size={20} /><span><strong>产品管理</strong><small>维护产品和提审状态</small></span></Link></>}<Link to="/account/login"><UserRound size={20} /><span><strong>账号中心</strong><small>登录和切换账号</small></span></Link></div>
    <header className="my-market-heading"><h2>我的发布</h2><Link className="primary-btn" to="/publish"><Plus size={16} />新增</Link></header>
    {!result ? <div className="state-page">正在加载…</div> : result.items.length ? <div className="market-post-grid">{result.items.map((post) => <MarketPostCard actions={<><Link className="outline-btn" to={`/publish/${post.id}`}>编辑</Link>{post.status !== "withdrawn" && <button className="outline-btn danger" onClick={() => void withdrawMarketPost(post.id).then(load)}>撤回</button>}</>} key={post.id} post={post} />)}</div> : <div className="state-page"><p>还没有发布供求信息</p><Link className="primary-btn" to="/publish">现在发布</Link></div>}
  </PageFrame>;
}
