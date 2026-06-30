import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { getVendor, listProducts } from "../api/public";
import { PageFrame } from "../components/public/PageFrame";
import { ErrorState, LoadingState } from "../components/public/StateViews";
import { VendorCover } from "../components/public/VendorCover";
import type { Product, Vendor } from "../types/api";

export function VendorDetailPage() {
  const { id = "" } = useParams();
  const [vendor, setVendor] = useState<Vendor | null>(null);
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const load = () => {
    setLoading(true);
    setError("");
    Promise.all([getVendor(id), listProducts({ vendorId: id, pageSize: 8 })])
      .then(([vendorData, productData]) => {
        setVendor(vendorData);
        setProducts(productData.items);
      })
      .catch(() => setError("厂商详情加载失败，请稍后重试"))
      .finally(() => setLoading(false));
  };

  useEffect(load, [id]);

  if (loading) return <PageFrame title="厂商详情"><LoadingState /></PageFrame>;
  if (error || !vendor) return <PageFrame title="厂商详情"><ErrorState text={error || "厂商不存在"} onRetry={load} /></PageFrame>;

  const region = [vendor.province, vendor.city, vendor.county].filter(Boolean).join(" ");
  return (
    <PageFrame title={vendor.name} subtitle={region || "源头农机配件厂商"}>
      <section className="vendor-detail-hero">
        <VendorCover vendor={vendor} variant="detail" />
        <div className="vendor-detail-info">
          <img alt={vendor.name} src={vendor.logo || "https://dummyimage.com/120x80/ffffff/0b5fea&text=DL"} />
          <div>
            <h2>{vendor.shortName || vendor.name}</h2>
            <p>{vendor.description || "专注农机配件生产与供应，支持批量采购和定制服务。"}</p>
            <div className="tag-row">
              {vendor.isVerified && <span>平台认证</span>}
              {vendor.isRecommended && <span>推荐厂商</span>}
              {vendor.tags?.map((tag) => (
                <span key={tag.id}>{tag.name}</span>
              ))}
            </div>
          </div>
        </div>
      </section>
      <section className="detail-grid">
        <article>
          <h3>主营产品</h3>
          <p>{vendor.mainProducts || "农机配件"}</p>
        </article>
        <article>
          <h3>服务优势</h3>
          <p>{vendor.serviceAdvantages || "质量稳定，发货及时，服务完善"}</p>
        </article>
        <article>
          <h3>联系方式</h3>
          <p>{vendor.contactName || "销售经理"} {vendor.phone || "请通过平台联系"}</p>
          {vendor.wechat && <p>微信：{vendor.wechat}</p>}
        </article>
      </section>
      <section className="section-block">
        <div className="section-title">
          <h2>关联产品</h2>
          <Link to={`/products?vendorId=${vendor.id}`}>查看全部</Link>
        </div>
        <div className="product-grid">
          {products.map((product) => (
            <article className="product-card" key={product.id}>
              <div className="product-image" style={{ backgroundImage: `url(${product.image || "https://dummyimage.com/480x300/eaf3ff/1f2a3d&text=Parts"})` }} />
              <div>
                <h3>{product.name}</h3>
                <p>{product.compatibleModels || product.description}</p>
              </div>
            </article>
          ))}
        </div>
      </section>
      {vendor.websiteUrl && (
        <a className="primary-btn" href={vendor.websiteUrl} rel="noreferrer" target="_blank">
          进入厂商官网
        </a>
      )}
    </PageFrame>
  );
}
