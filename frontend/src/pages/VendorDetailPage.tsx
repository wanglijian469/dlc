import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { getVendor } from "../api/public";
import type { Vendor } from "../types/api";

export function VendorDetailPage() {
  const { id = "" } = useParams();
  const [vendor, setVendor] = useState<Vendor | null>(null);
  useEffect(() => {
    getVendor(id).then(setVendor);
  }, [id]);
  if (!vendor) return <div className="state-page">加载中...</div>;
  return (
    <main className="plain-page">
      <Link to="/">返回首页</Link>
      <h1>{vendor.name}</h1>
      <p>{vendor.description}</p>
      <p>主营产品：{vendor.mainProducts}</p>
      <p>服务优势：{vendor.serviceAdvantages}</p>
      <p>联系人：{vendor.contactName} {vendor.phone}</p>
      <div className="tag-row">
        {vendor.tags?.map((tag) => (
          <span key={tag.id}>{tag.name}</span>
        ))}
      </div>
      {vendor.websiteUrl && (
        <a className="primary-btn" href={vendor.websiteUrl} rel="noreferrer" target="_blank">
          进入厂商官网
        </a>
      )}
    </main>
  );
}
