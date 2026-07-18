import { Link } from "react-router-dom";
import type { Menu } from "../../types/api";
import { MenuIcon } from "./MenuIcon";

export function MobileCategoryGrid({ menus }: { menus: Menu[] }) {
  const featuredMenus = menus.filter((menu) => menu.path !== "/products").slice(0, 7);
  const visibleMenus: Menu[] = [
    ...featuredMenus,
    { id: -1, name: "全部分类", icon: "grid", path: "/products" },
  ];
  return (
    <nav aria-label="常用配件分类" className="mobile-category-grid">
      {visibleMenus.map((menu) => (
        <Link key={menu.id} to={menu.path || "/products"}>
          <span><MenuIcon icon={menu.icon} size={18} /></span>
          <em>{menu.name}</em>
        </Link>
      ))}
    </nav>
  );
}
