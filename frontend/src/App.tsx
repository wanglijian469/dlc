import { Navigate, Route, Routes } from "react-router-dom";
import { ContentPage } from "./pages/ContentPage";
import { HomePage } from "./pages/HomePage";
import { ProductDetailPage } from "./pages/ProductDetailPage";
import { ProcessingServicesPage } from "./pages/ProcessingServicesPage";
import { ProductsPage } from "./pages/ProductsPage";
import { SearchPage } from "./pages/SearchPage";
import { NotFoundPage } from "./pages/NotFoundPage";
import { PlaceholderPage } from "./pages/PlaceholderPage";
import { VendorDetailPage } from "./pages/VendorDetailPage";
import { VendorsPage } from "./pages/VendorsPage";
import { AdminDashboardPage } from "./pages/admin/AdminDashboardPage";
import { AdminLoginPage } from "./pages/admin/AdminLoginPage";
import { AdminResourcePage } from "./pages/admin/AdminResourcePage";
import { ProtectedAdminRoute } from "./pages/admin/ProtectedAdminRoute";
import { AdminUsersPage } from "./pages/admin/AdminUsersPage";
import { VendorProfilePage } from "./pages/admin/VendorProfilePage";
import { VendorReviewsPage } from "./pages/admin/VendorReviewsPage";
import { AdminLogsPage } from "./pages/admin/AdminLogsPage";

export function App() {
  return (
    <Routes>
      <Route path="/" element={<HomePage />} />
      <Route path="/search" element={<SearchPage />} />
      <Route path="/vendors" element={<VendorsPage />} />
      <Route path="/vendors/:id" element={<VendorDetailPage />} />
      <Route path="/products" element={<ProductsPage />} />
      <Route path="/products/:id" element={<ProductDetailPage />} />
      <Route path="/join" element={<ContentPage slug="join" />} />
      <Route path="/about" element={<ContentPage slug="about" />} />
      <Route path="/service" element={<ProcessingServicesPage />} />
      <Route path="/purchase" element={<ContentPage slug="purchase" />} />
      <Route path="/links" element={<ContentPage slug="links" />} />
      <Route path="/contact" element={<PlaceholderPage title="联系我们" description="如需平台合作、资料更正或厂商认证，请通过平台运营方公布的联系方式与我们联系。" />} />
      <Route path="/feedback" element={<PlaceholderPage title="反馈建议" description="欢迎反馈错误资料、使用问题和功能建议。反馈入口将在运营联系方式配置后开放。" />} />
      <Route path="/admin/login" element={<AdminLoginPage />} />
      <Route
        path="/admin/dashboard"
        element={
          <ProtectedAdminRoute roles={["admin"]}>
            <AdminDashboardPage />
          </ProtectedAdminRoute>
        }
      />
      <Route path="/admin/vendor-profile" element={<ProtectedAdminRoute roles={["vendor"]}><VendorProfilePage /></ProtectedAdminRoute>} />
      <Route path="/admin/vendor-reviews" element={<ProtectedAdminRoute roles={["admin"]}><VendorReviewsPage /></ProtectedAdminRoute>} />
      <Route path="/admin/users" element={<ProtectedAdminRoute roles={["admin"]}><AdminUsersPage /></ProtectedAdminRoute>} />
      <Route path="/admin/operation-logs" element={<ProtectedAdminRoute roles={["admin"]}><AdminLogsPage /></ProtectedAdminRoute>} />
      <Route
        path="/admin/:resource"
        element={
          <ProtectedAdminRoute roles={["admin"]}>
            <AdminResourcePage />
          </ProtectedAdminRoute>
        }
      />
      <Route path="/admin" element={<Navigate to="/admin/dashboard" replace />} />
      <Route path="*" element={<NotFoundPage />} />
    </Routes>
  );
}
