import { useEffect, useState, type FormEvent } from "react";
import { AdminLayout } from "../../components/admin/AdminLayout";
import { createVendorPost, listVendorPosts, reviewVendorPost, submitVendorPost, updateVendorPost, uploadFile, withdrawVendorPost } from "../../api/admin";
import { getApiErrorMessage } from "../../api/client";
import type { VendorPost } from "../../types/api";

const empty = { postType: "update" as const, title: "", summary: "", content: "", coverImage: "" };
const statusText: Record<string, string> = { draft: "草稿", pending: "待审核", approved: "已发布", rejected: "已驳回", withdrawn: "已撤回" };

export function VendorPostsPage() {
  const role = localStorage.getItem("cms_role");
  const [items, setItems] = useState<VendorPost[]>([]);
  const [editing, setEditing] = useState<VendorPost>();
  const [form, setForm] = useState<Partial<VendorPost>>(empty);
  const [message, setMessage] = useState("");
  const load = () => listVendorPosts().then(setItems).catch((error) => setMessage(getApiErrorMessage(error, "内容加载失败")));
  useEffect(() => { void load(); }, []);
  if (role !== "vendor") return <AdminLayout title="厂商动态审核"><div className="admin-panel"><div className="vendor-post-list">{items.map((item) => <article key={item.id}><div><span>{item.postType === "case" ? "案例" : "动态"} · {statusText[item.status]}</span><h3>{item.title}</h3><p>{item.summary || item.content.slice(0, 100)}</p></div>{item.status === "pending" && <div><button className="primary-btn small" type="button" onClick={() => reviewVendorPost(item.id, "approved").then(load)}>通过</button><button className="outline-btn small" type="button" onClick={() => { const note = window.prompt("请填写驳回原因") || ""; if (note) reviewVendorPost(item.id, "rejected", note).then(load); }}>驳回</button></div>}</article>)}</div>{message && <p className="admin-message">{message}</p>}</div></AdminLayout>;
  const submit = (event: FormEvent) => {
    event.preventDefault();
    const action = editing ? updateVendorPost(editing.id, form) : createVendorPost(form);
    action.then(() => { setEditing(undefined); setForm(empty); setMessage("草稿已保存"); void load(); }).catch((error) => setMessage(getApiErrorMessage(error, "保存失败")));
  };
  return <AdminLayout title="企业动态与案例"><div className="admin-panel vendor-post-editor"><header><div><h2>主页内容</h2><p>企业动态和案例审核通过后展示在本厂官方主页。</p></div></header><form onSubmit={submit}><div className="auction-form-grid"><label>类型<select value={form.postType || "update"} onChange={(event) => setForm({ ...form, postType: event.target.value as "update" | "case" })}><option value="update">企业动态</option><option value="case">客户案例</option></select></label><label>标题<input required value={form.title || ""} onChange={(event) => setForm({ ...form, title: event.target.value })} /></label><label className="wide-field">摘要<textarea value={form.summary || ""} onChange={(event) => setForm({ ...form, summary: event.target.value })} /></label><label className="wide-field">封面图片<input accept="image/*" type="file" onChange={(event) => { const file = event.target.files?.[0]; if (!file) return; uploadFile(file).then((result) => setForm((value) => ({ ...value, coverImage: result.url, coverAssetId: result.assetId }))).catch((error) => setMessage(getApiErrorMessage(error, "封面上传失败"))); }} />{form.coverImage && <img className="auction-upload-preview" alt="动态封面预览" src={form.coverImage} />}</label><label className="wide-field">正文<textarea required rows={6} value={form.content || ""} onChange={(event) => setForm({ ...form, content: event.target.value })} /></label></div><button className="primary-btn" type="submit">保存草稿</button></form><div className="vendor-post-list">{items.map((item) => <article key={item.id}><div><span>{statusText[item.status]}</span><h3>{item.title}</h3>{item.reviewNote && <small>审核意见：{item.reviewNote}</small>}</div><div>{item.status !== "pending" && <button className="outline-btn small" type="button" onClick={() => { setEditing(item); setForm(item); }}>编辑</button>}{(item.status === "draft" || item.status === "rejected") && <button className="primary-btn small" type="button" onClick={() => submitVendorPost(item.id).then(load)}>提交审核</button>}{(item.status === "pending" || item.status === "approved") && <button className="outline-btn small" type="button" onClick={() => withdrawVendorPost(item.id).then(load)}>撤回</button>}</div></article>)}</div>{message && <p className="admin-message">{message}</p>}</div></AdminLayout>;
}
