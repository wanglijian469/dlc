import { useEffect, useMemo, useState, type FormEvent } from "react";
import { ImageUp, Pencil, Plus, Trash2, X } from "lucide-react";
import { checkOwnProductDuplicate, createOwnProduct, deleteOwnProduct, listOwnProducts, updateOwnProduct, updateOwnProductSubmission, uploadFile, withdrawOwnProductSubmission } from "../../api/admin";
import { getApiErrorMessage } from "../../api/client";
import { getFilterOptions } from "../../api/public";
import type { Category, VendorProductDuplicateResult, VendorProductRecord } from "../../types/api";
import { productFieldGuidance } from "../../config/formGuidance";
import { GalleryEditor, SpecsEditor } from "./StructuredEditors";
import { ProtectedMediaImage } from "./ProtectedMediaImage";

const emptyDraft: Partial<VendorProductRecord> = { vendorProductName: "", vendorModel: "", compatibleModels: "", description: "", detailContent: "", image: "", galleryRaw: "", specsRaw: "", priceNote: "", supplyAbility: "", inquiryText: "欢迎询价" };

export function VendorProductsEditor() {
  const [records, setRecords] = useState<VendorProductRecord[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [form, setForm] = useState<Partial<VendorProductRecord>>(emptyDraft);
  const [editing, setEditing] = useState<VendorProductRecord>();
  const [duplicate, setDuplicate] = useState<VendorProductDuplicateResult>({ exact: false, similar: [] });
  const [open, setOpen] = useState(false);
  const [message, setMessage] = useState("");

  const rootCategories = useMemo(() => categories.filter((item) => !item.parentId && item.isEnabled !== false), [categories]);
  const load = () => Promise.all([listOwnProducts(), getFilterOptions()]).then(([rows, filters]) => { setRecords(rows); setCategories(filters.categories); }).catch(() => setMessage("产品资料加载失败"));
  useEffect(() => { void load(); }, []);
  useEffect(() => {
    if (!open) return;
    const previousOverflow = document.body.style.overflow;
    const closeOnEscape = (event: KeyboardEvent) => { if (event.key === "Escape") setOpen(false); };
    document.body.style.overflow = "hidden";
    window.addEventListener("keydown", closeOnEscape);
    return () => { document.body.style.overflow = previousOverflow; window.removeEventListener("keydown", closeOnEscape); };
  }, [open]);
  useEffect(() => {
    if (!open || !String(form.vendorProductName || "").trim()) { setDuplicate({ exact: false, similar: [] }); return; }
    const timer = window.setTimeout(() => void checkOwnProductDuplicate({ name: String(form.vendorProductName), model: String(form.vendorModel || ""), excludeType: editing?.recordType, excludeId: editing?.id }).then(setDuplicate).catch(() => undefined), 300);
    return () => window.clearTimeout(timer);
  }, [open, form.vendorProductName, form.vendorModel, editing]);

  const startNew = () => { setEditing(undefined); setForm(emptyDraft); setDuplicate({ exact: false, similar: [] }); setOpen(true); };
  const startEdit = (record: VendorProductRecord) => { setEditing(record); setForm({ ...record }); setDuplicate({ exact: false, similar: [] }); setOpen(true); };
  const submit = (event: FormEvent) => {
    event.preventDefault();
    setMessage("");
    if (duplicate.exact) { setMessage("本厂已存在名称和型号相同的产品，请编辑原记录"); return; }
    const action = editing?.recordType === "submission" && editing.submissionId
      ? updateOwnProductSubmission(editing.submissionId, form)
      : editing?.recordType === "supplier" && editing.supplierId
        ? updateOwnProduct(editing.supplierId, form)
        : createOwnProduct(form);
    action.then(() => { setMessage("产品资料已提交平台审核"); setOpen(false); void load(); }).catch((error) => setMessage(getApiErrorMessage(error, "提交失败，请检查必填项或是否重复录入")));
  };
  const remove = (record: VendorProductRecord) => {
    const pending = record.recordType === "submission";
    if (!window.confirm(pending ? "确定撤回这条产品资料吗？" : "确定停止供应该产品吗？")) return;
    const action = pending && record.submissionId ? withdrawOwnProductSubmission(record.submissionId) : deleteOwnProduct(record.supplierId || record.id);
    action.then(() => void load()).catch((error) => setMessage(getApiErrorMessage(error, pending ? "撤回失败" : "停止供应失败")));
  };
  const uploadCover = (file: File) => uploadFile(file).then((result) => setForm((value) => ({ ...value, image: result.url })));

  return <section className="vendor-products-editor admin-panel">
    <header><div><h2>我的产品资料</h2><p>填写本厂真实产品名称、型号、参数、图片和供货信息；平台审核后统一整理分类并发布。</p></div><button className="primary-btn small" type="button" onClick={startNew}><Plus size={16} />添加本厂产品</button></header>
    {message && <p className="admin-message">{message}</p>}
    <div className="vendor-product-list">{records.map((record) => <article key={`${record.recordType}-${record.id}`}>
      <ProtectedMediaImage assetId={assetId(record.image)} alt={record.vendorProductName} src={record.image || ""} />
      <div><strong>{record.vendorProductName}</strong>{record.vendorModel && <em>{record.vendorModel}</em>}<span>{record.categoryName || rootCategories.find((item) => item.id === record.categoryId)?.name || "待平台分类"} · {statusLabel(record.status)}</span><p>{record.description || record.supplyAbility || "暂无产品说明"}</p>{record.reviewNote && <small>审核意见：{record.reviewNote}</small>}</div>
      <div><button aria-label="编辑本厂产品" type="button" onClick={() => startEdit(record)}><Pencil size={16} /></button><button aria-label={record.recordType === "submission" ? "撤回产品资料" : "停止供应"} type="button" onClick={() => remove(record)}><Trash2 size={16} /></button></div>
    </article>)}</div>
    {!records.length && <p className="structured-empty">尚未录入本厂产品，可点击“添加本厂产品”开始填写。</p>}
    {open && <div className="vendor-product-drawer-layer"><button aria-label="关闭产品编辑抽屉" className="vendor-product-drawer-backdrop" type="button" onClick={() => setOpen(false)} /><aside aria-label={editing ? "编辑本厂产品" : "添加本厂产品"} aria-modal="true" className="vendor-product-drawer" role="dialog"><form className="vendor-product-form" onSubmit={submit}>
      <header><div><span>本厂产品资料</span><strong>{editing ? "编辑本厂产品" : "添加本厂产品"}</strong><p>请按产品铭牌和真实供货情况填写，具体平台分类由运营审核确认。</p></div><button aria-label="关闭产品编辑抽屉" type="button" onClick={() => setOpen(false)}><X size={20} /></button></header>
      <div className="vendor-product-drawer-body"><div className="vendor-product-fields">
        <label>本厂产品名称<input required value={form.vendorProductName || ""} onChange={(event) => setForm({ ...form, vendorProductName: event.target.value })} /></label>
        <label>本厂型号<input value={form.vendorModel || ""} onChange={(event) => setForm({ ...form, vendorModel: event.target.value })} /></label>
        <label>产品大类<select required value={form.categoryId || ""} onChange={(event) => setForm({ ...form, categoryId: Number(event.target.value) || undefined })}><option value="">请选择大类</option>{rootCategories.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}</select></label>
        <label>适配信息<input value={form.compatibleModels || ""} onChange={(event) => setForm({ ...form, compatibleModels: event.target.value })} /></label>
        <label>价格说明<input value={form.priceNote || ""} onChange={(event) => setForm({ ...form, priceNote: event.target.value })} /></label>
        <label>供货能力<input value={form.supplyAbility || ""} onChange={(event) => setForm({ ...form, supplyAbility: event.target.value })} /></label>
        <label className="wide-field">产品说明<textarea className="writing-example" placeholder={productFieldGuidance.description} value={form.description || ""} onChange={(event) => setForm({ ...form, description: event.target.value })} /></label>
        <label className="wide-field">详细说明<textarea className="writing-example" placeholder={productFieldGuidance.detailContent} value={form.detailContent || ""} onChange={(event) => setForm({ ...form, detailContent: event.target.value })} /></label>
        <label className="wide-field">询价说明<input value={form.inquiryText || ""} onChange={(event) => setForm({ ...form, inquiryText: event.target.value })} /></label>
      </div>
      {duplicate.exact && <p className="admin-message danger">本厂已存在名称和型号相同的产品，请编辑原记录。</p>}
      {!duplicate.exact && duplicate.similar.length > 0 && <div className="vendor-duplicate-warning"><strong>本厂已有相似产品，请确认不是重复录入：</strong>{duplicate.similar.map((item) => <span key={`${item.recordType}-${item.id}`}>{item.name}{item.model ? ` · ${item.model}` : ""}</span>)}</div>}
      <SpecsEditor value={form.specsRaw || ""} onChange={(specsRaw) => setForm({ ...form, specsRaw })} />
      <GalleryEditor value={form.galleryRaw || ""} onChange={(galleryRaw) => setForm({ ...form, galleryRaw })} />
      <div className="wide-field product-cover-upload"><strong>产品主图</strong><label className="outline-btn small upload-button"><ImageUp size={15} />上传图片<input accept="image/jpeg,image/png,image/webp" type="file" onChange={(event) => { const file = event.target.files?.[0]; if (file) void uploadCover(file); }} /></label>{form.image && <ProtectedMediaImage assetId={assetId(form.image)} alt="产品图片预览" src={form.image} />}</div>
      </div>
      <footer><button className="outline-btn" type="button" onClick={() => setOpen(false)}>取消</button><button className="primary-btn" disabled={duplicate.exact} type="submit">提交审核</button></footer>
    </form></aside></div>}
  </section>;
}

function statusLabel(status: VendorProductRecord["status"]) { return ({ pending: "待审核", pending_update: "有待审核修改", approved: "已展示", rejected: "已驳回" })[status]; }
function assetId(url?: string) { const match = url?.match(/\/api\/media\/(\d+)/); return match ? Number(match[1]) : undefined; }
