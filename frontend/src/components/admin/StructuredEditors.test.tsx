import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { uploadFile } from "../../api/admin";
import { GalleryEditor, SpecsEditor } from "./StructuredEditors";

vi.mock("../../api/admin", () => ({
  uploadFile: vi.fn(),
}));

vi.mock("./ProtectedMediaImage", () => ({
  ProtectedMediaImage: ({ assetId, src }: { assetId?: number; src: string }) => <output data-asset-id={assetId}>{src}</output>,
}));

describe("GalleryEditor", () => {
  it("uses the authenticated media preview for an uploaded asset", () => {
    render(<GalleryEditor onChange={() => undefined} value={'["/api/media/42"]'} />);

    expect(screen.getByText("/api/media/42")).toHaveAttribute("data-asset-id", "42");
  });
});

describe("SpecsEditor", () => {
  it("stores a camera-uploaded photo with its specification", async () => {
    const onChange = vi.fn();
    vi.mocked(uploadFile).mockResolvedValue({ url: "/api/media/88" } as Awaited<ReturnType<typeof uploadFile>>);
    render(<SpecsEditor onChange={onChange} value={'[{"name":"材质","value":"20CrMnTi"}]'} />);

    const photoInput = screen.getByLabelText("参数图片 1");
    expect(photoInput).toHaveAttribute("capture", "environment");
    fireEvent.change(photoInput, { target: { files: [new File(["image"], "nameplate.jpg", { type: "image/jpeg" })] } });

    await waitFor(() => expect(uploadFile).toHaveBeenCalled());
    expect(onChange).toHaveBeenLastCalledWith('[{"name":"材质","value":"20CrMnTi","image":"/api/media/88"}]');
  });

  it("keeps photo deletion separate from parameter deletion", () => {
    const onChange = vi.fn();
    render(<SpecsEditor onChange={onChange} value={'[{"name":"材质","value":"20CrMnTi","image":"/api/media/1"},{"name":"重量","value":"2.6 kg"}]'} />);

    fireEvent.click(screen.getByRole("button", { name: "删除参数图片 1" }));
    expect(onChange).toHaveBeenLastCalledWith('[{"name":"材质","value":"20CrMnTi","image":""},{"name":"重量","value":"2.6 kg"}]');

    fireEvent.click(screen.getByRole("button", { name: "删除参数 2" }));
    expect(onChange).toHaveBeenLastCalledWith('[{"name":"材质","value":"20CrMnTi","image":"/api/media/1"}]');
  });
});
