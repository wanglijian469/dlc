import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createResource, deleteResource, listResource, updateResource } from "../../api/admin";
import { AdminResourcePage } from "./AdminResourcePage";

vi.mock("../../api/admin", () => ({
  createResource: vi.fn(),
  deleteResource: vi.fn(),
  listConfigs: vi.fn(),
  listResource: vi.fn(),
  updateConfig: vi.fn(),
  updateResource: vi.fn(),
}));

const mockedCreateResource = vi.mocked(createResource);
const mockedListResource = vi.mocked(listResource);

function renderVendorsAdmin() {
  return render(
    <MemoryRouter initialEntries={["/admin/vendors"]}>
      <Routes>
        <Route element={<AdminResourcePage />} path="/admin/:resource" />
      </Routes>
    </MemoryRouter>,
  );
}

describe("AdminResourcePage vendor form", () => {
  beforeEach(() => {
    mockedListResource.mockResolvedValue([]);
    mockedCreateResource.mockResolvedValue({ id: 1, name: "测试厂商" });
    vi.mocked(updateResource).mockResolvedValue({ id: 1, name: "测试厂商" });
    vi.mocked(deleteResource).mockResolvedValue({ deleted: true });
  });

  it("renders richer vendor profile fields and submits them", async () => {
    renderVendorsAdmin();

    expect(screen.getByLabelText("厂商官网 URL")).toBeInTheDocument();
    expect(screen.getByLabelText("成立年份")).toBeInTheDocument();
    expect(screen.getByLabelText("厂房面积")).toBeInTheDocument();
    expect(screen.getByLabelText("员工规模")).toBeInTheDocument();
    expect(screen.getByLabelText("年产能")).toBeInTheDocument();
    expect(screen.getByLabelText("主要设备")).toBeInTheDocument();
    expect(screen.getByLabelText("认证资质")).toBeInTheDocument();
    expect(screen.getByLabelText("质检能力")).toBeInTheDocument();
    expect(screen.getByLabelText("供货区域")).toBeInTheDocument();
    expect(screen.getByLabelText("合作方式")).toBeInTheDocument();
    expect(screen.getByLabelText("售后服务")).toBeInTheDocument();

    fireEvent.change(screen.getByLabelText("厂商名称"), { target: { value: "浙江汉丰农机有限公司" } });
    fireEvent.change(screen.getByLabelText("厂商官网 URL"), { target: { value: "https://vendor.example.com" } });
    fireEvent.change(screen.getByLabelText("年产能"), { target: { value: "年产液压件 20 万套" } });
    fireEvent.change(screen.getByLabelText("主要设备"), { target: { value: "数控车床、液压测试台" } });
    fireEvent.click(screen.getByRole("button", { name: "新增" }));

    await waitFor(() =>
      expect(mockedCreateResource).toHaveBeenCalledWith(
        "vendors",
        expect.objectContaining({
          name: "浙江汉丰农机有限公司",
          websiteUrl: "https://vendor.example.com",
          annualCapacity: "年产液压件 20 万套",
          equipment: "数控车床、液压测试台",
        }),
      ),
    );
  });
});
