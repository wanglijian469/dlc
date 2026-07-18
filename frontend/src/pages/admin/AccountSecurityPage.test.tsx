import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { changePassword } from "../../api/admin";
import { AccountSecurityPage } from "./AccountSecurityPage";

vi.mock("../../api/admin", () => ({ changePassword: vi.fn(), logoutSession: vi.fn() }));
const mockedChangePassword = vi.mocked(changePassword);

describe("AccountSecurityPage", () => {
  beforeEach(() => {
    localStorage.setItem("cms_authenticated", "true");
    localStorage.setItem("cms_role", "vendor");
    localStorage.setItem("cms_username", "vendor-a");
    mockedChangePassword.mockResolvedValue({ changed: true, reauthenticate: true });
  });
  afterEach(() => { cleanup(); localStorage.clear(); vi.clearAllMocks(); });

  it("validates confirmation before submitting", () => {
    render(<MemoryRouter><AccountSecurityPage /></MemoryRouter>);
    fireEvent.change(screen.getByLabelText("当前密码"), { target: { value: "old-password" } });
    fireEvent.change(screen.getByLabelText(/^新密码/), { target: { value: "new-password" } });
    fireEvent.change(screen.getByLabelText("确认新密码"), { target: { value: "different-password" } });
    fireEvent.click(screen.getByRole("button", { name: "修改密码并重新登录" }));
    expect(screen.getByRole("alert")).toHaveTextContent("两次输入的新密码不一致");
    expect(mockedChangePassword).not.toHaveBeenCalled();
  });

  it("clears the local session and requires login after success", async () => {
    render(<MemoryRouter initialEntries={["/admin/account-security"]}><Routes><Route path="/admin/account-security" element={<AccountSecurityPage />} /><Route path="/account/login" element={<div>请重新登录</div>} /></Routes></MemoryRouter>);
    fireEvent.change(screen.getByLabelText("当前密码"), { target: { value: "old-password" } });
    fireEvent.change(screen.getByLabelText(/^新密码/), { target: { value: "new-password" } });
    fireEvent.change(screen.getByLabelText("确认新密码"), { target: { value: "new-password" } });
    fireEvent.click(screen.getByRole("button", { name: "修改密码并重新登录" }));
    await screen.findByText("请重新登录");
    expect(mockedChangePassword).toHaveBeenCalledWith("old-password", "new-password");
    expect(localStorage.getItem("cms_authenticated")).toBeNull();
  });
});
