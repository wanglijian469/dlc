import { Link } from "react-router-dom";
import { PageFrame } from "../components/public/PageFrame";

export function NotFoundPage() {
  return <PageFrame title="页面未找到" subtitle="您访问的地址不存在或内容已经下线">
    <section className="not-found"><strong>404</strong><p>可以返回首页，或继续查找厂商和配件产品。</p><div className="card-actions"><Link className="primary-btn" to="/">返回首页</Link><Link className="outline-btn" to="/search">开始搜索</Link></div></section>
  </PageFrame>;
}
