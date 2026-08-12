import { useEffect, useState } from "react";
import { CheckCircle2, GitMerge, PlusCircle, Search, XCircle } from "lucide-react";
import { listProductSubmissionMatches, listProductSubmissions, reviewProductSubmission } from "../../api/admin";
import { getApiErrorMessage } from "../../api/client";
import { getFilterOptions } from "../../api/public";
import { AdminLayout } from "../../components/admin/AdminLayout";
import { Pagination } from "../../components/public/Pagination";
import { ProtectedMediaImage } from "../../components/admin/ProtectedMediaImage";
import type { Category, Product, ProductMatchSuggestion, ProductSubmission } from "../../types/api";
import { hierarchicalCategoryOptions } from "../../utils/categories";

type Resolution = "create_product" | "link_product";

export function ProductReviewsPage() {
  const [rows, setRows] = useState<ProductSubmission[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [status, setStatus] = useState("pending");
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [notes, setNotes] = useState<Record<number, string>>({});
  const [resolutions, setResolutions] = useState<Record<number, Resolution>>({});
  const [targets, setTargets] = useState<Record<number, number>>({});
  const [matches, setMatches] = useState<Record<number, ProductMatchSuggestion[]>>({});
  const [keywords, setKeywords] = useState<Record<number, string>>({});
  const [catalogDrafts, setCatalogDrafts] = useState<Record<number, Partial<Product>>>({});
  const [message, setMessage] = useState("");
  const pageSize = 10;
  const load = () => Promise.all([listProductSubmissions({ page, pageSize, status }), getFilterOptions()]).then(([result, filters]) => {
    setRows(result.items); setTotal(result.total); setCategories(filters.categories);
    setResolutions((current) => ({ ...Object.fromEntries(result.items.filter(isStandaloneNew).map((row) => [row.id, current[row.id] || "create_product"])), ...current }));
    setCatalogDrafts((current) => ({ ...Object.fromEntries(result.items.filter(isStandaloneNew).map((row) => [row.id, current[row.id] || { ...row.productDraft, name: row.supplierDraft?.vendorProductName || row.productDraft?.name }])) , ...current }));
    result.items.filter((row) => row.status === "pending" && isStandaloneNew(row)).forEach((row) => void loadMatches(row.id));
  }).catch(() => setMessage("产品审核队列加载失败"));
  useEffect(() => { void load(); }, [page, status]);

  const loadMatches = (id: number, keyword = "") => listProductSubmissionMatches(id, keyword).then((items) => setMatches((current) => ({ ...current, [id]: items }))).catch(() => setMessage("匹配建议加载失败"));
  const review = (row: ProductSubmission, next: "approved" | "rejected") => {
    const payload: Parameters<typeof reviewProductSubmission>[1] = { status: next, reviewNote: notes[row.id] || "" };
    if (next === "approved" && isStandaloneNew(row)) {
      payload.resolution = resolutions[row.id] || "create_product";
      if (payload.resolution === "link_product") payload.targetProductId = targets[row.id];
      else payload.catalogProduct = catalogDrafts[row.id];
      if (payload.resolution === "link_product" && !payload.targetProductId) { setMessage("请选择需要关联的标准产品"); return; }
      if (payload.resolution === "create_product" && !String(payload.catalogProduct?.name || "").trim()) { setMessage("请填写标准产品名称"); return; }
    }
    reviewProductSubmission(row.id, payload).then(() => { setMessage(next === "approved" ? "审核已通过" : "已驳回提交"); void load(); }).catch((error) => setMessage(getApiErrorMessage(error, "审核失败，可能存在版本冲突、重复关联或目标产品无效")));
  };
  const patchCatalog = (id: number, patch: Partial<Product>) => setCatalogDrafts((current) => ({ ...current, [id]: { ...current[id], ...patch } }));

  return <AdminLayout title="产品审核"><section className="admin-panel product-review-panel">
    <header><div><h2>厂商产品资料审核</h2><p>平台负责统一产品目录。请人工确认关联已有标准产品，或校正资料后创建新的标准产品。</p></div><select aria-label="审核状态" value={status} onChange={(event) => { setStatus(event.target.value); setPage(1); }}><option value="pending">待审核</option><option value="approved">已通过</option><option value="rejected">已驳回</option><option value="">全部</option></select></header>
    {message && <p className="admin-message">{message}</p>}
    <div className="review-list">{rows.map((row) => <article className="review-card" key={row.id}>
      <header><div><strong>{row.supplierDraft?.vendorProductName || row.productDraft?.name || row.product?.name || "本厂产品"}</strong><span>{typeLabel(row.submissionType)} · {row.vendor?.name || `厂商 #${row.vendorId}`}</span></div><span className={`status-badge ${row.status}`}>{statusLabel(row.status)}</span></header>
      <div className="review-diff-grid"><dl><div><dt>本厂产品名称</dt><dd>{row.supplierDraft?.vendorProductName || "—"}</dd></div><div><dt>厂商型号</dt><dd>{row.supplierDraft?.vendorModel || "—"}</dd></div><div><dt>适配信息</dt><dd>{row.supplierDraft?.compatibleModels || row.productDraft?.compatibleModels || "—"}</dd></div><div><dt>价格说明</dt><dd>{row.supplierDraft?.priceNote || "—"}</dd></div><div><dt>供货能力</dt><dd>{row.supplierDraft?.supplyAbility || "—"}</dd></div></dl>{row.supplierDraft?.image && <ProtectedMediaImage assetId={assetId(row.supplierDraft.image)} alt="厂商提交的产品图片" src={row.supplierDraft.image} />}</div>
      {row.status === "pending" && isStandaloneNew(row) && <div className="product-resolution-panel">
        <div className="resolution-tabs"><button className={resolutions[row.id] === "create_product" ? "selected" : ""} type="button" onClick={() => setResolutions({ ...resolutions, [row.id]: "create_product" })}><PlusCircle size={16} />创建标准产品</button><button className={resolutions[row.id] === "link_product" ? "selected" : ""} type="button" onClick={() => setResolutions({ ...resolutions, [row.id]: "link_product" })}><GitMerge size={16} />关联已有产品</button></div>
        {resolutions[row.id] === "link_product" ? <div className="product-match-panel"><div className="admin-search"><Search size={16} /><input aria-label="搜索标准产品" placeholder="按产品名称、型号或适配信息搜索" value={keywords[row.id] || ""} onChange={(event) => setKeywords({ ...keywords, [row.id]: event.target.value })} /><button className="outline-btn small" type="button" onClick={() => void loadMatches(row.id, keywords[row.id] || "")}>搜索</button></div><div className="catalog-results">{(matches[row.id] || []).map((item) => <button className={targets[row.id] === item.product.id ? "selected" : ""} disabled={item.vendorAlreadyLinked} key={item.product.id} type="button" onClick={() => setTargets({ ...targets, [row.id]: item.product.id! })}><strong>{item.product.name}</strong><span>{item.product.category?.name || "未分类"} · {item.product.compatibleModels || "无适配说明"}</span><small>{item.reasons.join("、") || "搜索结果"} · {item.product.supplierCount || 0} 家供应商{item.vendorAlreadyLinked ? " · 已关联该厂商" : ""}</small></button>)}</div>{!(matches[row.id] || []).length && <p className="structured-empty">没有匹配建议，可搜索其他标准产品；仍无结果时选择“创建标准产品”。</p>}</div>
          : <CatalogProductEditor categories={categories} value={catalogDrafts[row.id] || {}} onChange={(patch) => patchCatalog(row.id, patch)} />}
      </div>}
      {row.status === "pending" && <div className="review-actions"><label>审核意见<input value={notes[row.id] || ""} onChange={(event) => setNotes({ ...notes, [row.id]: event.target.value })} /></label><button className="primary-btn small" type="button" onClick={() => review(row, "approved")}><CheckCircle2 size={15} />通过</button><button className="outline-btn small danger" type="button" onClick={() => review(row, "rejected")}><XCircle size={15} />驳回</button></div>}
      {row.reviewNote && <p className="review-note">审核意见：{row.reviewNote}</p>}
    </article>)}</div>
    {!rows.length && <p className="structured-empty">当前没有符合条件的产品提交。</p>}
    <Pagination page={page} pageSize={pageSize} total={total} onChange={setPage} />
  </section></AdminLayout>;
}

function CatalogProductEditor({ value, categories, onChange }: { value: Partial<Product>; categories: Category[]; onChange: (patch: Partial<Product>) => void }) { return <div className="catalog-product-editor"><p>以下内容将作为平台统一产品资料发布，厂商原始名称和型号仍保留在供应信息中。</p><label>标准产品名称<input required value={value.name || ""} onChange={(event) => onChange({ name: event.target.value })} /></label><label>准确分类<select value={value.categoryId || ""} onChange={(event) => onChange({ categoryId: Number(event.target.value) || undefined })}><option value="">请选择分类</option>{hierarchicalCategoryOptions(categories).map((item) => <option key={item.value} value={item.value}>{item.label}</option>)}</select></label><label>通用适配信息<input value={value.compatibleModels || ""} onChange={(event) => onChange({ compatibleModels: event.target.value })} /></label><label className="wide-field">标准产品说明<textarea value={value.description || ""} onChange={(event) => onChange({ description: event.target.value })} /></label></div>; }
function typeLabel(type: ProductSubmission["submissionType"]) { return ({ new_product: "新增本厂产品", link_product: "历史关联记录", link_supplier: "历史关联记录", update_supplier: "修改供应信息", update_offer: "修改供应信息" } as Record<string, string>)[type] || type; }
function statusLabel(status: ProductSubmission["status"]) { return ({ pending: "待审核", approved: "已通过", rejected: "已驳回", superseded: "已被新版本替代" })[status]; }
function assetId(url?: string) { const match = url?.match(/\/api\/media\/(\d+)/); return match ? Number(match[1]) : undefined; }
function isStandaloneNew(row: ProductSubmission) { return row.submissionType === "new_product" && !row.supplierId; }
