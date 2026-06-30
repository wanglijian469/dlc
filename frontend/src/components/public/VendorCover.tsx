import type { ReactNode } from "react";
import { Cable, CircleDot, Cog, Droplets, Link as LinkIcon, Package, Sprout, Wheat, Wrench } from "lucide-react";
import type { Vendor } from "../../types/api";

type CoverKind = "gear" | "chain" | "bearing" | "hydraulic" | "electrical" | "harvester" | "seeding" | "general";

const coverIcons = {
  gear: Cog,
  chain: LinkIcon,
  bearing: CircleDot,
  hydraulic: Droplets,
  electrical: Cable,
  harvester: Wheat,
  seeding: Sprout,
  general: Package,
};

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
  const validImage = getValidCoverImage(vendor.coverImage);
  const sequence = String(number || vendor.sortOrder || vendor.id).padStart(2, "0").slice(-2);
  const badge = vendor.isVerified ? "平台认证" : vendor.tags?.[0]?.name || "源头厂商";

  if (validImage) {
    return (
      <div className={`vendor-cover vendor-cover-image vendor-cover-${variant}`} style={{ backgroundImage: `url(${validImage})` }}>
        <CoverBadges sequence={sequence} badge={badge} />
        {children}
      </div>
    );
  }

  const kind = getVendorCoverKind(vendor.mainProducts || vendor.description || "");
  const Icon = coverIcons[kind] || Wrench;
  return (
    <div className={`vendor-cover vendor-cover-default vendor-cover-${variant} vendor-cover-${kind}`}>
      <div className="vendor-cover-pattern" />
      <CoverBadges sequence={sequence} badge={badge} />
      <Icon className="vendor-cover-main-icon" aria-hidden="true" size={variant === "detail" ? 76 : 48} strokeWidth={1.7} />
      {children}
    </div>
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

export function getValidCoverImage(src?: string) {
  if (!src) return "";
  const value = src.trim();
  const normalized = value.toLowerCase();
  if (normalized.includes("dummyimage.com") && normalized.includes("factory")) return "";
  return value;
}

function getVendorCoverKind(text: string): CoverKind {
  if (/链条|链轮|链/.test(text)) return "chain";
  if (/轴承|轴套/.test(text)) return "bearing";
  if (/液压|油缸|油泵|油管/.test(text)) return "hydraulic";
  if (/电气|照明|线束|传感器|蓄电池|仪表/.test(text)) return "electrical";
  if (/收获|割台|割刀|脱粒|拨禾/.test(text)) return "harvester";
  if (/播种|施肥|排种|开沟/.test(text)) return "seeding";
  if (/变速|齿轮|减速|差速/.test(text)) return "gear";
  return "general";
}
