import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";
import type { Menu } from "../../types/api";
import { MobileCategoryGrid } from "./MobileCategoryGrid";
import { MobileBottomNav } from "./MobileBottomNav";
import { SidebarNav } from "./SidebarNav";

const menus: Menu[] = [
  { id: 1, name: "首页", icon: "home", path: "/" },
  { id: 2, name: "传动配件", icon: "cog", path: "/products?categoryId=2", isDefaultOpen: true, children: [
    { id: 21, name: "变速箱齿轮", icon: "dot", path: "/products?keyword=齿轮" },
    { id: 22, name: "后桥差速器", icon: "dot", path: "/products?keyword=差速器" },
  ] },
];
const renderSidebar = (path = "/") => render(<MemoryRouter initialEntries={[path]}><SidebarNav menus={menus} /></MemoryRouter>);
afterEach(cleanup);

describe("menu navigation", () => {
  it("renders semantic SVG icons", () => { const { container } = renderSidebar(); expect(container.querySelectorAll(".menu-icon svg")).toHaveLength(2); });
  it("renders mobile category icons", () => { const { container } = render(<MemoryRouter><MobileCategoryGrid menus={[{ id: 1, name: "液压系统", icon: "droplets", path: "/products" }]} /></MemoryRouter>); expect(container.querySelectorAll(".mobile-category-grid svg")).toHaveLength(1); });
  it("shows seven common categories plus the all-categories entry", () => {
    const manyMenus = Array.from({ length: 11 }, (_, index) => ({ id: index + 1, name: `分类 ${index + 1}`, path: `/products?categoryId=${index + 1}` }));
    const { container } = render(<MemoryRouter><MobileCategoryGrid menus={manyMenus} /></MemoryRouter>);
    expect(container.querySelectorAll(".mobile-category-grid a")).toHaveLength(8);
    expect(screen.getByRole("link", { name: "全部分类" })).toHaveAttribute("href", "/products");
  });
  it("uses the fixed mobile marketplace navigation", () => {
    render(<MemoryRouter><MobileBottomNav menus={[
      { id: 1, name: "首页", path: "/", icon: "home" },
      { id: 2, name: "厂商", path: "/vendors", icon: "factory" },
      { id: 3, name: "厂商", path: "/account/login", icon: "user" },
    ]} /></MemoryRouter>);
    expect(screen.getByRole("link", { name: "分类" })).toHaveAttribute("href", "/categories");
    expect(screen.getByRole("link", { name: "发布" })).toHaveAttribute("href", "/publish");
    expect(screen.getByRole("link", { name: "供求" })).toHaveAttribute("href", "/purchase");
    expect(screen.getByRole("link", { name: "我的" })).toHaveAttribute("href", "/account/posts");
  });
  it("honors isDefaultOpen and toggles the whole parent row", () => {
    renderSidebar(); expect(screen.getByText("变速箱齿轮")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "传动配件" })); expect(screen.queryByText("变速箱齿轮")).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "传动配件" })); expect(screen.getByText("变速箱齿轮")).toBeInTheDocument();
  });
  it("uses panel controls for the sidebar collapse state", () => {
	const { container } = renderSidebar();
	const toggle = screen.getByRole("button", { name: "收起分类" });
	expect(container.querySelector(".lucide-panel-left-close")).toBeInTheDocument();
	fireEvent.click(toggle);
	expect(screen.getByRole("button", { name: "展开分类" })).toBeInTheDocument();
	expect(container.querySelector(".lucide-panel-left-open")).toBeInTheDocument();
  });
  it("matches a structured parent URL", () => { const { container } = renderSidebar("/products?categoryId=2"); expect(container.querySelector(".menu-row.selected")).toHaveTextContent("传动配件"); });
  it("auto-expands and highlights a matching child URL", () => { renderSidebar("/products?keyword=齿轮"); expect(screen.getByText("变速箱齿轮").closest("a")).toHaveClass("active"); });
});
