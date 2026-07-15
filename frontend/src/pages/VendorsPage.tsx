import { FormEvent, useEffect, useMemo, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { getFilterOptions, listVendors, type VendorListParams } from "../api/public";
import { EmptyState, ErrorState, LoadingState } from "../components/public/StateViews";
import { PageFrame } from "../components/public/PageFrame";
import { VendorCard } from "../components/public/VendorCard";
import type { FilterOptions, PageResult, Vendor } from "../types/api";
import { SlidersHorizontal, X } from "lucide-react";
import { Pagination } from "../components/public/Pagination";
import { useSite } from "../contexts/SiteContext";
import { getMenuLabel } from "../utils/navigation";

const pageSize = 12;

export function VendorsPage() {
  const { layout } = useSite();
  const pageTitle = getMenuLabel(layout.topMenus, "/vendors", "厂商目录");
  const [params, setParams] = useSearchParams();
  const [keyword, setKeyword] = useState(params.get("keyword") || "");
  const [province, setProvince] = useState(params.get("province") || "");
  const [tagId, setTagId] = useState(params.get("tagId") || "");
  const [sort, setSort] = useState<"recommended" | "latest">((params.get("sort") as "recommended" | "latest") || "recommended");
  const [filters, setFilters] = useState<FilterOptions>({ provinces: [], categories: [], serviceTags: [] });
  const [result, setResult] = useState<PageResult<Vendor> | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [filterOpen, setFilterOpen] = useState(false);

  const query = useMemo<VendorListParams>(
    () => ({
      keyword: params.get("keyword") || undefined,
      province: params.get("province") || undefined,
      tagId: params.get("tagId") || undefined,
      sort: ((params.get("sort") as "recommended" | "latest") || "recommended"),
      page: Number(params.get("page") || 1),
      pageSize,
    }),
    [params],
  );

  const load = () => {
    setLoading(true);
    setError("");
    Promise.all([getFilterOptions(), listVendors(query)])
      .then(([filterOptions, vendors]) => {
        setFilters(filterOptions);
        setResult(vendors);
      })
      .catch(() => setError("厂商目录加载失败，请稍后重试"))
      .finally(() => setLoading(false));
  };

  useEffect(load, [query]);

  const submit = (event: FormEvent) => {
    event.preventDefault();
    const next = new URLSearchParams();
    if (keyword.trim()) next.set("keyword", keyword.trim());
    if (province) next.set("province", province);
    if (tagId) next.set("tagId", tagId);
    if (sort !== "recommended") next.set("sort", sort);
    next.set("page", "1");
    setParams(next);
  };

  const total = result?.total || 0;
  const currentPage = result?.page || 1;
  const activeFilters = [keyword.trim(), province, tagId ? filters.serviceTags.find((tag) => String(tag.id) === tagId)?.name : ""].filter(Boolean) as string[];
  const clearFilters = () => { setKeyword(""); setProvince(""); setTagId(""); setSort("recommended"); setParams(new URLSearchParams()); };

  return (
    <PageFrame title={pageTitle} subtitle="按地区、服务标签和关键词筛选源头农机配件厂商">
      <button aria-expanded={filterOpen} className="mobile-filter-toggle" type="button" onClick={() => setFilterOpen(!filterOpen)}><SlidersHorizontal size={17} />筛选与排序{activeFilters.length > 0 && <span>{activeFilters.length}</span>}</button>
      <form className={`filter-bar ${filterOpen ? "open" : ""}`} onSubmit={submit}>
        <input value={keyword} placeholder="搜索厂商名称、主营产品" onChange={(event) => setKeyword(event.target.value)} />
        <select value={province} onChange={(event) => setProvince(event.target.value)}>
          <option value="">全部地区</option>
          {filters.provinces.map((item) => (
            <option key={item} value={item}>
              {item}
            </option>
          ))}
        </select>
        <select value={tagId} onChange={(event) => setTagId(event.target.value)}>
          <option value="">全部服务</option>
          {filters.serviceTags.map((tag) => (
            <option key={tag.id} value={tag.id}>
              {tag.name}
            </option>
          ))}
        </select>
        <select value={sort} onChange={(event) => setSort(event.target.value as "recommended" | "latest")}>
          <option value="recommended">推荐优先</option>
          <option value="latest">最新入驻</option>
        </select>
        <button className="primary-btn" type="submit">
          搜索
        </button>
      </form>
      {activeFilters.length > 0 && <div className="active-filter-row">{activeFilters.map((item) => <span key={item}>{item}</span>)}<button type="button" onClick={clearFilters}><X size={14} />清除筛选</button></div>}
      {loading && <LoadingState />}
      {error && <ErrorState text={error} onRetry={load} />}
      {!loading && !error && result && (
        <>
          <div className="list-summary">共找到 {total} 家厂商</div>
          {result.items.length ? (
            <div className="vendor-grid directory-grid">
              {result.items.map((vendor) => (
                <VendorCard directory key={vendor.id} vendor={vendor} />
              ))}
            </div>
          ) : (
            <EmptyState text="暂无符合条件的厂商，请调整筛选条件" />
          )}
          <Pagination page={currentPage} pageSize={result.pageSize} total={result.total} onChange={(page) => { const next = new URLSearchParams(params); next.set("page", String(page)); setParams(next); window.scrollTo({ top: 0, behavior: "smooth" }); }} />
        </>
      )}
    </PageFrame>
  );
}
