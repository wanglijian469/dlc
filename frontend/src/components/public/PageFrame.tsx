import { Link } from "react-router-dom";
import type { ReactNode } from "react";
import { MobileBottomNav } from "./MobileBottomNav";
import { PublicHeader } from "./PublicHeader";
import type { Menu } from "../../types/api";

const fallbackMenus: Menu[] = [
  { id: 1, name: "首页", path: "/" },
  { id: 2, name: "配件产品", path: "/products" },
  { id: 3, name: "厂商目录", path: "/vendors" },
  { id: 4, name: "加工服务", path: "/service" },
  { id: 5, name: "采购信息", path: "/purchase" },
];

export function PageFrame({ title, subtitle, children }: { title: string; subtitle?: string; children: ReactNode }) {
  return (
    <div className="site-shell">
      <PublicHeader menus={fallbackMenus} />
      <main className="plain-page directory-page">
        <Link className="back-link" to="/">
          返回首页
        </Link>
        <header className="page-heading">
          <h1>{title}</h1>
          {subtitle && <p>{subtitle}</p>}
        </header>
        {children}
      </main>
      <MobileBottomNav />
    </div>
  );
}
