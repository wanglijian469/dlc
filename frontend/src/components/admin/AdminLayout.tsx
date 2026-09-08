import { BarChart3, Camera, ChevronDown, ChevronRight, ClipboardCheck, Factory, FileImage, FileText, Gavel, History, Home, KeyRound, LayoutDashboard, Link as LinkIcon, ListTree, LogOut, Menu, Newspaper, Package, Settings, Tags, Users, X } from "lucide-react";
import { NavLink, useLocation } from "react-router-dom";
import { useEffect, useState, type ComponentType, type ReactNode } from "react";
import type { LucideProps } from "lucide-react";
import { logoutSession, type AccountRole } from "../../api/admin";

type AdminLink = {
  label: string;
  path: string;
  icon: ComponentType<LucideProps>;
  group: "概览" | "业务内容" | "基础配置" | "系统管理";
  roles?: AccountRole[];
  subLinks?: Array<{ label: string; path: string }>;
};

const links: AdminLink[] = [
  { label: "厂商工作台", path: "/admin/vendor-workspace", icon: LayoutDashboard, group: "概览", roles: ["vendor"] },
  { label: "SEO 工作台", path: "/admin/seo", icon: BarChart3, group: "概览", roles: ["admin"] },
  { label: "控制台", path: "/admin/dashboard", icon: LayoutDashboard, group: "概览" },
  { label: "智能采集", path: "/admin/capture", icon: Camera, group: "业务内容", roles: ["admin", "vendor"] },
  { label: "厂商信息", path: "/admin/vendors", icon: Users, group: "业务内容", roles: ["admin", "editor"] },
  { label: "资料审核", path: "/admin/vendor-reviews", icon: ClipboardCheck, group: "业务内容" },
  { label: "配件产品", path: "/admin/products", icon: Package, group: "业务内容", roles: ["admin", "editor"] },
  { label: "产品审核", path: "/admin/product-reviews", icon: ClipboardCheck, group: "业务内容" },
  { label: "供求治理", path: "/admin/market-posts", icon: ClipboardCheck, group: "业务内容", roles: ["admin", "reviewer"] },
  { label: "厂商动态审核", path: "/admin/vendor-post-reviews", icon: Newspaper, group: "业务内容", roles: ["admin", "reviewer"] },
  { label: "采购竞价监管", path: "/admin/auctions", icon: Gavel, group: "业务内容", roles: ["admin", "reviewer"] },
  { label: "页面与快捷导航", path: "/admin/menus", icon: ListTree, group: "基础配置" },
  { label: "厂商标签", path: "/admin/tags", icon: Tags, group: "基础配置" },
  { label: "配件分类", path: "/admin/categories", icon: BarChart3, group: "基础配置", roles: ["admin", "editor"] },
  { label: "厂商分类", path: "/admin/vendor-categories", icon: ListTree, group: "基础配置", roles: ["admin", "editor"] },
  { label: "Banner 管理", path: "/admin/banners", icon: FileImage, group: "基础配置" },
  { label: "页面与行业文章", path: "/admin/pages", icon: FileText, group: "基础配置", roles: ["admin", "editor"] },
  { label: "友情链接", path: "/admin/friend-links", icon: LinkIcon, group: "基础配置" },
  { label: "平台配置", path: "/admin/configs", icon: Settings, group: "系统管理", subLinks: [{ label: "站点与页脚", path: "/admin/configs?section=site" }, { label: "首页展示", path: "/admin/configs?section=home" }, { label: "主题样式", path: "/admin/configs?section=theme" }, { label: "静态化与缓存", path: "/admin/configs?section=static" }, { label: "访问与采集防护", path: "/admin/configs?section=security" }, { label: "智能采集云服务", path: "/admin/configs?section=capture" }] },
  { label: "CMS 账号", path: "/admin/users", icon: KeyRound, group: "系统管理" },
	{ label: "操作日志", path: "/admin/operation-logs", icon: History, group: "系统管理" },
	{ label: "访问分析", path: "/admin/analytics", icon: BarChart3, group: "系统管理" },
  { label: "我的厂商资料", path: "/admin/vendor-profile", icon: Factory, group: "业务内容", roles: ["vendor"] },
  { label: "我的产品资料", path: "/admin/vendor-products", icon: Package, group: "业务内容", roles: ["vendor"] },
  { label: "企业动态与案例", path: "/admin/vendor-posts", icon: Newspaper, group: "业务内容", roles: ["vendor"] },
  { label: "访问数据", path: "/admin/vendor-analytics", icon: BarChart3, group: "业务内容", roles: ["vendor"] },
  { label: "账号安全", path: "/admin/account-security", icon: KeyRound, group: "系统管理", roles: ["vendor"] },
];

const groups: AdminLink["group"][] = ["概览", "业务内容", "基础配置", "系统管理"];

function matchesTarget(pathname: string, search: string, target: string) {
  const [targetPathname, query = ""] = target.split("?");
  return pathname === targetPathname && search === (query ? `?${query}` : "");
}

