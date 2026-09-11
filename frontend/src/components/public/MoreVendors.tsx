import { Link } from "react-router-dom";
import type { HomeSections, Vendor } from "../../types/api";
import { VendorCard } from "./VendorCard";

export function MoreVendors({ vendors, homeSections }: { vendors: Vendor[]; homeSections?: HomeSections }) {
  return (
    <section className="section-block home-scroll-section home-directory">
      <div className="section-title filters">
        <div><span className="section-eyebrow">SUPPLIER DIRECTORY</span><h2>{homeSections?.moreTitle || "更多厂商"}</h2></div>
        <div>
          <Link to={homeSections?.moreLink || "/vendors"}>全部地区</Link>
          <Link to="/vendors?sort=recommended">推荐优先</Link>
          <Link to="/vendors?sort=latest">最新入驻</Link>
          <Link to="/products">主营品类</Link>
        </div>
      </div>
      {vendors.length ? (
        <div className="vendor-grid home-card-rail">
          {vendors.map((vendor) => (
            <VendorCard key={vendor.id} compact vendor={vendor} />
          ))}
        </div>
      ) : (
        <div className="state-panel">暂无更多厂商</div>
      )}
    </section>
  );
}
