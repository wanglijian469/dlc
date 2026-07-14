import { Fragment, useEffect, useState, type ReactNode } from "react";
import { Link } from "react-router-dom";
import { useSite } from "../../contexts/SiteContext";
import { getHome } from "../../api/public";
import { MobileBottomNav } from "./MobileBottomNav";
import { MobileHeader } from "./MobileHeader";
import { PublicHeader } from "./PublicHeader";
import { SidebarNav } from "./SidebarNav";
import { StatsFooter } from "./StatsFooter";

type BreadcrumbItem = { label: string; path: string };

export function PageFrame({ title, subtitle, breadcrumbs = [], children }: { title: string; subtitle?: string; breadcrumbs?: BreadcrumbItem[]; children: ReactNode }) {
  const site = useSite();
  const [legacyLayout, setLegacyLayout] = useState(site.layout);
  useEffect(() => { if (!site.provided) getHome().then((home) => setLegacyLayout({ ...site.layout, siteMeta: home.siteMeta || site.layout.siteMeta, topMenus: home.topMenus.length ? home.topMenus : site.layout.topMenus, sidebarMenus: home.sidebarMenus, auxiliaryMenus: home.auxiliaryMenus, mobileMenus: home.mobileMenus, mobileBottomMenus: home.mobileBottomMenus || site.layout.mobileBottomMenus })).catch(() => undefined); }, [site.provided]);
  const layout = site.provided ? site.layout : legacyLayout;
  const breadcrumbTrail = [{ label: "首页", path: "/" }, ...breadcrumbs];
  return <div className="site-shell">
    <PublicHeader menus={layout.topMenus} siteMeta={layout.siteMeta} />
    <MobileHeader auxiliaryMenus={layout.auxiliaryMenus} menus={layout.sidebarMenus} siteMeta={layout.siteMeta} />
    <main className="site-body subpage-body">
      {layout.sidebarMenus.length ? <SidebarNav auxiliaryMenus={layout.auxiliaryMenus} menus={layout.sidebarMenus} /> : <aside aria-label="分类导航" className="sidebar sidebar-placeholder" />}
      <section className="content subpage-content"><div className="plain-page directory-page">
        <nav aria-label="面包屑" className="breadcrumbs">{breadcrumbTrail.map((item, index) => <Fragment key={`${item.path}-${item.label}`}>{index > 0 && <span aria-hidden="true">/</span>}<Link to={item.path}>{item.label}</Link></Fragment>)}<span aria-hidden="true">/</span><span aria-current="page">{title}</span></nav>
        <header className="page-heading"><h1>{title}</h1>{subtitle && <p>{subtitle}</p>}</header>{children}
      </div></section>
    </main>
    <StatsFooter siteMeta={layout.siteMeta} />
    <MobileBottomNav menus={layout.mobileBottomMenus} />
  </div>;
}
