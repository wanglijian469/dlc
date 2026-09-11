import { Wrench, Grid2X2, Home, Factory, UserRound } from "lucide-react";
import { Link, useLocation } from "react-router-dom";
import type { Menu } from "../../types/api";

const iconMap = { home: Home, grid: Grid2X2, factory: Factory, wrench: Wrench, user: UserRound };
const items: Menu[] = [
  { id: 1, name: "首页", path: "/", icon: "home" },
  { id: 2, name: "找产品", path: "/products", icon: "grid" },
  { id: 3, name: "找厂商", path: "/vendors", icon: "factory" },
  { id: 4, name: "加工服务", path: "/service", icon: "wrench" },
  { id: 5, name: "我的", path: "/account/profile", icon: "user" },
];

export function MobileBottomNav({ menus }: { menus?: Menu[] }) {
  void menus; // The mobile information architecture is intentionally fixed.
  const location = useLocation();
  return (
    <nav aria-label="移动端主导航" className="mobile-bottom-nav">
      {items.map((item) => {
        const Icon = iconMap[(item.icon || "") as keyof typeof iconMap] || Grid2X2;
        const path = item.path || "/";
        const active = path === "/" ? location.pathname === "/" : location.pathname.startsWith(path);
        return <Link className={active ? "active" : ""} key={item.id} to={path === "/account/profile" && localStorage.getItem("cms_role") === "vendor" ? "/admin/vendor-workspace" : path}><Icon size={20} /><span>{item.name}</span></Link>;
      })}
    </nav>
  );
}
