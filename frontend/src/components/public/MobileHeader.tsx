import { ChevronDown, ChevronRight, Menu as MenuIconButton, Search, X } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { Link, useLocation } from "react-router-dom";
import type { Menu, SiteMeta } from "../../types/api";
import { MenuIcon } from "./MenuIcon";

export function MobileHeader({ siteMeta, menus = [], auxiliaryMenus = [], navigationTitle = "全部分类" }: { siteMeta?: SiteMeta; menus?: Menu[]; auxiliaryMenus?: Menu[]; navigationTitle?: string }) {
  const [open, setOpen] = useState(false);
  const [expanded, setExpanded] = useState<Set<number>>(new Set());
  const location = useLocation();
  const triggerRef = useRef<HTMLButtonElement>(null);
  const drawerRef = useRef<HTMLElement>(null);
  const joinMenu = auxiliaryMenus.find((menu) => menu.path === "/join") || { id: -900001, name: "厂商入驻", path: "/join", icon: "clipboard-plus" };
  const secondaryMenus = auxiliaryMenus.filter((menu) => menu.path !== "/join");
  useEffect(() => { setExpanded(new Set(menus.filter((menu) => menu.isDefaultOpen || menu.contextActive || menu.children?.some((child) => child.contextActive)).map((menu) => menu.id))); }, [menus]);
  useEffect(() => setOpen(false), [location.pathname, location.search]);
  useEffect(() => {
    if (!open) return;
	const previousOverflow = document.body.style.overflow;
	document.body.style.overflow = "hidden";
	drawerRef.current?.querySelector<HTMLElement>("button")?.focus();
    const onKey = (event: KeyboardEvent) => event.key === "Escape" && setOpen(false);
    document.addEventListener("keydown", onKey);
	return () => { document.removeEventListener("keydown", onKey); document.body.style.overflow = previousOverflow; triggerRef.current?.focus(); };
  }, [open]);

  return (
    <>
      <header className="mobile-header">
        <button aria-expanded={open} aria-label="打开分类菜单" className="mobile-icon-button" ref={triggerRef} type="button" onClick={() => setOpen(true)}><MenuIconButton size={22} /></button>
        <Link className="brand mobile-brand" to="/"><span className="mobile-brand-visual"><span className="brand-mark">{siteMeta?.mobileBrandMark || siteMeta?.brandMark || "农"}</span>{siteMeta?.brandLogo && <img alt="" className="brand-logo" src={siteMeta.brandLogo} onError={(event) => { event.currentTarget.style.display = "none"; }} />}</span><span>{siteMeta?.mobileBrandName || siteMeta?.siteName || "大陆农机配件"}</span></Link>
        <Link aria-label="搜索" className="mobile-icon-button" to="/search"><Search size={22} /></Link>
      </header>
      {open && <button aria-label="关闭分类菜单" className="mobile-drawer-backdrop" type="button" onClick={() => setOpen(false)} />}
      <aside aria-label="移动端分类导航" aria-modal={open || undefined} className={`mobile-menu-drawer ${open ? "open" : ""}`} ref={drawerRef} role={open ? "dialog" : undefined}>
		<header><strong>{navigationTitle}</strong><button aria-label="关闭分类菜单" type="button" onClick={() => setOpen(false)}><X size={20} /></button></header>
        <nav>
          {menus.map((menu) => {
            const hasChildren = Boolean(menu.children?.length);
            const isOpen = expanded.has(menu.id);
            return <div className={`mobile-drawer-item ${menu.contextActive ? "context-active" : ""}`} key={menu.id}>
			  {hasChildren ? <button aria-expanded={isOpen} type="button" onClick={() => setExpanded((current) => { const next = new Set(current); if (next.has(menu.id)) next.delete(menu.id); else next.add(menu.id); return next; })}><span><MenuIcon icon={menu.icon} />{menu.name}{menu.contextActive && <small className="directory-context-badge">所属</small>}{menu.badge && <small className="directory-count-badge">{menu.badge}</small>}</span>{isOpen ? <ChevronDown size={16} /> : <ChevronRight size={16} />}</button> : <Link to={menu.path || "/"}><span><MenuIcon icon={menu.icon} />{menu.name}{menu.contextActive && <small className="directory-context-badge">所属</small>}</span>{menu.badge && <small className="directory-count-badge">{menu.badge}</small>}</Link>}
			  {isOpen && menu.children?.map((child) => <Link className={`mobile-drawer-child ${child.contextActive ? "context-active" : ""}`} key={child.id} to={child.path || "/"}><span>{child.name}{child.contextActive && <small className="directory-context-badge">所属</small>}</span>{child.badge && <small className="directory-count-badge">{child.badge}</small>}</Link>)}
            </div>;
          })}
        </nav>
        <footer className="mobile-drawer-footer">
		  <Link className="mobile-drawer-join" to={joinMenu.path || "/join"}><MenuIcon icon={joinMenu.icon || "clipboard-plus"} />{joinMenu.name}</Link>
		  {secondaryMenus.length > 0 && <div>{secondaryMenus.map((menu) => <Link key={menu.id} to={menu.path || "/"}>{menu.name}</Link>)}</div>}
		</footer>
      </aside>
    </>
  );
}
