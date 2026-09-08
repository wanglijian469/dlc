import { useEffect, useMemo, useState, useRef, type FormEvent } from "react";
import { BadgeDollarSign, Copy, Pencil, Plus, Trash2, X } from "lucide-react";
import { checkOwnProductDuplicate, deleteOwnProduct, listOwnProducts, updateOwnProductPrice, withdrawOwnProductSubmission } from "../../api/admin";
import { getApiErrorMessage } from "../../api/client";
import { getFilterOptions } from "../../api/public";
import type { Category, VendorProductDuplicateResult, VendorProductRecord } from "../../types/api";
import { productFieldGuidance } from "../../config/formGuidance";
import { SpecsEditor } from "./StructuredEditors";
import { ProtectedMediaImage } from "./ProtectedMediaImage";

import { commitWorkDraft, deleteWorkDraft, listWorkDrafts, updateShowroomOrder, type WorkDraft } from "../../api/workspace";
import { usePrivateDraft } from "../../hooks/usePrivateDraft";
import { ProductPhotosEditor } from "./ProductPhotosEditor";

const priceUnitOptions = ["件", "套", "台", "个", "支", "组", "箱", "公斤", "吨", "米"];
const minOrderOptions = [1, 5, 10, 20, 50, 100, 500, 1000];
const availableQuantityOptions = [0, 10, 50, 100, 500, 1000, 5000, 10000];
const leadTimeOptions = ["现货", "3天内", "7天内", "15天内", "30天内", "45天内", "60天内", "双方协商"];
const freightNoteOptions = ["按实际运费结算", "包邮", "卖家承担", "买家承担", "物流到付", "双方协商"];
const supplyAbilityOptions = ["按订单生产", "现货供应", "支持小批量供货", "支持批量供货", "支持来图来样定制", "产能面议"];

const createDefaultDraft = (): Partial<VendorProductRecord> => ({
  vendorProductName: "", vendorModel: "", compatibleModels: "", description: "", detailContent: "", image: "", galleryRaw: "", specsRaw: "", priceNote: "", inquiryText: "欢迎询价",
  priceUnit: "件", minOrderQuantity: 1, availableQuantity: 100, leadTime: "7天内", freightNote: "按实际运费结算", supplyAbility: "按订单生产",
});

