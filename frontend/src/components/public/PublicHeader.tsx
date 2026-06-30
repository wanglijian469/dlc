import { UserRound } from "lucide-react";
import { Link, useLocation } from "react-router-dom";
import type { Menu } from "../../types/api";

const BRAND_MARK = "\u519c";
const BRAND_NAME = "\u5927\u9646\u519c\u673a\u914d\u4ef6";
const SUBMIT_VENDOR = "\u63d0\u4ea4\u5382\u5546";
const ADMIN_LOGIN = "\u540e\u53f0\u767b\u5f55";

export function PublicHeader({ menus }: { menus: Menu[] }) {
  const location = useLocation();
  return (
    <header className="public-header">
      <Link className="brand" to="/">
        <span className="brand-mark">{BRAND_MARK}</span>
        <span>{BRAND_NAME}</span>
      </Link>
      <nav className="top-nav">
        {menus.map((menu) => (
          <Link className={isActiveMenu(location.pathname, menu.path || "/") ? "active" : ""} key={menu.id} to={menu.path || "/"}>
            {menu.name}
          </Link>
        ))}
      </nav>
      <div className="header-actions">
        <Link className="primary-btn" to="/join">
          {SUBMIT_VENDOR}
        </Link>
        <Link className="outline-btn" to="/admin/login">
          <UserRound size={16} />
          {ADMIN_LOGIN}
        </Link>
      </div>
    </header>
  );
}

function isActiveMenu(pathname: string, path: string) {
  if (path === "/") return pathname === "/";
  return pathname === path || pathname.startsWith(`${path}/`);
}
