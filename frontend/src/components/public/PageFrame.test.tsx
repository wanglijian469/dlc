import { cleanup, fireEvent, render, waitFor, within } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { HomePayload } from "../../types/api";
import { getHome } from "../../api/public";
import { PageFrame } from "./PageFrame";

vi.mock("../../api/public", () => ({
  getHome: vi.fn(),
}));

const homePayload: HomePayload = {
  topMenus: [],
  sidebarMenus: [
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
      path: "/products?keyword=传动配件",
      isDefaultOpen: true,
      children: [{ id: 3, name: "变速箱齿轮", icon: "dot", path: "/products?keyword=变速箱齿轮" }],
    },
  ],
  auxiliaryMenus: [{ id: 10, name: "提交厂商", path: "/submit" }],
  mobileMenus: [],
  banner: { title: "" },
  recommendedVendors: [],
  moreVendors: [],
  stats: [],
  safeguards: [],
  join: { text: "", buttonText: "", path: "/" },
};

describe("PageFrame", () => {
  beforeEach(() => {
    vi.mocked(getHome).mockResolvedValue(homePayload);
  });

  afterEach(() => cleanup());

  it.each([
    ["/products", "配件产品"],
    ["/vendors", "厂商目录"],
    ["/service", "加工服务"],
    ["/purchase", "采购信息"],
  ])("renders Chinese main navigation and highlights %s", (path, activeLabel) => {
    const { container } = render(
      <MemoryRouter initialEntries={[path]}>
        <PageFrame title="配件产品">
          <div>列表内容</div>
        </PageFrame>
      </MemoryRouter>,
    );

    const topNav = container.querySelector(".top-nav") as HTMLElement;
    expect(topNav.querySelector("a.active")).toHaveTextContent(activeLabel);
    expect(topNav).toHaveTextContent("配件产品");
    expect(topNav).toHaveTextContent("厂商目录");
    expect(topNav).toHaveTextContent("加工服务");
    expect(topNav).toHaveTextContent("采购信息");
  });

  it("renders the same collapsed sidebar categories on desktop subpages and lets parent rows toggle", async () => {
    const { container } = render(
      <MemoryRouter initialEntries={["/products"]}>
        <PageFrame title="配件产品">
          <div>产品列表</div>
        </PageFrame>
      </MemoryRouter>,
    );

    await waitFor(() => expect(getHome).toHaveBeenCalled());

    const sidebar = container.querySelector(".sidebar") as HTMLElement;
    const parent = within(sidebar).getByRole("button", { name: "传动配件" });
    expect(parent).toBeInTheDocument();
    expect(within(sidebar).queryByRole("link", { name: /变速箱齿轮/ })).not.toBeInTheDocument();
    fireEvent.click(parent);
    expect(within(sidebar).getByRole("link", { name: /变速箱齿轮/ })).toBeInTheDocument();
    expect(within(sidebar).getByRole("link", { name: "提交厂商" })).toBeInTheDocument();
  });
});
