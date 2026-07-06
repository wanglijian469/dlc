import { Grid2X2, Home, Settings, UserRound, Warehouse } from "lucide-react";
import { Link, useLocation } from "react-router-dom";
import type { Menu } from "../../types/api";

const iconMap = { home: Home, grid: Grid2X2, factory: Warehouse, warehouse: Warehouse, settings: Settings, user: UserRound };
const fallbackItems: Menu[] = [
  { id: 1, name: "首页", path: "/", icon: "home" },
  { id: 2, name: "分类", path: "/products", icon: "grid" },
  { id: 3, name: "厂商", path: "/vendors", icon: "factory" },
  { id: 4, name: "加工服务", path: "/service", icon: "settings" },
  { id: 5, name: "我的", path: "/admin/login", icon: "user" },
];

export function MobileBottomNav({ menus }: { menus?: Menu[] }) {
  const location = useLocation();
  const items = menus?.length ? menus : fallbackItems;
  return (
    <nav className="mobile-bottom-nav">
      {items.map((item) => {
        const Icon = iconMap[(item.icon || "") as keyof typeof iconMap] || Grid2X2;
        const path = item.path || "/";
        const active = path === "/" ? location.pathname === "/" : location.pathname.startsWith(path);
        return (
          <Link className={active ? "active" : ""} key={item.id || item.name} to={path}>
            <Icon size={20} />
            <span>{item.name}</span>
          </Link>
        );
      })}
    </nav>
  );
}
