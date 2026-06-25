import { Navigate } from "react-router-dom";
import type { ReactNode } from "react";

export function ProtectedAdminRoute({ children }: { children: ReactNode }) {
  const token = localStorage.getItem("admin_token");
  if (!token) {
    return <Navigate to="/admin/login" replace />;
  }
  return <>{children}</>;
}
