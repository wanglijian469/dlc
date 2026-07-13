import { render, screen, within } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";
import { StatsFooter } from "./StatsFooter";

describe("StatsFooter", () => {
  it("renders the platform statistics as clear icon cards", () => {
    const { container } = render(
      <MemoryRouter>
        <StatsFooter stats={[
          { label: "入驻厂商", value: "10" },
          { label: "配件产品", value: "14" },
          { label: "加工服务厂商", value: "4" },
          { label: "覆盖省份", value: "6" },
        ]} />
      </MemoryRouter>,
    );

    const overview = screen.getByRole("region", { name: "平台数据概览" });
    for (const label of ["入驻厂商", "配件产品", "加工服务厂商", "覆盖省份"]) {
      expect(within(overview).getByText(label)).toBeInTheDocument();
    }
    expect(container.querySelectorAll(".footer-stat-card")).toHaveLength(4);
    expect(container.querySelectorAll(".footer-stat-icon svg")).toHaveLength(4);
  });
});
