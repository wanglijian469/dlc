import { FormEvent, useEffect, useState } from "react";
import { Factory, Grid2X2, Search, Wrench } from "lucide-react";
import { Link, useNavigate } from "react-router-dom";
import { listProducts } from "../../api/public";
import type { HomePayload, Product } from "../../types/api";

export function MobileDirectoryHome({ home }: { home: HomePayload }) {
  const [keyword, setKeyword] = useState("");
  const [products, setProducts] = useState<Product[]>([]);
  const navigate = useNavigate();
  useEffect(() => {
    void listProducts({ recommended: true, pageSize: 6 }).then((result) => setProducts(result.items)).catch(() => undefined);
  }, []);
  const search = (event: FormEvent) => { event.preventDefault(); if (keyword.trim()) navigate(`/search?keyword=${encodeURIComponent(keyword.trim())}`); };
  const vendors = [...home.recommendedVendors, ...home.moreVendors].filter((vendor, index, list) => list.findIndex((candidate) => candidate.id === vendor.id) === index).slice(0, 6);
  return <div className="mobile-market-home">
    <form className="mobile-market-search" onSubmit={search}><Search size={20} /><input aria-label="全站搜索" placeholder="搜索配件、型号、厂家" value={keyword} onChange={(event) => setKeyword(event.target.value)} /><button>搜索</button></form>
    <nav className="mobile-market-shortcuts"><Link to="/products"><Grid2X2 /><span>找配件</span></Link><Link to="/vendors"><Factory /><span>找厂家</span></Link><Link to="/service"><Wrench /><span>加工服务</span></Link><Link to="/join"><Factory /><span>厂商入驻</span></Link></nav>
    {products.length > 0 && <section className="mobile-feed-section"><header><h2>推荐配件</h2><Link to="/products">更多配件</Link></header><div className="mobile-product-masonry">{products.map((product) => <Link key={product.id} to={`/products/${product.slug || product.id}`}>{product.image ? <img alt="" src={product.image} /> : <span className="product-fallback">农机配件</span>}<strong>{product.name}</strong><small>{product.priceNote || "价格面议"}</small></Link>)}</div></section>}
    {vendors.length > 0 && <section className="mobile-feed-section"><header><h2>源头厂家</h2><Link to="/vendors">厂家目录</Link></header><div className="mobile-vendor-list">{vendors.map((vendor) => <Link key={vendor.id} to={`/v/${vendor.slug || vendor.id}`}>{vendor.logo ? <img alt="" src={vendor.logo} /> : <Factory size={22} />}<span><strong>{vendor.shortName || vendor.name}</strong><small>{[vendor.province, vendor.mainProducts].filter(Boolean).join(" · ")}</small></span></Link>)}</div></section>}
  </div>;
}
