import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { readFileSync } from "node:fs";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";
import type { Menu } from "../../types/api";
import { MobileCategoryGrid } from "./MobileCategoryGrid";
import { SidebarNav } from "./SidebarNav";

const sidebarMenus: Menu[] = [
  {
    id: 1,
    name: "首页",
    icon: "home",
    path: "/",
  },
  {
    id: 2,
    name: "传动配件",
    icon: "cog",
    path: "/products?categoryId=2",
    isDefaultOpen: true,
    children: [
      { id: 21, name: "变速箱齿轮", icon: "dot", path: "/products?keyword=齿轮" },
      { id: 22, name: "后桥差速器", icon: "dot", path: "/products?keyword=差速器" },
    ],
  },
];

function renderSidebar(initialEntry = "/") {
  return render(
    <MemoryRouter initialEntries={[initialEntry]}>
      <SidebarNav auxiliaryMenus={[]} menus={sidebarMenus} />
    </MemoryRouter>,
  );
}

afterEach(() => cleanup());

describe("menu navigation", () => {
  it("renders sidebar menu icons as SVGs instead of icon-key letters", () => {
    const { container } = renderSidebar();

    expect(container.querySelectorAll(".menu-icon svg")).toHaveLength(2);
    expect(screen.queryByText("c")).not.toBeInTheDocument();
  });

  it("renders mobile category icons as SVGs instead of icon-key letters", () => {
    const { container } = render(
      <MemoryRouter>
        <MobileCategoryGrid menus={[{ id: 1, name: "液压系统", icon: "droplets", path: "/products?keyword=液压" }]} />
      </MemoryRouter>,
    );

    expect(container.querySelectorAll(".mobile-category-grid svg")).toHaveLength(1);
    expect(screen.queryByText("d")).not.toBeInTheDocument();
  });

  it("keeps parent menus collapsed by default even when isDefaultOpen is true", () => {
    renderSidebar("/");

    expect(screen.getByText("传动配件")).toBeInTheDocument();
    expect(screen.queryByText("变速箱齿轮")).not.toBeInTheDocument();
  });

  it("toggles a parent menu by clicking the whole parent row", () => {
    renderSidebar("/");

    fireEvent.click(screen.getByRole("button", { name: "传动配件" }));
    expect(screen.getByText("变速箱齿轮")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "传动配件" }));
    expect(screen.queryByText("变速箱齿轮")).not.toBeInTheDocument();
  });

  it("marks parent menus selected when pathname and search match the parent path", () => {
    const { container } = renderSidebar("/products?categoryId=2");

    const selectedRow = container.querySelector(".menu-row.selected");
    expect(selectedRow).toHaveTextContent("传动配件");
  });

  it("auto-expands and highlights a parent when the current URL matches a child path", () => {
    const { container } = renderSidebar("/products?keyword=齿轮");

    expect(screen.getByText("变速箱齿轮")).toBeInTheDocument();
    expect(container.querySelector(".menu-row.selected")).toHaveTextContent("传动配件");
    expect(screen.getByText("变速箱齿轮")).toHaveClass("active");
  });

  it("does not style default-open sidebar menus as the selected page", () => {
    const css = readFileSync("src/styles/global.css", "utf8");

    expect(css).not.toContain(".sidebar-item.open > .menu-row");
  });
});
