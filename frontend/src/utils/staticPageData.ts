import type { LayoutConfig, Product, ProductSupplier, Vendor, VendorPost } from "../types/api";

export interface StaticPageData {
  kind: "vendor" | "product" | "supplier";
  slug: string;
  layout: LayoutConfig;
  vendor?: Vendor;
  products?: Product[];
  product?: Product;
  suppliers?: ProductSupplier[];
 supplier?: ProductSupplier;
  related?: Product[];
  posts?: VendorPost[];
}

export function getStaticPageData(kind?: StaticPageData["kind"], slug?: string) {
  const node = document.getElementById("static-page-data");
  if (!node?.textContent) return null;
  try {
    const data = JSON.parse(node.textContent) as StaticPageData;
    if (kind && data.kind !== kind) return null;
    if (slug && data.slug !== slug) return null;
    return data;
  } catch {
    return null;
  }
}
