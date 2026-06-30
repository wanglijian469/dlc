import { describe, expect, it } from "vitest";
import { getProductIndustryKind, getValidCoverImage, getVendorIndustryKind } from "./IndustryCover";

describe("industry cover matching", () => {
  it("treats old Factory and Parts dummy URLs as invalid cover images", () => {
    expect(getValidCoverImage("https://dummyimage.com/600x320/eaf3ff/1f2a3d&text=Factory+01")).toBe("");
    expect(getValidCoverImage("https://dummyimage.com/480x300/eaf3ff/1f2a3d&text=Parts")).toBe("");
    expect(getValidCoverImage("https://img.example.com/real-part.jpg")).toBe("https://img.example.com/real-part.jpg");
  });

  it("matches products to the nine main menu industry kinds by category id", () => {
    expect(getProductIndustryKind({ id: 1, name: "滤芯套件", categoryId: 1 })).toBe("wearing");
    expect(getProductIndustryKind({ id: 2, name: "齿轮箱", categoryId: 2 })).toBe("transmission");
    expect(getProductIndustryKind({ id: 3, name: "履带总成", categoryId: 4 })).toBe("chassis");
    expect(getProductIndustryKind({ id: 4, name: "液压油缸", categoryId: 5 })).toBe("hydraulic");
    expect(getProductIndustryKind({ id: 5, name: "发动机滤芯", categoryId: 6 })).toBe("engine");
    expect(getProductIndustryKind({ id: 6, name: "制动盘", categoryId: 7 })).toBe("brake");
    expect(getProductIndustryKind({ id: 7, name: "线束插头", categoryId: 8 })).toBe("electrical");
    expect(getProductIndustryKind({ id: 8, name: "割台刀片", categoryId: 9 })).toBe("harvester");
    expect(getProductIndustryKind({ id: 9, name: "排种器", categoryId: 10 })).toBe("seeding");
  });

  it("matches vendors to main menu industry kinds from main products", () => {
    expect(getVendorIndustryKind({ id: 1, name: "测试厂商", mainProducts: "变速箱、链条、轴承" })).toBe("transmission");
    expect(getVendorIndustryKind({ id: 2, name: "测试厂商", mainProducts: "液压油泵、高压油管" })).toBe("hydraulic");
  });
});
