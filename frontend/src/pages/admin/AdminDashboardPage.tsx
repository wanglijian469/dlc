import { FileImage, FileText, LayoutDashboard, Link as LinkIcon, ListTree, Package, Settings, Tags, Users } from "lucide-react";
import { Link } from "react-router-dom";
import { AdminLayout } from "../../components/admin/AdminLayout";

const quickEntries = [
  { label: "厂商信息", path: "/admin/vendors", icon: Users, meta: "维护供应商资料、认证和加工能力" },
  { label: "配件产品", path: "/admin/products", icon: Package, meta: "管理产品、分类关联和详情内容" },
  { label: "平台配置", path: "/admin/configs", icon: Settings, meta: "调整首页模块、站点信息和统计文案" },
];

const modules = [
  { label: "导航菜单", icon: ListTree, value: "5 类入口" },
  { label: "厂商标签", icon: Tags, value: "支持筛选" },
  { label: "Banner 管理", icon: FileImage, value: "首页主视觉" },
  { label: "内容页面", icon: FileText, value: "SEO 页面" },
  { label: "友情链接", icon: LinkIcon, value: "外部合作" },
];

export function AdminDashboardPage() {
  return (
    <AdminLayout title="控制台">
      <section className="admin-dashboard-hero">
        <div>
          <span className="admin-eyebrow">Console</span>
          <h2>内容运营工作台</h2>
          <p>统一维护导航、厂商、标签、分类、产品、Banner 和平台配置。后台保存后，前台首页、目录页和搜索页会读取同一套数据。</p>
        </div>
        <div className="admin-hero-badge">
          <LayoutDashboard aria-hidden="true" size={30} />
          <strong>CMS</strong>
          <span>管理端</span>
        </div>
      </section>

      <section className="admin-panel">
        <div className="admin-section-title">
          <h2>快捷入口</h2>
          <span>常用管理任务</span>
        </div>
        <div className="admin-quick-grid">
          {quickEntries.map(({ icon: Icon, label, meta, path }) => (
            <Link aria-label={label} className="admin-quick-card" key={path} to={path}>
              <span>
                <Icon aria-hidden="true" size={22} />
              </span>
              <strong>{label}</strong>
              <small>{meta}</small>
            </Link>
          ))}
        </div>
      </section>

      <section className="admin-panel">
        <div className="admin-section-title">
          <h2>模块概览</h2>
          <span>后台维护范围</span>
        </div>
        <div className="admin-module-grid">
          {modules.map(({ icon: Icon, label, value }) => (
            <article className="admin-module-card" key={label}>
              <Icon aria-hidden="true" size={20} />
              <div>
                <strong>{label}</strong>
                <span>{value}</span>
              </div>
            </article>
          ))}
        </div>
      </section>
    </AdminLayout>
  );
}
