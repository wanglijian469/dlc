import type { ReactNode } from "react";
import type { Vendor } from "../../types/api";
import { getVendorIndustryKind, IndustryCover } from "./IndustryCover";

export function VendorCover({
  vendor,
  number,
  variant = "card",
  children,
}: {
  vendor: Vendor;
  number?: number;
  variant?: "card" | "detail";
  children?: ReactNode;
}) {
  const sequence = String(number || vendor.sortOrder || vendor.id).padStart(2, "0").slice(-2);
  const badge = vendor.isVerified ? "平台认证" : vendor.tags?.[0]?.name || "源头厂商";

  return (
    <IndustryCover
      className={`vendor-cover vendor-cover-${variant}`}
      fallbackClassName="vendor-cover-default"
      iconSize={variant === "detail" ? 76 : 48}
      image={vendor.coverImage}
      kind={getVendorIndustryKind(vendor)}
    >
      <CoverBadges sequence={sequence} badge={badge} />
      {children}
    </IndustryCover>
  );
}

function CoverBadges({ sequence, badge }: { sequence: string; badge: string }) {
  return (
    <>
      <span className="vendor-cover-index">{sequence}</span>
      <span className="vendor-cover-badge">{badge}</span>
    </>
  );
}
