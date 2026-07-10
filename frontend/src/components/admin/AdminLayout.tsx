import { BarChart3, FileImage, FileText, Home, LayoutDashboard, Link as LinkIcon, ListTree, Package, Settings, Tags, Users } from "lucide-react";
import { NavLink } from "react-router-dom";
import type { ComponentType, ReactNode } from "react";
import type { LucideProps } from "lucide-react";

type AdminLink = {
  label: string;
  path: string;
  icon: ComponentType<LucideProps>;
  group: "概览" | "内容" | "系统";
};

const links: AdminLink[] = [
  { label: "控制台", path: "/admin/dashboard", icon: LayoutDashboard, group: "概览" },
  { label: "导航菜单", path: "/admin/menus", icon: ListTree, group: "内容" },
  { label: "厂商信息", path: "/admin/vendors", icon: Users, group: "内容" },
  { label: "厂商标签", path: "/admin/tags", icon: Tags, group: "内容" },
  { label: "配件分类", path: "/admin/categories", icon: BarChart3, group: "内容" },
  { label: "配件产品", path: "/admin/products", icon: Package, group: "内容" },
  { label: "Banner 管理", path: "/admin/banners", icon: FileImage, group: "内容" },
  { label: "内容页面", path: "/admin/pages", icon: FileText, group: "内容" },
  { label: "友情链接", path: "/admin/friend-links", icon: LinkIcon, group: "内容" },
  { label: "平台配置", path: "/admin/configs", icon: Settings, group: "系统" },
];

const groups: AdminLink["group"][] = ["概览", "内容", "系统"];

export function AdminLayout({ title, children }: { title: string; children: ReactNode }) {
  return (
    <div className="admin-shell">
      <aside className="admin-sidebar" aria-label="后台导航">
        <div className="admin-brand-block">
          <span className="admin-brand-mark">农</span>
          <div>
            <strong>大陆农机配件 CMS</strong>
            <span>Gin-Vue-Admin 融合后台</span>
          </div>
        </div>
        <nav className="admin-nav">
          {groups.map((group) => (
            <div className="admin-nav-group" key={group}>
              <span className="admin-nav-group-title">{group}</span>
              {links
                .filter((link) => link.group === group)
                .map(({ icon: Icon, label, path }) => (
                  <NavLink aria-label={`导航：${label}`} className={({ isActive }) => (isActive ? "active" : "")} key={path} to={path}>
                    <Icon aria-hidden="true" size={17} />
                    <span>{label}</span>
                  </NavLink>
                ))}
            </div>
          ))}
        </nav>
      </aside>
      <main className="admin-main">
        <header className="admin-topbar">
          <div>
            <span className="admin-breadcrumb">后台管理 / {title}</span>
            <h1>{title}</h1>
          </div>
          <div className="admin-topbar-actions">
            <span className="admin-user-pill">admin</span>
            <a className="outline-btn small" href="/">
              <Home aria-hidden="true" size={15} />
              返回前台
            </a>
          </div>
        </header>
        <section className="admin-content">{children}</section>
      </main>
    </div>
  );
}
