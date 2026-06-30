import { Link } from "react-router-dom";
import type { Vendor } from "../../types/api";
import { getVendorEntryTarget } from "../../utils/navigation";

const logoFallback = "https://dummyimage.com/96x96/ffffff/0b5fea&text=DL";

interface VendorCardProps {
  vendor: Vendor;
  compact?: boolean;
  directory?: boolean;
}

export function VendorCard({ vendor, compact = false, directory = false }: VendorCardProps) {
  const target = getVendorEntryTarget(vendor);
  const region = [vendor.province, vendor.city].filter(Boolean).join(" · ");
  const mainProducts = vendor.mainProducts || "农机配件";
  const advantages = vendor.serviceAdvantages || "源头工厂、支持定制、现货供应、交付稳定";

  return (
    <article className={compact ? "vendor-card compact" : "vendor-card"}>
      <div className="vendor-card-header">
        <img
          alt={vendor.name}
          className="vendor-logo"
          src={vendor.logo || logoFallback}
          onError={(event) => {
            event.currentTarget.src = logoFallback;
          }}
        />
        <div className="vendor-title-block">
          <h3>{vendor.name}</h3>
          <div className="tag-row">
            {vendor.isVerified && <span className="tag-blue">平台认证</span>}
            <span className="tag-green">源头厂商</span>
          </div>
        </div>
      </div>

      <div className="vendor-body">
        <p className="vendor-line">地区：{region || "全国供应"}</p>
        <p className="vendor-line">主营：{mainProducts}</p>
        <p className="vendor-line">优势：{advantages}</p>
        <div className="tag-row service-tags">
          {vendor.tags?.slice(0, compact ? 2 : 4).map((tag) => (
            <span className="tag-green" key={tag.id}>
              {tag.name}
            </span>
          ))}
        </div>
      </div>

      <div className="card-actions">
        {target.type === "external" ? (
          <a className="primary-btn small" href={target.href} rel="noreferrer" target="_blank">
            查看厂商
          </a>
        ) : (
          <Link className="primary-btn small" to={target.href}>
            查看厂商
          </Link>
        )}
        <Link className="outline-btn small" to={`/products?vendorId=${vendor.id}`}>
          查看产品
        </Link>
        {directory && (
          <Link className="outline-btn small" to={`/vendors/${vendor.id}`}>
            在线询价
          </Link>
        )}
      </div>
    </article>
  );
}