export function AdminLayout({ title, children }: { title: string; children: ReactNode }) {
  const location = useLocation();
  const role = (localStorage.getItem("cms_role") || "admin") as AccountRole;
  const username = localStorage.getItem("cms_username") || (role === "vendor" ? "厂商用户" : "admin");
  const visibleLinks = links.filter((link) => link.roles ? link.roles.includes(role) : role === "admin");
  const [menuOpen, setMenuOpen] = useState(false);
  const [openSubmenuPaths, setOpenSubmenuPaths] = useState<Set<string>>(() => new Set(links.filter((link) => link.subLinks?.some((subLink) => matchesTarget(location.pathname, location.search, subLink.path))).map((link) => link.path)));
  const linkClass = (target: string) => matchesTarget(location.pathname, location.search, target) ? "active" : "";
  useEffect(() => {
    const activeParents = links.filter((link) => link.subLinks?.some((subLink) => matchesTarget(location.pathname, location.search, subLink.path))).map((link) => link.path);
    if (activeParents.length) setOpenSubmenuPaths((current) => new Set([...current, ...activeParents]));
  }, [location.pathname, location.search]);
  const toggleSubmenu = (path: string) => setOpenSubmenuPaths((current) => {
    const next = new Set(current);
    next.has(path) ? next.delete(path) : next.add(path);
    return next;
  });
  return (
    <div className={role === "vendor" ? "admin-shell vendor-workspace-shell" : "admin-shell"}>
      {menuOpen && <button aria-label="关闭后台导航" className="admin-nav-backdrop" type="button" onClick={() => setMenuOpen(false)} />}
      <aside className={`admin-sidebar ${menuOpen ? "open" : ""}`} aria-label="后台导航">
        <div className="admin-brand-block">
          <img alt="" className="admin-brand-mark" src="/favicon.svg?v=2" />
          <div>
            <strong>{role === "vendor" ? "大陆农机 · 厂商中心" : "大陆农机配件 CMS"}</strong>
          </div>
          <button aria-label="关闭后台导航" className="admin-nav-close" type="button" onClick={() => setMenuOpen(false)}><X size={19} /></button>
        </div>
        <nav className={`admin-nav ${role === "vendor" ? "admin-nav-vendor" : ""}`}>
          {role === "vendor" ? <div className="admin-nav-vendor-list">
            {visibleLinks.map(({ icon: Icon, label, path }) => <div className="admin-nav-entry" key={path}>
              <NavLink aria-label={`导航：${label}`} className={() => linkClass(path)} to={path} onClick={() => setMenuOpen(false)}><Icon aria-hidden="true" size={17} /><span>{label}</span></NavLink>
            </div>)}
          </div> : groups.map((group) => (
            <div className="admin-nav-group" key={group}>
              <span className="admin-nav-group-title">{group}</span>
              {visibleLinks
                .filter((link) => link.group === group)
                .map(({ icon: Icon, label, path, subLinks }) => {
                  const isOpen = openSubmenuPaths.has(path);
                  return <div className="admin-nav-entry" key={path}>
                    {subLinks ? <button aria-expanded={isOpen} className={`admin-nav-parent ${linkClass(path)}`} type="button" onClick={() => toggleSubmenu(path)}><span><Icon aria-hidden="true" size={17} /><span>{label}</span></span>{isOpen ? <ChevronDown aria-hidden="true" size={16} /> : <ChevronRight aria-hidden="true" size={16} />}</button> : <NavLink aria-label={`导航：${label}`} className={() => linkClass(path)} to={path} onClick={() => setMenuOpen(false)}><Icon aria-hidden="true" size={17} /><span>{label}</span></NavLink>}
                    {subLinks && isOpen && <div aria-label={`${label}子菜单`} className="admin-nav-submenu">{subLinks.map((subLink) => <NavLink aria-label={`导航：${subLink.label}`} className={() => linkClass(subLink.path)} key={subLink.path} to={subLink.path} onClick={() => setMenuOpen(false)}>{subLink.label}</NavLink>)}</div>}
                  </div>;
                })}
            </div>
          ))}
        </nav>
      </aside>
      <main className="admin-main">
        <header className="admin-topbar">
          <button aria-label="打开后台导航" className="admin-menu-button" type="button" onClick={() => setMenuOpen(true)}><Menu size={20} /></button>
          <div>
            <span className="admin-breadcrumb">{role === "vendor" ? "厂商中心" : "后台管理"} / {title}</span>
            <h1>{title}</h1>
          </div>
          <div className="admin-topbar-actions">
            <span className="admin-user-pill">{username} · {role === "vendor" ? "厂商" : "管理员"}</span>
            <a className="outline-btn small" href="/">
              <Home aria-hidden="true" size={15} />
              返回前台
            </a>
            <button className="outline-btn small" type="button" onClick={() => {
              void logoutSession().catch(() => undefined);
              localStorage.removeItem("cms_authenticated");
              localStorage.removeItem("cms_role");
              localStorage.removeItem("cms_username");
              window.location.href = role === "vendor" ? "/account/login" : "/admin/login";
            }}>
              <LogOut aria-hidden="true" size={15} />
              退出
            </button>
          </div>
        </header>
        <section className="admin-content">{children}</section>
      </main>
      {role === "vendor" && <nav className="vendor-workspace-bottom" aria-label="厂商手机导航">{[["工作台", "/admin/vendor-workspace"], ["产品", "/admin/vendor-products"], ["发布", "/admin/vendor-products?new=1"], ["推广", "/admin/vendor-analytics"], ["我的", "/admin/vendor-profile"]].map(([label, path]) => <NavLink key={label} className={() => linkClass(path)} to={path}>{label}</NavLink>)}</nav>}
    </div>
  );
}
