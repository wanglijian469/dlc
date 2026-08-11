import { ClipboardList, Grid2X2, Home, Plus, UserRound } from "lucide-react";
import { Link, useLocation } from "react-router-dom";
import type { Menu } from "../../types/api";

const iconMap = { home: Home, grid: Grid2X2, plus: Plus, clipboard: ClipboardList, user: UserRound };
const items: Menu[] = [
  { id: 1, name: "首页", path: "/", icon: "home" },
  { id: 2, name: "分类", path: "/categories", icon: "grid" },
  { id: 3, name: "发布", path: "/publish", icon: "plus" },
  { id: 4, name: "供求", path: "/purchase", icon: "clipboard" },
  { id: 5, name: "我的", path: "/account/posts", icon: "user" },
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
        return <Link className={`${active ? "active" : ""} ${path === "/publish" ? "publish-item" : ""}`} key={item.id} to={path}><Icon size={20} /><span>{item.name}</span></Link>;
      })}
    </nav>
  );
}
