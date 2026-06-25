import { AdminLayout } from "../../components/admin/AdminLayout";

export function AdminDashboardPage() {
  return (
    <AdminLayout title="控制台">
      <div className="admin-card">
        <h2>内容管理</h2>
        <p>可维护导航菜单、厂商、标签、分类、产品、Banner 和平台配置。后台保存后，前台首页、目录页和搜索页会读取同一套数据。</p>
      </div>
    </AdminLayout>
  );
}
