import { Search } from "lucide-react";
import { useState } from "react";
import { useNavigate } from "react-router-dom";
import type { Banner } from "../../types/api";
import { getSearchPath } from "../../utils/navigation";

export function HeroSearch({ banner }: { banner: Banner }) {
  const [keyword, setKeyword] = useState("");
  const navigate = useNavigate();
  const doSearch = (value = keyword) => {
    const path = getSearchPath(value);
    if (path) navigate(path);
  };
  return (
    <section className="hero-search">
      <div className="hero-copy">
        <h1>{banner.title}</h1>
        <p>{banner.subtitle}</p>
        <div className="search-box">
          <Search size={18} />
          <input
            value={keyword}
            onChange={(event) => setKeyword(event.target.value)}
            onKeyDown={(event) => event.key === "Enter" && doSearch()}
            placeholder={banner.searchPlaceholder || "搜索配件名称、农机型号、厂商名称等"}
          />
          <button type="button" onClick={() => doSearch()}>
            搜索
          </button>
        </div>
        <div className="hot-keywords">
          <span>热门搜索：</span>
          {banner.hotKeywords?.map((item) => (
            <button key={item} type="button" onClick={() => doSearch(item)}>
              {item}
            </button>
          ))}
        </div>
      </div>
      <div className="hero-machine" />
    </section>
  );
}
