import { Link } from "react-router-dom";
import type { Product } from "../../types/api";

const invalidImagePattern = /dummyimage\.com.*(?:Parts|Product)/i;

function isRealImage(url?: string) {
  return Boolean(url && !invalidImagePattern.test(url));
}

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
  const text = `${product.name || ""} ${product.description || ""} ${product.compatibleModels || ""}`;
  if (/液压|油缸|油泵|油管/.test(text)) return "液压系统配件";
  if (/发动机|滤芯|散热器|喷油/.test(text)) return "动力发动机配件";
  if (/制动|刹车|换挡|离合/.test(text)) return "制动换挡配件";
  if (/电气|照明|线束|传感器/.test(text)) return "电气照明配件";
  if (/收获|割台|割刀|搅龙/.test(text)) return "收获割台配件";
  if (/播种|施肥|排种|开沟/.test(text)) return "播种施肥配件";
  if (/底盘|履带|轮毂|支重轮/.test(text)) return "行走底盘配件";
  if (/齿轮|链条|轴承|变速箱|传动/.test(text)) return "传动配件";
  if (/刀片|皮带|油封|螺栓|易损/.test(text)) return "农机易损件";
  return "农机配件";
}

export function ProductCard({ product, compact = false }: { product: Product; compact?: boolean }) {
  const vendor = product.vendor;
  const region = [vendor?.province, vendor?.city].filter(Boolean).join(" · ");
  const showImage = isRealImage(product.image);

  return (
    <article className={compact ? "product-card compact" : "product-card"}>
      {showImage && <img alt={product.name} className="product-thumb" src={product.image} />}
      <div className="product-card-main">
        <div className="product-card-title">
          <h3>{product.name}</h3>
          <div className="tag-row">
            {product.isHot && <span className="tag-orange">热销</span>}
            {product.isRecommended && <span className="tag-blue">推荐</span>}
            <span className="tag-green">支持定制</span>
          </div>
        </div>
        <p className="product-line">型号：{product.description || "按需匹配"}</p>
        <p className="product-line">适配机型：{product.compatibleModels || "通用农机配件"}</p>
        <p className="product-line">分类：{categoryName(product)}</p>
        <p className="product-line">供应商：{vendor?.name || "平台供应商"}</p>
        {region && <p className="product-line">地区：{region}</p>}
        <p className="product-line">价格：面议 / 批量报价</p>
        <div className="card-actions">
          <Link className="outline-btn small" to={`/products?keyword=${encodeURIComponent(product.name)}`}>
            查看详情
          </Link>
          {vendor?.id ? (
            <Link className="primary-btn small" to={`/vendors/${vendor.id}`}>
              联系供应商
            </Link>
          ) : (
            <Link className="primary-btn small" to="/vendors">
              联系供应商
            </Link>
          )}
        </div>
      </div>
    </article>
  );
}
