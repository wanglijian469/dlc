import { Link } from "react-router-dom";
import type { Product } from "../../types/api";
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

export function ProductCard({ product, compact = false }: { product: Product; compact?: boolean }) {
  const vendor = product.vendor;
  const region = [vendor?.province, vendor?.city].filter(Boolean).join(" · ");
  return (
    <article className={compact ? "product-card compact" : "product-card"}>
      <ProductCover product={product} />
      <div className="product-card-main">
        <div className="product-card-title">
          <h3>{product.name}</h3>
          <div className="tag-row">
            {product.isHot && <span className="tag-orange">热销</span>}
            {product.isRecommended && <span className="tag-blue">推荐</span>}
          </div>
        </div>
        <p className="product-line">型号：{product.description || "按需匹配"}</p>
        <p className="product-line">适配机型：{product.compatibleModels || "通用农机配件"}</p>
        <p className="product-line">分类：{categoryName(product)}</p>
        <p className="product-line">供应商：{vendor?.name || "平台供应商"}</p>
        {region && <p className="product-line">地区：{region}</p>}
        <p className="product-line">价格：{product.priceNote || "面议 / 批量报价"}</p>
        <div className="card-actions">
          <Link className="outline-btn small" to={`/products/${product.id}`}>厂商信息</Link>
          {vendor?.id ? (
            <Link className="primary-btn small" to={`/vendors/${vendor.id}`}>联系供应商</Link>
          ) : (
            <Link className="primary-btn small" to="/vendors">联系供应商</Link>
          )}
        </div>
      </div>
    </article>
  );
}
