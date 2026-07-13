import { cleanup, render, screen, within } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { AdminLayout } from "./AdminLayout";

describe("AdminLayout navigation", () => {
  beforeEach(() => {
    localStorage.clear();
    localStorage.setItem("cms_role", "admin");
  });

  afterEach(() => cleanup());

  it("separates daily business work from platform configuration", () => {
    render(<MemoryRouter><AdminLayout title="控制台"><div>内容</div></AdminLayout></MemoryRouter>);

    const businessGroup = screen.getByText("业务内容").closest(".admin-nav-group") as HTMLElement;
    const configGroup = screen.getByText("基础配置").closest(".admin-nav-group") as HTMLElement;

    for (const name of ["厂商信息", "资料审核", "产品审核", "配件产品"]) {
      expect(within(businessGroup).getByRole("link", { name: `导航：${name}` })).toBeInTheDocument();
    }
    for (const name of ["导航菜单", "厂商标签", "配件分类", "Banner 管理", "内容页面", "友情链接"]) {
      expect(within(configGroup).getByRole("link", { name: `导航：${name}` })).toBeInTheDocument();
    }
  });
});
