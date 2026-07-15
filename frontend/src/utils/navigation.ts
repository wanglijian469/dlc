import type { Menu } from "../types/api";

export function getSearchPath(keyword: string): string | null {
  const value = keyword.trim();
  if (!value) return null;
  return `/search?keyword=${encodeURIComponent(value)}`;
}

/** Returns the current label configured for a public navigation entry. */
export function getMenuLabel(menus: Menu[], path: string, fallback: string): string {
  return menus.find((menu) => menu.path === path)?.name.trim() || fallback;
}

export function getVendorEntryTarget(vendor: { id: number; websiteUrl?: string | null }) {
  const websiteUrl = vendor.websiteUrl?.trim();
  if (websiteUrl) {
    return { type: "external" as const, href: websiteUrl };
  }
  return { type: "internal" as const, href: `/vendors/${vendor.id}` };
}
