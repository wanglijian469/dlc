import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { getProduct } from "../api/public";
import { PageFrame } from "../components/public/PageFrame";
import { ErrorState, LoadingState } from "../components/public/StateViews";
import type { Product } from "../types/api";

export function ProductDetailPage() {
  const { id = "" } = useParams();
  const [product, setProduct] = useState<Product | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const load = () => {
    setLoading(true);
    setError("");
    getProduct(id)
      .then(setProduct)
      .catch(() => setError("产品详情加载失败，请稍后重试"))
      .finally(() => setLoading(false));
  };

  useEffect(load, [id]);

  if (loading) return <PageFrame title="产品详情"><LoadingState /></PageFrame>;
  if (error || !product) return <PageFrame title="产品详情"><ErrorState text={error || "产品不存在"} onRetry={load} /></PageFrame>;

  const inquiryText = product.inquiryText || "联系供应商";
  const inquiryPath = product.inquiryPath || (product.vendor?.id ? `/vendors/${product.vendor.id}` : "/vendors");

  return (
    <PageFrame title={product.name} subtitle={product.category?.name || "农机配件产品"}>
      <section className="vendor-profile-hero product-detail-hero">
        <div className="vendor-profile-main">
          <div className="vendor-profile-lines">
            <p>适配机型：{product.compatibleModels || "通用农机配件"}</p>
            <p>所属分类：{product.category?.name || "农机配件"}</p>
            <p>供应厂商：{product.vendor?.name || "平台供应商"}</p>
            <p>价格：{product.priceNote || "面议 / 批量报价"}</p>
          </div>
          {product.detailContent && <p className="product-detail-copy">{product.detailContent}</p>}
        </div>
        <aside className="vendor-website-card">
          <span>采购咨询</span>
          <h3>{product.priceNote || "面议 / 批量报价"}</h3>
          <p>{product.description || "可联系供应商确认型号、库存、交期和批量报价。"}</p>
          <Link className="primary-btn vendor-website-main" to={inquiryPath}>{inquiryText}</Link>
        </aside>
      </section>

      {product.gallery?.length ? (
        <section className="section-block">
          <div className="section-title"><h2>产品图片</h2></div>
          <div className="product-gallery">
            {product.gallery.map((src) => <img alt={product.name} key={src} src={src} />)}
          </div>
        </section>
      ) : null}

      {product.specs?.length ? (
        <section className="vendor-profile-card">
          <h3>规格参数</h3>
          <div className="vendor-detail-list">
            {product.specs.map((spec) => <p key={spec.name}>{spec.name}：{spec.value}</p>)}
          </div>
        </section>
      ) : null}
    </PageFrame>
  );
}
