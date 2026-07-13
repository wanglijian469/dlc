import { Clock3, Factory, PackageSearch, Search, Trash2 } from "lucide-react";
import { useEffect, useState, type CSSProperties } from "react";
import { useNavigate } from "react-router-dom";
import type { Banner } from "../../types/api";

type SearchType = "all" | "vendors" | "products";
const historyKey = "dalu_search_history";

export function HeroSearch({ banner }: { banner: Banner }) {
  const [keyword, setKeyword] = useState("");
  const [type, setType] = useState<SearchType>("all");
  const [history, setHistory] = useState<string[]>([]);
  const navigate = useNavigate();

  useEffect(() => {
    try { setHistory(JSON.parse(localStorage.getItem(historyKey) || "[]")); }
    catch { setHistory([]); }
  }, []);

  const doSearch = (raw = keyword) => {
    const value = raw.trim();
    if (!value) return;
    const next = [value, ...history.filter((item) => item !== value)].slice(0, 6);
    setHistory(next);
    localStorage.setItem(historyKey, JSON.stringify(next));
    const query = `keyword=${encodeURIComponent(value)}`;
    navigate(type === "vendors" ? `/vendors?${query}` : type === "products" ? `/products?${query}` : `/search?${query}`);
  };

  const heroImage = banner.backgroundImage || "/images/industry/hero-marketplace.jpg";
  return (
    <section className="hero-search marketplace-hero" style={{ "--hero-image": `url(${heroImage})` } as CSSProperties}>
      <div className="hero-copy">
        <h1>{banner.title || "查农机配件，找公开厂商资料"}</h1>
        <div className="hero-search-panel">
          <div className="search-type-tabs" role="tablist" aria-label="搜索范围">
            <button aria-selected={type === "all"} onClick={() => setType("all")} role="tab" type="button"><Search size={15} />综合搜索</button>
            <button aria-selected={type === "vendors"} onClick={() => setType("vendors")} role="tab" type="button"><Factory size={15} />找厂商</button>
            <button aria-selected={type === "products"} onClick={() => setType("products")} role="tab" type="button"><PackageSearch size={15} />找配件</button>
          </div>
          <div className="search-box">
            <Search size={20} />
            <input aria-label="搜索关键词" value={keyword} onChange={(event) => setKeyword(event.target.value)} onKeyDown={(event) => event.key === "Enter" && doSearch()} placeholder={banner.searchPlaceholder || "输入厂商名称、配件名称或农机型号"} />
            <button type="button" onClick={() => doSearch()}>立即查找</button>
          </div>
          <div className="hot-keywords"><span>热门：</span>{banner.hotKeywords?.slice(0, 6).map((item) => <button key={item} type="button" onClick={() => doSearch(item)}>{item}</button>)}</div>
          {history.length > 0 && <div className="search-history"><span><Clock3 size={14} />最近搜索</span>{history.map((item) => <button key={item} type="button" onClick={() => doSearch(item)}>{item}</button>)}<button aria-label="清空搜索历史" className="clear-history" type="button" onClick={() => { setHistory([]); localStorage.removeItem(historyKey); }}><Trash2 size={14} /></button></div>}
        </div>
      </div>
    </section>
  );
}
