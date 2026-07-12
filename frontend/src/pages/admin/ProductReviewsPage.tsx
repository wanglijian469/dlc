import { useEffect, useState } from "react";
import { CheckCircle2, GitMerge, XCircle } from "lucide-react";
import { listProductSubmissions, reviewProductSubmission } from "../../api/admin";
import { AdminLayout } from "../../components/admin/AdminLayout";
import { Pagination } from "../../components/public/Pagination";
import { ProtectedMediaImage } from "../../components/admin/ProtectedMediaImage";
import type { ProductSubmission } from "../../types/api";

export function ProductReviewsPage() {
  const [rows, setRows] = useState<ProductSubmission[]>([]);
  const [status, setStatus] = useState("pending");
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [notes, setNotes] = useState<Record<number, string>>({});
  const [mergeTargets, setMergeTargets] = useState<Record<number, string>>({});
  const [message, setMessage] = useState("");
  const pageSize = 10;
  const load = () => listProductSubmissions({ page, pageSize, status }).then((result) => { setRows(result.items); setTotal(result.total); }).catch(() => setMessage("产品审核队列加载失败"));
  useEffect(() => { void load(); }, [page, status]);
  const review = (row: ProductSubmission, next: "approved" | "rejected") => {
    const mergeProductId = Number(mergeTargets[row.id]) || undefined;
    reviewProductSubmission(row.id, next, notes[row.id] || "", mergeProductId).then(() => { setMessage(next === "approved" ? "审核已通过" : "已驳回提交"); void load(); }).catch(() => setMessage("审核失败，可能存在版本冲突或目标产品无效"));
  };
  return <AdminLayout title="产品审核"><section className="admin-panel product-review-panel">
    <header><div><h2>产品候选与供应关系审核</h2><p>审核新产品目录、已有产品关联和厂商供应信息修改。新产品候选可合并到已有目录产品。</p></div><select aria-label="审核状态" value={status} onChange={(event) => { setStatus(event.target.value); setPage(1); }}><option value="pending">待审核</option><option value="approved">已通过</option><option value="rejected">已驳回</option><option value="">全部</option></select></header>
    {message && <p className="admin-message">{message}</p>}
    <div className="review-list">{rows.map((row) => <article className="review-card" key={row.id}>
      <header><div><strong>{row.productDraft?.name || row.product?.name || "已有目录产品"}</strong><span>{typeLabel(row.submissionType)} · {row.vendor?.name || `厂商 #${row.vendorId}`}</span></div><span className={`status-badge ${row.status}`}>{statusLabel(row.status)}</span></header>
      <div className="review-diff-grid"><dl><div><dt>本厂产品名</dt><dd>{row.supplierDraft?.vendorProductName || "—"}</dd></div><div><dt>厂商型号</dt><dd>{row.supplierDraft?.vendorModel || "—"}</dd></div><div><dt>适配信息</dt><dd>{row.supplierDraft?.compatibleModels || row.productDraft?.compatibleModels || "—"}</dd></div><div><dt>价格说明</dt><dd>{row.supplierDraft?.priceNote || "—"}</dd></div></dl>{row.supplierDraft?.image && <ProtectedMediaImage assetId={assetId(row.supplierDraft.image)} alt="提交的产品图片" src={row.supplierDraft.image} />}</div>
      {row.status === "pending" && <div className="review-actions"><label>审核意见<input value={notes[row.id] || ""} onChange={(event) => setNotes({ ...notes, [row.id]: event.target.value })} /></label>{row.submissionType === "new_product" && <label><GitMerge size={15} />合并到已有产品 ID（留空则新建）<input inputMode="numeric" value={mergeTargets[row.id] || ""} onChange={(event) => setMergeTargets({ ...mergeTargets, [row.id]: event.target.value })} /></label>}<button className="primary-btn small" type="button" onClick={() => review(row, "approved")}><CheckCircle2 size={15} />通过</button><button className="outline-btn small danger" type="button" onClick={() => review(row, "rejected")}><XCircle size={15} />驳回</button></div>}
      {row.reviewNote && <p className="review-note">审核意见：{row.reviewNote}</p>}
    </article>)}</div>
    {!rows.length && <p className="structured-empty">当前没有符合条件的产品提交。</p>}
    <Pagination page={page} pageSize={pageSize} total={total} onChange={setPage} />
  </section></AdminLayout>;
}

function typeLabel(type: ProductSubmission["submissionType"]) { return ({ new_product: "新增产品候选", link_product: "关联已有产品", link_supplier: "关联已有产品", update_supplier: "修改供应信息", update_offer: "修改供应信息" } as Record<string, string>)[type] || type; }
function statusLabel(status: ProductSubmission["status"]) { return ({ pending: "待审核", approved: "已通过", rejected: "已驳回", superseded: "已被新版本替代" })[status]; }
function assetId(url?: string) { const match = url?.match(/\/api\/media\/(\d+)/); return match ? Number(match[1]) : undefined; }
