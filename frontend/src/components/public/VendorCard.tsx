import { Link } from "react-router-dom";
import type { Vendor } from "../../types/api";
import { getVendorEntryTarget } from "../../utils/navigation";
import { VendorCover } from "./VendorCover";

const logoFallback = "https://dummyimage.com/96x64/fff/0b5fea&text=DL";

export function VendorCard({ vendor, compact = false }: { vendor: Vendor; compact?: boolean }) {
  const target = getVendorEntryTarget(vendor);
  const region = [vendor.province, vendor.city].filter(Boolean).join(" ");
  return (
    <article className={compact ? "vendor-card compact" : "vendor-card"}>
      <VendorCover vendor={vendor}>
        <img alt="" className="vendor-logo" src={vendor.logo || logoFallback} onError={(event) => { event.currentTarget.src = logoFallback; }} />
      </VendorCover>
      <div className="vendor-body">
        <h3>{vendor.name}</h3>
        <p className="vendor-meta">
          {region || "全国供应"}
          {vendor.isVerified ? " · 平台认证" : ""}
        </p>
        <p>主营：{vendor.mainProducts || "农机配件"}</p>
        <p>优势：{vendor.serviceAdvantages || "质量稳定，发货及时"}</p>
        <div className="tag-row">
          {vendor.tags?.slice(0, 3).map((tag) => (
            <span key={tag.id}>{tag.name}</span>
          ))}
        </div>
        <div className="card-actions">
          {target.type === "external" ? (
            <a className="primary-btn small" href={target.href} rel="noreferrer" target="_blank">
              进入厂商
            </a>
          ) : (
            <Link className="primary-btn small" to={target.href}>
              进入厂商
            </Link>
          )}
          {!compact && (
            <Link className="outline-btn small" to={`/vendors/${vendor.id}`}>
              查看详情
            </Link>
          )}
        </div>
      </div>
    </article>
  );
}
