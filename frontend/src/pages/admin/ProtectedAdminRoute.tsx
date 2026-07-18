import { Navigate } from "react-router-dom";
import { useEffect, useState, type ReactNode } from "react";
import { getCurrentSession, type AccountRole, type LoginResponse } from "../../api/admin";

export function ProtectedAdminRoute({ children, roles }: { children: ReactNode; roles?: AccountRole[] }) {
  const [session, setSession] = useState<LoginResponse | null>(() => {
    if (localStorage.getItem("cms_authenticated") !== "true") return null;
    return { username: localStorage.getItem("cms_username") || "", role: (localStorage.getItem("cms_role") || "vendor") as AccountRole };
  });
  const [checking, setChecking] = useState(true);

  useEffect(() => {
    getCurrentSession().then((current) => {
      setSession(current);
      localStorage.setItem("cms_authenticated", "true");
      localStorage.setItem("cms_role", current.role);
      localStorage.setItem("cms_username", current.username);
    }).catch(() => {
      setSession(null);
      localStorage.removeItem("cms_authenticated");
      localStorage.removeItem("cms_role");
      localStorage.removeItem("cms_username");
    }).finally(() => setChecking(false));
  }, []);

  if (checking && session) return <>{children}</>;
  if (checking) return <div className="admin-route-loading" role="status">正在验证登录状态…</div>;
  if (!session) {
    return <Navigate to="/admin/login" replace />;
  }
  const role = session.role;
  if (roles && !roles.includes(role)) {
    return <Navigate to={role === "vendor" ? "/admin/vendor-profile" : "/admin/dashboard"} replace />;
  }
  return <>{children}</>;
}
