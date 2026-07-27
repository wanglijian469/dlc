import { Building2, CheckCircle2, MapPin, Phone, Settings2 } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { getProduct, getProductSuppliers, listProducts } from "../api/public";
import { getProductIndustryKind, IndustryCover } from "../components/public/IndustryCover";
import { PageFrame } from "../components/public/PageFrame";
import { ProductCard } from "../components/public/ProductCard";
import { ErrorState, LoadingState } from "../components/public/StateViews";
import type { Product, ProductSupplier } from "../types/api";
import { useSite } from "../contexts/SiteContext";
import { getMenuLabel } from "../utils/navigation";
import { vendorPath } from "../utils/vendorPath";

export function ProductDetailPage() {
  const { layout } = useSite();
  const productsLabel = getMenuLabel(layout.topMenus, "/products", "配件产品");
  const { id = "" } = useParams();
  const [product, setProduct] = useState<Product | null>(null);
  const [suppliers, setSuppliers] = useState<ProductSupplier[]>([]);
  const [related, setRelated] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [activeImage, setActiveImage] = useState("");
  const load = () => {
    setLoading(true); setError("");
    Promise.all([getProduct(id), getProductSuppliers(id)]).then(([item, supplierRows]) => {
      setProduct(item); setSuppliers(supplierRows);
      return listProducts({ categoryId: item.categoryId, pageSize: 4 }).then((result) => setRelated(result.items.filter((entry) => String(entry.id) !== id).slice(0, 3))).catch(() => setRelated([]));
    }).catch(() => setError("产品详情加载失败或产品暂未上架")).finally(() => setLoading(false));
  };
  useEffect(load, [id]);
  const gallery = useMemo(() => product ? Array.from(new Set([product.image, ...(product.gallery || [])].filter(Boolean) as string[])) : [], [product]);
  useEffect(() => { setActiveImage(gallery[0] || ""); }, [gallery]);

  if (loading) return <PageFrame breadcrumbs={[{ label: productsLabel, path: "/products" }]} title="产品详情"><LoadingState /></PageFrame>;
  if (error || !product) return <PageFrame breadcrumbs={[{ label: productsLabel, path: "/products" }]} title="产品详情"><ErrorState text={error || "产品不存在"} onRetry={load} /></PageFrame>;
  const productBreadcrumbs = [{ label: productsLabel, path: "/products" }];
  if (product.category?.name) productBreadcrumbs.push({ label: product.category.name, path: product.category.slug ? `/products/category/${product.category.slug}` : `/products?categoryId=${product.categoryId}` });
  return <PageFrame breadcrumbs={productBreadcrumbs} title={product.name} subtitle={product.category?.name || "农机配件产品目录"}>
    <section className="product-showcase">
      <IndustryCover className="product-main-media" image={activeImage} imageAlt={product.name} iconSize={86} kind={getProductIndustryKind(product)}>{gallery.length > 1 && <div aria-label="产品图片选择" className="product-thumbnails" role="list">{gallery.map((src, index) => <button aria-label={`查看第 ${index + 1} 张产品图片`} aria-pressed={src === activeImage} key={src} onClick={() => setActiveImage(src)} type="button"><img alt="" src={src} onError={(event) => { event.currentTarget.parentElement!.style.display = "none"; }} /></button>)}</div>}</IndustryCover>
      <div className="product-showcase-copy"><div className="tag-row">{product.isHot && <span className="tag-orange">热销</span>}{product.isRecommended && <span className="tag-blue">推荐</span>}</div><h2>{product.name}</h2><p className="product-lead">{product.description || "平台公共产品资料；具体型号、价格、库存与交期请向下方供应商确认。"}</p><dl><div><dt>适配机型</dt><dd>{product.compatibleModels || "通用农机配件"}</dd></div><div><dt>所属分类</dt><dd>{product.category?.name || "农机配件"}</dd></div><div><dt>支持供应商</dt><dd>{product.supplierCount || suppliers.length} 家</dd></div><div><dt>供应地区</dt><dd>{product.supplierRegions?.join(" · ") || "以供应商说明为准"}</dd></div></dl><div className="contact-actions"><a className="primary-btn" href="#suppliers"><Building2 size={16} />查看供应商</a><Link className="outline-btn" to="/products">返回产品目录</Link></div></div>
    </section>
    {gallery.length > 1 && <section className="vendor-section-card"><header><CheckCircle2 size={20} /><h2>产品图片</h2></header><div className="product-gallery">{gallery.map((src) => <img alt={product.name} loading="lazy" key={src} src={src} />)}</div></section>}
    <section className="product-detail-grid">{product.detailContent && <article className="vendor-section-card"><header><CheckCircle2 size={20} /><h2>产品说明</h2></header><p>{product.detailContent}</p></article>}{product.specs?.length ? <article className="vendor-section-card"><header><Settings2 size={20} /><h2>规格参数</h2></header><dl className="spec-table">{product.specs.map((spec) => <div key={spec.name}><dt>{spec.name}</dt><dd>{spec.value}</dd></div>)}</dl></article> : null}</section>
    <section className="vendor-section-card product-suppliers" id="suppliers"><header><Building2 size={20} /><h2>供应商（{suppliers.length} 家）</h2></header><div className="supplier-list">{suppliers.map((supplier) => <article key={supplier.id}><div><strong>{supplier.vendor?.name || `供应商 #${supplier.vendorId}`}</strong><span><MapPin size={14} />{[supplier.vendor?.province, supplier.vendor?.city].filter(Boolean).join(" · ") || "供应区域请咨询厂商"}</span></div><dl><div><dt>厂商型号</dt><dd>{supplier.vendorModel || "按需匹配"}</dd></div><div><dt>适配信息</dt><dd>{supplier.compatibleModels || product.compatibleModels || "请咨询厂商"}</dd></div><div><dt>价格说明</dt><dd>{supplier.priceNote || "面议 / 批量报价"}</dd></div></dl><p>{supplier.description || "该厂商可供应此产品，具体库存和交期请直接联系。"}</p><Link className="primary-btn small" to={vendorPath(supplier.vendor || { id: supplier.vendorId })}><Phone size={15} />{supplier.inquiryText || "联系该厂商"}</Link></article>)}</div>{!suppliers.length && <p className="structured-empty">暂无已通过审核的供应商。</p>}</section>
    {related.length > 0 && <section className="section-block"><div className="section-title"><h2>相关产品</h2><Link to={product.category?.slug ? `/products/category/${product.category.slug}` : `/products?categoryId=${product.categoryId}`}>查看更多</Link></div><div className="product-grid related-products">{related.map((item) => <ProductCard key={item.id} product={item} />)}</div></section>}
    <div className="mobile-product-action"><a className="primary-btn" href="#suppliers"><Building2 size={17} />查看供应商（{suppliers.length}）</a></div>
  </PageFrame>;
}
