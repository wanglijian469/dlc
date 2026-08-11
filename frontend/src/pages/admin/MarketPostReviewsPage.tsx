import { useEffect, useState } from "react";
import { ExternalLink, RotateCcw, ShieldX } from "lucide-react";
import { Link } from "react-router-dom";
import { getApiErrorMessage } from "../../api/client";
import { listAdminMarketPosts, updateAdminMarketPostStatus } from "../../api/admin";
import { AdminLayout } from "../../components/admin/AdminLayout";
import { Pagination } from "../../components/public/Pagination";
import type { MarketPost, PageResult } from "../../types/api";

export function MarketPostReviewsPage() {
  const [status, setStatus] = useState("");
  const [page, setPage] = useState(1);
  const [result, setResult] = useState<PageResult<MarketPost> | null>(null);
  const [message, setMessage] = useState("");
  const load = () => listAdminMarketPosts({ page, pageSize: 20, status: status || undefined }).then(setResult).catch((reason) => setMessage(getApiErrorMessage(reason, "供求信息加载失败")));
  useEffect(() => { void load(); }, [page, status]);
  const changeStatus = (post: MarketPost, next: "published" | "removed") => {
    const reason = next === "removed" ? window.prompt("请输入下架原因", "内容不符合平台发布规范") : "";
    if (next === "removed" && reason === null) return;
    void updateAdminMarketPostStatus(post.id, next, reason || "").then(() => { setMessage(next === "removed" ? "已下架" : "已恢复公开"); load(); }).catch((error) => setMessage(getApiErrorMessage(error, "状态更新失败")));
  };
  return <AdminLayout title="供求治理">
    <p className="admin-page-description">处理已公开供求信息的事后治理、下架和恢复。</p>
    <div className="admin-toolbar"><select value={status} onChange={(event) => { setStatus(event.target.value); setPage(1); }}><option value="">全部状态</option><option value="published">公开中</option><option value="removed">已下架</option><option value="withdrawn">用户撤回</option></select></div>
    {message && <p className="admin-message">{message}</p>}
    <div className="admin-table-wrap"><table className="admin-table"><thead><tr><th>类型</th><th>标题</th><th>发布者</th><th>状态</th><th>发布时间</th><th>操作</th></tr></thead><tbody>{result?.items.map((post) => <tr key={post.id}><td>{post.type === "demand" ? "求购" : "供应"}</td><td>{post.title}</td><td>{post.publisherName}</td><td>{post.status}</td><td>{new Date(post.createdAt).toLocaleString("zh-CN")}</td><td><span className="admin-row-actions"><Link target="_blank" to={`/purchase/${post.id}`}><ExternalLink size={15} />查看</Link>{post.status === "published" ? <button onClick={() => changeStatus(post, "removed")}><ShieldX size={15} />下架</button> : post.status === "removed" && <button onClick={() => changeStatus(post, "published")}><RotateCcw size={15} />恢复</button>}</span></td></tr>)}</tbody></table></div>
    {result && <Pagination page={result.page} pageSize={result.pageSize} total={result.total} onChange={setPage} />}
  </AdminLayout>;
}
