import { Grid2X2, Home, Settings, UserRound, Warehouse } from "lucide-react";
import { Link } from "react-router-dom";

export function MobileBottomNav() {
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
        return (
          <Link className={item.path === "/" ? "active" : ""} key={item.label} to={item.path}>
            <Icon size={20} />
            <span>{item.label}</span>
          </Link>
        );
      })}
    </nav>
  );
}
