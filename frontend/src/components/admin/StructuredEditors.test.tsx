import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { GalleryEditor } from "./StructuredEditors";

vi.mock("./ProtectedMediaImage", () => ({
  ProtectedMediaImage: ({ assetId, src }: { assetId?: number; src: string }) => <output data-asset-id={assetId}>{src}</output>,
}));

describe("GalleryEditor", () => {
  it("uses the authenticated media preview for an uploaded asset", () => {
    render(<GalleryEditor onChange={() => undefined} value={'["/api/media/42"]'} />);

    expect(screen.getByText("/api/media/42")).toHaveAttribute("data-asset-id", "42");
  });
});
