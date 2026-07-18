import { Link } from "react-router-dom";
import { Factory, MapPinned, Package, Wrench, type LucideIcon } from "lucide-react";
import type { SiteMeta, StatItem } from "../../types/api";

export function StatsFooter({ stats = [], siteMeta }: { stats?: StatItem[]; siteMeta?: SiteMeta }) {
  const copyrightYear = siteMeta?.copyrightYear || String(new Date().getFullYear());
  const copyrightOwner = siteMeta?.copyrightOwner || siteMeta?.siteName || "大陆农机配件";
  const filingNumber = siteMeta?.filingNumber || "待运营方配置";
  return <footer className="stats-footer public-site-footer">
    {stats.length > 0 && <section aria-label="平台数据概览" className="stats-row">{stats.map((item) => {
      const Icon = statIcon(item.label);
      return <article className="footer-stat-card" key={item.label}><span aria-hidden="true" className="footer-stat-icon"><Icon size={22} /></span><div className="footer-stat-copy"><strong>{item.value}</strong><span>{item.label}</span></div></article>;
    })}</section>}
    <div className="footer-directory">
      <div><strong>平台服务</strong><Link to="/about">关于平台</Link><Link to="/join">厂商入驻</Link><Link to="/service">加工服务</Link><Link to="/privacy">隐私说明</Link></div>
      <div><strong>联系与反馈</strong><Link to="/contact">联系我们</Link><Link to="/feedback">反馈建议</Link><Link to="/links">友情链接</Link></div>
      <div><strong>友情链接</strong><Link to="/links">查看合作伙伴与行业服务入口</Link><span>合作链接由后台统一维护</span></div>
      <div className="footer-trust"><strong>信息说明</strong><span>厂商资料经审核后公开展示</span><span>联系方式仅向已登录厂商账号完整展示</span></div>
    </div>
    <div className="footer-legal"><span>© {copyrightYear} {copyrightOwner}</span><span>版权所有</span><span>备案号：{filingNumber}</span></div>
  </footer>;
}

function statIcon(label: string): LucideIcon {
  if (label.includes("产品") || label.includes("配件")) return Package;
  if (label.includes("加工")) return Wrench;
  if (label.includes("省份") || label.includes("地区")) return MapPinned;
  return Factory;
}
