import type { Product } from "../../types/api";
import { getProductIndustryKind, IndustryCover } from "./IndustryCover";

export function ProductCover({ product }: { product: Product }) {
  return <IndustryCover className="product-image" image={product.image} kind={getProductIndustryKind(product)} iconSize={46} />;
}
