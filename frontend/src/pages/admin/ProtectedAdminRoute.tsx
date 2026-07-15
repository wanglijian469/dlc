import { Navigate } from "react-router-dom";
import type { ReactNode } from "react";
import type { AccountRole } from "../../api/admin";

export function ProtectedAdminRoute({ children, roles }: { children: ReactNode; roles?: AccountRole[] }) {
  const token = localStorage.getItem("admin_token");
  if (!token) {
    return <Navigate to="/admin/login" replace />;
  }
  const role = (localStorage.getItem("cms_role") || "user") as AccountRole;
  if (roles && !roles.includes(role)) {
    return <Navigate to={role === "vendor" ? "/admin/vendor-profile" : role === "admin" ? "/admin/dashboard" : "/"} replace />;
  }
  return <>{children}</>;
}
