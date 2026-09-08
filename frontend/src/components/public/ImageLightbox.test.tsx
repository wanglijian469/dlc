import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";
import { useState } from "react";
import { ImageLightbox, type LightboxImage } from "./ImageLightbox";

const images: LightboxImage[] = [
  { src: "/api/media/1?wm=7", alt: "厂房图片", caption: "厂房 · 生产车间" },
  { src: "/api/media/2?wm=7", alt: "设备图片", caption: "设备 · 加工中心" },
];

function Harness({ single = false }: { single?: boolean }) {
  const [open, setOpen] = useState(false);
  const [index, setIndex] = useState(0);
  const rows = single ? images.slice(0, 1) : images;
  return <><button type="button" onClick={() => setOpen(true)}>打开预览</button>{open && <ImageLightbox images={rows} index={index} onIndexChange={setIndex} onClose={() => setOpen(false)} />}</>;
}

describe("ImageLightbox", () => {
  afterEach(() => { cleanup(); document.body.style.overflow = ""; });

  it("cycles with controls and keyboard while keeping watermarked URLs", async () => {
    render(<Harness />);
    fireEvent.click(screen.getByRole("button", { name: "打开预览" }));
    expect(screen.getByRole("img", { name: "厂房图片" })).toHaveAttribute("src", "/api/media/1?wm=7");
    expect(screen.getByText("1 / 2")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "上一张图片" }));
    expect(screen.getByRole("img", { name: "设备图片" })).toBeInTheDocument();
    fireEvent.keyDown(document, { key: "ArrowRight" });
    expect(screen.getByRole("img", { name: "厂房图片" })).toBeInTheDocument();
    fireEvent.keyDown(document, { key: "ArrowLeft" });
    expect(screen.getByRole("img", { name: "设备图片" })).toBeInTheDocument();
  });

  it("supports mobile swiping and keeps short touches on the same image", () => {
    const { container } = render(<Harness />);
    fireEvent.click(screen.getByRole("button", { name: "打开预览" }));
    const figure = container.querySelector(".image-lightbox figure")!;
    fireEvent.touchStart(figure, { touches: [{ clientX: 220 }] });
    fireEvent.touchEnd(figure, { changedTouches: [{ clientX: 200 }] });
    expect(screen.getByRole("img", { name: "厂房图片" })).toBeInTheDocument();
    fireEvent.touchStart(figure, { touches: [{ clientX: 220 }] });
    fireEvent.touchEnd(figure, { changedTouches: [{ clientX: 120 }] });
    expect(screen.getByRole("img", { name: "设备图片" })).toBeInTheDocument();
  });

  it("locks scrolling, closes with Escape and restores opener focus", async () => {
    render(<Harness />);
    const opener = screen.getByRole("button", { name: "打开预览" });
    opener.focus();
    fireEvent.click(opener);
    await waitFor(() => expect(screen.getByRole("button", { name: "关闭图片预览" })).toHaveFocus());
    expect(document.body).toHaveStyle({ overflow: "hidden" });
    fireEvent.keyDown(document, { key: "Escape" });
    expect(screen.queryByRole("dialog", { name: "图片预览" })).not.toBeInTheDocument();
    expect(document.body.style.overflow).toBe("");
    expect(opener).toHaveFocus();
  });

  it("closes from the backdrop and handles single-image load failure", () => {
    const { container } = render(<Harness single />);
    fireEvent.click(screen.getByRole("button", { name: "打开预览" }));
    expect(screen.queryByRole("button", { name: "上一张图片" })).not.toBeInTheDocument();
    expect(screen.queryByText("1 / 1")).not.toBeInTheDocument();
    fireEvent.error(screen.getByRole("img", { name: "厂房图片" }));
    expect(screen.getByRole("status")).toHaveTextContent("图片加载失败");
    fireEvent.mouseDown(container.querySelector(".image-lightbox-backdrop")!);
    expect(screen.queryByRole("dialog", { name: "图片预览" })).not.toBeInTheDocument();
  });
});
