import { ChevronDown, ChevronRight } from "lucide-react";
import { Link } from "react-router-dom";
import type { Menu } from "../../types/api";

export function SidebarNav({ menus, auxiliaryMenus }: { menus: Menu[]; auxiliaryMenus: Menu[] }) {
  return (
    <aside className="sidebar">
      <div className="sidebar-main-scroll">
        {menus.map((menu) => (
          <div className={`sidebar-item ${menu.isDefaultOpen ? "open" : ""}`} key={menu.id}>
            <Link className={menu.path === "/" ? "selected menu-row" : "menu-row"} to={menu.path || "/"}>
              <span className="menu-icon">{menu.icon?.slice(0, 1) || "•"}</span>
              <span>{menu.name}</span>
              {menu.children?.length ? menu.isDefaultOpen ? <ChevronDown size={16} /> : <ChevronRight size={16} /> : null}
            </Link>
            {menu.isDefaultOpen && menu.children?.length ? (
              <div className="submenu">
                {menu.children.map((child) => (
                  <Link key={child.id} to={child.path || "/"}>
                    {child.name}
                  </Link>
                ))}
              </div>
            ) : null}
          </div>
        ))}
      </div>
      <div className="sidebar-links">
        {auxiliaryMenus.map((menu) => (
          <Link key={menu.id} to={menu.path || "/"}>
            {menu.name}
          </Link>
        ))}
      </div>
    </aside>
  );
}
