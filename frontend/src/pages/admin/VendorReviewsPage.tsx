import { useEffect, useMemo, useState } from "react";
import { Check, ExternalLink, X } from "lucide-react";
import { listVendorSubmissionPage, reviewVendorSubmission, type VendorSubmission } from "../../api/admin";
import { AdminLayout } from "../../components/admin/AdminLayout";
import { Pagination } from "../../components/public/Pagination";
import type { Vendor } from "../../types/api";
import { ProtectedMediaImage } from "../../components/admin/ProtectedMediaImage";

const comparedFields: Array<{ key: keyof Vendor; label: string }> = [
  { key: "name", label: "厂商名称" }, { key: "shortName", label: "简称" }, { key: "logo", label: "Logo" }, { key: "coverImage", label: "封面图" },
  { key: "province", label: "省份" }, { key: "city", label: "城市" }, { key: "county", label: "区县" }, { key: "address", label: "地址" },
  { key: "mainProducts", label: "主营产品" }, { key: "serviceModels", label: "适配机型" }, { key: "serviceAdvantages", label: "服务优势" }, { key: "description", label: "公司简介" },
  { key: "establishedYear", label: "成立年份" }, { key: "factoryArea", label: "厂房面积" }, { key: "employeeCount", label: "员工规模" }, { key: "annualCapacity", label: "年产能" },
  { key: "equipment", label: "主要设备" }, { key: "certifications", label: "认证资质" }, { key: "qualityControl", label: "质检能力" }, { key: "supplyRegions", label: "供货区域" },
  { key: "cooperationTerms", label: "合作方式" }, { key: "afterSalesService", label: "售后服务" }, { key: "websiteUrl", label: "官网" }, { key: "contactName", label: "联系人" }, { key: "phone", label: "电话" }, { key: "wechat", label: "微信" },
  { key: "providesProcessing", label: "提供加工服务" }, { key: "processingServices", label: "加工能力" }, { key: "processingMaterials", label: "加工材料" }, { key: "processingEquipment", label: "加工设备" }, { key: "processingCapacity", label: "加工产能" }, { key: "processingRegions", label: "加工区域" }, { key: "processingNotes", label: "接单说明" },
];

export function VendorReviewsPage() {
  const [rows, setRows] = useState<VendorSubmission[]>([]);
  const [filter, setFilter] = useState("pending");
  const [notes, setNotes] = useState<Record<number, string>>({});
  const [message, setMessage] = useState("");
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [processingId, setProcessingId] = useState<number | null>(null);
  const load = () => listVendorSubmissionPage({ page, pageSize: 20, status: filter || undefined }).then((result) => { setRows(result.items); setTotal(result.total); }).catch(() => setMessage("审核列表加载失败"));
  useEffect(() => { void load(); }, [filter, page]);

  const review = (row: VendorSubmission, status: "approved" | "rejected") => {
    if (status === "rejected" && !notes[row.id]?.trim()) {
      setMessage("驳回时请填写审核意见");
      return;
    }
    if (status === "approved" && !window.confirm("确认发布本次厂商资料变更？已隐藏厂商仍会保持隐藏。")) return;
    setProcessingId(row.id);
    reviewVendorSubmission(row.id, status, notes[row.id] || "").then(() => {
      setMessage(status === "approved" ? "审核通过，前台资料已更新" : "已驳回，厂商可修改后重新提交");
      void load();
    }).catch((error: { response?: { status?: number } }) => setMessage(error.response?.status === 409 ? "正式资料版本已变化，请重新比较并要求厂商重新提交" : "审核操作失败，该记录可能已被处理")).finally(() => setProcessingId(null));
  };

  return (
    <AdminLayout title="厂商资料审核">
      {message && <p className="admin-message">{message}</p>}
      <div className="admin-toolbar admin-panel compact review-toolbar">
        <label>审核状态<select value={filter} onChange={(e) => { setFilter(e.target.value); setPage(1); }}><option value="pending">待审核</option><option value="approved">已通过</option><option value="rejected">已驳回</option><option value="">全部</option></select></label>
        <span>共 {total} 条提交</span>
      </div>
      <div className="vendor-review-list">
        {rows.map((row) => <ReviewCard disabled={processingId === row.id} key={row.id} row={row} note={notes[row.id] || ""} setNote={(value) => setNotes({ ...notes, [row.id]: value })} review={review} />)}
        {!rows.length && <div className="admin-panel">当前没有相关审核记录</div>}
      </div>
      <Pagination onChange={setPage} page={page} pageSize={20} total={total} />
    </AdminLayout>
  );
}

