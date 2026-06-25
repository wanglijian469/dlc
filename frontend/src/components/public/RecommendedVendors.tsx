import { Link } from "react-router-dom";
import type { Vendor } from "../../types/api";
import { VendorCard } from "./VendorCard";

export function RecommendedVendors({ vendors }: { vendors: Vendor[] }) {
  return (
    <section className="section-block">
      <div className="section-title">
        <h2>推荐厂商</h2>
        <Link to="/vendors">更多</Link>
      </div>
      <div className="vendor-grid recommended">
        {vendors.map((vendor) => (
          <VendorCard key={vendor.id} vendor={vendor} />
        ))}
      </div>
    </section>
  );
}
