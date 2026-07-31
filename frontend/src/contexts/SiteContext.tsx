import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from "react";
import { getLayoutConfig } from "../api/public";
import type { LayoutConfig, Menu } from "../types/api";
import { getStaticPageData } from "../utils/staticPageData";

const menus: Menu[] = [{ id: -1, name: "首页", path: "/" }, { id: -2, name: "配件产品", path: "/products" }, { id: -3, name: "厂商目录", path: "/vendors" }, { id: -4, name: "加工服务", path: "/service" }, { id: -5, name: "采购信息", path: "/purchase" }];
const fallback: LayoutConfig = { siteMeta: { siteName: "大陆农机配件", brandMark: "农", submitVendorText: "厂商入驻", adminLoginText: "后台登录", mobileBrandName: "大陆农机配件", mobileBrandMark: "农", copyrightOwner: "大陆农机配件", copyrightYear: String(new Date().getFullYear()), filingNumber: "待运营方配置" }, theme: { primaryColor: "#1559c7", accentColor: "#0d8b6f" }, topMenus: menus, sidebarMenus: [], auxiliaryMenus: [], mobileMenus: [], mobileBottomMenus: menus, version: "fallback" };
type Value = { layout: LayoutConfig; loading: boolean; reload: () => void; provided: boolean };
const Context = createContext<Value>({ layout: fallback, loading: true, reload: () => undefined, provided: false });

export function SiteProvider({ children }: { children: ReactNode }) {
  const staticLayout = getStaticPageData()?.layout;
  const [layout, setLayout] = useState(staticLayout || fallback);
  const [loading, setLoading] = useState(!staticLayout);
  const [nonce, setNonce] = useState(0);
  useEffect(() => {
    if (staticLayout && nonce === 0) return;
    let active = true;
    setLoading(true);
    getLayoutConfig().then((value) => active && setLayout(value)).catch(() => undefined).finally(() => active && setLoading(false));
    return () => { active = false; };
  }, [nonce, staticLayout]);
  useEffect(() => { document.documentElement.style.setProperty("--primary-blue", layout.theme.primaryColor); document.documentElement.style.setProperty("--industrial-green", layout.theme.accentColor); }, [layout.theme]);
  const value = useMemo(() => ({ layout, loading, reload: () => setNonce((value) => value + 1), provided: true }), [layout, loading]);
  return <Context.Provider value={value}>{children}</Context.Provider>;
}
export const useSite = () => useContext(Context);
