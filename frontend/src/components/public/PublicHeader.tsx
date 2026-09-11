import { UserRound } from "lucide-react";
import { useState } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { logoutSession, type AccountRole } from "../../api/admin";
import type { Menu, SiteMeta } from "../../types/api";
import { trackAnalytics } from "../../analytics";
import { normalizeTopMenus } from "../../utils/navigation";

const defaultMeta: SiteMeta = {
  brandMark: "农",
  siteName: "大陆农机配件",
  submitVendorText: "厂商入驻",
  adminLoginText: "厂商登录",
  mobileBrandName: "大陆农机配件",
  mobileBrandMark: "农",
};

export function PublicHeader({ menus, siteMeta }: { menus: Menu[]; siteMeta?: SiteMeta }) {
  const location = useLocation();
  const navigate = useNavigate();
  const [session, setSession] = useState(() => readSession());
  const meta = { ...defaultMeta, ...siteMeta };
  const publicMenus = normalizeTopMenus(menus);
  const logout = () => {
    void logoutSession().catch(() => undefined);
    localStorage.removeItem("cms_authenticated");
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
        {publicMenus.map((menu) => (
          <Link className={isActiveMenu(location.pathname, menu.path || "/") ? "active" : ""} key={menu.id} to={menu.path || "/"}>
            {menu.name}
          </Link>
        ))}
      </nav>
      <div className="header-actions">
        {!publicMenus.some((menu) => menu.path === "/about") && <Link className="header-text-link header-about-link" to="/about">关于平台</Link>}
        <Link className="primary-btn" to="/join" onClick={() => trackAnalytics({ eventType: "join_cta_click", path: "/join" })}>{meta.submitVendorText}</Link>
        {session ? (
          <span className="header-account">
            <Link className="outline-btn" to={session.role === "vendor" ? "/admin/vendor-profile" : session.role === "buyer" ? "/account/profile" : "/admin/dashboard"}><UserRound size={16} />{session.role === "vendor" ? "厂商工作台" : session.role === "buyer" ? "账号资料" : "管理中心"}</Link>
            <button className="header-text-link" type="button" onClick={logout}>退出</button>
          </span>
        ) : (
          <Link className="outline-btn" to="/account/login">
            <UserRound size={16} />
            厂商登录
          </Link>
        )}
      </div>
    </header>
  );
}

function readSession() {
  const username = localStorage.getItem("cms_username");
  if (localStorage.getItem("cms_authenticated") !== "true" || !username) return null;
  const role = localStorage.getItem("cms_role") as AccountRole | null;
  if (role !== "vendor" && role !== "buyer" && role !== "admin" && role !== "editor" && role !== "reviewer") return null;
  return { username, role };
}

function isActiveMenu(pathname: string, path: string) {
  if (path === "/") return pathname === "/";
  return pathname === path || pathname.startsWith(`${path}/`);
}
