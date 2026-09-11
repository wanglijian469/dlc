import { Link } from "react-router-dom";
import type { HomeSections, Vendor } from "../../types/api";
import { VendorCard } from "./VendorCard";

export function RecommendedVendors({ vendors, homeSections }: { vendors: Vendor[]; homeSections?: HomeSections }) {
  return (
    <section className="section-block home-scroll-section home-recommended">
      <div className="section-title">
        <div><span className="section-eyebrow">SELECTED SUPPLIERS</span><h2>{homeSections?.recommendedTitle || "推荐厂商"}</h2></div>
        <Link to={homeSections?.recommendedLink || "/vendors"}>浏览全部厂商 <span aria-hidden="true">↗</span></Link>
      </div>
      <div className="vendor-grid recommended home-card-rail">
        {vendors.map((vendor) => (
          <VendorCard key={vendor.id} vendor={vendor} />
        ))}
      </div>
    </section>
  );
}
