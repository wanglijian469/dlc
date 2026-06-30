import { Cable, Cog, Disc3, Droplets, Gauge, Package, Sprout, Truck, Wheat, Wrench, type LucideIcon } from "lucide-react";
import type { ReactNode } from "react";
import type { Product, Vendor } from "../../types/api";

export type IndustryKind =
  | "wearing"
  | "transmission"
  | "chassis"
  | "hydraulic"
  | "engine"
  | "brake"
  | "electrical"
  | "harvester"
  | "seeding"
  | "general";

const industryIcons: Record<IndustryKind, LucideIcon> = {
  wearing: Wrench,
  transmission: Cog,
  chassis: Truck,
  hydraulic: Droplets,
  engine: Gauge,
  brake: Disc3,
  electrical: Cable,
  harvester: Wheat,
  seeding: Sprout,
  general: Package,
};

const categoryIdMap: Record<number, IndustryKind> = {
  1: "wearing",
  2: "transmission",
  3: "transmission",
  4: "chassis",
  5: "hydraulic",
  6: "engine",
  7: "brake",
  8: "electrical",
  9: "harvester",
  10: "seeding",
};

export function IndustryCover({
  image,
  kind,
  className,
  fallbackClassName = "",
  iconSize = 48,
  children,
}: {
  image?: string;
  kind: IndustryKind;
  className: string;
  fallbackClassName?: string;
  iconSize?: number;
  children?: ReactNode;
}) {
  const validImage = getValidCoverImage(image);
  if (validImage) {
    return (
      <div className={`${className} industry-cover-image`} style={{ backgroundImage: `url(${validImage})` }}>
        {children}
      </div>
    );
  }

  const Icon = industryIcons[kind] || Package;
  return (
    <div className={`${className} ${fallbackClassName} industry-cover-default industry-cover-${kind}`}>
      <div className="industry-cover-pattern" />
      <Icon className="industry-cover-main-icon" aria-hidden="true" size={iconSize} strokeWidth={1.7} />
      {children}
    </div>
  );
}

export function getValidCoverImage(src?: string) {
  if (!src) return "";
  const value = src.trim();
  const normalized = value.toLowerCase();
  if (normalized.includes("dummyimage.com") && (normalized.includes("factory") || normalized.includes("parts"))) return "";
  return value;
}

export function getVendorIndustryKind(vendor: Vendor): IndustryKind {
  return getIndustryKindFromText(`${vendor.mainProducts || ""} ${vendor.description || ""}`);
}

export function getProductIndustryKind(product: Product): IndustryKind {
  if (product.categoryId && categoryIdMap[Number(product.categoryId)]) return categoryIdMap[Number(product.categoryId)];
  return getIndustryKindFromText(`${product.category?.name || ""} ${product.name || ""} ${product.description || ""}`);
}

export function getIndustryKindFromText(text: string): IndustryKind {
  if (/刀片|刀架|滤芯|皮带张紧|张紧轮|密封圈|油封|螺栓|销轴|易损/.test(text)) return "wearing";
  if (/履带|支重轮|托链轮|驱动轮|引导轮|轮胎|轮毂|底盘/.test(text)) return "chassis";
  if (/液压|油缸|油泵|油管|多路阀|分配阀|接头/.test(text)) return "hydraulic";
  if (/发动机|喷油|水泵|散热器|起动机|发电机|活塞|缸套/.test(text)) return "engine";
  if (/制动|刹车|换挡|离合拉杆|拉线|踏板|操纵阀/.test(text)) return "brake";
  if (/电气|照明|线束|传感器|蓄电池|仪表|开关|灯具/.test(text)) return "electrical";
  if (/收获|割台|割刀|护刃|拨禾|搅龙|脱粒/.test(text)) return "harvester";
  if (/播种|施肥|排种|开沟|镇压|播种盘/.test(text)) return "seeding";
  if (/传动|变速|齿轮|链条|链轮|轴承|轴套|离合器|差速|减速|半轴/.test(text)) return "transmission";
  return "general";
}
