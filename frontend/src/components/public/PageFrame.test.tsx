import { render } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";
import { PageFrame } from "./PageFrame";

describe("PageFrame", () => {
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
});
