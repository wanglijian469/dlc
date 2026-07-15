import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { getFriendLinks, getPage } from "../api/public";
import { PageFrame } from "../components/public/PageFrame";
import { EmptyState, ErrorState, LoadingState } from "../components/public/StateViews";
import type { ContentPageRecord, FriendLink } from "../types/api";
import type { ContentBlock } from "../types/api";

export function ContentPage({ slug }: { slug: string }) {
  const [page, setPage] = useState<ContentPageRecord | null>(null);
  const [links, setLinks] = useState<FriendLink[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const load = () => {
    setLoading(true);
    setError("");
    Promise.all([getPage(slug), slug === "links" ? getFriendLinks() : Promise.resolve([])])
      .then(([pageData, linkData]) => {
        setPage(pageData);
        setLinks(linkData);
      })
      .catch(() => setError("页面内容加载失败，请稍后重试"))
      .finally(() => setLoading(false));
  };

  useEffect(load, [slug]);

  if (loading) return <PageFrame title="页面内容"><LoadingState /></PageFrame>;
  if (error || !page) return <PageFrame title="页面内容"><ErrorState text={error || "页面不存在"} onRetry={load} /></PageFrame>;

  return (
    <PageFrame title={page.title} subtitle={page.summary}>
      <article className="cms-page-content">
        {page.blocks?.length ? page.blocks.map((block, index) => <ContentBlockView block={block} key={`${block.type}-${index}`} />) : page.content ? <section className="content-text-block"><p>{page.content}</p></section> : <EmptyState text="该页面暂未填写正文内容" />}
      </article>
      {slug === "links" && links.length > 0 && (
        <div className="search-card-list">
          {links.map((link) => (
            <a className="search-card" href={link.url} key={link.id} rel="noreferrer" target="_blank">
              <div className="search-card-head">
                <strong>{link.name}</strong>
                <span className="tag-blue">友情链接</span>
              </div>
              {link.logo && <img alt="" className="link-logo" src={link.logo} />}
              <p>{link.url}</p>
            </a>
          ))}
        </div>
      )}
    </PageFrame>
  );
}

export function ContentBlockView({ block }: { block: ContentBlock }) {
  if (block.type === "hero") return <section className="content-hero-block"><span>大陆农机配件</span><h2>{block.title}</h2><p>{block.text}</p>{block.buttonPath && <Link className="primary-btn" to={block.buttonPath}>{block.buttonText || "立即查看"}</Link>}</section>;
  if (block.type === "steps") return <section className="content-section-block"><h2>{block.title}</h2><div className="content-steps">{block.items?.map((item, index) => <article key={item}><strong>{String(index + 1).padStart(2, "0")}</strong><p>{item}</p></article>)}</div></section>;
  if (block.type === "faq") return <section className="content-section-block"><h2>{block.title}</h2><div className="content-faq">{block.items?.map((item) => { const [question, answer] = item.split("|"); return <details key={item}><summary>{question}</summary><p>{answer || "请联系平台运营人员了解详情。"}</p></details>; })}</div></section>;
  if (block.type === "cta") return <section className="content-cta-block"><div><h2>{block.title}</h2><p>{block.text}</p></div>{block.buttonPath && <Link className="primary-btn" to={block.buttonPath}>{block.buttonText || "立即进入"}</Link>}</section>;
  if (block.type === "contact") return <section className="content-section-block contact-block"><h2>{block.title || "联系平台"}</h2><p>{block.text}</p><div className="contact-actions">{block.phone && <a className="primary-btn" href={`tel:${block.phone}`}>拨打 {block.phone}</a>}{block.wechat && <span>微信：{block.wechat}</span>}</div></section>;
  return <section className="content-section-block"><h2>{block.title}</h2><p>{block.text}</p>{block.items?.length ? <ul>{block.items.map((item) => <li key={item}>{item}</li>)}</ul> : null}</section>;
}
