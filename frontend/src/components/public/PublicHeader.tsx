import { UserRound } from "lucide-react";
import { Link, useLocation } from "react-router-dom";
import type { Menu, SiteMeta } from "../../types/api";

const defaultMeta: SiteMeta = {
  brandMark: "农",
  siteName: "大陆农机配件",
  submitVendorText: "提交厂商",
  adminLoginText: "后台登录",
  mobileBrandName: "大陆农机配件",
  mobileBrandMark: "农",
};

export function PublicHeader({ menus, siteMeta }: { menus: Menu[]; siteMeta?: SiteMeta }) {
  const location = useLocation();
  const meta = { ...defaultMeta, ...siteMeta };
  return (
    <header className="public-header">
      <Link className="brand" to="/">
        <span className="brand-mark">{meta.brandMark}</span>
        <span>{meta.siteName}</span>
      </Link>
      <nav className="top-nav">
        {menus.map((menu) => (
          <Link className={isActiveMenu(location.pathname, menu.path || "/") ? "active" : ""} key={menu.id} to={menu.path || "/"}>
            {menu.name}
          </Link>
        ))}
      </nav>
      <div className="header-actions">
        {!menus.some((menu) => menu.path === "/about") && <Link className="header-text-link" to="/about">关于平台</Link>}
        <Link className="primary-btn" to="/join">{meta.submitVendorText}</Link>
        <Link className="outline-btn" to="/admin/login">
          <UserRound size={16} />
          {meta.adminLoginText}
        </Link>
      </div>
    </header>
  );
}

function isActiveMenu(pathname: string, path: string) {
  if (path === "/") return pathname === "/";
  return pathname === path || pathname.startsWith(`${path}/`);
}
