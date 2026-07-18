import { fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter, useLocation } from "react-router-dom";
import { describe, expect, it } from "vitest";
import { HeroSearch } from "./HeroSearch";

function Location() { const location = useLocation(); return <output aria-label="current-location">{location.pathname}{location.search}</output>; }
describe("HeroSearch", () => {
  it("uses the CMS banner image and falls back to the marketplace image", () => {
    const { container, rerender, unmount } = render(<MemoryRouter><HeroSearch banner={{ title: "市场入口", backgroundImage: "/api/media/123" }} /></MemoryRouter>);
    expect(container.querySelector(".hero-search")).toHaveStyle({ "--hero-image": "url(/api/media/123)" });

    rerender(<MemoryRouter><HeroSearch banner={{ title: "市场入口" }} /></MemoryRouter>);
    expect(container.querySelector(".hero-search")).toHaveStyle({ "--hero-image": "url(/images/industry/hero-marketplace.jpg)" });
    unmount();
  });

  it("routes directly to product results and stores search history", () => {
    render(<MemoryRouter><HeroSearch banner={{ title: "快速找厂找货", subtitle: "不再展示的副标题", hotKeywords: ["齿轮"] }} /><Location /></MemoryRouter>);
    expect(screen.queryByText("不再展示的副标题")).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("tab", { name: "找配件" }));
    fireEvent.change(screen.getByLabelText("搜索关键词"), { target: { value: "液压泵" } });
    fireEvent.click(screen.getByRole("button", { name: "立即查找" }));
    expect(screen.getByLabelText("current-location")).toHaveTextContent("/products?keyword=%E6%B6%B2%E5%8E%8B%E6%B3%B5");
    expect(localStorage.getItem("dalu_search_history")).toContain("液压泵");
  });
});
