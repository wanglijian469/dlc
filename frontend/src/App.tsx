import { Navigate, Route, Routes, useLocation } from "react-router-dom";
import { lazy, Suspense, useEffect } from "react";
import { trackRoute } from "./analytics";
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
import { ProtectedAdminRoute } from "./pages/admin/ProtectedAdminRoute";
import { ArticlePage, ArticlesPage } from "./pages/ArticlesPage";

const AdminDashboardPage = lazy(() => import("./pages/admin/AdminDashboardPage").then((module) => ({ default: module.AdminDashboardPage })));
const AdminLoginPage = lazy(() => import("./pages/admin/AdminLoginPage").then((module) => ({ default: module.AdminLoginPage })));
const AdminResourcePage = lazy(() => import("./pages/admin/AdminResourcePage").then((module) => ({ default: module.AdminResourcePage })));
const AdminUsersPage = lazy(() => import("./pages/admin/AdminUsersPage").then((module) => ({ default: module.AdminUsersPage })));
const VendorProfilePage = lazy(() => import("./pages/admin/VendorProfilePage").then((module) => ({ default: module.VendorProfilePage })));
const VendorReviewsPage = lazy(() => import("./pages/admin/VendorReviewsPage").then((module) => ({ default: module.VendorReviewsPage })));
const AdminLogsPage = lazy(() => import("./pages/admin/AdminLogsPage").then((module) => ({ default: module.AdminLogsPage })));
const ProductReviewsPage = lazy(() => import("./pages/admin/ProductReviewsPage").then((module) => ({ default: module.ProductReviewsPage })));
const AdminAnalyticsPage = lazy(() => import("./pages/admin/AdminAnalyticsPage").then((module) => ({ default: module.AdminAnalyticsPage })));
const AdminSEOPage = lazy(() => import("./pages/admin/AdminSEOPage").then((module) => ({ default: module.AdminSEOPage })));

export function App() {
  const location = useLocation();
  useEffect(() => { trackRoute(location.pathname, location.search); }, [location.pathname, location.search]);
  return (
    <Suspense fallback={<div className="state-page">正在加载页面…</div>}>
    <Routes>
      <Route path="/" element={<HomePage />} />
      <Route path="/search" element={<SearchPage />} />
      <Route path="/vendors" element={<VendorsPage />} />
      <Route path="/vendors/:id" element={<VendorDetailPage />} />
      <Route path="/products" element={<ProductsPage />} />
      <Route path="/products/category/:slug" element={<ProductsPage />} />
      <Route path="/products/:id" element={<ProductDetailPage />} />
      <Route path="/join" element={<ContentPage slug="join" />} />
      <Route path="/about" element={<ContentPage slug="about" />} />
      <Route path="/service" element={<ProcessingServicesPage />} />
      <Route path="/purchase" element={<ContentPage slug="purchase" />} />
      <Route path="/links" element={<ContentPage slug="links" />} />
      <Route path="/privacy" element={<ContentPage slug="privacy" />} />

      <Route path="/guides" element={<ArticlesPage />} />
      <Route path="/guides/:slug" element={<GuideRoute />} />
      <Route path="/contact" element={<PlaceholderPage title="联系我们" description="如需平台合作、资料更正或厂商认证，请通过平台运营方公布的联系方式与我们联系。" />} />
      <Route path="/feedback" element={<PlaceholderPage title="反馈建议" description="欢迎反馈错误资料、使用问题和功能建议。反馈入口将在运营联系方式配置后开放。" />} />
      <Route path="/account/login" element={<AdminLoginPage />} />
      <Route path="/admin/login" element={<AdminLoginPage staffOnly />} />
      <Route
        path="/admin/dashboard"
        element={
          <ProtectedAdminRoute roles={["admin", "editor", "reviewer"]}>
            <AdminDashboardPage />
          </ProtectedAdminRoute>
        }
      />
      <Route path="/admin/vendor-profile" element={<ProtectedAdminRoute roles={["vendor"]}><VendorProfilePage /></ProtectedAdminRoute>} />
      <Route path="/admin/vendor-reviews" element={<ProtectedAdminRoute roles={["admin"]}><VendorReviewsPage /></ProtectedAdminRoute>} />
      <Route path="/admin/product-reviews" element={<ProtectedAdminRoute roles={["admin"]}><ProductReviewsPage /></ProtectedAdminRoute>} />
      <Route path="/admin/users" element={<ProtectedAdminRoute roles={["admin"]}><AdminUsersPage /></ProtectedAdminRoute>} />
      <Route path="/admin/operation-logs" element={<ProtectedAdminRoute roles={["admin"]}><AdminLogsPage /></ProtectedAdminRoute>} />
      <Route path="/admin/analytics" element={<ProtectedAdminRoute roles={["admin"]}><AdminAnalyticsPage /></ProtectedAdminRoute>} />
      <Route path="/admin/seo" element={<ProtectedAdminRoute roles={["admin"]}><AdminSEOPage /></ProtectedAdminRoute>} />
      {(["vendors", "products", "categories", "pages"] as const).map((resource) => (
        <Route
          key={resource}
          path={`/admin/${resource}`}
          element={<ProtectedAdminRoute roles={["admin", "editor"]}><AdminResourcePage resourceName={resource} /></ProtectedAdminRoute>}
        />
      ))}
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
    </Suspense>
  );
}

function GuideRoute() {
  const slug = window.location.pathname.split("/").pop() || "";
  return <ArticlePage slug={slug} />;
}
