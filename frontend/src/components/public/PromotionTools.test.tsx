import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { PromotionTools, promotionURL } from "./PromotionTools";
vi.mock("qrcode", () => ({default:{toDataURL:vi.fn().mockResolvedValue("data:image/png;base64,QA")}}));
afterEach(cleanup);
describe("public promotion",()=>{
 it("strips contact and unrelated query data and rejects private/external paths",()=>{
  expect(promotionURL("/v/factory/products/7?contact=wechat&private=secret#details","qr")).toBe(window.location.origin+"/v/factory/products/7?from=qr");
  for(const path of ["/api/media/7","/admin/capture","https://external.example/v/factory"])expect(()=>promotionURL(path,"share")).toThrow();
 });
 it("generates QR locally, copies the public URL, and restores focus",async()=>{
  const writeText=vi.fn().mockResolvedValue(undefined);
  Object.defineProperty(navigator,"clipboard",{configurable:true,value:{writeText}});
  render(<PromotionTools path="/v/factory" title="厂商公开名称"/>);
  const opener=screen.getByRole("button",{name:"分享推广"});opener.focus();fireEvent.click(opener);
  expect(await screen.findByAltText("推广二维码")).toHaveAttribute("src","data:image/png;base64,QA");
  fireEvent.click(screen.getByRole("button",{name:"复制链接"}));
  await waitFor(()=>expect(writeText).toHaveBeenCalledWith(window.location.origin+"/v/factory?from=share"));
  fireEvent.keyDown(document,{key:"Escape"});
  expect(screen.queryByRole("dialog")).not.toBeInTheDocument();expect(opener).toHaveFocus();expect(document.body.style.overflow).toBe("");
 });
});
