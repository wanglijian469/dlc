import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";
import { ProductCard } from "./ProductCard";

describe("ProductCard", () => {
  it("renders the industrial fallback when a product image is missing", () => {
    const { container } = render(
      <MemoryRouter>
        <ProductCard
          product={{
            id: 9,
            name: "变速箱齿轮总成",
            categoryId: 2,
            compatibleModels: "联合收割机、拖拉机",
            isHot: true,
            category: { id: 2, name: "传动配件" },
            vendor: { id: 1, name: "河北金瑞农机制造有限公司", province: "河北" },
          }}
        />
      </MemoryRouter>,
    );

    expect(container.querySelector(".industry-cover-default.product-image")).toBeInTheDocument();
    expect(container.textContent).not.toContain("Parts");
    expect(screen.queryByRole("img")).not.toBeInTheDocument();
    expect(screen.getByText("变速箱齿轮总成")).toBeInTheDocument();
    expect(screen.getByText("热销")).toHaveClass("tag-orange");
    expect(screen.getByText(/适配机型：联合收割机、拖拉机/)).toBeInTheDocument();
    expect(screen.getByText(/分类：传动配件/)).toBeInTheDocument();
    expect(screen.getByText(/供应商：河北金瑞农机制造有限公司/)).toBeInTheDocument();
    expect(screen.getByText(/价格：面议 \/ 批量报价/)).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "联系供应商" })).toHaveAttribute("href", "/vendors/1");
    expect(screen.getByRole("link", { name: "厂商信息" })).toHaveAttribute("href", "/products/9");
  });

  it("uses the real product image as the card cover", () => {
    const { container } = render(
      <MemoryRouter>
        <ProductCard product={{ id: 10, name: "液压油缸", image: "https://img.example.com/cylinder.jpg" }} />
      </MemoryRouter>,
    );

    const cover = container.querySelector(".product-image.industry-cover-image");
    const image = container.querySelector(".industry-cover-photo");
    expect(cover).toBeInTheDocument();
    expect(image).toHaveAttribute("src", "https://img.example.com/cylinder.jpg");
  });
});
