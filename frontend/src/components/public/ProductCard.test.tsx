import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";
import { ProductCard } from "./ProductCard";

describe("ProductCard", () => {
  it("renders product information without a large top image when image is missing", () => {
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

    expect(container.querySelector(".industry-cover-default")).not.toBeInTheDocument();
    expect(container.querySelector(".product-image")).not.toBeInTheDocument();
    expect(container.textContent).not.toContain("Parts");
    expect(screen.queryByRole("img")).not.toBeInTheDocument();
    expect(screen.getByText("变速箱齿轮总成")).toBeInTheDocument();
    expect(screen.getByText("热销")).toHaveClass("tag-orange");
    expect(screen.getByText(/适配机型：联合收割机、拖拉机/)).toBeInTheDocument();
    expect(screen.getByText(/分类：传动配件/)).toBeInTheDocument();
    expect(screen.getByText(/供应商：河北金瑞农机制造有限公司/)).toBeInTheDocument();
    expect(screen.getByText(/价格：面议 \/ 批量报价/)).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "联系供应商" })).toHaveAttribute("href", "/vendors/1");
  });

  it("uses only a small thumbnail when a real product image exists", () => {
    const { container } = render(
      <MemoryRouter>
        <ProductCard product={{ id: 10, name: "液压油缸", image: "https://img.example.com/cylinder.jpg" }} />
      </MemoryRouter>,
    );

    const image = screen.getByRole("img", { name: "液压油缸" });
    expect(image).toHaveClass("product-thumb");
    expect(image).toHaveAttribute("src", "https://img.example.com/cylinder.jpg");
    expect(container.querySelector(".product-image")).not.toBeInTheDocument();
  });
});
