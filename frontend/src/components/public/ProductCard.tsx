import { Link } from "react-router-dom";
import type { Product } from "../../types/api";
import { vendorPath } from "../../utils/vendorPath";
import { ProductCover } from "./ProductCover";

function categoryName(product: Product) {
  if (product.category?.name) return product.category.name;
  const categoryMap: Record<number, string> = {
    1: "农机易损件",
    2: "传动配件",
    3: "传动配件",
    4: "行走底盘配件",
    5: "液压系统配件",
    6: "动力发动机配件",
    7: "制动换挡配件",
    8: "电气照明配件",
    9: "收获割台配件",
    10: "播种施肥配件",
  };
  if (product.categoryId && categoryMap[Number(product.categoryId)]) return categoryMap[Number(product.categoryId)];
  return "农机配件";
}

function supplierPrice(product: Product) {
  const supplier = product.supplier;
  if (!supplier) return product.priceNote || "面议 / 批量报价";
  if (supplier.priceValidUntil && new Date(supplier.priceValidUntil).getTime() < Date.now()) return "价格已过期，请询价";
  if (supplier.negotiable || !supplier.unitPriceCents) return supplier.priceNote || "面议";
  return "¥" + (supplier.unitPriceCents / 100).toFixed(2) + (supplier.priceUnit ? " / " + supplier.priceUnit : "");
}

export function ProductCard({ product, compact = false }: { product: Product; compact?: boolean }) {
  const supplier = product.supplier;
  const vendor = supplier?.vendor || product.vendor;
  const detailPath = supplier && vendor?.slug ? `/v/${vendor.slug}/products/${supplier.id}` : `/products/${product.slug || product.id}`;
  const directoryMode = !supplier && !product.vendor;
  const regions = product.supplierRegions?.filter(Boolean).join(" · ");
  return (
    <article className={compact ? "product-card compact" : "product-card"}>
      <Link className="product-cover-link" aria-label={"查看产品：" + (supplier?.vendorProductName || product.name)} to={detailPath}><ProductCover product={supplier ? { ...product, image: supplier.image } : product} /></Link>
      <div className="product-card-main">
        <div className="product-card-title">
          <h3><Link to={detailPath}>{supplier?.vendorProductName || product.name}</Link></h3>
          <div className="tag-row">
            {product.isHot && <span className="tag-orange">热销</span>}
            {product.isRecommended && <span className="tag-blue">推荐</span>}
          </div>
        </div>
        <p className="product-line">型号：{supplier?.vendorModel || "按供应商型号匹配"}</p>
        <p className="product-line">适配机型：{supplier?.compatibleModels || product.compatibleModels || "通用农机配件"}</p>
        <p className="product-line">分类：{categoryName(product)}</p>
        {directoryMode ? <p className="product-line supplier-count">支持供应商：{product.supplierCount || 0} 家</p> : <p className="product-line">供应厂商：{vendor?.name || "供应厂商"}</p>}
        {directoryMode && regions && <p className="product-line">供应地区：{regions}</p>}
        <p className="product-line product-live-price">价格：{supplierPrice(product)}</p>
        {supplier?.priceUpdatedAt && <p className="product-price-time">更新于 {new Date(supplier.priceUpdatedAt).toLocaleString()}</p>}
        <div className="card-actions">
          <Link className="outline-btn small" to={detailPath}>产品详情</Link>
          <Link className="primary-btn small" to={directoryMode || !vendor?.id ? `/products/${product.id}#suppliers` : vendorPath(vendor)}>供应厂商</Link>
        </div>
      </div>
    </article>
  );
}
