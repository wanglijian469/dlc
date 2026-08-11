import { useEffect, useState } from "react";
import { CalendarDays, MapPin, Phone, ShieldCheck } from "lucide-react";
import { Link, useLocation, useNavigate, useParams } from "react-router-dom";
import { getApiErrorMessage } from "../api/client";
import { getMarketPost, getMarketPostContact } from "../api/market";
import { PageFrame } from "../components/public/PageFrame";
import type { MarketPost, MarketPostContact } from "../types/api";

export function MarketPostDetailPage() {
  const { id = "" } = useParams();
  const [post, setPost] = useState<MarketPost | null>(null);
  const [contact, setContact] = useState<MarketPostContact | null>(null);
  const [error, setError] = useState("");
  const navigate = useNavigate();
  const location = useLocation();
  useEffect(() => { getMarketPost(id).then(setPost).catch(() => setError("供求信息不存在或已下架")); }, [id]);
  const reveal = () => {
    if (localStorage.getItem("cms_authenticated") !== "true") {
      navigate(`/account/login?returnTo=${encodeURIComponent(location.pathname)}`);
      return;
    }
    getMarketPostContact(id).then(setContact).catch((reason) => setError(getApiErrorMessage(reason, "联系方式加载失败")));
  };
  if (error && !post) return <div className="state-page"><p>{error}</p><Link to="/purchase">返回供求大厅</Link></div>;
  if (!post) return <div className="state-page">正在加载…</div>;
  return <PageFrame title={post.title} breadcrumbs={[{ label: "供求信息", path: "/purchase" }]}>
    <article className="market-detail">
      <div className="market-detail-gallery">{post.images.length ? post.images.map((image) => <img alt="" key={image} src={image} />) : <div className="market-detail-placeholder">{post.type === "demand" ? "求购" : "供应"}</div>}</div>
      <div className="market-detail-main"><span className={`market-post-type ${post.type}`}>{post.type === "demand" ? "求购信息" : "供应信息"}</span><h2>{post.title}</h2><div className="market-detail-meta"><span><MapPin size={16} />{[post.province, post.city].filter(Boolean).join(" · ") || "地区面议"}</span><span><CalendarDays size={16} />有效至 {new Date(post.expiresAt).toLocaleDateString("zh-CN")}</span></div><dl><div><dt>数量</dt><dd>{post.quantity || "面议"}</dd></div><div><dt>适配机型</dt><dd>{post.compatibleModels || "请联系发布者确认"}</dd></div><div><dt>交期说明</dt><dd>{post.deliveryNote || "双方协商"}</dd></div></dl><section><h3>详细说明</h3><p>{post.description}</p></section><p className="market-publisher">发布者：{post.publisherName}</p></div>
      <aside className="market-contact-card"><ShieldCheck size={24} /><h3>平台联系方式保护</h3><p>登录后查看联系方式，平台会记录访问并限制异常频率。</p>{contact ? <div className="market-contact-result"><strong>{contact.contactName}</strong><a href={`tel:${contact.phone}`}><Phone size={17} />{contact.phone}</a></div> : <button className="primary-btn" onClick={reveal}><Phone size={17} />登录后查看电话</button>}{error && <p className="form-error">{error}</p>}</aside>
    </article>
    <div className="mobile-market-action">{contact ? <a className="primary-btn" href={`tel:${contact.phone}`}><Phone size={17} />立即联系</a> : <button className="primary-btn" onClick={reveal}><Phone size={17} />查看联系方式</button>}</div>
  </PageFrame>;
}
