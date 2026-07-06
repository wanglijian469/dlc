import { useEffect, useState } from "react";
import { getFriendLinks, getPage } from "../api/public";
import { PageFrame } from "../components/public/PageFrame";
import { EmptyState, ErrorState, LoadingState } from "../components/public/StateViews";
import type { ContentPageRecord, FriendLink } from "../types/api";

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
        {page.content ? <p>{page.content}</p> : <EmptyState text="该页面暂未填写正文内容" />}
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
