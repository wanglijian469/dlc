import { useEffect, useState, type FormEvent } from "react";
import { ImageUp, Pencil, Plus, Trash2 } from "lucide-react";
import { createOwnProduct, deleteOwnProduct, listOwnProducts, updateOwnProduct, uploadFile } from "../../api/admin";
import { getFilterOptions } from "../../api/public";
import type { Category, Product } from "../../types/api";
import { GalleryEditor, SpecsEditor } from "./StructuredEditors";
import { ProtectedMediaImage } from "./ProtectedMediaImage";

const emptyProduct: Partial<Product> = { name: "", categoryId: undefined, compatibleModels: "", description: "", detailContent: "", image: "", galleryRaw: "[]", specsRaw: "[]", priceNote: "" };

export function VendorProductsEditor() {
  const [products, setProducts] = useState<Product[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [form, setForm] = useState<Partial<Product>>(emptyProduct);
  const [editingId, setEditingId] = useState<number>();
  const [message, setMessage] = useState("");
  const [open, setOpen] = useState(false);
  const load = () => Promise.all([listOwnProducts(), getFilterOptions()]).then(([rows, filters]) => { setProducts(rows); setCategories(filters.categories); }).catch(() => setMessage("产品资料加载失败"));
  useEffect(() => { void load(); }, []);

  const submit = (event: FormEvent) => {
    event.preventDefault();
    setMessage("");
    const action = editingId ? updateOwnProduct(editingId, form) : createOwnProduct(form);
    action.then(() => { setMessage("产品资料已保存，管理员审核上架后将在前台展示"); setForm(emptyProduct); setEditingId(undefined); setOpen(false); void load(); }).catch(() => setMessage("保存失败，请检查产品名称和图片"));
  };
  const edit = (product: Product) => { setEditingId(product.id); setForm(product); setOpen(true); };
  const remove = (id: number) => { if (!window.confirm("确定删除这条产品资料吗？")) return; deleteOwnProduct(id).then(() => void load()).catch(() => setMessage("删除失败")); };
  const uploadCover = (file: File) => uploadFile(file).then((result) => setForm((current) => ({ ...current, image: result.url })));

  return <section className="vendor-products-editor admin-panel">
    <header><div><h2>厂商产品与说明</h2><p>上传产品照片、填写适配机型和产品说明；新内容需管理员审核后上架。</p></div><button className="primary-btn small" type="button" onClick={() => { setForm(emptyProduct); setEditingId(undefined); setOpen(true); }}><Plus size={16} />新增产品</button></header>
    {message && <p className="admin-message">{message}</p>}
    <div className="vendor-product-list">{products.map((product) => <article key={product.id}><ProtectedMediaImage assetId={assetId(product.image)} alt={product.name} src={product.image || ""} /><div><strong>{product.name}</strong><span>{product.category?.name || "未分类"} · {product.status === 1 ? "已上架" : "待审核"}</span><p>{product.description || "暂无产品说明"}</p></div><div><button aria-label="编辑产品" type="button" onClick={() => edit(product)}><Pencil size={16} /></button><button aria-label="删除产品" type="button" onClick={() => remove(product.id)}><Trash2 size={16} /></button></div></article>)}</div>
    {!products.length && <p className="structured-empty">暂无产品资料，可点击“新增产品”录入。</p>}
    {open && <form className="vendor-product-form" onSubmit={submit}>
      <header><strong>{editingId ? "编辑产品" : "新增产品"}</strong><button type="button" onClick={() => setOpen(false)}>关闭</button></header>
      <div className="vendor-product-fields">
        <label>产品名称<input required value={form.name || ""} onChange={(event) => setForm({ ...form, name: event.target.value })} /></label>
        <label>产品分类<select value={form.categoryId || ""} onChange={(event) => setForm({ ...form, categoryId: Number(event.target.value) || undefined })}><option value="">请选择</option>{categories.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}</select></label>
        <label>适配机型<input value={form.compatibleModels || ""} onChange={(event) => setForm({ ...form, compatibleModels: event.target.value })} /></label>
        <label>价格说明<input value={form.priceNote || ""} onChange={(event) => setForm({ ...form, priceNote: event.target.value })} /></label>
        <label className="wide-field">列表简介<textarea value={form.description || ""} onChange={(event) => setForm({ ...form, description: event.target.value })} /></label>
        <label className="wide-field">产品详细说明<textarea value={form.detailContent || ""} onChange={(event) => setForm({ ...form, detailContent: event.target.value })} /></label>
        <div className="wide-field product-cover-upload"><strong>产品主图</strong><label className="outline-btn small upload-button"><ImageUp size={15} />上传主图<input accept="image/jpeg,image/png,image/webp" type="file" onChange={(event) => { const file = event.target.files?.[0]; if (file) void uploadCover(file); }} /></label>{form.image && <ProtectedMediaImage assetId={assetId(form.image)} alt="产品主图预览" src={form.image} />}</div>
        <div className="wide-field"><GalleryEditor value={form.galleryRaw || "[]"} onChange={(galleryRaw) => setForm({ ...form, galleryRaw })} /></div>
        <div className="wide-field"><SpecsEditor value={form.specsRaw || "[]"} onChange={(specsRaw) => setForm({ ...form, specsRaw })} /></div>
      </div>
      <button className="primary-btn" type="submit">保存产品资料</button>
    </form>}
  </section>;
}

function assetId(url?: string) { const match = url?.match(/\/api\/media\/(\d+)/); return match ? Number(match[1]) : undefined; }
