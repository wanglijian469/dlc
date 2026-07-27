import { useEffect, useState, type FormEvent } from "react";
import { Plus, Trash2 } from "lucide-react";
import { deleteProductSupplier, listProductSuppliers, saveProductSupplier } from "../../api/admin";
import { getApiErrorMessage } from "../../api/client";
import type { ProductSupplier, VendorOption } from "../../types/api";
import { VendorOptionSearch } from "./VendorOptionSearch";

export function AdminProductSuppliersEditor({ productId, onChanged }: { productId: number; onChanged?: () => void }) {
  const [rows, setRows] = useState<ProductSupplier[]>([]);
  const [selectedVendor, setSelectedVendor] = useState<VendorOption | null>(null);
  const [form, setForm] = useState<Partial<ProductSupplier>>({ vendorId: 0, status: "approved", inquiryText: "欢迎询价" });
  const [message, setMessage] = useState("");
  const load = () => listProductSuppliers(productId).then(setRows);

  useEffect(() => {
    void load();
  }, [productId]);

  const submit = (event: FormEvent) => {
    event.preventDefault();
    if (!form.vendorId) return;
    setMessage("");
    void saveProductSupplier(productId, form)
      .then(() => {
        setMessage("厂商关联已保存并立即生效");
        setSelectedVendor(null);
        setForm({ vendorId: 0, status: "approved", inquiryText: "欢迎询价" });
        void load();
        onChanged?.();
      })
      .catch((error) => setMessage(getApiErrorMessage(error, "厂商关联保存失败")));
  };

  const remove = (row: ProductSupplier) => {
    if (!window.confirm(`确认删除“${row.vendor?.name || `厂商 #${row.vendorId}`}”的供应关联？`)) return;
    setMessage("");
    void deleteProductSupplier(productId, row.id)
      .then(() => {
        setMessage("厂商关联已删除");
        void load();
        onChanged?.();
      })
      .catch((error) => setMessage(getApiErrorMessage(error, "删除厂商关联失败")));
  };

  return (
    <section className="admin-product-suppliers">
      <h3>关联厂商</h3>
      <p>支持按名称、简称、地区和主营产品搜索；管理员维护的供应关系立即生效并记录操作日志。</p>
      {message && <p className="admin-message">{message}</p>}
      <div className="supplier-admin-list">
        {rows.map((row) => (
          <article key={row.id}>
            <div>
              <strong>{row.vendor?.name || `厂商 #${row.vendorId}`}</strong>
              <span>
                {[row.vendor?.province, row.vendor?.city].filter(Boolean).join(" · ") || "地区未填写"}
                {" · "}{row.vendor?.publicationStatus === "published" && row.vendor?.isVisible ? "前台已发布" : "未发布"}
              </span>
              <span>厂商型号：{row.vendorModel || "未填写"} · 价格说明：{row.priceNote || "未填写"}</span>
            </div>
            <button aria-label="删除供应关系" title="删除关联" type="button" onClick={() => remove(row)}>
              <Trash2 size={15} />
            </button>
          </article>
        ))}
      </div>
      {!rows.length && <p className="structured-empty">暂无关联厂商。</p>}
      <VendorOptionSearch
        excludedIds={[...rows.map((row) => row.vendorId), ...(selectedVendor ? [selectedVendor.id] : [])]}
        onSelect={(vendor) => {
          setSelectedVendor(vendor);
          setForm((current) => ({ ...current, vendorId: vendor.id }));
        }}
      />
      {selectedVendor && (
        <form className="supplier-admin-form" onSubmit={submit}>
          <div className="supplier-selected-vendor">
            <strong>{selectedVendor.name}</strong>
            <small>{[selectedVendor.province, selectedVendor.city].filter(Boolean).join(" · ") || "地区未填写"}</small>
          </div>
          <input aria-label="厂商型号" placeholder="厂商型号" value={form.vendorModel || ""} onChange={(event) => setForm({ ...form, vendorModel: event.target.value })} />
          <input aria-label="价格说明" placeholder="价格说明" value={form.priceNote || ""} onChange={(event) => setForm({ ...form, priceNote: event.target.value })} />
          <button className="outline-btn small" type="submit"><Plus size={15} />添加厂商</button>
        </form>
      )}
    </section>
  );
}
