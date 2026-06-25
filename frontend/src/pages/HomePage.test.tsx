import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";
import { HomeView } from "./HomePage";

describe("HomeView", () => {
  it("renders the required homepage sections", () => {
    render(
      <MemoryRouter>
        <HomeView
          home={{
            topMenus: [{ id: 1, name: "首页", path: "/" }],
            sidebarMenus: [{ id: 2, name: "传动配件", isDefaultOpen: true, children: [{ id: 3, name: "变速箱齿轮" }] }],
            auxiliaryMenus: [{ id: 4, name: "提交厂商", path: "/join" }],
            mobileMenus: [{ id: 5, name: "农机易损件", icon: "wrench" }],
            banner: { title: "找农机配件，查源头厂商", subtitle: "原厂品质", searchPlaceholder: "搜索配件名称", hotKeywords: ["收割机"] },
            recommendedVendors: [{ id: 1, name: "山东沃得农机配件有限公司", mainProducts: "变速箱、链条", tags: [{ id: 1, name: "源头厂商" }] }],
            moreVendors: [],
            stats: [{ label: "入驻厂商", value: "2000+" }],
            safeguards: ["平台审核", "安心认证"],
            join: { text: "入驻成为厂商，展示您的产品与实力，获取更多采购商机", buttonText: "立即入驻", path: "/join" },
          }}
        />
      </MemoryRouter>,
    );

    expect(screen.getByText("找农机配件，查源头厂商")).toBeInTheDocument();
    expect(screen.getByText("推荐厂商")).toBeInTheDocument();
    expect(screen.getByText("更多厂商")).toBeInTheDocument();
    expect(screen.queryByText("最新求购")).not.toBeInTheDocument();
  });
});
