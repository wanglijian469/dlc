import { useEffect, useMemo, useState, type FormEvent } from "react";
import { ArrowLeft, ArrowRight, Camera, Check, ClipboardCheck, FileImage, LoaderCircle, Plus, RefreshCw, Save, Send, Trash2 } from "lucide-react";
import { AdminLayout } from "../../components/admin/AdminLayout";
import { VendorOptionSearch } from "../../components/admin/VendorOptionSearch";
import {
  commitCapturePackage, createCapturePackage, createVendorInvitation, deleteCaptureDocument, deleteCapturePackage,
  getCapturePackage, listCapturePackages, recognizeCapturePackage, updateCaptureDocument, updateCaptureDraft, uploadCaptureDocuments,
  type CaptureDraft, type CaptureField, type CapturePackage, type CapturePackageDetail, type CaptureProductDraft,
} from "../../api/admin";
import { API_BASE_URL, getApiErrorMessage } from "../../api/client";
import type { VendorOption } from "../../types/api";

const statusLabels: Record<string, string> = {
  uploading: "待上传", queued: "排队中", processing: "识别中", needs_review: "待校对", ready: "可提交", committed: "已提交审核", failed: "识别失败",
};

const vendorFields = [
  ["name", "厂商全称"], ["shortName", "厂商简称"], ["province", "省份"], ["city", "城市"], ["county", "区县"], ["address", "详细地址"],
  ["websiteUrl", "官网"], ["contactName", "联系人"], ["phone", "联系电话"], ["wechat", "微信"], ["mainProducts", "主营产品"],
  ["serviceAdvantages", "服务优势"], ["description", "企业简介"], ["equipment", "主要设备"], ["certifications", "资质认证"],
] as const;

const productFields = [["name", "产品名称"], ["model", "厂商型号"], ["compatibleModels", "适配机型"], ["priceNote", "价格说明"], ["supplyAbility", "供货能力"], ["description", "产品简介"], ["detailContent", "详细说明"]] as const;

const blankField = (): CaptureField => ({ value: "", confidence: 1, confirmed: false });
const blankDraft = (): CaptureDraft => ({ vendorFields: { name: blankField() }, products: [], warnings: [] });
const blankProduct = (number: number): CaptureProductDraft => ({ key: `manual-${Date.now()}-${number}`, fields: { name: blankField(), model: blankField(), categoryName: blankField() }, selectedCropIds: [], specs: [] });

