import { useEffect, useState } from "react";
import { listResourcePage } from "../../api/admin";
import type { Product } from "../../types/api";

export function AdminVendorProductsPanel({ vendorId }: { vendorId: number }) {
  const [rows, setRows] = useState<Product[]>([]);
  useEffect(() => { void listResourcePage<Product & { id: number }>("products", { page: 1, pageSize: 100, vendorId }).then((result) => setRows(result.items)); }, [vendorId]);
  return <section className="admin-product-suppliers"><h3>关联产品</h3><p>以下产品通过供应关系关联到该厂商；新增或调整关联请进入“配件产品”编辑页。</p><div className="supplier-admin-list">{rows.map((product) => <article key={product.id}><div><strong>{product.name}</strong><span>{product.category?.name || "农机配件"} · {product.publicationStatus === "published" ? "目录已发布" : "目录未发布"}</span></div></article>)}</div>{!rows.length && <p className="structured-empty">该厂商暂无已关联产品。</p>}</section>;
}
