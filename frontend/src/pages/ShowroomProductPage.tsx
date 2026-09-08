import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { getShowroomProduct } from "../api/workspace";
import type { ProductSupplier } from "../types/api";
import { PageFrame } from "../components/public/PageFrame";
import { ImageLightbox } from "../components/public/ImageLightbox";
import { ErrorState, LoadingState } from "../components/public/StateViews";
import { VendorContactActions } from "../components/public/VendorContactActions";
import { PromotionTools } from "../components/public/PromotionTools";
import { publicMediaURL } from "../utils/publicMedia";
import { getStaticPageData } from "../utils/staticPageData";
export function ShowroomProductPage() {
 const { slug = "", supplierId = "" } = useParams();
 const initial = getStaticPageData("supplier", slug + "/products/" + supplierId)?.supplier || null;
 const [supplier, setSupplier] = useState<ProductSupplier | null>(initial), [error, setError] = useState(""), [index, setIndex] = useState<number | null>(null);
 useEffect(() => { let active = true; setSupplier(initial); setIndex(null); setError(""); getShowroomProduct(slug, supplierId).then(row => { if (active) setSupplier(row); }).catch(() => { if (active) { setSupplier(null); setError("本厂产品不存在或暂未公开"); } }); return () => { active = false; }; }, [slug, supplierId]);
 const vendor = supplier?.vendor;
 if (error) return <PageFrame title="本厂产品"><ErrorState text={error} onRetry={() => window.location.reload()}/></PageFrame>;
 if (!supplier || !vendor) return <PageFrame title="本厂产品"><LoadingState/></PageFrame>;
 const path = "/v/" + slug + "/products/" + supplierId;
 const images = [supplier.image, ...(supplier.gallery || []), ...(supplier.specs || []).map(s => s.image)].filter(Boolean).map(src => ({ src: publicMediaURL(src), alt: supplier.vendorProductName || "本厂产品" }));
 const expired = supplier.priceValidUntil && new Date(supplier.priceValidUntil) < new Date();
 return <PageFrame title={supplier.vendorProductName || "本厂产品"} breadcrumbs={[{ label: vendor.name, path: "/v/" + slug }]} subtitle={"本厂型号：" + (supplier.vendorModel || "请联系厂商确认")}>
  <section className="own-product-hero"><div className="own-product-media">{images[0] ? <button className="image-zoom-trigger" aria-label="放大本厂产品主图" onClick={() => setIndex(0)}><img src={images[0].src} alt={images[0].alt}/></button> : <div className="image-placeholder">本厂产品图片待补充</div>}</div><div className="own-product-copy"><Link className="showroom-company" to={"/v/" + slug}>{vendor.name} · 企业展厅</Link><h2>{supplier.vendorProductName}</h2><p>{supplier.description}</p><strong className="own-product-price">{expired ? "价格已过期，请询价" : supplier.negotiable || !supplier.unitPriceCents ? supplier.priceNote || "价格面议" : "¥" + (supplier.unitPriceCents / 100).toFixed(2) + " / " + (supplier.priceUnit || "件")}</strong><dl>{[["型号", supplier.vendorModel], ["适配信息", supplier.compatibleModels], ["起订量", supplier.minOrderQuantity], ["可供应量", supplier.availableQuantity === 0 ? "按需供应" : supplier.availableQuantity], ["交期", supplier.leadTime], ["运费", supplier.freightNote], ["供货能力", supplier.supplyAbility]].map(([label, value]) => <div key={label}><dt>{label}</dt><dd>{value ?? "请联系厂商确认"}</dd></div>)}</dl><small>资料更新：{supplier.updatedAt ? new Date(supplier.updatedAt).toLocaleDateString() : "以厂商最新确认为准"}</small><VendorContactActions vendor={vendor} path={path} sticky/><PromotionTools path={path} title={vendor.name + " · " + supplier.vendorProductName} description={supplier.description} image={supplier.image}/></div></section>
  <section className="vendor-section-card"><h2>产品介绍与参数</h2><p className="pre-line">{supplier.detailContent || "详细资料请联系本厂确认。"}</p><dl className="supplier-specs">{supplier.specs?.map((spec, i) => <div key={i}><dt>{spec.name}</dt><dd>{spec.value}</dd></div>)}</dl><div className="vendor-media-grid">{images.slice(1).map((img, i) => <button key={i} className="image-zoom-trigger" aria-label={"放大本厂产品图片 " + (i + 2)} onClick={() => setIndex(i + 1)}><img loading="lazy" src={img.src} alt={img.alt}/></button>)}</div></section>
  {index !== null && <ImageLightbox images={images} index={index} onIndexChange={setIndex} onClose={() => setIndex(null)}/>}
 </PageFrame>;
}