export function CaptureWorkbenchPage() {
  const role = localStorage.getItem("cms_role") || "admin";
  const [packages, setPackages] = useState<CapturePackage[]>([]);
  const [detail, setDetail] = useState<CapturePackageDetail | null>(null);
  const [title, setTitle] = useState("");
  const [vendor, setVendor] = useState<VendorOption | null>(null);
  const [selectedDocument, setSelectedDocument] = useState<number>();
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState(false);
  const [inviteURL, setInviteURL] = useState("");

  const loadList = () => listCapturePackages().then(setPackages).catch((error) => setMessage(getApiErrorMessage(error, "资料包列表加载失败")));
  const loadDetail = (id: number, quiet = false) => getCapturePackage(id).then((value) => { setDetail(value); setSelectedDocument((current) => current || value.package.documents?.[0]?.id); return value; }).catch((error) => { if (!quiet) setMessage(getApiErrorMessage(error, "资料包加载失败")); throw error; });
  useEffect(() => { void loadList(); }, []);
  useEffect(() => {
    const current = detail?.package;
    if (!current || (current.status !== "queued" && current.status !== "processing")) return;
    const timer = window.setInterval(() => void loadDetail(current.id, true).then(loadList).catch(() => undefined), 2000);
    return () => window.clearInterval(timer);
  }, [detail?.package.id, detail?.package.status]);

  const create = (event: FormEvent) => {
    event.preventDefault(); if (!title.trim()) return;
    setBusy(true); setMessage("");
    createCapturePackage({ title: title.trim(), vendorId: role === "admin" ? vendor?.id : undefined }).then((value) => { setTitle(""); setVendor(null); setDetail(value); return loadList(); }).catch((error) => setMessage(getApiErrorMessage(error, "资料包创建失败"))).finally(() => setBusy(false));
  };
  const run = (action: () => Promise<CapturePackageDetail>, fallback: string) => { setBusy(true); setMessage(""); action().then((value) => { setDetail(value); void loadList(); }).catch((error) => setMessage(getApiErrorMessage(error, fallback))).finally(() => setBusy(false)); };
  const changeDraft = (mutate: (draft: CaptureDraft) => void) => setDetail((current) => { if (!current) return current; const draft = JSON.parse(JSON.stringify(current.draft || blankDraft())) as CaptureDraft; mutate(draft); return { ...current, draft }; });
  const updateVendorField = (key: string, value: string, confirm = true) => changeDraft((draft) => { const previous = draft.vendorFields[key] || blankField(); draft.vendorFields[key] = { ...previous, value, confirmed: confirm || previous.confirmed }; });
  const updateProductField = (index: number, key: string, value: string, confirm = true) => changeDraft((draft) => { const previous = draft.products[index].fields[key] || blankField(); draft.products[index].fields[key] = { ...previous, value, confirmed: confirm || previous.confirmed }; });
  const save = () => { if (!detail) return; run(() => updateCaptureDraft(detail.package.id, detail.draft), "草稿保存失败"); };
  const changeDocument = (id: number, sortOrder: number, documentType: "unknown" | "business_card" | "brochure") => updateCaptureDocument(id, { sortOrder, documentType }).then(setDetail).catch((error) => setMessage(getApiErrorMessage(error, "资料设置保存失败")));
  const moveDocument = (id: number, offset: -1 | 1) => {
    if (!detail) return;
    const rows = [...(detail.package.documents || [])].sort((a, b) => a.sortOrder - b.sortOrder);
    const index = rows.findIndex((row) => row.id === id); const target = rows[index + offset]; if (index < 0 || !target) return;
    const current = rows[index]; setBusy(true);
    Promise.all([updateCaptureDocument(current.id, { sortOrder: target.sortOrder, documentType: current.documentType }), updateCaptureDocument(target.id, { sortOrder: current.sortOrder, documentType: target.documentType })]).then(() => loadDetail(detail.package.id)).catch((error) => setMessage(getApiErrorMessage(error, "资料顺序保存失败"))).finally(() => setBusy(false));
  };
  const activeDocument = detail?.package.documents?.find((row) => row.id === selectedDocument) || detail?.package.documents?.[0];
  const hasDraft = Boolean(detail && (Object.keys(detail.draft.vendorFields || {}).length || detail.draft.products?.length));
  const canEdit = detail && !["queued", "processing", "committed"].includes(detail.package.status);

  return <AdminLayout title="智能采集">
    <div className="capture-workbench">
      <aside className="capture-package-sidebar admin-panel">
        <form onSubmit={create}>
          <h2>新建厂商资料包</h2>
          <input maxLength={150} placeholder="例如：2026 展会－某某机械" required value={title} onChange={(event) => setTitle(event.target.value)} />
          {role === "admin" && <VendorOptionSearch label="可选：关联已有厂商" onSelect={setVendor} />}
          {vendor && <small>已关联：{vendor.name}</small>}
          <button className="primary-btn" disabled={busy} type="submit"><Plus size={16} />新建资料包</button>
        </form>
        <div className="capture-package-list">
          {packages.map((item) => <button className={detail?.package.id === item.id ? "active" : ""} key={item.id} type="button" onClick={() => { setInviteURL(""); setMessage(""); void loadDetail(item.id); }}><span>{item.title}</span><small>{item.vendor?.name || "尚未关联厂商"} · {statusLabels[item.status]}</small></button>)}
          {!packages.length && <p>暂无资料包。请按“一家厂商一个资料包”开始拍摄。</p>}
        </div>
      </aside>

      <section className="capture-main">
        {message && <p className="admin-message" role="alert">{message}</p>}
        {!detail ? <div className="admin-panel capture-empty"><Camera size={42} /><h2>选择或新建资料包</h2><p>连续拍摄名片、产品彩页，系统会自动生成待校对草稿。</p></div> : <>
          <header className="admin-panel capture-header">
            <div><span className={`capture-status ${detail.package.status}`}>{statusLabels[detail.package.status]}</span><h2>{detail.package.title}</h2><p>{detail.package.vendor?.name || "识别后匹配或新建厂商"} · {detail.package.documents?.length || 0} 张资料</p></div>
            <div>
              {canEdit && <label className="primary-btn capture-camera"><Camera size={17} />连续拍摄/选图<input accept="image/jpeg,image/png,image/webp" capture="environment" multiple type="file" onChange={(event) => { const files = Array.from(event.target.files || []); event.target.value = ""; if (!files.length) return; setBusy(true); uploadCaptureDocuments(detail.package.id, files).then(() => loadDetail(detail.package.id)).then(loadList).catch((error) => setMessage(getApiErrorMessage(error, "资料上传失败"))).finally(() => setBusy(false)); }} /></label>}
              {canEdit && <button className="outline-btn" disabled={busy || !(detail.package.documents?.length)} type="button" onClick={() => run(() => recognizeCapturePackage(detail.package.id), "识别任务启动失败")}><RefreshCw size={16} />{detail.package.recognitionVersion ? "重新识别" : "开始识别"}</button>}
              {canEdit && <button className="outline-btn danger" type="button" onClick={() => { if (!window.confirm("删除整个资料包？原件将进入待清理状态。")) return; setBusy(true); deleteCapturePackage(detail.package.id).then(() => { setDetail(null); return loadList(); }).finally(() => setBusy(false)); }}><Trash2 size={16} />删除</button>}
            </div>
          </header>
          {!detail.aiEnabled && canEdit && <p className="capture-ai-warning">腾讯云智能识别尚未启用。可继续上传、建立空白草稿或使用现有 XLSX 导入。</p>}
          {detail.package.errorMessage && <p className="form-error">{detail.package.errorMessage}</p>}
          {(detail.package.status === "queued" || detail.package.status === "processing") && <div className="admin-panel capture-processing"><LoaderCircle className="spin" /><strong>正在识别资料</strong><p>可以离开页面，任务完成后会保留结果。</p></div>}
          <div className="capture-review-grid">
            <div className="admin-panel capture-source-panel">
              <h3>原始材料</h3>
              {activeDocument ? <><img alt={activeDocument.asset.originalName} src={`${API_BASE_URL}/api/admin/capture-documents/${activeDocument.id}/content`} /><div className="capture-active-document-actions"><select aria-label="资料类型" disabled={!canEdit} value={activeDocument.documentType} onChange={(event) => void changeDocument(activeDocument.id, activeDocument.sortOrder, event.target.value as typeof activeDocument.documentType)}><option value="unknown">自动判断</option><option value="business_card">名片</option><option value="brochure">产品彩页</option></select>{canEdit && <><button aria-label="前移资料" className="outline-btn small" type="button" onClick={() => moveDocument(activeDocument.id, -1)}><ArrowLeft size={14} /></button><button aria-label="后移资料" className="outline-btn small" type="button" onClick={() => moveDocument(activeDocument.id, 1)}><ArrowRight size={14} /></button><button aria-label="删除资料图片" className="outline-btn small danger" type="button" onClick={() => { if (!window.confirm("删除这张资料图片？")) return; void deleteCaptureDocument(activeDocument.id).then(() => loadDetail(detail.package.id)); }}><Trash2 size={14} /></button></>}</div><div className="capture-document-tabs">{detail.package.documents?.slice().sort((a, b) => a.sortOrder - b.sortOrder).map((row) => <button className={row.id === activeDocument.id ? "active" : ""} key={row.id} onClick={() => setSelectedDocument(row.id)} type="button"><FileImage size={14} />第 {row.sortOrder} 张<small>{row.documentType === "business_card" ? "名片" : row.documentType === "brochure" ? "彩页" : "待识别"}</small></button>)}</div>{activeDocument.ocrText && <details><summary>查看 OCR 原文</summary><pre>{activeDocument.ocrText}</pre></details>}</> : <p>请先拍摄或选择资料图片。</p>}
            </div>
            <div className="admin-panel capture-draft-panel">
              <div className="capture-draft-title"><div><h3>识别草稿</h3><p>黄色字段建议复核，红色及关键字段必须确认。</p></div>{canEdit && !hasDraft && <button className="outline-btn" type="button" onClick={() => setDetail({ ...detail, draft: blankDraft() })}>建立空白草稿</button>}</div>
              {hasDraft && <>
                {detail.draft.warnings?.map((warning) => <p className="capture-warning" key={warning}>{warning}</p>)}
                <FieldSection title="厂商资料" fields={vendorFields} values={detail.draft.vendorFields} onFocus={setSelectedDocument} onChange={updateVendorField} />
                <section className="capture-product-section"><header><h4>产品资料</h4>{canEdit && <button className="outline-btn small" type="button" onClick={() => changeDraft((draft) => draft.products.push(blankProduct(draft.products.length + 1)))}><Plus size={14} />添加产品</button>}</header>
                  {detail.draft.products.map((product, index) => <article className="capture-product-card" key={product.key}><header><strong>产品 {index + 1}</strong>{canEdit && <button aria-label="删除产品" type="button" onClick={() => changeDraft((draft) => draft.products.splice(index, 1))}><Trash2 size={15} /></button>}</header>
                    <div className="capture-fields">{productFields.map(([key, label]) => <CaptureInput field={product.fields[key]} key={key} label={label} name={key} onFocus={setSelectedDocument} onChange={(value) => updateProductField(index, key, value)} onConfirm={() => updateProductField(index, key, product.fields[key]?.value || "", true)} />)}
                      <label className="capture-field"><span>产品分类</span><select value={product.fields.categoryName?.value || ""} onChange={(event) => updateProductField(index, "categoryName", event.target.value)}><option value="">请选择</option>{detail.categories?.map((category) => <option key={category.id} value={category.name}>{category.name}</option>)}</select></label>
                    </div>
                    <ProductCrops detail={detail} product={product} onChange={(ids) => changeDraft((draft) => { draft.products[index].selectedCropIds = ids; })} />
                  </article>)}
                </section>
                {canEdit && <div className="capture-actions"><button className="outline-btn" disabled={busy} type="button" onClick={save}><Save size={16} />保存校对</button><button className="primary-btn" disabled={busy || detail.package.status !== "ready"} type="button" onClick={() => run(() => commitCapturePackage(detail.package.id), "提交审核失败")}><Send size={16} />提交现有审核流程</button></div>}
                {detail.package.status === "committed" && <div className="capture-committed"><Check size={18} /><span>厂商与产品草稿已进入现有审核流程。</span>{role === "admin" && detail.package.vendorId && <button className="outline-btn small" type="button" onClick={() => createVendorInvitation(detail.package.vendorId!).then((result) => { setInviteURL(result.inviteUrl); void navigator.clipboard?.writeText(result.inviteUrl); }).catch((error) => setMessage(getApiErrorMessage(error, "邀请链接生成失败")))}>生成厂商接管邀请</button>}{inviteURL && <code>{inviteURL}</code>}</div>}
              </>}
            </div>
          </div>
        </>}
      </section>
    </div>
  </AdminLayout>;
}

