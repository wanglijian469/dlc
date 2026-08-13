import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { register } from "../api/admin";
import { AccountLoginPage } from "./AccountLoginPage";

vi.mock("../api/admin", () => ({ login: vi.fn(), logoutSession: vi.fn(), register: vi.fn() }));

describe("AccountLoginPage", () => {
  beforeEach(() => { localStorage.clear(); vi.mocked(register).mockResolvedValue({ username: "buyer-a", role: "buyer" }); });
  it("registers a buyer and enters their posts", async () => {
    render(<MemoryRouter initialEntries={["/account/login"]}><Routes><Route path="/account/login" element={<AccountLoginPage />} /><Route path="/account/posts" element={<div>我的发布页面</div>} /></Routes></MemoryRouter>);
    fireEvent.click(screen.getByRole("button", { name: "注册" }));
	fireEvent.click(screen.getByRole("button", { name: "我是采购商" }));
    fireEvent.change(screen.getByLabelText("账号"), { target: { value: "buyer-a" } });
    const passwords = screen.getAllByLabelText(/密码/);
    fireEvent.change(passwords[0], { target: { value: "secret1" } });
    fireEvent.change(passwords[1], { target: { value: "secret1" } });
    fireEvent.click(screen.getByRole("button", { name: "注册采购商账号" }));
    await waitFor(() => expect(register).toHaveBeenCalledWith(expect.objectContaining({ username: "buyer-a", role: "buyer" })));
    expect(await screen.findByText("我的发布页面")).toBeInTheDocument();
  });

  it("puts the vendor role first and selects it by default", () => {
	render(<MemoryRouter initialEntries={["/account/login"]}><AccountLoginPage /></MemoryRouter>);
	fireEvent.click(screen.getByRole("button", { name: "注册" }));
	const roles = screen.getByRole("button", { name: "我是厂商" }).parentElement?.querySelectorAll("button");
	expect(roles?.[0]).toHaveTextContent("我是厂商");
	expect(screen.getByRole("button", { name: "我是厂商" })).toHaveClass("active");
	expect(screen.getByLabelText("公司全称")).toBeInTheDocument();
	fireEvent.click(screen.getByRole("button", { name: "我是采购商" }));
	expect(screen.getByLabelText("称呼")).toBeInTheDocument();
  });
});
