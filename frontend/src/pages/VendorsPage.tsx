import { FormEvent, useEffect, useMemo, useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { getFilterOptions, listVendors, type VendorListParams } from "../api/public";
import { EmptyState, ErrorState, LoadingState } from "../components/public/StateViews";
import { PageFrame } from "../components/public/PageFrame";
import { VendorCard } from "../components/public/VendorCard";
import type { FilterOptions, PageResult, Vendor, VendorCategory } from "../types/api";
import { SlidersHorizontal, X } from "lucide-react";
import { Pagination } from "../components/public/Pagination";
import { useSite } from "../contexts/SiteContext";
import { getMenuLabel } from "../utils/navigation";
import { MobileDirectorySearch } from "../components/public/MobileDirectorySearch";
import { buildVendorNavigationMenus } from "../utils/vendorNavigation";

const pageSize = 12;

export function VendorsPage() {
  const { layout, vendorCategories } = useSite();
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
	  vendorCategoryId: params.get("vendorCategoryId") || undefined,
	  newlyJoined: params.get("newlyJoined") === "true" || params.get("newlyJoined") === "1" || undefined,
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
	if (params.get("vendorCategoryId")) next.set("vendorCategoryId", params.get("vendorCategoryId") || "");
	if (params.get("newlyJoined")) next.set("newlyJoined", params.get("newlyJoined") || "");
    if (sort !== "recommended") next.set("sort", sort);
    next.set("page", "1");
    setParams(next);
  };

  const total = result?.total || 0;
  const currentPage = result?.page || 1;
	const selectedVendorCategoryId = params.get("vendorCategoryId") || "";
	const selectedVendorCategory = findVendorCategory(vendorCategories, selectedVendorCategoryId);
	const newlyJoinedSelected = params.get("newlyJoined") === "true" || params.get("newlyJoined") === "1";
	const activeFilters = [keyword.trim(), province, tagId ? filters.serviceTags.find((tag) => String(tag.id) === tagId)?.name : ""].filter(Boolean) as string[];
  const clearFilters = () => { setKeyword(""); setProvince(""); setTagId(""); setSort("recommended"); setParams(new URLSearchParams()); };
	const clearVendorCategory = () => { const next = new URLSearchParams(params); next.delete("vendorCategoryId"); next.delete("page"); setParams(next); };
	const clearNewlyJoined = () => { const next = new URLSearchParams(params); next.delete("newlyJoined"); next.delete("page"); setParams(next); };
	const vendorCategoryPath = (categoryId?: number) => {
	  const next = new URLSearchParams(params);
	  if (categoryId) next.set("vendorCategoryId", String(categoryId)); else next.delete("vendorCategoryId");
	  next.delete("newlyJoined");
	  next.delete("page");
	  return `/vendors${next.toString() ? `?${next.toString()}` : ""}`;
	};
	const newlyJoinedPath = () => {
	  const next = new URLSearchParams(params);
	  next.delete("vendorCategoryId");
	  next.set("newlyJoined", "true");
	  next.set("sort", "latest");
	  next.delete("page");
	  return `/vendors?${next.toString()}`;
	};
	const navigationMenus = useMemo(() => buildVendorNavigationMenus(vendorCategories, { categoryPath: vendorCategoryPath, newlyJoinedPath }), [params, vendorCategories]);

  return (
	<PageFrame navigationMenus={navigationMenus} navigationTitle="厂商分类" title={pageTitle} subtitle="按厂商分类、地区、服务标签和关键词筛选源头农机配件厂商">
      <MobileDirectorySearch value={keyword} placeholder="搜索厂商名称、主营产品" onChange={setKeyword} onSubmit={submit} />
	  <button aria-expanded={filterOpen} className="mobile-filter-toggle" type="button" onClick={() => setFilterOpen(!filterOpen)}><SlidersHorizontal size={17} />筛选与排序{activeFilters.length + (selectedVendorCategory ? 1 : 0) + (newlyJoinedSelected ? 1 : 0) > 0 && <span>{activeFilters.length + (selectedVendorCategory ? 1 : 0) + (newlyJoinedSelected ? 1 : 0)}</span>}</button>
      <form className={`filter-bar ${filterOpen ? "open" : ""}`} onSubmit={submit}>
        <input className="filter-keyword" value={keyword} placeholder="搜索厂商名称、主营产品" onChange={(event) => setKeyword(event.target.value)} />
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
	  {(activeFilters.length > 0 || selectedVendorCategory || newlyJoinedSelected) && <div className="active-filter-row">{selectedVendorCategory && <button className="filter-chip" type="button" onClick={clearVendorCategory}>{selectedVendorCategory.name}<X size={14} /><span className="sr-only">清除厂商分类</span></button>}{newlyJoinedSelected && <button className="filter-chip" type="button" onClick={clearNewlyJoined}>新入驻厂商<X size={14} /><span className="sr-only">清除新入驻筛选</span></button>}{activeFilters.map((item) => <span key={item}>{item}</span>)}<button type="button" onClick={clearFilters}><X size={14} />清除全部</button></div>}
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
		  ) : newlyJoinedSelected ? (
			<div className="vendor-category-empty"><EmptyState text="近 90 天暂无新入驻厂商" /><Link className="primary-btn" to="/join">申请厂商入驻</Link></div>
		  ) : selectedVendorCategory ? (
			<div className="vendor-category-empty"><EmptyState text={`“${selectedVendorCategory.name}”暂无入驻企业，招商进行中`} /><Link className="primary-btn" to="/join">申请厂商入驻</Link></div>
		  ) : (
			<EmptyState text="暂无符合条件的厂商，请调整筛选条件" />
		  )}
          <Pagination page={currentPage} pageSize={result.pageSize} total={result.total} onChange={(page) => { const next = new URLSearchParams(params); next.set("page", String(page)); setParams(next); window.scrollTo({ top: 0, behavior: "smooth" }); }} />
        </>
      )}
    </PageFrame>
  );
}

function findVendorCategory(categories: VendorCategory[], id: string) {
	for (const category of categories) {
		if (String(category.id) === id) return category;
		const child = category.children?.find((item) => String(item.id) === id);
		if (child) return child;
	}
	return undefined;
}
