import { Link } from "react-router-dom";
import type { Vendor } from "../../types/api";
import { VendorCard } from "./VendorCard";

export function MoreVendors({ vendors }: { vendors: Vendor[] }) {
  return (
    <section className="section-block">
      <div className="section-title filters">
        <h2>更多厂商</h2>
        <div>
          <Link to="/vendors">全部地区</Link>
          <Link to="/vendors?sort=recommended">推荐优先</Link>
          <Link to="/vendors?sort=latest">最新入驻</Link>
          <Link to="/products">主营品类</Link>
        </div>
      </div>
      {vendors.length ? (
        <div className="vendor-grid">
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