function ReviewCard({ row, note, setNote, review, disabled }: { row: VendorSubmission; note: string; setNote: (value: string) => void; review: (row: VendorSubmission, status: "approved" | "rejected") => void; disabled: boolean }) {
  const changes = useMemo(() => comparedFields.filter(({ key }) => String(row.vendor?.[key] ?? "") !== String(row.draft?.[key] ?? "")), [row]);
  const mediaChanged = JSON.stringify(normalizeMedia(row.vendor?.media || [])) !== JSON.stringify(normalizeMedia(row.draft?.media || []));
  return (
    <article className="admin-panel vendor-review-card">
      <header><div><span className={`review-status ${row.status}`}>{statusLabel(row.status)}</span><h2>{row.draft.name || row.vendor.name}</h2><p>提交账号：{row.submittedBy} · {new Date(row.updatedAt).toLocaleString()}</p></div><div className="review-summary"><strong>{changes.length + (mediaChanged ? 1 : 0)} 项变更</strong><a className="outline-btn small" href={`/vendors/${row.vendorId}`} rel="noreferrer" target="_blank"><ExternalLink size={14} />查看当前页面</a></div></header>
      <div className="review-diff-table">
        <div className="diff-head">字段</div><div className="diff-head">当前前台内容</div><div className="diff-head">厂商提交内容</div>
        {(changes.length ? changes : comparedFields.slice(0, 3)).map(({ key, label }) => <div className="diff-row" key={key}><strong>{label}</strong><span>{formatValue(row.vendor?.[key])}</span><span>{formatValue(row.draft?.[key])}</span></div>)}
      </div>
      {mediaChanged && <div className="review-media-diff"><section><strong>当前企业图集</strong><div>{row.vendor?.media?.map((media) => <ProtectedMediaImage alt={media.caption || "当前图片"} assetId={media.assetId} key={media.id || media.url} src={media.url} />)}</div></section><section><strong>提交后的企业图集</strong><div>{row.draft?.media?.map((media) => <ProtectedMediaImage alt={media.caption || "提交图片"} assetId={media.assetId} key={media.assetId || media.id || media.url} src={media.url} />)}</div></section></div>}
      {row.status === "pending" ? <div className="review-actions"><textarea placeholder="审核意见（驳回时必填）" value={note} onChange={(e) => setNote(e.target.value)} /><button className="primary-btn" disabled={disabled} onClick={() => review(row, "approved")}><Check size={16} />{disabled ? "处理中…" : "审核通过并发布"}</button><button className="outline-btn danger" disabled={disabled} onClick={() => review(row, "rejected")}><X size={16} />驳回</button></div> : row.reviewNote && <p className="review-note">审核意见：{row.reviewNote}</p>}
    </article>
  );
}

function formatValue(value: unknown) { if (typeof value === "boolean") return value ? "是" : "否"; return String(value || "—"); }
function normalizeMedia(rows: NonNullable<Vendor["media"]>) { return rows.map((row) => ({ assetId: row.assetId || null, kind: row.kind, caption: row.caption || "", sortOrder: row.sortOrder || 0 })).sort((a, b) => a.sortOrder - b.sortOrder); }
function statusLabel(status: string) { return ({ pending: "待审核", approved: "已通过", rejected: "已驳回" } as Record<string, string>)[status] || status; }
