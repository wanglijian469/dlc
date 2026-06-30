import { render } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";
import { VendorCard } from "./VendorCard";

describe("VendorCard", () => {
  it("does not render old Factory dummy covers", () => {
    const { container } = render(
      <MemoryRouter>
        <VendorCard
          vendor={{
            id: 1,
            name: "山东测试农机配件有限公司",
            coverImage: "https://dummyimage.com/600x320/eaf3ff/1f2a3d&text=Factory+01",
            mainProducts: "变速箱、齿轮、链条",
            tags: [{ id: 1, name: "源头厂商" }],
          }}
        />
      </MemoryRouter>,
    );

    const cover = container.querySelector(".vendor-cover") as HTMLElement;
    expect(cover.style.backgroundImage).not.toContain("Factory");
    expect(cover.textContent).not.toContain("Factory");
    expect(cover).toHaveClass("industry-cover-transmission");
  });

  it("keeps real uploaded cover images", () => {
    const { container } = render(
      <MemoryRouter>
        <VendorCard
          vendor={{
            id: 2,
            name: "河北测试农机配件有限公司",
            coverImage: "https://img.example.com/workshop.jpg",
            mainProducts: "液压油泵、油缸",
          }}
        />
      </MemoryRouter>,
    );

    const cover = container.querySelector(".vendor-cover") as HTMLElement;
    expect(cover.style.backgroundImage).toContain("workshop.jpg");
  });
});
