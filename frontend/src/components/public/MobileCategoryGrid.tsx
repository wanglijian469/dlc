import { Link } from "react-router-dom";
import type { Menu } from "../../types/api";

export function MobileCategoryGrid({ menus }: { menus: Menu[] }) {
  return (
    <div className="mobile-category-grid">
      {menus.slice(0, 10).map((menu) => (
        <Link key={menu.id} to={menu.path || "/"}>
          <span>{menu.icon?.slice(0, 1) || "类"}</span>
          <em>{menu.name}</em>
        </Link>
      ))}
    </div>
  );
}
