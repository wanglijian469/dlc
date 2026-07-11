import { Navigate } from "react-router-dom";
import type { ReactNode } from "react";

export function ProtectedAdminRoute({ children, roles }: { children: ReactNode; roles?: Array<"admin" | "vendor"> }) {
  const token = localStorage.getItem("admin_token");
  if (!token) {
    return <Navigate to="/admin/login" replace />;
  }
  const role = (localStorage.getItem("cms_role") || "admin") as "admin" | "vendor";
  if (roles && !roles.includes(role)) {
    return <Navigate to={role === "vendor" ? "/admin/vendor-profile" : "/admin/dashboard"} replace />;
  }
  return <>{children}</>;
}
