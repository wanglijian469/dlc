import { Clock3, MapPin, PackageSearch } from "lucide-react";
import { Link } from "react-router-dom";
import type { MarketPost } from "../../types/api";

export function MarketPostCard({ post, actions }: { post: MarketPost; actions?: React.ReactNode }) {
  const place = [post.province, post.city].filter(Boolean).join(" · ") || "地区面议";
  return <article className="market-post-card">
    <Link className="market-post-cover" to={`/purchase/${post.id}`}>
      {post.images?.[0] ? <img alt="" src={post.images[0]} /> : <PackageSearch aria-hidden="true" size={34} />}
      <span className={`market-post-type ${post.type}`}>{post.type === "demand" ? "求购" : "供应"}</span>
    </Link>
    <div className="market-post-card-body">
      <Link to={`/purchase/${post.id}`}><h2>{post.title}</h2></Link>
      <p>{post.description}</p>
      <div className="market-post-meta"><span><MapPin size={14} />{place}</span><span><Clock3 size={14} />{new Date(post.createdAt).toLocaleDateString("zh-CN")}</span></div>
      <div className="market-post-footer"><strong>{post.quantity || "数量面议"}</strong><span>{post.publisherName}</span></div>
      {actions && <div className="market-post-actions">{actions}</div>}
    </div>
  </article>;
}
