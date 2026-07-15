import { UserRound } from "lucide-react";
import { useState } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import type { AccountRole } from "../../api/admin";
import type { Menu, SiteMeta } from "../../types/api";
import { trackAnalytics } from "../../analytics";

const defaultMeta: SiteMeta = {
  brandMark: "农",
  siteName: "大陆农机配件",
  submitVendorText: "厂商入驻",
  adminLoginText: "后台登录",
  mobileBrandName: "大陆农机配件",
  mobileBrandMark: "农",
};

export function PublicHeader({ menus, siteMeta }: { menus: Menu[]; siteMeta?: SiteMeta }) {
  const location = useLocation();
  const navigate = useNavigate();
  const [session, setSession] = useState(() => readSession());
  const meta = { ...defaultMeta, ...siteMeta };
  const logout = () => {
    localStorage.removeItem("admin_token");
    localStorage.removeItem("cms_role");
    localStorage.removeItem("cms_username");
    setSession(null);
    navigate("/");
  };
  return (
    <header className="public-header">
      <Link className="brand" to="/">
        {meta.brandLogo ? <img alt="" className="brand-logo" src={meta.brandLogo} /> : <span className="brand-mark">{meta.brandMark}</span>}
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
        <Link className="primary-btn" to="/join" onClick={() => trackAnalytics({ eventType: "join_cta_click", path: "/join" })}>{meta.submitVendorText}</Link>
        {session ? (
          <span className="header-account">
            {session.role === "user"
              ? <span className="header-account-name"><UserRound size={16} />{session.username}</span>
              : <Link className="outline-btn" to={session.role === "vendor" ? "/admin/vendor-profile" : "/admin/dashboard"}><UserRound size={16} />管理中心</Link>}
            <button className="header-text-link" type="button" onClick={logout}>退出</button>
          </span>
        ) : (
          <Link className="outline-btn" to="/admin/login">
            <UserRound size={16} />
            {meta.adminLoginText}
          </Link>
        )}
      </div>
    </header>
  );
}

function readSession() {
  const token = localStorage.getItem("admin_token");
  const username = localStorage.getItem("cms_username");
  if (!token || !username) return null;
  return { username, role: (localStorage.getItem("cms_role") || "user") as AccountRole };
}

function isActiveMenu(pathname: string, path: string) {
  if (path === "/") return pathname === "/";
  return pathname === path || pathname.startsWith(`${path}/`);
}
