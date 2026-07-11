import { ExternalLink } from "lucide-react";
import { Link } from "react-router-dom";
import type { Vendor } from "../../types/api";

interface VendorCardProps {
  vendor: Vendor;
  compact?: boolean;
  directory?: boolean;
}

export function VendorCard({ vendor, compact = false, directory = false }: VendorCardProps) {
  const region = [vendor.province, vendor.city].filter(Boolean).join(" · ");
  return (
    <article className={`${compact ? "vendor-card compact" : "vendor-card"} ${directory ? "directory-card" : ""}`}>
      <div className="vendor-card-header">
        <div className="vendor-title-block">
          <h3 title={vendor.name}>{vendor.name}</h3>
          <div className="tag-row">
            {vendor.isVerified && <span className="tag-blue">平台认证</span>}
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
        <Link className="primary-btn small" to={`/vendors/${vendor.id}`}>查看详情</Link>
        <Link className="outline-btn small" to={`/products?vendorId=${vendor.id}`}>查看产品</Link>
        {vendor.websiteUrl
          ? <a aria-label={`${vendor.name} 访问官网`} className="outline-btn small website-action" href={vendor.websiteUrl} rel="noreferrer" target="_blank"><ExternalLink size={14} />访问官网</a>
          : <span aria-disabled="true" className="outline-btn small website-action disabled">暂无官网</span>}
      </div>
    </article>
  );
}
