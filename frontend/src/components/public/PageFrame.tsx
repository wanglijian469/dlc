import { useEffect, useState, type ReactNode } from "react";
import { Link } from "react-router-dom";
import { getHome } from "../../api/public";
import type { HomePayload, Menu } from "../../types/api";
import { MobileBottomNav } from "./MobileBottomNav";
import { MobileHeader } from "./MobileHeader";
import { PublicHeader } from "./PublicHeader";
import { SidebarNav } from "./SidebarNav";

const fallbackMenus: Menu[] = [
  { id: 1, name: "首页", path: "/" },
  { id: 2, name: "配件产品", path: "/products" },
  { id: 3, name: "厂商目录", path: "/vendors" },
  { id: 4, name: "加工服务", path: "/service" },
  { id: 5, name: "采购信息", path: "/purchase" },
];

export function PageFrame({ title, subtitle, children }: { title: string; subtitle?: string; children: ReactNode }) {
  const [home, setHome] = useState<HomePayload | null>(null);

  useEffect(() => {
    let ignore = false;
    getHome()
      .then((payload) => {
        if (!ignore) setHome(payload);
      })
      .catch(() => {
        if (!ignore) setHome(null);
      });
    return () => {
      ignore = true;
    };
  }, []);

  const topMenus = home?.topMenus?.length ? home.topMenus : fallbackMenus;

  return (
    <div className="site-shell">
      <PublicHeader menus={topMenus} />
      <MobileHeader />
      <main className="site-body subpage-body">
        {home?.sidebarMenus?.length ? (
          <SidebarNav auxiliaryMenus={home.auxiliaryMenus || []} menus={home.sidebarMenus} />
        ) : (
          <aside aria-label="分类导航加载中" className="sidebar sidebar-placeholder" />
        )}
        <section className="content subpage-content">
          <div className="plain-page directory-page">
            <Link className="back-link" to="/">
              返回首页
            </Link>
            <header className="page-heading">
              <h1>{title}</h1>
              {subtitle && <p>{subtitle}</p>}
            </header>
            {children}
          </div>
        </section>
      </main>
      <MobileBottomNav />
    </div>
  );
}
