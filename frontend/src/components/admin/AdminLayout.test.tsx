import { cleanup, fireEvent, render, screen, within } from "@testing-library/react";
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
    render(<MemoryRouter initialEntries={["/admin/configs?section=site"]}><AdminLayout title="控制台"><div>内容</div></AdminLayout></MemoryRouter>);

    const businessGroup = screen.getByText("业务内容").closest(".admin-nav-group") as HTMLElement;
    const configGroup = screen.getByText("基础配置").closest(".admin-nav-group") as HTMLElement;

    for (const name of ["厂商信息", "资料审核", "产品审核", "配件产品"]) {
      expect(within(businessGroup).getByRole("link", { name: `导航：${name}` })).toBeInTheDocument();
    }
    for (const name of ["页面与快捷导航", "厂商标签", "配件分类", "Banner 管理", "页面与行业文章", "友情链接"]) {
      expect(within(configGroup).getByRole("link", { name: `导航：${name}` })).toBeInTheDocument();
    }
    const systemGroup = screen.getByText("系统管理").closest(".admin-nav-group") as HTMLElement;
    const configSubmenu = within(systemGroup).getByLabelText("平台配置子菜单");
    for (const name of ["站点与页脚", "首页展示", "主题样式"]) {
      expect(within(configSubmenu).getByRole("link", { name: `导航：${name}` })).toBeInTheDocument();
    }
  });

  it("highlights only the exact platform configuration submenu", () => {
    render(<MemoryRouter initialEntries={["/admin/configs?section=home"]}><AdminLayout title="平台配置"><div>内容</div></AdminLayout></MemoryRouter>);

    expect(screen.getByRole("button", { name: "平台配置" })).not.toHaveClass("active");
    expect(screen.getByRole("link", { name: "导航：站点与页脚" })).not.toHaveClass("active");
    expect(screen.getByRole("link", { name: "导航：首页展示" })).toHaveClass("active");
    expect(screen.getByRole("link", { name: "导航：主题样式" })).not.toHaveClass("active");
  });

  it("collapses and expands the platform configuration submenu", () => {
    render(<MemoryRouter><AdminLayout title="控制台"><div>内容</div></AdminLayout></MemoryRouter>);

    const trigger = screen.getByRole("button", { name: "平台配置" });
    expect(trigger).toHaveAttribute("aria-expanded", "false");
    expect(screen.queryByLabelText("平台配置子菜单")).not.toBeInTheDocument();

    fireEvent.click(trigger);
    expect(trigger).toHaveAttribute("aria-expanded", "true");
    expect(screen.getByLabelText("平台配置子菜单")).toBeInTheDocument();

    fireEvent.click(trigger);
    expect(screen.queryByLabelText("平台配置子菜单")).not.toBeInTheDocument();
  });
});
