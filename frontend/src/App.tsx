import { Navigate, Route, Routes } from "react-router-dom";
import { HomePage } from "./pages/HomePage";
import { PlaceholderPage } from "./pages/PlaceholderPage";
import { SearchPage } from "./pages/SearchPage";
import { VendorDetailPage } from "./pages/VendorDetailPage";
import { AdminDashboardPage } from "./pages/admin/AdminDashboardPage";
import { AdminLoginPage } from "./pages/admin/AdminLoginPage";
import { AdminResourcePage } from "./pages/admin/AdminResourcePage";

export function App() {
  return (
    <Routes>
      <Route path="/" element={<HomePage />} />
      <Route path="/search" element={<SearchPage />} />
      <Route path="/vendors/:id" element={<VendorDetailPage />} />
      <Route path="/vendors" element={<PlaceholderPage title="厂商目录" />} />
      <Route path="/products" element={<PlaceholderPage title="配件产品" />} />
      <Route path="/products/:id" element={<PlaceholderPage title="产品详情" />} />
      <Route path="/join" element={<PlaceholderPage title="提交厂商" />} />
      <Route path="/about" element={<PlaceholderPage title="关于平台" />} />
      <Route path="/service" element={<PlaceholderPage title="加工服务" />} />
      <Route path="/purchase" element={<PlaceholderPage title="采购信息" />} />
      <Route path="/admin/login" element={<AdminLoginPage />} />
      <Route path="/admin/dashboard" element={<AdminDashboardPage />} />
      <Route path="/admin/:resource" element={<AdminResourcePage />} />
      <Route path="/admin" element={<Navigate to="/admin/dashboard" replace />} />
    </Routes>
  );
}
