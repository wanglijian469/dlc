import { Factory, Grid2X2, Search } from "lucide-react";
import { Link } from "react-router-dom";
import { PageFrame } from "../components/public/PageFrame";
import { useSite } from "../contexts/SiteContext";

export function MobileCategoriesPage() {
  const { layout, vendorCategories } = useSite();
  return <PageFrame title="分类查找" subtitle="按配件分类或厂家行业快速浏览">
    <div className="mobile-category-switch"><Link to="/products"><Grid2X2 size={24} /><strong>配件分类</strong><span>按品类、机型查找产品</span></Link><Link to="/vendors"><Factory size={24} /><strong>厂家目录</strong><span>按地区和主营产品查找</span></Link><Link to="/search"><Search size={24} /><strong>全站搜索</strong><span>同时搜索产品与厂家</span></Link></div>
    <section className="mobile-category-links"><h2>常用配件分类</h2><div>{layout.sidebarMenus.slice(0, 12).map((menu) => <Link key={menu.id} to={menu.path || "/products"}>{menu.name}</Link>)}</div></section>
    <section className="mobile-category-links"><h2>厂家行业</h2><div>{vendorCategories.slice(0, 12).map((category) => <Link key={category.id} to={`/vendors?vendorCategoryId=${category.id}`}>{category.name}</Link>)}</div></section>
  </PageFrame>;
}
