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
  const badge = vendor.isVerified ? "平台认证" : vendor.tags?.[0]?.name;

  return (
    <IndustryCover
      className={`vendor-cover vendor-cover-${variant}`}
      fallbackClassName="vendor-cover-default"
      iconSize={variant === "detail" ? 76 : 48}
      image={vendor.coverImage}
      kind={getVendorIndustryKind(vendor)}
    >
      <CoverBadges badge={badge} />
      {children}
    </IndustryCover>
  );
}

function CoverBadges({ badge }: { badge?: string }) {
  return (
    <>
      {badge && <span className="vendor-cover-badge">{badge}</span>}
    </>
  );
}
