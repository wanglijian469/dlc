import { ArrowRight, Wrench } from "lucide-react";
import { Link } from "react-router-dom";
import type { HomeModule, Vendor } from "../../types/api";
import { VendorCard } from "./VendorCard";

export function ProcessingSection({ module, vendors }: { module: HomeModule; vendors: Vendor[] }) {
  return <section className="section-block marketplace-section processing-showcase">
    <div className="processing-visual">
      <img alt="农机配件加工服务" loading="lazy" src={module.image || "/images/industry/processing-service.jpg"} />
      <div><span><Wrench size={17} />按加工能力找厂商</span><h2>{module.title || "加工服务厂商"}</h2><p>{module.subtitle || "查找支持来图、来样、数控加工与批量代工的生产企业。"}</p><Link className="primary-btn warm-btn" to={module.path || "/service"}>查看加工服务厂商 <ArrowRight size={16} /></Link></div>
    </div>
    {vendors.length > 0 && <div className="vendor-grid processing-vendor-grid home-card-rail">{vendors.slice(0, module.limit || 4).map((item) => <VendorCard compact key={item.id} vendor={item} />)}</div>}
  </section>;
}
