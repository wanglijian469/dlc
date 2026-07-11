import { fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter, useLocation } from "react-router-dom";
import { describe, expect, it } from "vitest";
import { HeroSearch } from "./HeroSearch";

function Location() { const location = useLocation(); return <output aria-label="current-location">{location.pathname}{location.search}</output>; }
describe("HeroSearch", () => {
  it("routes directly to product results and stores search history", () => {
    render(<MemoryRouter><HeroSearch banner={{ title: "快速找厂找货", hotKeywords: ["齿轮"] }} /><Location /></MemoryRouter>);
    fireEvent.click(screen.getByRole("tab", { name: "找配件" }));
    fireEvent.change(screen.getByLabelText("搜索关键词"), { target: { value: "液压泵" } });
    fireEvent.click(screen.getByRole("button", { name: "立即查找" }));
    expect(screen.getByLabelText("current-location")).toHaveTextContent("/products?keyword=%E6%B6%B2%E5%8E%8B%E6%B3%B5");
    expect(localStorage.getItem("dalu_search_history")).toContain("液压泵");
  });
});
