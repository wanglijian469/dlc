import { BarChart3, ClipboardCheck, Factory, FileImage, FileText, History, Home, KeyRound, LayoutDashboard, Link as LinkIcon, ListTree, LogOut, Menu, Package, Settings, Tags, Users, X } from "lucide-react";
import { NavLink } from "react-router-dom";
import { useState, type ComponentType, type ReactNode } from "react";
import type { LucideProps } from "lucide-react";

type AdminLink = {
  label: string;
  path: string;
  icon: ComponentType<LucideProps>;
  group: "概览" | "业务内容" | "基础配置" | "系统管理";
  roles?: Array<"admin" | "vendor">;
};

const links: AdminLink[] = [
  { label: "控制台", path: "/admin/dashboard", icon: LayoutDashboard, group: "概览" },
  { label: "厂商信息", path: "/admin/vendors", icon: Users, group: "业务内容" },
  { label: "资料审核", path: "/admin/vendor-reviews", icon: ClipboardCheck, group: "业务内容" },
  { label: "产品审核", path: "/admin/product-reviews", icon: ClipboardCheck, group: "业务内容" },
  { label: "配件产品", path: "/admin/products", icon: Package, group: "业务内容" },
  { label: "导航菜单", path: "/admin/menus", icon: ListTree, group: "基础配置" },
  { label: "厂商标签", path: "/admin/tags", icon: Tags, group: "基础配置" },
  { label: "配件分类", path: "/admin/categories", icon: BarChart3, group: "基础配置" },
  { label: "Banner 管理", path: "/admin/banners", icon: FileImage, group: "基础配置" },
  { label: "内容页面", path: "/admin/pages", icon: FileText, group: "基础配置" },
  { label: "友情链接", path: "/admin/friend-links", icon: LinkIcon, group: "基础配置" },
  { label: "平台配置", path: "/admin/configs", icon: Settings, group: "系统管理" },
  { label: "CMS 账号", path: "/admin/users", icon: KeyRound, group: "系统管理" },
	{ label: "操作日志", path: "/admin/operation-logs", icon: History, group: "系统管理" },
  { label: "我的厂商资料", path: "/admin/vendor-profile", icon: Factory, group: "业务内容", roles: ["vendor"] },
];

const groups: AdminLink["group"][] = ["概览", "业务内容", "基础配置", "系统管理"];

export function AdminLayout({ title, children }: { title: string; children: ReactNode }) {
  const role = (localStorage.getItem("cms_role") || "admin") as "admin" | "vendor";
  const username = localStorage.getItem("cms_username") || (role === "vendor" ? "厂商用户" : "admin");
  const visibleLinks = links.filter((link) => link.roles ? link.roles.includes(role) : role === "admin");
  const [menuOpen, setMenuOpen] = useState(false);
  return (
    <div className="admin-shell">
      {menuOpen && <button aria-label="关闭后台导航" className="admin-nav-backdrop" type="button" onClick={() => setMenuOpen(false)} />}
      <aside className={`admin-sidebar ${menuOpen ? "open" : ""}`} aria-label="后台导航">
        <div className="admin-brand-block">
          <span className="admin-brand-mark">农</span>
          <div>
            <strong>大陆农机配件 CMS</strong>
          </div>
          <button aria-label="关闭后台导航" className="admin-nav-close" type="button" onClick={() => setMenuOpen(false)}><X size={19} /></button>
        </div>
        <nav className="admin-nav">
          {groups.map((group) => (
            <div className="admin-nav-group" key={group}>
              <span className="admin-nav-group-title">{group}</span>
              {visibleLinks
                .filter((link) => link.group === group)
                .map(({ icon: Icon, label, path }) => (
                  <NavLink aria-label={`导航：${label}`} className={({ isActive }) => (isActive ? "active" : "")} key={path} to={path} onClick={() => setMenuOpen(false)}>
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
          <button aria-label="打开后台导航" className="admin-menu-button" type="button" onClick={() => setMenuOpen(true)}><Menu size={20} /></button>
          <div>
            <span className="admin-breadcrumb">后台管理 / {title}</span>
            <h1>{title}</h1>
          </div>
          <div className="admin-topbar-actions">
            <span className="admin-user-pill">{username} · {role === "vendor" ? "厂商" : "管理员"}</span>
            <a className="outline-btn small" href="/">
              <Home aria-hidden="true" size={15} />
              返回前台
            </a>
            <button className="outline-btn small" type="button" onClick={() => {
              localStorage.removeItem("admin_token");
              localStorage.removeItem("cms_role");
              localStorage.removeItem("cms_username");
              window.location.href = "/admin/login";
            }}>
              <LogOut aria-hidden="true" size={15} />
              退出
            </button>
          </div>
        </header>
        <section className="admin-content">{children}</section>
      </main>
    </div>
  );
}
