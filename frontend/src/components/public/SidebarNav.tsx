import { ChevronDown, ChevronRight } from "lucide-react";
import { useState } from "react";
import { Link, useLocation } from "react-router-dom";
import type { Menu } from "../../types/api";

export function SidebarNav({ menus, auxiliaryMenus }: { menus: Menu[]; auxiliaryMenus: Menu[] }) {
  const location = useLocation();
  const [openIds, setOpenIds] = useState(() => new Set(menus.filter((menu) => menu.isDefaultOpen).map((menu) => menu.id)));

  const toggle = (id: number) => {
    setOpenIds((current) => {
      const next = new Set(current);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  return (
    <aside className="sidebar">
      <div className="sidebar-main-scroll">
        {menus.map((menu) => {
          const open = openIds.has(menu.id);
          const selected = location.pathname === (menu.path || "/");
          return (
            <div className={`sidebar-item ${open ? "open" : ""}`} key={menu.id}>
              <div className={selected ? "selected menu-row" : "menu-row"}>
                <Link to={menu.path || "/"}>
                  <span className="menu-icon">{menu.icon?.slice(0, 1) || "类"}</span>
                  <span>{menu.name}</span>
                </Link>
                {menu.children?.length ? (
                  <button aria-label={`${open ? "收起" : "展开"}${menu.name}`} type="button" onClick={() => toggle(menu.id)}>
                    {open ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
                  </button>
                ) : null}
              </div>
              {open && menu.children?.length ? (
                <div className="submenu">
                  {menu.children.map((child) => (
                    <Link key={child.id} to={child.path || "/"}>
                      {child.name}
                    </Link>
                  ))}
                </div>
              ) : null}
            </div>
          );
        })}
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
