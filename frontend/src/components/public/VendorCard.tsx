import { Link } from "react-router-dom";
import type { Vendor } from "../../types/api";
import { getVendorEntryTarget } from "../../utils/navigation";

export function VendorCard({ vendor, compact = false }: { vendor: Vendor; compact?: boolean }) {
  const target = getVendorEntryTarget(vendor);
  const image = vendor.coverImage || "https://dummyimage.com/600x320/eaf3ff/1f2a3d&text=Factory";
  return (
    <article className={compact ? "vendor-card compact" : "vendor-card"}>
      <div className="vendor-cover" style={{ backgroundImage: `url(${image})` }}>
        <img alt="" className="vendor-logo" src={vendor.logo || "https://dummyimage.com/96x64/fff/0b5fea&text=DL"} />
      </div>
      <div className="vendor-body">
        <h3>{vendor.name}</h3>
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
