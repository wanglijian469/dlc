import { Navigate, Route, Routes } from "react-router-dom";
import { ContentPage } from "./pages/ContentPage";
import { HomePage } from "./pages/HomePage";
import { ProductDetailPage } from "./pages/ProductDetailPage";
import { ProcessingServicesPage } from "./pages/ProcessingServicesPage";
import { ProductsPage } from "./pages/ProductsPage";
import { SearchPage } from "./pages/SearchPage";
import { VendorDetailPage } from "./pages/VendorDetailPage";
import { VendorsPage } from "./pages/VendorsPage";
import { AdminDashboardPage } from "./pages/admin/AdminDashboardPage";
import { AdminLoginPage } from "./pages/admin/AdminLoginPage";
import { AdminResourcePage } from "./pages/admin/AdminResourcePage";
import { ProtectedAdminRoute } from "./pages/admin/ProtectedAdminRoute";

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
      <Route path="/admin/login" element={<AdminLoginPage />} />
      <Route
        path="/admin/dashboard"
        element={
          <ProtectedAdminRoute>
            <AdminDashboardPage />
          </ProtectedAdminRoute>
        }
      />
      <Route
        path="/admin/:resource"
        element={
          <ProtectedAdminRoute>
            <AdminResourcePage />
          </ProtectedAdminRoute>
        }
      />
      <Route path="/admin" element={<Navigate to="/admin/dashboard" replace />} />
    </Routes>
  );
}
