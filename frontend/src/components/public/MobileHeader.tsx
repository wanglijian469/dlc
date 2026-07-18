import { ChevronDown, ChevronRight, Menu as MenuIconButton, Search, X } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { Link, useLocation } from "react-router-dom";
import type { Menu, SiteMeta } from "../../types/api";
import { MenuIcon } from "./MenuIcon";

export function MobileHeader({ siteMeta, menus = [], auxiliaryMenus = [] }: { siteMeta?: SiteMeta; menus?: Menu[]; auxiliaryMenus?: Menu[] }) {
  const [open, setOpen] = useState(false);
  const [expanded, setExpanded] = useState<Set<number>>(new Set());
  const location = useLocation();
  const triggerRef = useRef<HTMLButtonElement>(null);
  const drawerRef = useRef<HTMLElement>(null);
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
        <header><strong>全部分类</strong><button aria-label="关闭分类菜单" type="button" onClick={() => setOpen(false)}><X size={20} /></button></header>
        <nav>
          {menus.map((menu) => {
            const hasChildren = Boolean(menu.children?.length);
            const isOpen = expanded.has(menu.id);
            return <div className="mobile-drawer-item" key={menu.id}>
              {hasChildren ? <button aria-expanded={isOpen} type="button" onClick={() => setExpanded((current) => { const next = new Set(current); if (next.has(menu.id)) next.delete(menu.id); else next.add(menu.id); return next; })}><span><MenuIcon icon={menu.icon} />{menu.name}</span>{isOpen ? <ChevronDown size={16} /> : <ChevronRight size={16} />}</button> : <Link to={menu.path || "/"}><MenuIcon icon={menu.icon} />{menu.name}</Link>}
              {isOpen && menu.children?.map((child) => <Link className="mobile-drawer-child" key={child.id} to={child.path || "/"}>{child.name}</Link>)}
            </div>;
          })}
        </nav>
        <footer>{auxiliaryMenus.map((menu) => <Link key={menu.id} to={menu.path || "/"}>{menu.name}</Link>)}</footer>
      </aside>
    </>
  );
}
