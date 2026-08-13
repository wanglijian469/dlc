import type { Menu } from "../types/api";

const requiredTopMenus: Menu[] = [
  { id: -1, name: "首页", path: "/", icon: "home" },
  { id: -2, name: "厂商资源", path: "/vendors", icon: "factory" },
  { id: -3, name: "配件货源", path: "/products", icon: "grid" },
  { id: -4, name: "加工服务", path: "/service", icon: "settings" },
  { id: -5, name: "供求信息", path: "/purchase", icon: "clipboard" },
];

/** Keeps required public destinations present even for stale static payloads. */
export function normalizeTopMenus(menus: Menu[] = []): Menu[] {
  const byPath = new Map(menus.map((menu) => [menu.path, menu]));
  const requiredPaths = new Set(requiredTopMenus.map((menu) => menu.path));
  return [
    ...requiredTopMenus.map((required) => {
      const existing = byPath.get(required.path);
      return { ...existing, ...required, id: existing?.id ?? required.id };
    }),
    ...menus.filter((menu) => !requiredPaths.has(menu.path)),
  ];
}

export function getSearchPath(keyword: string): string | null {
  const value = keyword.trim();
  if (!value) return null;
  return `/search?keyword=${encodeURIComponent(value)}`;
}

/** Returns the current label configured for a public navigation entry. */
export function getMenuLabel(menus: Menu[], path: string, fallback: string): string {
  return menus.find((menu) => menu.path === path)?.name.trim() || fallback;
}

export function getVendorEntryTarget(vendor: { id: number; slug?: string | null; websiteUrl?: string | null }) {
  const websiteUrl = vendor.websiteUrl?.trim();
  if (websiteUrl) {
    return { type: "external" as const, href: websiteUrl };
  }
  return { type: "internal" as const, href: `/v/${vendor.slug || vendor.id}` };
}
