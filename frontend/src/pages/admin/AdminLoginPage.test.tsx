import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { register } from "../../api/admin";
import { AdminLoginPage } from "./AdminLoginPage";

vi.mock("../../api/admin", async () => {
  const actual = await vi.importActual<typeof import("../../api/admin")>("../../api/admin");
  return { ...actual, login: vi.fn(), register: vi.fn() };
});

describe("AdminLoginPage", () => {
  afterEach(cleanup);
  beforeEach(() => {
    localStorage.clear();
    vi.clearAllMocks();
  });

  it("renders login form", () => {
    render(<MemoryRouter><AdminLoginPage /></MemoryRouter>);
    expect(screen.getByText("大陆农机配件 CMS")).toBeInTheDocument();
    expect(screen.getByLabelText("用户名")).toBeInTheDocument();
    expect(screen.getByLabelText("密码")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "登录" })).toBeInTheDocument();
  });

  it("registers a vendor, stores the session, and opens the vendor CMS", async () => {
    vi.mocked(register).mockResolvedValue({ token: "token-1", username: "vendor-a", role: "vendor", vendorId: 9 });
    render(
      <MemoryRouter initialEntries={["/admin/login"]}>
        <Routes>
          <Route path="/admin/login" element={<AdminLoginPage />} />
          <Route path="/admin/vendor-profile" element={<div>厂商资料页</div>} />
        </Routes>
      </MemoryRouter>,
    );
    fireEvent.click(screen.getByRole("tab", { name: "注册" }));
    fireEvent.click(screen.getByLabelText("厂商用户"));
    fireEvent.change(screen.getByLabelText("用户名"), { target: { value: "vendor-a" } });
    fireEvent.change(screen.getByLabelText("公司全称"), { target: { value: "测试农机有限公司" } });
    fireEvent.change(screen.getByLabelText("密码"), { target: { value: "secret1" } });
    fireEvent.change(screen.getByLabelText("确认密码"), { target: { value: "secret1" } });
    fireEvent.click(screen.getByRole("button", { name: "注册并登录" }));

    await waitFor(() => expect(register).toHaveBeenCalledWith({ username: "vendor-a", password: "secret1", role: "vendor", companyName: "测试农机有限公司" }));
    expect(await screen.findByText("厂商资料页")).toBeInTheDocument();
    expect(localStorage.getItem("admin_token")).toBe("token-1");
  });

  it("validates matching registration passwords before calling the API", () => {
    render(<MemoryRouter><AdminLoginPage /></MemoryRouter>);
    fireEvent.click(screen.getByRole("tab", { name: "注册" }));
    fireEvent.change(screen.getByLabelText("用户名"), { target: { value: "member-a" } });
    fireEvent.change(screen.getByLabelText("密码"), { target: { value: "secret1" } });
    fireEvent.change(screen.getByLabelText("确认密码"), { target: { value: "secret2" } });
    fireEvent.click(screen.getByRole("button", { name: "注册并登录" }));
    expect(screen.getByText("两次输入的密码不一致")).toBeInTheDocument();
    expect(register).not.toHaveBeenCalled();
  });
});
