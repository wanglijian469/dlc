import { FormEvent, useEffect, useMemo, useState } from "react";
import { useParams, useSearchParams } from "react-router-dom";
import { getFilterOptions, listProducts, type ProductListParams } from "../api/public";
import { PageFrame } from "../components/public/PageFrame";
import { ProductCard } from "../components/public/ProductCard";
import { EmptyState, ErrorState, LoadingState } from "../components/public/StateViews";
import type { FilterOptions, PageResult, Product } from "../types/api";
import { hierarchicalCategoryOptions } from "../utils/categories";
import { SlidersHorizontal, X } from "lucide-react";
import { Pagination } from "../components/public/Pagination";
import { useSite } from "../contexts/SiteContext";
import { getMenuLabel } from "../utils/navigation";
import { MobileDirectorySearch } from "../components/public/MobileDirectorySearch";

const pageSize = 12;

export function ProductsPage() {
  const { layout } = useSite();
  const pageTitle = getMenuLabel(layout.topMenus, "/products", "配件产品");
  const [params, setParams] = useSearchParams();
  const { slug: categorySlug } = useParams();
  const [keyword, setKeyword] = useState(params.get("keyword") || "");
  const [categoryId, setCategoryId] = useState(params.get("categoryId") || params.get("category") || "");
  const [onlyHot, setOnlyHot] = useState(params.get("hot") === "true");
  const [filters, setFilters] = useState<FilterOptions>({ provinces: [], categories: [], serviceTags: [] });
  const [result, setResult] = useState<PageResult<Product> | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [filterOpen, setFilterOpen] = useState(false);

  const query = useMemo<ProductListParams>(
    () => ({
      keyword: params.get("keyword") || undefined,
      categoryId: params.get("categoryId") || params.get("category") || undefined,
      categorySlug: categorySlug || undefined,
      vendorId: params.get("vendorId") || undefined,
      hot: params.get("hot") === "true" || undefined,
      page: Number(params.get("page") || 1),
      pageSize,
    }),
    [params, categorySlug],
  );

  const load = () => {
    setLoading(true);
    setError("");
    Promise.all([getFilterOptions(), listProducts(query)])
      .then(([filterOptions, products]) => {
        setFilters(filterOptions);
        setResult(products);
      })
      .catch(() => setError("产品列表加载失败，请稍后重试"))
      .finally(() => setLoading(false));
  };

  useEffect(load, [query]);

  const submit = (event: FormEvent) => {
    event.preventDefault();
    const next = new URLSearchParams();
    if (keyword.trim()) next.set("keyword", keyword.trim());
    if (categoryId) next.set("categoryId", categoryId);
    if (onlyHot) next.set("hot", "true");
    next.set("page", "1");
    setParams(next);
  };

  const total = result?.total || 0;
  const currentPage = result?.page || 1;
  const activeFilters = [keyword.trim(), categoryId ? filters.categories.find((category) => String(category.id) === categoryId)?.name : "", onlyHot ? "只看热门" : ""].filter(Boolean) as string[];
  const clearFilters = () => { setKeyword(""); setCategoryId(""); setOnlyHot(false); setParams(new URLSearchParams()); };

  return (
    <PageFrame title={pageTitle} subtitle={`按分类、关键词和热门标识查找${pageTitle}`}>
      <MobileDirectorySearch value={keyword} placeholder="搜索配件名称、适配机型" onChange={setKeyword} onSubmit={submit} />
      <button aria-expanded={filterOpen} className="mobile-filter-toggle" type="button" onClick={() => setFilterOpen(!filterOpen)}><SlidersHorizontal size={17} />筛选产品{activeFilters.length > 0 && <span>{activeFilters.length}</span>}</button>
      <form className={`filter-bar ${filterOpen ? "open" : ""}`} onSubmit={submit}>
        <input className="filter-keyword" value={keyword} placeholder="搜索配件名称、适配机型" onChange={(event) => setKeyword(event.target.value)} />
        <select value={categoryId} onChange={(event) => setCategoryId(event.target.value)}>
          <option value="">全部分类</option>
          {hierarchicalCategoryOptions(filters.categories).map((category) => (
            <option key={category.value} value={category.value}>{category.label}</option>
          ))}
        </select>
        <label className="inline-check">
          <input checked={onlyHot} type="checkbox" onChange={(event) => setOnlyHot(event.target.checked)} />
          只看热门
        </label>
        <button className="primary-btn" type="submit">搜索</button>
      </form>
      {activeFilters.length > 0 && <div className="active-filter-row">{activeFilters.map((item) => <span key={item}>{item}</span>)}<button type="button" onClick={clearFilters}><X size={14} />清除筛选</button></div>}
      {loading && <LoadingState />}
      {error && <ErrorState text={error} onRetry={load} />}
      {!loading && !error && result && (
        <>
          <div className="list-summary">共找到 {total} 个产品</div>
          {result.items.length ? (
            <div className="product-grid">
              {result.items.map((product) => <ProductCard key={product.id} product={product} />)}
            </div>
          ) : (
            <EmptyState text="暂无符合条件的产品，请调整筛选条件" />
          )}
          <Pagination page={currentPage} pageSize={result.pageSize} total={result.total} onChange={(page) => { const next = new URLSearchParams(params); next.set("page", String(page)); setParams(next); window.scrollTo({ top: 0, behavior: "smooth" }); }} />
        </>
      )}
    </PageFrame>
  );
}
