import { useEffect, useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { search } from "../api/public";
import type { Product, Vendor } from "../types/api";

export function SearchPage() {
  const [params] = useSearchParams();
  const keyword = params.get("keyword") || "";
  const [vendors, setVendors] = useState<Vendor[]>([]);
  const [products, setProducts] = useState<Product[]>([]);

  useEffect(() => {
    search(keyword).then((result) => {
      setVendors(result.vendors);
      setProducts(result.products);
    });
  }, [keyword]);

  return (
    <main className="plain-page">
      <Link to="/">返回首页</Link>
      <h1>搜索：{keyword}</h1>
      <section>
        <h2>厂商结果</h2>
        {vendors.map((vendor) => (
          <Link className="result-row" key={vendor.id} to={`/vendors/${vendor.id}`}>
            {vendor.name}
          </Link>
        ))}
      </section>
      <section>
        <h2>产品结果</h2>
        {products.map((product) => (
          <div className="result-row" key={product.id}>
            {product.name}
          </div>
        ))}
      </section>
    </main>
  );
}
