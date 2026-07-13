import { useEffect, useState, type FormEvent } from "react";
import { Plus, Trash2 } from "lucide-react";
import { deleteProductSupplier, listProductSuppliers, listResource, saveProductSupplier } from "../../api/admin";
import type { ProductSupplier, Vendor } from "../../types/api";

export function AdminProductSuppliersEditor({ productId }: { productId: number }) {
  const [rows, setRows] = useState<ProductSupplier[]>([]);
  const [vendors, setVendors] = useState<Vendor[]>([]);
  const [form, setForm] = useState<Partial<ProductSupplier>>({ vendorId: 0, status: "approved", inquiryText: "欢迎询价" });
  const [message, setMessage] = useState("");
  const load = () => Promise.all([listProductSuppliers(productId), listResource<Vendor & { id: number }>("vendors")]).then(([suppliers, vendorRows]) => { setRows(suppliers); setVendors(vendorRows); });
  useEffect(() => { void load(); }, [productId]);
  const submit = (event: FormEvent) => { event.preventDefault(); if (!form.vendorId) return; saveProductSupplier(productId, form).then(() => { setMessage("供应商关联已保存并立即通过"); setForm({ vendorId: 0, status: "approved", inquiryText: "欢迎询价" }); void load(); }).catch(() => setMessage("供应商关联保存失败")); };
  const remove = (row: ProductSupplier) => {
    if (!window.confirm(`确认删除“${row.vendor?.name || `厂商 #${row.vendorId}`}”的供应关联？`)) return;
    deleteProductSupplier(productId, row.id).then(() => { setMessage("供应商关联已删除"); void load(); }).catch(() => setMessage("删除供应商关联失败"));
  };
  return <section className="admin-product-suppliers"><h3>关联供应商</h3><p>管理员直接维护的供应关系立即生效，并记录操作日志。</p>{message && <p className="admin-message">{message}</p>}<div className="supplier-admin-list">{rows.map((row) => <article key={row.id}><div><strong>{row.vendor?.name || `厂商 #${row.vendorId}`}</strong><span>{row.vendorModel || "未填写厂商型号"} · {row.status}</span></div><button aria-label="删除供应关系" title="删除关联" type="button" onClick={() => remove(row)}><Trash2 size={15} /></button></article>)}</div>{!rows.length && <p className="structured-empty">暂无关联供应商。</p>}<form className="supplier-admin-form" onSubmit={submit}><select required value={form.vendorId || ""} onChange={(event) => setForm({ ...form, vendorId: Number(event.target.value) })}><option value="">选择厂商</option>{vendors.map((vendor) => <option key={vendor.id} value={vendor.id}>{vendor.name}</option>)}</select><input placeholder="厂商型号" value={form.vendorModel || ""} onChange={(event) => setForm({ ...form, vendorModel: event.target.value })} /><input placeholder="价格说明" value={form.priceNote || ""} onChange={(event) => setForm({ ...form, priceNote: event.target.value })} /><button className="outline-btn small" type="submit"><Plus size={15} />添加供应商</button></form></section>;
}
