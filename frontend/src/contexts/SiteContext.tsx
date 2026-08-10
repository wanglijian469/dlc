import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from "react";
import { getLayoutConfig, getVendorCategories } from "../api/public";
import type { LayoutConfig, Menu, VendorCategory } from "../types/api";
import { getStaticPageData } from "../utils/staticPageData";

const topMenus: Menu[] = [{ id: -1, name: "首页", path: "/" }, { id: -3, name: "厂商资源", path: "/vendors" }, { id: -2, name: "配件货源", path: "/products" }, { id: -4, name: "加工服务", path: "/service" }, { id: -5, name: "采购信息", path: "/purchase" }];
const mobileBottomMenus: Menu[] = [{ id: -1, name: "首页", path: "/", icon: "home" }, { id: -3, name: "厂商", path: "/vendors", icon: "factory" }, { id: -2, name: "配件", path: "/products", icon: "grid" }, { id: -4, name: "加工服务", path: "/service", icon: "settings" }, { id: -6, name: "厂商", path: "/account/login", icon: "user" }];
const fallback: LayoutConfig = { siteMeta: { siteName: "大陆农机配件", brandMark: "农", submitVendorText: "厂商入驻", adminLoginText: "后台登录", mobileBrandName: "大陆农机配件", mobileBrandMark: "农", copyrightOwner: "大陆农机配件", copyrightYear: String(new Date().getFullYear()), filingNumber: "待运营方配置" }, theme: { primaryColor: "#1559c7", accentColor: "#0d8b6f" }, topMenus, sidebarMenus: [], auxiliaryMenus: [], mobileMenus: [], mobileBottomMenus, version: "fallback" };
type Value = { layout: LayoutConfig; vendorCategories: VendorCategory[]; loading: boolean; reload: () => void; provided: boolean };
const Context = createContext<Value>({ layout: fallback, vendorCategories: [], loading: true, reload: () => undefined, provided: false });

export function SiteProvider({ children }: { children: ReactNode }) {
  const staticLayout = getStaticPageData()?.layout;
  const [layout, setLayout] = useState(staticLayout || fallback);
  const [vendorCategories, setVendorCategories] = useState<VendorCategory[]>([]);
  const [loading, setLoading] = useState(!staticLayout);
  const [nonce, setNonce] = useState(0);
  useEffect(() => {
    if (staticLayout && nonce === 0) return;
    let active = true;
    setLoading(true);
    getLayoutConfig().then((value) => active && setLayout(value)).catch(() => undefined).finally(() => active && setLoading(false));
    return () => { active = false; };
  }, [nonce, staticLayout]);
  useEffect(() => {
    let active = true;
    getVendorCategories().then((categories) => { if (active) setVendorCategories(categories); }).catch(() => { if (active) setVendorCategories([]); });
    return () => { active = false; };
  }, [nonce]);
  useEffect(() => { document.documentElement.style.setProperty("--primary-blue", layout.theme.primaryColor); document.documentElement.style.setProperty("--industrial-green", layout.theme.accentColor); }, [layout.theme]);
  const value = useMemo(() => ({ layout, vendorCategories, loading, reload: () => setNonce((value) => value + 1), provided: true }), [layout, vendorCategories, loading]);
  return <Context.Provider value={value}>{children}</Context.Provider>;
}
export const useSite = () => useContext(Context);
