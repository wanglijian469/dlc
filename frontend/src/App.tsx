import { Navigate, Route, Routes } from "react-router-dom";
import { HomePage } from "./pages/HomePage";
import { PlaceholderPage } from "./pages/PlaceholderPage";
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
      <Route path="/products/:id" element={<PlaceholderPage title="产品详情" />} />
      <Route path="/join" element={<PlaceholderPage title="提交厂商" description="请联系平台运营人员提交厂商资料，后续版本将开放在线入驻表单。" />} />
      <Route path="/about" element={<PlaceholderPage title="关于平台" description="大陆农机配件聚合源头厂商、配件产品和加工服务信息，帮助维修与采购用户快速找厂。" />} />
      <Route path="/service" element={<PlaceholderPage title="加工服务" description="加工服务栏目将聚合定制加工、来图加工和批量配套能力。" />} />
      <Route path="/purchase" element={<PlaceholderPage title="采购信息" description="采购信息保留入口，首页暂不展示最新求购。" />} />
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
