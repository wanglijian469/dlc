import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { getArticle, listArticles } from "../api/public";
import { PageFrame } from "../components/public/PageFrame";
import { EmptyState, ErrorState, LoadingState } from "../components/public/StateViews";
import type { ContentPageRecord } from "../types/api";
import { ContentBlockView } from "./ContentPage";

export function ArticlesPage() {
  const [items, setItems] = useState<ContentPageRecord[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const load = () => { setLoading(true); setError(""); void listArticles().then((result) => setItems(result.items)).catch(() => setError("行业指南加载失败，请稍后重试")).finally(() => setLoading(false)); };
  useEffect(load, []);
  return <PageFrame title="农机配件行业指南" subtitle="围绕选型、适配、保养与采购验收整理真实业务知识">
    {loading && <LoadingState />}{error && <ErrorState text={error} onRetry={load} />}
    {!loading && !error && (items.length ? <div className="search-card-list">{items.map((item) => <Link className="search-card" key={item.id} to={`/guides/${item.slug}`}><div className="search-card-head"><strong>{item.title}</strong><span className="tag-blue">行业指南</span></div>{item.coverImage && <img alt="" className="link-logo" src={item.coverImage} />}<p>{item.summary || item.seoDescription || "查看完整行业指南"}</p><small>{item.authorName || "大陆农机配件"}{item.publishedAt ? ` · ${item.publishedAt.slice(0, 10)}` : ""}</small></Link>)}</div> : <EmptyState text="暂无已发布的行业指南" />)}
  </PageFrame>;
}

export function ArticlePage({ slug }: { slug: string }) {
  const [article, setArticle] = useState<ContentPageRecord | null>(null);
  const [error, setError] = useState("");
  useEffect(() => { setError(""); void getArticle(slug).then(setArticle).catch(() => setError("行业文章不存在或暂未发布")); }, [slug]);
  if (error) return <PageFrame title="行业指南"><ErrorState text={error} onRetry={() => window.location.reload()} /></PageFrame>;
  if (!article) return <PageFrame title="行业指南"><LoadingState /></PageFrame>;
  return <PageFrame title={article.title} subtitle={article.summary} breadcrumbs={[{ label: "行业指南", path: "/guides" }]}><article className="cms-page-content"><p className="article-meta">{article.authorName || "大陆农机配件"}{article.publishedAt ? ` · ${article.publishedAt.slice(0, 10)}` : ""}</p>{article.coverImage && <img alt={article.title} className="article-cover" src={article.coverImage} />}{article.blocks?.length ? article.blocks.map((block, index) => <ContentBlockView block={block} key={`${block.type}-${index}`} />) : article.content ? <section className="content-text-block"><p>{article.content}</p></section> : <EmptyState text="该文章暂未填写正文内容" />}<div className="article-related-links">{article.relatedCategoryId && <Link to={`/products?categoryId=${article.relatedCategoryId}`}>查看相关配件分类</Link>}{article.relatedProductId && <Link to={`/products/${article.relatedProductId}`}>查看相关产品</Link>}{article.relatedVendorId && <Link to={`/vendors/${article.relatedVendorId}`}>查看相关厂商</Link>}</div><Link className="primary-btn" to="/join">厂商入驻</Link></article></PageFrame>;
}
