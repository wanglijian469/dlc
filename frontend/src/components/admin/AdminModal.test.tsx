import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { AdminModal } from "./AdminModal";

describe("AdminModal", () => {
  it("traps focus, closes with Escape and restores background scrolling", () => {
    const close = vi.fn();
    render(<AdminModal label="编辑厂商" onClose={close}><input aria-label="厂商名称" /><button type="button">保存</button></AdminModal>);
    expect(screen.getByRole("dialog", { name: "编辑厂商" })).toHaveAttribute("aria-modal", "true");
    expect(document.body.style.overflow).toBe("hidden");
    fireEvent.keyDown(document, { key: "Escape" });
    expect(close).toHaveBeenCalledOnce();
  });
});
