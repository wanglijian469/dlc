import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";
import { trackAnalytics } from "../../analytics";
import { VendorContactActions } from "./VendorContactActions";
vi.mock("../../analytics",()=>({trackAnalytics:vi.fn()}));
vi.mock("../../api/public",()=>({getVendorContact:vi.fn().mockRejectedValue({response:{status:401}}),getVendorContactQRCode:vi.fn()}));
afterEach(()=>{cleanup();vi.clearAllMocks();});
const vendor={id:7,name:"本厂",wechat:"public-account",wechatPublic:true,wechatQrCode:"/api/media/7"};
describe("contact analytics",()=>{
 it("counts a successful copy, not merely opening the contact dialog",async()=>{
  Object.defineProperty(navigator,"clipboard",{configurable:true,value:{writeText:vi.fn().mockResolvedValue(undefined)}});
  render(<MemoryRouter><VendorContactActions vendor={vendor} path="/v/factory/products/8"/></MemoryRouter>);
  fireEvent.click(screen.getByRole("button",{name:"微信联系"}));expect(trackAnalytics).not.toHaveBeenCalled();
  fireEvent.click(screen.getByRole("button",{name:"复制微信号"}));
  await waitFor(()=>expect(trackAnalytics).toHaveBeenCalledWith({eventType:"contact_wechat_copy",path:"/v/factory/products/8"}));
 });
 it("does not count failed copies and only counts a QR after the image loads",async()=>{
  Object.defineProperty(navigator,"clipboard",{configurable:true,value:{writeText:vi.fn().mockRejectedValue(new Error("denied"))}});
  render(<MemoryRouter><VendorContactActions vendor={vendor} path="/v/factory"/></MemoryRouter>);
  fireEvent.click(screen.getByRole("button",{name:"微信联系"}));fireEvent.click(screen.getByRole("button",{name:"复制微信号"}));
  await screen.findByText("复制失败，请长按微信号复制");expect(trackAnalytics).not.toHaveBeenCalled();
  fireEvent.load(screen.getByAltText("本厂 微信二维码"));
  expect(trackAnalytics).toHaveBeenCalledWith({eventType:"contact_wechat_qr_view",path:"/v/factory"});
 });
 it("never reveals a non-public phone locally and preserves promotion source on login return",async()=>{
  render(<MemoryRouter initialEntries={["/v/factory?from=qr"]}><VendorContactActions vendor={{id:7,name:"本厂",phone:"13900000000",phonePublic:false,phoneAvailable:true}} path="/v/factory"/></MemoryRouter>);
  expect(screen.queryByText(/13900000000/)).not.toBeInTheDocument();
  fireEvent.click(screen.getByRole("button",{name:"查看联系电话"}));
  const login=await screen.findByRole("link",{name:"前往登录"});expect(decodeURIComponent(login.getAttribute("href")!)).toContain("/v/factory?from=qr&contact=phone");
  expect(trackAnalytics).not.toHaveBeenCalled();
 });
});
