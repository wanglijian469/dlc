import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";
import { AdminLoginPage } from "./AdminLoginPage";

describe("AdminLoginPage", () => {
  it("renders login form", () => {
    render(
      <MemoryRouter>
        <AdminLoginPage />
      </MemoryRouter>,
    );
    expect(screen.getByLabelText("用户名")).toBeInTheDocument();
    expect(screen.getByLabelText("密码")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "登录" })).toBeInTheDocument();
  });
});