export function VendorProductsEditor() {
  const actionLock = useRef(false);
  const [records, setRecords] = useState<VendorProductRecord[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [form, setForm] = useState<Partial<VendorProductRecord>>(createDefaultDraft);
  const [editing, setEditing] = useState<VendorProductRecord>();
  const [priceEditing, setPriceEditing] = useState<VendorProductRecord>();
  const [priceForm, setPriceForm] = useState<Partial<VendorProductRecord>>({});
  const [duplicate, setDuplicate] = useState<VendorProductDuplicateResult>({ exact: false, similar: [] });
  const [open, setOpen] = useState(false);
  const [message, setMessage] = useState("");
  const [step, setStep] = useState(1), [submitting, setSubmitting] = useState(false), [uploading, setUploading] = useState(false);
  const [keyword, setKeyword] = useState(""), [filter, setFilter] = useState(() => new URLSearchParams(window.location.search).get("status") || "");
  const [drafts, setDrafts] = useState<WorkDraft<Partial<VendorProductRecord>>[]>([]);
  const privateDraft = usePrivateDraft("product", form, open);
  const loadDrafts = () => listWorkDrafts<Partial<VendorProductRecord>>().then(rows => setDrafts(rows.filter(row => row.kind === "product"))).catch(() => setMessage("私有草稿加载失败，请重试"));
  const closeEditor = () => {
    if (submitting || uploading) { setMessage("请等待上传或提交完成"); return; }
    if (privateDraft.dirty && !window.confirm("还有未保存修改，确定关闭？")) return;
    setOpen(false); void loadDrafts();
  };
  const visibleRecords = records.filter(row => (row.vendorProductName + (row.vendorModel || "")).toLowerCase().includes(keyword.toLowerCase()) && (!filter || (filter === "pending" ? ["pending", "pending_update"].includes(row.status) : filter === "rejected" ? ["rejected", "rejected_update"].includes(row.status) : filter === "approved" ? ["approved","pending_update","rejected_update"].includes(row.status) : filter === "expired" ? !!row.priceValidUntil && new Date(row.priceValidUntil) < new Date() : row.status === filter)));

  const rootCategories = useMemo(() => categories.filter((item) => !item.parentId && item.isEnabled !== false), [categories]);
  const load = () => Promise.all([listOwnProducts(), getFilterOptions()]).then(([rows, filters]) => { setRecords(rows); setCategories(filters.categories); }).catch(() => setMessage("产品资料加载失败"));
  useEffect(() => { void load(); void loadDrafts(); }, []);
  useEffect(() => { if (new URLSearchParams(window.location.search).get("new") === "1") startNew(); }, [window.location.search]);
  useEffect(() => {
    if (!open) return;
    const previousOverflow = document.body.style.overflow;
    const closeOnEscape = (event: KeyboardEvent) => { if (event.key === "Escape") closeEditor(); };
    document.body.style.overflow = "hidden";
    window.addEventListener("keydown", closeOnEscape);
    return () => { document.body.style.overflow = previousOverflow; window.removeEventListener("keydown", closeOnEscape); };
  }, [open, privateDraft.dirty, uploading, submitting]);
  useEffect(() => {
    if (!open || !String(form.vendorProductName || "").trim()) { setDuplicate({ exact: false, similar: [] }); return; }
    let cancelled = false;
    const timer = window.setTimeout(() => void checkOwnProductDuplicate({ name: String(form.vendorProductName), model: String(form.vendorModel || ""), excludeType: editing?.recordType, excludeId: editing?.id }).then(result => { if (!cancelled) setDuplicate(result); }).catch(() => undefined), 300);
    return () => { cancelled = true; window.clearTimeout(timer); };
  }, [open, form.vendorProductName, form.vendorModel, editing]);

  const startNew = () => { privateDraft.activate(); setStep(1); setMessage(""); setEditing(undefined); setForm(createDefaultDraft()); setDuplicate({ exact: false, similar: [] }); setOpen(true); };
  const startEdit = (record: VendorProductRecord) => { const saved = drafts.find(d => d.targetType === record.recordType && d.targetId === record.id); privateDraft.activate(saved, record.recordType, record.id); setStep(1); setMessage(""); setEditing(record); setForm(saved?.payload || { ...record }); setDuplicate({ exact: false, similar: [] }); setOpen(true); };
  const openedReview = useRef("");
  useEffect(() => {
    const review = new URLSearchParams(window.location.search).get("review");
    if (!review || openedReview.current === review) return;
    const record = records.find(r => r.submissionId === Number(review));
    if (record) { openedReview.current = review; startEdit(record); }
  }, [records, window.location.search]);
  const startPrice = (record: VendorProductRecord) => { setPriceEditing(record); setPriceForm({
    ...record,
    priceUnit: record.priceUnit ?? "",
    minOrderQuantity: record.minOrderQuantity ?? 1,
    availableQuantity: record.availableQuantity ?? 100,
    leadTime: record.leadTime ?? "",
    freightNote: record.freightNote ?? "",
    supplyAbility: record.supplyAbility ?? "",
  }); };
  const savePrice = (event: FormEvent) => {
    event.preventDefault();
    if (!priceEditing?.supplierId) return;
    updateOwnProductPrice(priceEditing.supplierId, {
      unitPriceCents: Number(priceForm.unitPriceCents || 0),
      priceUnit: priceForm.priceUnit || "",
      minOrderQuantity: Number(priceForm.minOrderQuantity || 0),
      taxIncluded: Boolean(priceForm.taxIncluded),
      freightNote: priceForm.freightNote || "",
      availableQuantity: Number(priceForm.availableQuantity || 0),
      leadTime: priceForm.leadTime || "",
      priceValidUntil: priceForm.priceValidUntil || undefined,
      negotiable: Boolean(priceForm.negotiable),
      supplyAbility: priceForm.supplyAbility ?? "",
      expectedVersion: priceEditing.priceVersion || 1,
    }).then(() => { setPriceEditing(undefined); setMessage("价格、库存和交期已即时更新"); void load(); }).catch((error) => setMessage(getApiErrorMessage(error, "价格更新失败，请刷新后重试")));
  };
  const submit = async (event: FormEvent) => {
    event.preventDefault(); if (actionLock.current || submitting || uploading) return;
    if (step < 3) { nextStep(); return; }
    setMessage("");
    if (duplicate.exact) { setMessage("本厂已存在名称和型号相同的产品，请编辑原记录"); return; }
    actionLock.current = true; setSubmitting(true);
    try { const saved = await privateDraft.save(); await commitWorkDraft(saved.id, saved.version); setOpen(false); setMessage("产品资料已提交平台审核"); void load(); void loadDrafts(); }
    catch (error) { setMessage(getApiErrorMessage(error, "提交失败，请检查草稿和网络")); }
    finally { actionLock.current = false; setSubmitting(false); }
  };
  const nextStep = () => {
    if (step === 1 && (!form.vendorProductName?.trim() || !form.categoryId)) { setMessage("请填写本厂产品名称并选择产品大类"); return; }
    setMessage(""); setStep(value => Math.min(3, value + 1));
  };
  const resume = (row: WorkDraft<Partial<VendorProductRecord>>) => {
    privateDraft.activate(row); setForm(row.payload); setEditing(row.targetType ? records.find(r => r.recordType === row.targetType && r.id === row.targetId) : undefined); setStep(1); setMessage(""); setOpen(true);
  };
  const copyRecord = (record: VendorProductRecord) => {
    const { id: _id, supplierId: _supplier, submissionId: _submission, recordType: _type, status: _status, reviewNote: _note, ...payload } = record;
    privateDraft.activate(); setEditing(undefined); setForm({ ...payload, vendorProductName: "", vendorModel: "" }); setStep(1); setOpen(true); setMessage("已复制资料，请填写并确认新产品名称与型号");
  };
  const remove = (record: VendorProductRecord) => {
    const pending = record.recordType === "submission";
    if (!window.confirm(pending ? "确定撤回这条产品资料吗？" : "确定停止供应该产品吗？")) return;
    const action = pending && record.submissionId ? withdrawOwnProductSubmission(record.submissionId) : deleteOwnProduct(record.supplierId || record.id);
    action.then(() => void load()).catch((error) => setMessage(getApiErrorMessage(error, pending ? "撤回失败" : "停止供应失败")));
  };

  return <section className="vendor-products-editor admin-panel">
    <header><div><h2>我的产品资料</h2><p>填写本厂真实产品名称、型号、参数、图片和供货信息；平台审核后统一整理分类并发布。</p></div><button className="primary-btn small" type="button" onClick={startNew}><Plus size={16} />添加本厂产品</button></header>
    {message && <p className="admin-message">{message}</p>}
    <div className="workspace-filter"><input aria-label="搜索本厂产品" placeholder="搜索产品名称、型号" value={keyword} onChange={e => setKeyword(e.target.value)}/><select aria-label="产品状态筛选" value={filter} onChange={e => setFilter(e.target.value)}><option value="">全部产品</option><option value="draft">草稿</option><option value="pending">待审核</option><option value="approved">已发布</option><option value="rejected">需修改</option><option value="expired">价格待更新</option></select><button className="outline-btn" onClick={() => { setKeyword(""); setFilter(""); }}>清除筛选</button></div>
    {(!filter || filter === "draft") && drafts.length > 0 && <section className="private-draft-list"><h3>未完成草稿 · 仅本人可见</h3>{drafts.filter(row => (row.payload.vendorProductName || "").includes(keyword)).map(row => <article key={row.id}><div><strong>{row.payload.vendorProductName || "未命名产品"}</strong><small>保存于 {new Date(row.updatedAt).toLocaleString()}</small></div><button className="primary-btn small" onClick={() => resume(row)}>继续编辑</button><button className="outline-btn small" onClick={() => { if (window.confirm("删除这份私有草稿？公开产品不受影响。")) void deleteWorkDraft(row).then(loadDrafts).catch(e => setMessage(getApiErrorMessage(e, "删除失败"))); }}>删除草稿</button></article>)}</section>}
    <div className="vendor-product-list">{visibleRecords.map((record) => <article key={`${record.recordType}-${record.id}`}>
      <ProtectedMediaImage assetId={assetId(record.image)} alt={record.vendorProductName} src={record.image || ""} />
      <div><strong>{record.vendorProductName}</strong>{record.vendorModel && <em>{record.vendorModel}</em>}<span>{record.categoryName || rootCategories.find((item) => item.id === record.categoryId)?.name || "待平台分类"} · {statusLabel(record.status)}</span><p>{record.description || record.supplyAbility || "暂无产品说明"}</p>{["approved","pending_update","rejected_update"].includes(record.status) && <small>{record.negotiable || !record.unitPriceCents ? "价格面议" : "当前价 ¥" + (record.unitPriceCents / 100).toFixed(2) + " / " + (record.priceUnit || "件")} · 库存/供应量 {record.availableQuantity || 0}</small>}{record.reviewNote && <small>审核意见：{record.reviewNote}</small>}</div>
      <div><button aria-label="复制为新产品" type="button" onClick={() => copyRecord(record)}><Copy size={16}/></button>{["approved","pending_update","rejected_update"].includes(record.status) && <><button className="outline-btn small" onClick={() => void updateShowroomOrder(record.supplierId || record.id, !record.showroomFeatured, record.showroomOrder || 0).then(load).catch(e => setMessage(getApiErrorMessage(e,"排序失败")))}>{record.showroomFeatured ? "取消置顶" : "展厅置顶"}</button><label>展厅排序<select aria-label={"展厅排序：" + record.vendorProductName} value={record.showroomOrder || 0} onChange={e => void updateShowroomOrder(record.supplierId || record.id, !!record.showroomFeatured, Number(e.target.value)).then(load).catch(() => setMessage("排序保存失败"))}>{Array.from(new Set([record.showroomOrder || 0, 0, 1, 2, 3, 5, 10, 20, 50])).sort((a,b)=>a-b).map(n=><option key={n} value={n}>{n === 0 ? "默认" : n}</option>)}</select></label></>}{["approved","pending_update","rejected_update"].includes(record.status) && <button aria-label="即时更新价格库存" title="即时更新价格库存" type="button" onClick={() => startPrice(record)}><BadgeDollarSign size={16} /></button>}<button aria-label="编辑本厂产品" type="button" onClick={() => startEdit(record)}><Pencil size={16} /></button><button aria-label={record.recordType === "submission" ? "撤回产品资料" : "停止供应"} type="button" onClick={() => remove(record)}><Trash2 size={16} /></button></div>
    </article>)}</div>
    {!records.length && <p className="structured-empty">尚未录入本厂产品，可点击“添加本厂产品”开始填写。</p>}
    {priceEditing && <div className="vendor-product-drawer-layer"><button aria-label="关闭价格编辑" className="vendor-product-drawer-backdrop" type="button" onClick={() => setPriceEditing(undefined)} /><aside aria-modal="true" className="vendor-product-drawer compact-price-drawer" role="dialog"><form className="vendor-product-form" onSubmit={savePrice}><header><div><span>即时生效</span><strong>更新价格、库存与交期</strong><p>{priceEditing.vendorProductName}。本次修改不改变产品名称、分类、参数和图片。</p></div><button aria-label="关闭" type="button" onClick={() => setPriceEditing(undefined)}><X size={20} /></button></header><div className="vendor-product-drawer-body"><div className="vendor-product-fields"><label>单价（元）<input disabled={priceForm.negotiable} min="0" step="0.01" type="number" value={Number(priceForm.unitPriceCents || 0) / 100} onChange={(event) => setPriceForm({ ...priceForm, unitPriceCents: Math.round(Number(event.target.value) * 100) })} /></label><StringPresetSelect label="计价单位" options={priceUnitOptions} value={priceForm.priceUnit} onChange={(priceUnit) => setPriceForm({ ...priceForm, priceUnit })} /><NumberPresetSelect label="起订量" options={minOrderOptions} value={priceForm.minOrderQuantity} onChange={(minOrderQuantity) => setPriceForm({ ...priceForm, minOrderQuantity })} /><NumberPresetSelect label="库存 / 可供应量" options={availableQuantityOptions} value={priceForm.availableQuantity} zeroLabel="按需供应" onChange={(availableQuantity) => setPriceForm({ ...priceForm, availableQuantity })} /><StringPresetSelect label="交期" options={leadTimeOptions} value={priceForm.leadTime} onChange={(leadTime) => setPriceForm({ ...priceForm, leadTime })} /><label>价格有效期<input type="date" value={priceForm.priceValidUntil?.slice(0, 10) || ""} onChange={(event) => setPriceForm({ ...priceForm, priceValidUntil: event.target.value ? new Date(event.target.value + "T23:59:59").toISOString() : undefined })} /></label><StringPresetSelect label="运费说明" options={freightNoteOptions} value={priceForm.freightNote} onChange={(freightNote) => setPriceForm({ ...priceForm, freightNote })} /><StringPresetSelect label="供货能力" options={supplyAbilityOptions} value={priceForm.supplyAbility} onChange={(supplyAbility) => setPriceForm({ ...priceForm, supplyAbility })} /><label className="checkbox-label"><input checked={Boolean(priceForm.taxIncluded)} type="checkbox" onChange={(event) => setPriceForm({ ...priceForm, taxIncluded: event.target.checked })} />价格含税</label><label className="checkbox-label"><input checked={Boolean(priceForm.negotiable)} type="checkbox" onChange={(event) => setPriceForm({ ...priceForm, negotiable: event.target.checked })} />价格面议</label></div></div><footer><button className="outline-btn" type="button" onClick={() => setPriceEditing(undefined)}>取消</button><button className="primary-btn" type="submit">立即更新</button></footer></form></aside></div>}
    {open && <div className="vendor-product-drawer-layer"><button aria-label="关闭产品编辑抽屉" className="vendor-product-drawer-backdrop" type="button" onClick={closeEditor} /><aside aria-label={editing ? "编辑本厂产品" : "添加本厂产品"} aria-modal="true" className="vendor-product-drawer" role="dialog"><form className="vendor-product-form" onSubmit={submit}>
      <header><div><span>本厂产品资料</span><strong>{editing ? "编辑本厂产品" : "添加本厂产品"}</strong><p>请按产品铭牌和真实供货情况填写，具体平台分类由运营审核确认。</p></div><button aria-label="关闭产品编辑抽屉" type="button" onClick={closeEditor}><X size={20} /></button></header>
      <nav className="product-stepper" aria-label="产品录入步骤">{["产品资料", "供货信息", "预览提交"].map((label,i) => <button key={label} type="button" aria-current={step === i + 1 ? "step" : undefined} disabled={i + 1 > step} onClick={() => setStep(i + 1)}><b>{i + 1}</b>{label}</button>)}</nav>
      <div className="draft-save-status" role="status">{privateDraft.status}</div>
      {message && <p className="admin-message" role="alert">{message}</p>}
      <div className="vendor-product-drawer-body">
      <section hidden={step !== 1}><ProductPhotosEditor value={form} onChange={patch => setForm(value => ({ ...value, ...patch }))} onBusy={setUploading}/><div className="vendor-product-fields">
        <label>本厂产品名称<input required value={form.vendorProductName || ""} onChange={(event) => setForm({ ...form, vendorProductName: event.target.value })} /></label>
        <label>本厂型号<input value={form.vendorModel || ""} onChange={(event) => setForm({ ...form, vendorModel: event.target.value })} /></label>
        <label>产品大类<select required value={form.categoryId || ""} onChange={(event) => setForm({ ...form, categoryId: Number(event.target.value) || undefined })}><option value="">请选择大类</option>{rootCategories.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}</select></label>
        <label>适配信息<input value={form.compatibleModels || ""} onChange={(event) => setForm({ ...form, compatibleModels: event.target.value })} /></label>

      </div>      {duplicate.exact && <p className="admin-message danger">本厂已存在名称和型号相同的产品，请编辑原记录。</p>}
      {!duplicate.exact && duplicate.similar.length > 0 && <div className="vendor-duplicate-warning"><strong>本厂已有相似产品，请确认不是重复录入：</strong>{duplicate.similar.map((item) => <span key={`${item.recordType}-${item.id}`}>{item.name}{item.model ? ` · ${item.model}` : ""}</span>)}</div>}
</section>
      <section hidden={step !== 2}><div className="vendor-product-fields">
        <label>价格说明<input value={form.priceNote || ""} onChange={(event) => setForm({ ...form, priceNote: event.target.value })} /></label>
        <label className="checkbox-label"><input checked={Boolean(form.negotiable)} type="checkbox" onChange={(event) => setForm({ ...form, negotiable: event.target.checked })} />价格面议</label>
        <label>单价（元）<input disabled={form.negotiable} min="0" step="0.01" type="number" value={Number(form.unitPriceCents || 0) / 100} onChange={(event) => setForm({ ...form, unitPriceCents: Math.round(Number(event.target.value) * 100) })} /></label>
        <StringPresetSelect label="计价单位" options={priceUnitOptions} value={form.priceUnit} onChange={(priceUnit) => setForm({ ...form, priceUnit })} />
        <NumberPresetSelect label="起订量" options={minOrderOptions} value={form.minOrderQuantity} onChange={(minOrderQuantity) => setForm({ ...form, minOrderQuantity })} />
        <NumberPresetSelect label="库存 / 可供应量" options={availableQuantityOptions} value={form.availableQuantity} zeroLabel="按需供应" onChange={(availableQuantity) => setForm({ ...form, availableQuantity })} />
        <StringPresetSelect label="交期" options={leadTimeOptions} value={form.leadTime} onChange={(leadTime) => setForm({ ...form, leadTime })} />
        <StringPresetSelect label="运费说明" options={freightNoteOptions} value={form.freightNote} onChange={(freightNote) => setForm({ ...form, freightNote })} />
        <label className="checkbox-label"><input checked={Boolean(form.taxIncluded)} type="checkbox" onChange={(event) => setForm({ ...form, taxIncluded: event.target.checked })} />价格含税</label>
        <StringPresetSelect label="供货能力" options={supplyAbilityOptions} value={form.supplyAbility} onChange={(supplyAbility) => setForm({ ...form, supplyAbility })} />

      </div><details><summary>补充说明与规格参数（选填）</summary><div className="vendor-product-fields">        <label className="wide-field">产品说明<textarea className="writing-example" placeholder={productFieldGuidance.description} value={form.description || ""} onChange={(event) => setForm({ ...form, description: event.target.value })} /></label>
        <label className="wide-field">详细说明<textarea className="writing-example" placeholder={productFieldGuidance.detailContent} value={form.detailContent || ""} onChange={(event) => setForm({ ...form, detailContent: event.target.value })} /></label>
        <label className="wide-field">询价说明<input value={form.inquiryText || ""} onChange={(event) => setForm({ ...form, inquiryText: event.target.value })} /></label>
</div><SpecsEditor value={form.specsRaw || ""} onChange={specsRaw => setForm({ ...form, specsRaw })}/></details></section>
      {step === 3 && <section className="product-submit-preview"><span>客户看到的本厂产品资料 · 审核后公开</span>{form.image && <ProtectedMediaImage assetId={assetId(form.image)} src={form.image} alt="提交预览主图"/>}<h2>{form.vendorProductName}</h2><p>型号：{form.vendorModel || "待确认"}</p><p>{form.description}</p><dl>{[["计价单位",form.priceUnit],["起订量",form.minOrderQuantity],["库存 / 可供应量",form.availableQuantity === 0 ? "按需供应" : form.availableQuantity],["交期",form.leadTime],["运费说明",form.freightNote],["供货能力",form.supplyAbility]].map(([label,value])=><div key={label}><dt>{label}</dt><dd>{value}</dd></div>)}</dl><p className="admin-message">请确认图片属于本厂，并核对库存、起订量和交期是否符合实际供货情况。</p><label className="checkbox-label"><input required type="checkbox"/>我已确认产品资料及供货信息真实准确</label></section>}
      </div>
      <footer><button className="outline-btn" type="button" onClick={closeEditor} disabled={submitting || uploading}>取消</button><button className="outline-btn" type="button" disabled={submitting || uploading} onClick={() => void privateDraft.save().then(() => loadDrafts()).catch(() => {})}>保存草稿</button>{step > 1 && <button className="outline-btn" type="button" disabled={submitting} onClick={() => setStep(step - 1)}>上一步</button>}<button className="primary-btn" disabled={duplicate.exact || submitting || uploading} type="submit">{submitting ? "提交中…" : step < 3 ? "下一步" : "确认并提交审核"}</button></footer>
    </form></aside></div>}
  </section>;
}

function statusLabel(status: VendorProductRecord["status"]) { return ({ pending: "待审核", pending_update: "有待审核修改", approved: "已发布", rejected_update: "需修改（原版仍公开）", rejected: "需修改" })[status]; }
function assetId(url?: string) { const match = url?.match(/\/api\/media\/(\d+)/); return match ? Number(match[1]) : undefined; }

function StringPresetSelect({ label, options, value, onChange }: { label: string; options: string[]; value?: string; onChange: (value: string) => void }) {
  const current = value || "";
  const legacy = !options.includes(current);
  return <label>{label}<select aria-label={label} value={current} onChange={(event) => onChange(event.target.value)}>{legacy && <option value={current}>当前值：{current || "未填写"}</option>}{options.map((option) => <option key={option} value={option}>{option}</option>)}</select></label>;
}

function NumberPresetSelect({ label, options, value, zeroLabel, onChange }: { label: string; options: number[]; value?: number; zeroLabel?: string; onChange: (value: number) => void }) {
  const current = value ?? options[0];
  const legacy = !options.includes(current);
  return <label>{label}<select aria-label={label} value={current} onChange={(event) => onChange(Number(event.target.value))}>{legacy && <option value={current}>当前值：{current}</option>}{options.map((option) => <option key={option} value={option}>{option === 0 && zeroLabel ? zeroLabel : option}</option>)}</select></label>;
}
