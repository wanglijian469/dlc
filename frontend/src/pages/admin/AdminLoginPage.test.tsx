import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { logoutSession, register, staffLogin } from "../../api/admin";
import { AdminLoginPage } from "./AdminLoginPage";

vi.mock("../../api/admin", async () => {
  const actual = await vi.importActual<typeof import("../../api/admin")>("../../api/admin");
  return { ...actual, login: vi.fn(), staffLogin: vi.fn(), register: vi.fn(), logoutSession: vi.fn() };
});

describe("AdminLoginPage", () => {
  afterEach(cleanup);
  beforeEach(() => {
    localStorage.clear();
    vi.clearAllMocks();
  });

  it("renders a vendor-only login form", () => {
    render(<MemoryRouter><AdminLoginPage /></MemoryRouter>);
    expect(screen.getByText("厂商账号登录")).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: "厂商登录" })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: "厂商入驻" })).toBeInTheDocument();
    expect(screen.queryByText("普通用户")).not.toBeInTheDocument();
  });

  it("registers a vendor and opens the vendor workspace", async () => {
    vi.mocked(register).mockResolvedValue({ username: "vendor-a", role: "vendor", vendorId: 9 });
    render(
      <MemoryRouter initialEntries={["/account/login"]}>
        <Routes>
          <Route path="/account/login" element={<AdminLoginPage />} />
          <Route path="/admin/vendor-workspace" element={<div>厂商资料页</div>} />
        </Routes>
      </MemoryRouter>,
    );
    fireEvent.click(screen.getByRole("tab", { name: "厂商入驻" }));
    fireEvent.change(screen.getByLabelText("用户名"), { target: { value: "vendor-a" } });
    fireEvent.change(screen.getByLabelText("公司全称"), { target: { value: "测试农机有限公司" } });
    fireEvent.change(screen.getByLabelText("密码"), { target: { value: "secret1" } });
    fireEvent.change(screen.getByLabelText("确认密码"), { target: { value: "secret1" } });
    fireEvent.click(screen.getByRole("button", { name: "提交入驻并登录" }));

    await waitFor(() => expect(register).toHaveBeenCalledWith({ username: "vendor-a", password: "secret1", role: "vendor", companyName: "测试农机有限公司" }));
    expect(await screen.findByText("厂商资料页")).toBeInTheDocument();
    expect(localStorage.getItem("cms_authenticated")).toBe("true");
  });

  it("validates matching onboarding passwords before calling the API", () => {
    render(<MemoryRouter><AdminLoginPage /></MemoryRouter>);
    fireEvent.click(screen.getByRole("tab", { name: "厂商入驻" }));
    fireEvent.change(screen.getByLabelText("用户名"), { target: { value: "vendor-a" } });
    fireEvent.change(screen.getByLabelText("公司全称"), { target: { value: "测试农机有限公司" } });
    fireEvent.change(screen.getByLabelText("密码"), { target: { value: "secret1" } });
    fireEvent.change(screen.getByLabelText("确认密码"), { target: { value: "secret2" } });
    fireEvent.click(screen.getByRole("button", { name: "提交入驻并登录" }));
    expect(screen.getByText("两次输入的密码不一致")).toBeInTheDocument();
    expect(register).not.toHaveBeenCalled();
  });

  it("shows a separate session card and allows account switching", () => {
    localStorage.setItem("cms_authenticated", "true");
    localStorage.setItem("cms_role", "vendor");
    localStorage.setItem("cms_username", "vendor-a");
    vi.mocked(logoutSession).mockResolvedValue({ loggedOut: true });
    render(<MemoryRouter><AdminLoginPage /></MemoryRouter>);
    expect(screen.getByText("当前已登录")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "进入厂商工作台" })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "切换账号" }));
    expect(screen.getByText("厂商账号登录")).toBeInTheDocument();
    expect(logoutSession).toHaveBeenCalled();
  });

  it("keeps the staff route separate from vendor onboarding", () => {
    render(<MemoryRouter><AdminLoginPage staffOnly /></MemoryRouter>);
    expect(screen.getByText("CMS 员工登录")).toBeInTheDocument();
    expect(screen.queryByRole("tab", { name: "厂商入驻" })).not.toBeInTheDocument();
    expect(staffLogin).not.toHaveBeenCalled();
  });
});
