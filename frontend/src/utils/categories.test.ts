import { describe, expect, it } from "vitest";
import { hierarchicalCategoryOptions } from "./categories";

describe("hierarchicalCategoryOptions", () => {
  it("places child categories directly below their parent", () => {
    expect(hierarchicalCategoryOptions([
      { id: 2, name: "旋耕机/耕作机械", parentId: 1, sortOrder: 1 },
      { id: 1, name: "播种施肥配件", parentId: 0, sortOrder: 10 },
    ])).toEqual([
      { label: "播种施肥配件", value: 1 },
      { label: "　└ 旋耕机/耕作机械", value: 2 },
    ]);
  });
});
