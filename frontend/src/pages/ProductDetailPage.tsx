import { CheckCircle2, Phone, Settings2 } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { getProduct, listProducts } from "../api/public";
import { getProductIndustryKind, IndustryCover } from "../components/public/IndustryCover";
import { PageFrame } from "../components/public/PageFrame";
import { ProductCard } from "../components/public/ProductCard";
import { ErrorState, LoadingState } from "../components/public/StateViews";
import type { Product } from "../types/api";

export function ProductDetailPage() {
  const { id = "" } = useParams();
  const [product, setProduct] = useState<Product | null>(null);
  const [related, setRelated] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
	const [activeImage, setActiveImage] = useState("");
  const load = () => { setLoading(true); setError(""); getProduct(id).then((item) => { setProduct(item); return listProducts({ categoryId: item.categoryId, pageSize: 4 }).then((result) => setRelated(result.items.filter((relatedItem) => String(relatedItem.id) !== id).slice(0, 3))).catch(() => setRelated([])); }).catch(() => setError("产品详情加载失败或产品暂未上架")).finally(() => setLoading(false)); };
  useEffect(load, [id]);
  const gallery = useMemo(() => product ? Array.from(new Set([product.image, ...(product.gallery || [])].filter(Boolean) as string[])) : [], [product]);
	useEffect(() => { setActiveImage(gallery[0] || ""); }, [gallery]);

  if (loading) return <PageFrame title="产品详情"><LoadingState /></PageFrame>;
  if (error || !product) return <PageFrame title="产品详情"><ErrorState text={error || "产品不存在"} onRetry={load} /></PageFrame>;
  const vendorPath = product.vendor?.id ? `/vendors/${product.vendor.id}` : "/vendors";

  return <PageFrame title={product.name} subtitle={product.category?.name || "农机配件产品"}>
    <section className="product-showcase">
      <div className="product-main-media">
		<IndustryCover className="product-detail-fallback" iconSize={86} kind={getProductIndustryKind(product)} />
        {activeImage && <img alt={product.name} src={activeImage} onError={(event) => { event.currentTarget.style.display = "none"; }} />}
		{gallery.length > 1 && <div aria-label="产品图片选择" className="product-thumbnails" role="list">{gallery.map((src, index) => <button aria-label={`查看第 ${index + 1} 张产品图片`} aria-pressed={src === activeImage} key={src} onClick={() => setActiveImage(src)} type="button"><img alt="" src={src} /></button>)}</div>}
      </div>
      <div className="product-showcase-copy">
        <div className="tag-row">{product.isHot && <span className="tag-orange">热销</span>}{product.isRecommended && <span className="tag-blue">推荐</span>}</div>
        <h2>{product.name}</h2><p className="product-lead">{product.description || "可联系供应商确认型号、库存、交期和批量报价。"}</p>
        <dl><div><dt>适配机型</dt><dd>{product.compatibleModels || "通用农机配件"}</dd></div><div><dt>所属分类</dt><dd>{product.category?.name || "农机配件"}</dd></div><div><dt>供应厂商</dt><dd><Link to={vendorPath}>{product.vendor?.name || "平台供应商"}</Link></dd></div><div><dt>价格说明</dt><dd>{product.priceNote || "面议 / 批量报价"}</dd></div></dl>
        <div className="contact-actions"><Link className="primary-btn" to={vendorPath}><Phone size={16} />联系供应商</Link><Link className="outline-btn" to={`/products?vendorId=${product.vendor?.id || ""}`}>查看该厂商产品</Link></div>
      </div>
    </section>

    {gallery.length > 1 && <section className="vendor-section-card"><header><CheckCircle2 size={20} /><h2>产品图片</h2></header><div className="product-gallery">{gallery.map((src) => <img alt={product.name} loading="lazy" key={src} src={src} />)}</div></section>}
    <section className="product-detail-grid">
      {product.detailContent && <article className="vendor-section-card"><header><CheckCircle2 size={20} /><h2>产品说明</h2></header><p>{product.detailContent}</p></article>}
      {product.specs?.length ? <article className="vendor-section-card"><header><Settings2 size={20} /><h2>规格参数</h2></header><dl className="spec-table">{product.specs.map((spec) => <div key={spec.name}><dt>{spec.name}</dt><dd>{spec.value}</dd></div>)}</dl></article> : null}
    </section>
    {related.length > 0 && <section className="section-block"><div className="section-title"><h2>相关产品</h2><Link to={`/products?categoryId=${product.categoryId}`}>查看更多</Link></div><div className="product-grid related-products">{related.map((item) => <ProductCard key={item.id} product={item} />)}</div></section>}
  </PageFrame>;
}
