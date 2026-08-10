import { ChevronDown, ChevronLeft, ChevronRight } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { Link, useLocation } from "react-router-dom";
import type { Menu } from "../../types/api";
import { MenuIcon } from "./MenuIcon";

export function SidebarNav({ menus, title }: { menus: Menu[]; title?: string }) {
  const location = useLocation();
  const currentPath = normalizePath(`${location.pathname}${location.search}`);
  const activeParentIds = useMemo(() => findActiveParentIds(menus, currentPath), [menus, currentPath]);
  const [openIds, setOpenIds] = useState<Set<number>>(() => new Set([...activeParentIds, ...menus.filter((menu) => menu.isDefaultOpen).map((menu) => menu.id)]));
  const [collapsed, setCollapsed] = useState(() => localStorage.getItem("directory_sidebar_collapsed") === "true");

  useEffect(() => { setOpenIds((current) => { const next = new Set(current); activeParentIds.forEach((id) => next.add(id)); return next; }); }, [activeParentIds]);
  const toggle = (id: number) => setOpenIds((current) => { const next = new Set(current); next.has(id) ? next.delete(id) : next.add(id); return next; });

  return <aside aria-label={title || "分类导航"} className={`sidebar ${collapsed ? "collapsed" : ""}`}>
    <button className="sidebar-collapse" title={collapsed ? "展开分类" : "收起分类"} type="button" aria-label={collapsed ? "展开分类" : "收起分类"} onClick={() => { const next = !collapsed; setCollapsed(next); localStorage.setItem("directory_sidebar_collapsed", String(next)); }}>{collapsed ? <ChevronRight size={17} /> : <ChevronLeft size={17} />}</button>
    <div className="sidebar-main-scroll">
	  {title && <div className="sidebar-directory-title">{title}</div>}
      {menus.map((menu) => {
        const open = openIds.has(menu.id);
        const hasChildren = Boolean(menu.children?.length);
        const selected = isMenuActive(menu, currentPath) || Boolean(menu.contextActive) || Boolean(menu.children?.some((child) => isMenuActive(child, currentPath) || child.contextActive));
        return <div className={`sidebar-item ${open ? "open" : ""}`} key={menu.id}>
          <div className={selected ? "selected menu-row" : "menu-row"}>
			{hasChildren ? <button aria-expanded={open} className="menu-row-button" type="button" onClick={() => toggle(menu.id)}><span className="menu-row-label"><span className="menu-icon"><MenuIcon icon={menu.icon} /></span><span>{menu.name}</span>{menu.contextActive && <small className="directory-context-badge">所属</small>}{menu.badge && <small className="directory-count-badge">{menu.badge}</small>}</span>{open ? <ChevronDown size={16} /> : <ChevronRight size={16} />}</button> : <Link to={menu.path || "/"}><span className="menu-icon"><MenuIcon icon={menu.icon} /></span><span className="menu-name">{menu.name}</span>{menu.contextActive && <small className="directory-context-badge">所属</small>}{menu.badge && <small className="directory-count-badge">{menu.badge}</small>}</Link>}
          </div>
		  {open && menu.children?.length ? <div className="submenu">{menu.children.map((child) => <Link className={`${isMenuActive(child, currentPath) ? "active" : ""} ${child.contextActive ? "context-active" : ""}`.trim()} key={child.id} to={child.path || "/"}><MenuIcon className="submenu-icon" icon={child.icon || "dot"} size={12} /><span>{child.name}</span>{child.contextActive && <small className="directory-context-badge">所属</small>}{child.badge && <small className="directory-count-badge">{child.badge}</small>}</Link>)}</div> : null}
        </div>;
      })}
    </div>
  </aside>;
}

function findActiveParentIds(menus: Menu[], currentPath: string) { const ids = new Set<number>(); menus.forEach((menu) => { if (menu.isDefaultOpen || menu.contextActive || menu.children?.some((child) => isMenuActive(child, currentPath) || child.contextActive)) ids.add(menu.id); }); return ids; }
function isMenuActive(menu: Menu, currentPath: string) { return normalizePath(menu.path || "/") === currentPath; }
function normalizePath(path: string) { try { return decodeURI(path); } catch { return path; } }
