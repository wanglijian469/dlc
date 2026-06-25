import { Grid2X2, Home, Settings, UserRound, Warehouse } from "lucide-react";
import { Link, useLocation } from "react-router-dom";

export function MobileBottomNav() {
  const location = useLocation();
  const items = [
    { label: "首页", path: "/", icon: Home },
    { label: "分类", path: "/products", icon: Grid2X2 },
    { label: "厂商", path: "/vendors", icon: Warehouse },
    { label: "加工服务", path: "/service", icon: Settings },
    { label: "我的", path: "/admin/login", icon: UserRound },
  ];
  return (
    <nav className="mobile-bottom-nav">
      {items.map((item) => {
        const Icon = item.icon;
        const active = item.path === "/" ? location.pathname === "/" : location.pathname.startsWith(item.path);
        return (
          <Link className={active ? "active" : ""} key={item.label} to={item.path}>
            <Icon size={20} />
            <span>{item.label}</span>
          </Link>
        );
      })}
    </nav>
  );
}
