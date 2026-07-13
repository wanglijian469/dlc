import { AlertTriangle, BarChart3, ClipboardCheck, FileImage, FileText, LayoutDashboard, Link as LinkIcon, ListTree, Package, Tags, Users } from "lucide-react";
import { Link } from "react-router-dom";
import { useEffect, useState } from "react";
import { AdminLayout } from "../../components/admin/AdminLayout";
import { getDashboardStats } from "../../api/admin";

const quickEntries = [
  { label: "厂商信息", path: "/admin/vendors", icon: Users, meta: "维护供应商资料、认证和加工能力" },
  { label: "资料审核", path: "/admin/vendor-reviews", icon: ClipboardCheck, meta: "审核厂商提交的展示资料和能力信息" },
  { label: "产品审核", path: "/admin/product-reviews", icon: ClipboardCheck, meta: "处理厂商产品新增与更新申请" },
  { label: "配件产品", path: "/admin/products", icon: Package, meta: "管理产品、分类关联和详情内容" },
];

const modules = [
  { label: "导航菜单", icon: ListTree, value: "5 类入口" },
  { label: "厂商标签", icon: Tags, value: "支持筛选" },
  { label: "配件分类", icon: BarChart3, value: "产品目录" },
  { label: "Banner 管理", icon: FileImage, value: "首页主视觉" },
  { label: "内容页面", icon: FileText, value: "SEO 页面" },
  { label: "友情链接", icon: LinkIcon, value: "外部合作" },
];

export function AdminDashboardPage() {
  const [stats, setStats] = useState({ vendors: 0, products: 0, pendingReviews: 0, pendingProductReviews: 0, missingImages: 0 });
	const [statsError, setStatsError] = useState("");
	const loadStats = () => { setStatsError(""); return getDashboardStats().then(setStats).catch(() => setStatsError("控制台统计加载失败，请重试")); };
  useEffect(() => { void loadStats(); }, []);
  return (
    <AdminLayout title="控制台">
      <section className="admin-dashboard-hero">
        <div>
          <span className="admin-eyebrow">Console</span>
          <h2>内容运营工作台</h2>
          <p>日常重点维护厂商信息、资料审核、产品审核和配件产品；导航、标签、分类及页面内容归入基础配置，按需调整。</p>
        </div>
        <div className="admin-hero-badge">
          <LayoutDashboard aria-hidden="true" size={30} />
          <strong>CMS</strong>
          <span>管理端</span>
        </div>
      </section>

      <section className="admin-metric-grid">
        <Link to="/admin/vendors"><Users size={20} /><span>厂商总数<strong>{stats.vendors}</strong></span></Link>
        <Link to="/admin/products"><Package size={20} /><span>产品总数<strong>{stats.products}</strong></span></Link>
        <Link to="/admin/vendor-reviews"><ClipboardCheck size={20} /><span>厂商资料待审<strong>{stats.pendingReviews}</strong></span></Link>
        <Link to="/admin/product-reviews"><ClipboardCheck size={20} /><span>产品待审<strong>{stats.pendingProductReviews}</strong></span></Link>
        <Link to="/admin/vendors"><AlertTriangle size={20} /><span>缺少图片<strong>{stats.missingImages}</strong></span></Link>
      </section>
	  {statsError && <p className="admin-message" role="alert">{statsError} <button onClick={() => void loadStats()} type="button">重试</button></p>}

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
          <h2>基础配置</h2>
          <span>网站平台内容，通常按需维护</span>
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
