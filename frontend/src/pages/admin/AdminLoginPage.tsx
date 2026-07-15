import { FormEvent, useState } from "react";
import { Building2, Lock, ShieldCheck, UserPlus, UserRound } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { login, register, type AccountRole, type LoginResponse } from "../../api/admin";
import { getApiErrorMessage } from "../../api/client";
import { trackAnalytics } from "../../analytics";

type Mode = "login" | "register";
type RegistrationRole = "user" | "vendor";

export function AdminLoginPage() {
  const [mode, setMode] = useState<Mode>("login");
  const [username, setUsername] = useState("admin");
  const [password, setPassword] = useState("admin123");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [role, setRole] = useState<RegistrationRole>("user");
  const [companyName, setCompanyName] = useState("");
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [existingSession, setExistingSession] = useState(() => {
    const token = localStorage.getItem("admin_token");
    const storedUsername = localStorage.getItem("cms_username");
    return token && storedUsername ? storedUsername : "";
  });
  const navigate = useNavigate();

  const switchMode = (next: Mode) => {
    setMode(next);
    setError("");
    if (next === "register" && username === "admin" && password === "admin123") {
      setUsername("");
      setPassword("");
    }
  };

  const completeLogin = (result: LoginResponse) => {
    localStorage.setItem("admin_token", result.token);
    localStorage.setItem("cms_role", result.role);
    localStorage.setItem("cms_username", result.username);
    setExistingSession(result.username);
    navigate(destinationForRole(result.role));
  };

  const logout = () => {
    localStorage.removeItem("admin_token");
    localStorage.removeItem("cms_role");
    localStorage.removeItem("cms_username");
    setExistingSession("");
  };

  const submit = (event: FormEvent) => {
    event.preventDefault();
    setError("");
    if (mode === "register" && password !== confirmPassword) {
      setError("两次输入的密码不一致");
      return;
    }
    if (mode === "register" && role === "vendor" && !companyName.trim()) {
      setError("厂商用户必须填写公司全称");
      return;
    }
    setSubmitting(true);
    const action = mode === "login"
      ? login(username, password)
      : register({ username, password, role, companyName: role === "vendor" ? companyName.trim() : undefined });
    action
      .then((result) => { if (mode === "register") trackAnalytics({ eventType: "vendor_register_success", path: "/join" }); completeLogin(result); })
      .catch((reason) => setError(getApiErrorMessage(reason, mode === "login" ? "登录失败，请检查账号密码" : "注册失败，请检查填写内容")))
      .finally(() => setSubmitting(false));
  };

  return (
    <main className="admin-login">
      <section className="admin-login-visual" aria-label="后台品牌">
        <div className="admin-login-logo">农</div>
        <h1>大陆农机配件 CMS</h1>
        <span>普通用户可登录查看完整联系方式，厂商用户可维护并提交企业资料。</span>
      </section>
      <form className="admin-login-card" onSubmit={submit}>
        {existingSession && <div className="existing-session"><span>当前已登录：{existingSession}</span><button type="button" onClick={logout}>退出登录</button></div>}
        <div className="admin-auth-tabs" role="tablist" aria-label="账号入口">
          <button aria-selected={mode === "login"} className={mode === "login" ? "active" : ""} role="tab" type="button" onClick={() => switchMode("login")}>登录</button>
          <button aria-selected={mode === "register"} className={mode === "register" ? "active" : ""} role="tab" type="button" onClick={() => switchMode("register")}>注册</button>
        </div>
        <div className="admin-login-heading">
          <span>{mode === "login" ? "Welcome back" : "Create account"}</span>
          <h2>{mode === "login" ? "账号登录" : "账号注册"}</h2>
        </div>
        {error && <p className="form-error">{error}</p>}
        {mode === "register" && (
          <fieldset className="registration-role-picker">
            <legend>注册类型</legend>
            <label><input checked={role === "user"} name="registration-role" type="radio" onChange={() => setRole("user")} />普通用户</label>
            <label><input checked={role === "vendor"} name="registration-role" type="radio" onChange={() => setRole("vendor")} />厂商用户</label>
          </fieldset>
        )}
        <label>
          用户名
          <span className="admin-input-icon">
            <UserRound aria-hidden="true" size={17} />
            <input aria-label="用户名" autoComplete="username" required value={username} onChange={(event) => setUsername(event.target.value)} />
          </span>
        </label>
        {mode === "register" && role === "vendor" && (
          <label>
            公司全称
            <span className="admin-input-icon">
              <Building2 aria-hidden="true" size={17} />
              <input aria-label="公司全称" required value={companyName} onChange={(event) => setCompanyName(event.target.value)} />
            </span>
          </label>
        )}
        <label>
          密码{mode === "register" && "（至少 6 位）"}
          <span className="admin-input-icon">
            <Lock aria-hidden="true" size={17} />
            <input aria-label="密码" autoComplete={mode === "login" ? "current-password" : "new-password"} minLength={mode === "register" ? 6 : undefined} required type="password" value={password} onChange={(event) => setPassword(event.target.value)} />
          </span>
        </label>
        {mode === "register" && (
          <label>
            确认密码
            <span className="admin-input-icon">
              <Lock aria-hidden="true" size={17} />
              <input aria-label="确认密码" autoComplete="new-password" minLength={6} required type="password" value={confirmPassword} onChange={(event) => setConfirmPassword(event.target.value)} />
            </span>
          </label>
        )}
        <button className="primary-btn" disabled={submitting} type="submit">
          {mode === "login" ? <ShieldCheck aria-hidden="true" size={17} /> : <UserPlus aria-hidden="true" size={17} />}
          {submitting ? "正在提交…" : mode === "login" ? "登录" : "注册并登录"}
        </button>
        {mode === "register" && role === "vendor" && <p className="registration-hint">公司将同步进入后台厂商资源，完善资料并通过审核后才会在前台展示。</p>}
      </form>
    </main>
  );
}

function destinationForRole(role: AccountRole) {
  if (role === "admin") return "/admin/dashboard";
  if (role === "vendor") return "/admin/vendor-profile";
  return "/";
}
