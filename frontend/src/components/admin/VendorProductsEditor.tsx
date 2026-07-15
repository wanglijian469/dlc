import { useEffect, useState, type FormEvent } from "react";
import { ImageUp, Link2, Pencil, Plus, Search, Trash2, X } from "lucide-react";
import { createOwnProduct, deleteOwnProduct, linkOwnProduct, listOwnProducts, searchVendorProductCatalog, updateOwnProduct, uploadFile } from "../../api/admin";
import { getFilterOptions } from "../../api/public";
import type { Category, Product, ProductSupplier, VendorProductRecord } from "../../types/api";
import { hierarchicalCategoryOptions } from "../../utils/categories";
import { ProtectedMediaImage } from "./ProtectedMediaImage";

const emptyCandidate: Partial<Product> = { name: "", compatibleModels: "", description: "", detailContent: "", image: "", priceNote: "" };
const emptyOffer: Partial<ProductSupplier> = { vendorProductName: "", vendorModel: "", compatibleModels: "", description: "", image: "", priceNote: "", inquiryText: "欢迎询价" };

export function VendorProductsEditor() {
  const [records, setRecords] = useState<VendorProductRecord[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [mode, setMode] = useState<"link" | "new" | "edit">("link");
  const [candidate, setCandidate] = useState<Partial<Product>>(emptyCandidate);
  const [offer, setOffer] = useState<Partial<ProductSupplier>>(emptyOffer);
  const [editingSupplierId, setEditingSupplierId] = useState<number>();
  const [catalog, setCatalog] = useState<Product[]>([]);
  const [selectedProduct, setSelectedProduct] = useState<Product>();
  const [keyword, setKeyword] = useState("");
  const [open, setOpen] = useState(false);
  const [message, setMessage] = useState("");

  const load = () => Promise.all([listOwnProducts(), getFilterOptions()]).then(([rows, filters]) => {
    setRecords(rows); setCategories(filters.categories);
  }).catch(() => setMessage("产品资料加载失败"));
  useEffect(() => { void load(); }, []);
  useEffect(() => {
    if (!open) return;
    const previousOverflow = document.body.style.overflow;
    const closeOnEscape = (event: KeyboardEvent) => { if (event.key === "Escape") setOpen(false); };
    document.body.style.overflow = "hidden";
    window.addEventListener("keydown", closeOnEscape);
    return () => { document.body.style.overflow = previousOverflow; window.removeEventListener("keydown", closeOnEscape); };
  }, [open]);

  const searchCatalog = () => searchVendorProductCatalog(keyword).then(setCatalog).catch(() => setMessage("平台产品目录搜索失败"));
  const reset = (nextMode: "link" | "new") => { setMode(nextMode); setCandidate(emptyCandidate); setOffer(emptyOffer); setSelectedProduct(undefined); setEditingSupplierId(undefined); setCatalog([]); setKeyword(""); setOpen(true); };
  const edit = (record: VendorProductRecord) => { setMode("edit"); setCandidate(record.product); setOffer(record.supplier); setEditingSupplierId(record.supplier.id); setOpen(true); };
  const submit = (event: FormEvent) => {
    event.preventDefault(); setMessage("");
    let action: Promise<unknown>;
    if (mode === "new") action = createOwnProduct(candidate);
    else if (mode === "edit" && editingSupplierId) action = updateOwnProduct(editingSupplierId, offer);
    else if (selectedProduct) action = linkOwnProduct(selectedProduct.id, offer);
    else { setMessage("请先选择一个平台产品"); return; }
    action.then(() => { setMessage("已提交产品审核；已发布的旧供应信息会继续展示到新版本通过审核"); setOpen(false); void load(); }).catch(() => setMessage("提交失败，请检查必填项或是否已关联该产品"));
  };
  const remove = (supplierId: number) => { if (!window.confirm("确定停止供应该产品吗？")) return; deleteOwnProduct(supplierId).then(() => void load()).catch(() => setMessage("停止供应失败")); };
  const uploadCover = (file: File) => uploadFile(file).then((result) => mode === "new" ? setCandidate((value) => ({ ...value, image: result.url })) : setOffer((value) => ({ ...value, image: result.url })));

  return <section className="vendor-products-editor admin-panel">
    <header><div><h2>厂商产品与供应信息</h2><p>优先关联平台已有产品；搜索不到时可提交新的产品候选。所有新增和修改均进入独立产品审核队列。</p></div><div className="card-actions"><button className="outline-btn small" type="button" onClick={() => reset("link")}><Link2 size={16} />关联平台产品</button><button className="primary-btn small" type="button" onClick={() => reset("new")}><Plus size={16} />提交新产品</button></div></header>
    {message && <p className="admin-message">{message}</p>}
    <div className="vendor-product-list">{records.map((record) => <article key={record.supplier.id}>
      <ProtectedMediaImage assetId={assetId(record.supplier.image || record.product.image)} alt={record.product.name} src={record.supplier.image || record.product.image || ""} />
      <div><strong>{record.supplier.vendorProductName || record.product.name}</strong><span>{record.product.category?.name || "未分类"} · {statusLabel(record.supplier.status)}{record.latestSubmission?.status === "pending" ? " · 有待审核修改" : ""}</span><p>{record.supplier.description || record.product.description || "暂无供应说明"}</p></div>
      <div><button aria-label="编辑供应信息" type="button" onClick={() => edit(record)}><Pencil size={16} /></button><button aria-label="停止供应" type="button" onClick={() => remove(record.supplier.id)}><Trash2 size={16} /></button></div>
    </article>)}</div>
    {!records.length && <p className="structured-empty">暂无产品供应信息，可先搜索平台产品并建立供应关联。</p>}
    {open && <div className="vendor-product-drawer-layer"><button aria-label="关闭产品编辑抽屉" className="vendor-product-drawer-backdrop" type="button" onClick={() => setOpen(false)} /><aside aria-label={drawerTitle(mode)} aria-modal="true" className="vendor-product-drawer" role="dialog"><form className="vendor-product-form" onSubmit={submit}>
      <header><div><span>{mode === "edit" ? "供应信息维护" : "产品资料提交"}</span><strong>{drawerTitle(mode)}</strong><p>{drawerHint(mode)}</p></div><button aria-label="关闭产品编辑抽屉" type="button" onClick={() => setOpen(false)}><X size={20} /></button></header>
      <div className="vendor-product-drawer-body">
        {mode === "link" && <div className="catalog-picker"><div className="admin-search"><Search size={16} /><input placeholder="搜索产品名称、适配机型" value={keyword} onChange={(event) => setKeyword(event.target.value)} /><button className="outline-btn small" type="button" onClick={searchCatalog}>搜索</button></div><div className="catalog-results">{catalog.map((product) => <button className={selectedProduct?.id === product.id ? "selected" : ""} key={product.id} type="button" onClick={() => { setSelectedProduct(product); setOffer((value) => ({ ...value, vendorProductName: value.vendorProductName || product.name, compatibleModels: value.compatibleModels || product.compatibleModels })); }}><strong>{product.name}</strong><span>{product.category?.name || "农机配件"} · 已有 {product.supplierCount || 0} 家供应商</span></button>)}</div>{!catalog.length && <p className="structured-empty">输入名称或适配机型搜索平台产品；搜索不到时关闭抽屉后选择“提交新产品”。</p>}</div>}
        {mode === "new" ? <CandidateFields form={candidate} categories={categories} onChange={setCandidate} /> : <OfferFields product={selectedProduct || candidate as Product} form={offer} onChange={setOffer} />}
        <div className="wide-field product-cover-upload"><strong>{mode === "new" ? "产品标准配图" : "本厂产品配图"}</strong><label className="outline-btn small upload-button"><ImageUp size={15} />上传图片<input accept="image/jpeg,image/png,image/webp" type="file" onChange={(event) => { const file = event.target.files?.[0]; if (file) void uploadCover(file); }} /></label>{(mode === "new" ? candidate.image : offer.image) && <ProtectedMediaImage assetId={assetId(mode === "new" ? candidate.image : offer.image)} alt="产品图片预览" src={(mode === "new" ? candidate.image : offer.image) || ""} />}</div>
      </div>
      <footer><button className="outline-btn" type="button" onClick={() => setOpen(false)}>取消</button><button className="primary-btn" type="submit">提交审核</button></footer>
    </form></aside></div>}
  </section>;
}

function CandidateFields({ form, categories, onChange }: { form: Partial<Product>; categories: Category[]; onChange: (value: Partial<Product>) => void }) { return <div className="vendor-product-fields">
  <label>产品名称<input required value={form.name || ""} onChange={(event) => onChange({ ...form, name: event.target.value })} /></label>
  <label>产品分类<select value={form.categoryId || ""} onChange={(event) => onChange({ ...form, categoryId: Number(event.target.value) || undefined })}><option value="">请选择</option>{hierarchicalCategoryOptions(categories).map((item) => <option key={item.value} value={item.value}>{item.label}</option>)}</select></label>
  <label>通用适配机型<input value={form.compatibleModels || ""} onChange={(event) => onChange({ ...form, compatibleModels: event.target.value })} /></label>
  <label>本厂价格说明<input value={form.priceNote || ""} onChange={(event) => onChange({ ...form, priceNote: event.target.value })} /></label>
  <label className="wide-field">公共产品说明<textarea value={form.description || ""} onChange={(event) => onChange({ ...form, description: event.target.value })} /></label>
  <label className="wide-field">产品详细说明<textarea value={form.detailContent || ""} onChange={(event) => onChange({ ...form, detailContent: event.target.value })} /></label>
</div>; }

function OfferFields({ product, form, onChange }: { product?: Product; form: Partial<ProductSupplier>; onChange: (value: Partial<ProductSupplier>) => void }) { return <div className="vendor-product-fields">
  {product?.name && <p className="wide-field admin-message">关联目录：{product.name}</p>}
  <label>本厂产品名称<input value={form.vendorProductName || ""} onChange={(event) => onChange({ ...form, vendorProductName: event.target.value })} /></label>
  <label>本厂型号<input value={form.vendorModel || ""} onChange={(event) => onChange({ ...form, vendorModel: event.target.value })} /></label>
  <label>适配信息<input value={form.compatibleModels || ""} onChange={(event) => onChange({ ...form, compatibleModels: event.target.value })} /></label>
  <label>价格说明<input value={form.priceNote || ""} onChange={(event) => onChange({ ...form, priceNote: event.target.value })} /></label>
  <label className="wide-field">本厂供应说明<textarea value={form.description || ""} onChange={(event) => onChange({ ...form, description: event.target.value })} /></label>
  <label>询价文案<input value={form.inquiryText || ""} onChange={(event) => onChange({ ...form, inquiryText: event.target.value })} /></label>
</div>; }

function statusLabel(status: ProductSupplier["status"]) { return ({ pending: "待审核", approved: "已展示", rejected: "已驳回", disabled: "已停供" })[status]; }
function drawerTitle(mode: "link" | "new" | "edit") { return mode === "new" ? "提交新产品候选" : mode === "edit" ? "编辑本厂供应信息" : "关联平台产品"; }
function drawerHint(mode: "link" | "new" | "edit") { return mode === "new" ? "补充公共产品资料和标准配图，提交后由管理员审核。" : mode === "edit" ? "修改仅影响本厂供应信息，已发布版本会保留到新版本审核通过。" : "先查找平台已有产品，再填写本厂型号、适配信息和价格说明。"; }
function assetId(url?: string) { const match = url?.match(/\/api\/media\/(\d+)/); return match ? Number(match[1]) : undefined; }
