import type { Vendor } from "../../types/api";
import { VendorCard } from "./VendorCard";

export function MoreVendors({ vendors }: { vendors: Vendor[] }) {
  return (
    <section className="section-block">
      <div className="section-title filters">
        <h2>更多厂商</h2>
        <div>
          <button>所在地区</button>
          <button>主营品类</button>
          <button>全部服务</button>
          <button>综合排序</button>
        </div>
      </div>
      <div className="vendor-grid">
        {vendors.map((vendor) => (
          <VendorCard key={vendor.id} compact vendor={vendor} />
        ))}
      </div>
    </section>
  );
}
