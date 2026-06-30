import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";
import { MobileCategoryGrid } from "./MobileCategoryGrid";
import { SidebarNav } from "./SidebarNav";

describe("menu navigation icons", () => {
  it("renders sidebar menu icons as SVGs instead of icon-key letters", () => {
    const { container } = render(
      <MemoryRouter>
        <SidebarNav
          auxiliaryMenus={[]}
          menus={[
            {
              id: 1,
              name: "农机易损件",
              icon: "wrench",
              path: "/products?keyword=易损件",
              children: [{ id: 2, name: "刀片刀架", icon: "dot", path: "/products?keyword=刀片" }],
              isDefaultOpen: true,
            },
          ]}
        />
      </MemoryRouter>,
    );

    expect(container.querySelectorAll(".menu-icon svg")).toHaveLength(1);
    expect(screen.queryByText("w")).not.toBeInTheDocument();
    expect(screen.getByText("刀片刀架")).toHaveAttribute("href", "/products?keyword=刀片");
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
});
