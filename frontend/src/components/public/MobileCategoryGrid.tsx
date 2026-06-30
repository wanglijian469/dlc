import { Link } from "react-router-dom";
import type { Menu } from "../../types/api";
import { MenuIcon } from "./MenuIcon";

export function MobileCategoryGrid({ menus }: { menus: Menu[] }) {
  return (
    <div className="mobile-category-grid">
      {menus.slice(0, 10).map((menu) => (
        <Link key={menu.id} to={menu.path || "/products"}>
          <span><MenuIcon icon={menu.icon} size={18} /></span>
          <em>{menu.name}</em>
        </Link>
      ))}
    </div>
  );
}
