import { Link } from "react-router-dom";
import type { SiteMeta, StatItem } from "../../types/api";

export function StatsFooter({ stats = [], safeguards = [], siteMeta }: { stats?: StatItem[]; safeguards?: string[]; siteMeta?: SiteMeta }) {
  return <footer className="stats-footer public-site-footer">
    {stats.length > 0 && <div className="stats-row">{stats.map((item) => <div key={item.label}><strong>{item.value}</strong><span>{item.label}</span></div>)}</div>}
    <div className="footer-directory">
      <div><strong>平台服务</strong><Link to="/about">关于平台</Link><Link to="/join">提交厂商</Link><Link to="/service">加工服务</Link></div>
      <div><strong>联系与反馈</strong><Link to="/contact">联系我们</Link><Link to="/feedback">反馈建议</Link><Link to="/links">友情链接</Link></div>
      <div><strong>友情链接</strong><Link to="/links">查看合作伙伴与行业服务入口</Link><span>合作链接由后台统一维护</span></div>
      <div className="footer-trust"><strong>信息说明</strong><span>{safeguards.length ? safeguards.join(" · ") : "厂商资料经审核后公开展示"}</span><span>联系电话仅向已登录用户完整展示</span></div>
    </div>
    <div className="footer-legal"><span>© {new Date().getFullYear()} {siteMeta?.siteName || "大陆农机配件"}</span><span>版权所有</span><span>备案号：待运营方配置</span></div>
  </footer>;
}
