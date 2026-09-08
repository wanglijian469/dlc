import { cleanup, render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";
import { getShowroomProduct } from "../api/workspace";
import { ShowroomProductPage } from "./ShowroomProductPage";
vi.mock("../api/workspace",()=>({getShowroomProduct:vi.fn()}));
vi.mock("../components/public/PageFrame",()=>({PageFrame:({children}: {children:React.ReactNode})=><main>{children}</main>}));
vi.mock("../components/public/PromotionTools",()=>({PromotionTools:()=>null}));
vi.mock("../components/public/VendorContactActions",()=>({VendorContactActions:()=>null}));
afterEach(()=>{cleanup();vi.clearAllMocks();});
function renderPage(){return render(<MemoryRouter initialEntries={["/v/factory/products/8"]}><Routes><Route path="/v/:slug/products/:supplierId" element={<ShowroomProductPage/>}/></Routes></MemoryRouter>);}
describe("own vendor product",()=>{
 it("does not borrow a catalog or another vendor image when own image is missing",async()=>{
  vi.mocked(getShowroomProduct).mockResolvedValue({id:8,vendorId:7,vendor:{id:7,name:"本厂名称"},vendorProductName:"本厂油缸",vendorModel:"A-1",product:{id:9,image:"/other-supplier.jpg"}} as never);
  const {container}=renderPage();
  expect(await screen.findByText("本厂产品图片待补充")).toBeInTheDocument();
  expect(getShowroomProduct).toHaveBeenCalledWith("factory","8");
  expect(screen.getByRole("link",{name:"本厂名称 · 企业展厅"})).toHaveAttribute("href","/v/factory");
  expect(container.querySelector('img[src="/other-supplier.jpg"]')).toBeNull();
 });
 it("shows unavailable state when the server rejects an unowned/unpublished detail",async()=>{
  vi.mocked(getShowroomProduct).mockRejectedValue(new Error("not public"));
  renderPage();expect(await screen.findByText("本厂产品不存在或暂未公开")).toBeInTheDocument();
  expect(screen.queryByRole("button",{name:"放大本厂产品主图"})).not.toBeInTheDocument();
 });
});
