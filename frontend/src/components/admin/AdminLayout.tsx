import { NavLink } from "react-router-dom";
import type { ReactNode } from "react";

const links = [
  ["控制台", "/admin/dashboard"],
  ["导航菜单", "/admin/menus"],
  ["厂商信息", "/admin/vendors"],
  ["厂商标签", "/admin/tags"],
  ["配件分类", "/admin/categories"],
  ["配件产品", "/admin/products"],
  ["Banner 管理", "/admin/banners"],
  ["平台配置", "/admin/configs"],
];

export function AdminLayout({ title, children }: { title: string; children: ReactNode }) {
  return (
    <div className="admin-shell">
      <aside className="admin-sidebar">
        <strong>大陆农机配件 CMS</strong>
        {links.map(([label, path]) => (
          <NavLink key={path} to={path}>
            {label}
          </NavLink>
        ))}
      </aside>
      <main className="admin-main">
        <header>
          <h1>{title}</h1>
          <a href="/">返回前台</a>
        </header>
        {children}
      </main>
    </div>
  );
}
