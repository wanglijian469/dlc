import { ExternalLink } from "lucide-react";
import { useState } from "react";
import { Link } from "react-router-dom";
import type { Vendor } from "../../types/api";
import { vendorPath } from "../../utils/vendorPath";

interface VendorCardProps {
  vendor: Vendor;
  compact?: boolean;
  directory?: boolean;
}

export function VendorCard({ vendor, compact = false, directory = false }: VendorCardProps) {
  const region = [vendor.province, vendor.city].filter(Boolean).join(" · ");
  const [failedCover, setFailedCover] = useState<string>();
  const [failedLogo, setFailedLogo] = useState<string>();
  return (
    <article className={`${compact ? "vendor-card compact" : "vendor-card"} ${directory ? "directory-card" : ""}`}>
      <Link className="vendor-card-visual" to={vendorPath(vendor)}>{vendor.coverImage && vendor.coverImage !== failedCover ? <img loading="lazy" src={vendor.coverImage} alt={vendor.name + " 企业展示"} onError={() => setFailedCover(vendor.coverImage)} /> : <span title={vendor.name}>{vendor.shortName || vendor.name} · 企业展厅</span>}</Link>
      <div className="vendor-card-header">
        {vendor.logo && vendor.logo !== failedLogo && <img className="vendor-card-logo" src={vendor.logo} alt={vendor.name + " Logo"} onError={() => setFailedLogo(vendor.logo)} />}
        <div className="vendor-title-block">
          <h3 title={vendor.name}>{vendor.name}</h3>
          <div className="tag-row">

            {vendor.isRecommended && <span className="tag-green">推荐厂商</span>}
          </div>
        </div>
      </div>
      <div className="vendor-body">
        <p className="vendor-line"><strong>地区：</strong>{region || "全国供应"}</p>
        <p className="vendor-line"><strong>主营：</strong>{vendor.mainProducts || "农机配件"}</p>
        <p className="vendor-line"><strong>优势：</strong>{vendor.serviceAdvantages || "企业资料待完善"}</p>
        <div className="tag-row service-tags">{vendor.tags?.slice(0, compact ? 2 : 3).map((tag) => <span className="tag-green" key={tag.id}>{tag.name}</span>)}</div>
      </div>
      <div className="card-actions vendor-card-actions">
        <Link className="primary-btn small" to={vendorPath(vendor)}>进入展厅</Link>
        <Link className="outline-btn small" to={vendorPath(vendor) + "#showroom-products"}>查看产品</Link>
        {vendor.websiteUrl
          ? <a aria-label={`${vendor.name} 访问官网`} className="outline-btn small website-action" href={vendor.websiteUrl} rel="noreferrer" target="_blank"><ExternalLink size={14} />访问官网</a>
          : null}
      </div>
    </article>
  );
}