function FieldSection({ title, fields, values, onChange, onFocus }: { title: string; fields: readonly (readonly [string, string])[]; values: Record<string, CaptureField>; onChange: (key: string, value: string, confirm?: boolean) => void; onFocus: (id?: number) => void }) {
  return <section className="capture-field-section"><h4>{title}</h4><div className="capture-fields">{fields.map(([key, label]) => <CaptureInput field={values[key]} key={key} label={label} name={key} onFocus={onFocus} onChange={(value) => onChange(key, value)} onConfirm={() => onChange(key, values[key]?.value || "", true)} />)}</div></section>;
}

function CaptureInput({ field = blankField(), label, name, onChange, onConfirm, onFocus }: { field?: CaptureField; label: string; name: string; onChange: (value: string) => void; onConfirm: () => void; onFocus: (id?: number) => void }) {
  const critical = ["name", "model", "contactName", "phone", "wechat"].includes(name);
  const level = field.confidence < .7 ? "low" : field.confidence < .9 ? "medium" : "high";
  const needsConfirm = Boolean(field.value && !field.confirmed && (critical || field.confidence < .7));
  const multiline = ["description", "detailContent", "mainProducts", "serviceAdvantages", "equipment", "certifications"].includes(name);
  return <label className={`capture-field ${field.value ? level : ""} ${needsConfirm ? "needs-confirm" : ""}`} onFocus={() => onFocus(field.sourceDocumentId)}><span>{label}{field.value && <small>{Math.round(field.confidence * 100)}%</small>}</span>{multiline ? <textarea rows={3} value={field.value} onChange={(event) => onChange(event.target.value)} /> : <input value={field.value} onChange={(event) => onChange(event.target.value)} />}{needsConfirm && <button className="capture-confirm" type="button" onClick={onConfirm}><ClipboardCheck size={14} />确认无误</button>}</label>;
}

function ProductCrops({ detail, product, onChange }: { detail: CapturePackageDetail; product: CaptureProductDraft; onChange: (ids: number[]) => void }) {
  const crops = useMemo(() => detail.package.crops?.filter((crop) => crop.productKey === product.key) || [], [detail.package.crops, product.key]);
  if (!crops.length) return null;
  const selected = product.selectedCropIds || [];
  return <div className="capture-crops"><strong>候选产品图片</strong><div>{crops.map((crop) => <button className={selected.includes(crop.id) ? "selected" : ""} key={crop.id} type="button" onClick={() => onChange(selected.includes(crop.id) ? selected.filter((id) => id !== crop.id) : [...selected, crop.id])}><img alt="候选产品" src={`${API_BASE_URL}/api/admin/capture-crops/${crop.id}/content`} />{selected.includes(crop.id) && <Check size={17} />}</button>)}</div></div>;
}
