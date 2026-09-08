import { FormEvent, useState } from "react";
import { Building2, Lock, UserRound } from "lucide-react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { login, logoutSession, register, type LoginResponse } from "../api/admin";
import { getApiErrorMessage } from "../api/client";

type Mode = "login" | "register";
type RegistrationRole = "buyer" | "vendor";

export function AccountLoginPage() {
  const [mode, setMode] = useState<Mode>("login");
  const [role, setRole] = useState<RegistrationRole>("vendor");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [companyName, setCompanyName] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [contactName, setContactName] = useState("");
  const [phone, setPhone] = useState("");
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [params] = useSearchParams();
  const navigate = useNavigate();
  const currentRole = localStorage.getItem("cms_role");
  const currentUser = localStorage.getItem("cms_username");

  const finish = (result: LoginResponse) => {
    localStorage.setItem("cms_authenticated", "true");
    localStorage.setItem("cms_role", result.role);
    localStorage.setItem("cms_username", result.username);
    const requested = params.get("returnTo") || "";
    const safeReturn = requested.startsWith("/") && !requested.startsWith("//") && !requested.includes("\\") && !/[\u0000-\u001f]/.test(requested) ? requested : "";
    navigate(safeReturn || (result.role === "vendor" ? "/admin/vendor-workspace" : "/account/posts"), { replace: true });
  };

  const submit = (event: FormEvent) => {
    event.preventDefault();
    setError("");
    if (mode === "register" && password !== confirmPassword) return setError("两次输入的密码不一致");
    if (mode === "register" && role === "vendor" && !companyName.trim()) return setError("厂商入驻需要填写公司全称");
    setSubmitting(true);
    const action = mode === "login" ? login(username, password) : register({ username, password, role, companyName: role === "vendor" ? companyName.trim() : undefined, displayName, contactName, phone });
    action.then(finish).catch((reason) => setError(getApiErrorMessage(reason, mode === "login" ? "登录失败" : "注册失败"))).finally(() => setSubmitting(false));
  };

  if (currentUser && (currentRole === "buyer" || currentRole === "vendor")) {
    return <main className="account-auth-page"><section className="account-auth-card account-session-card"><span>当前已登录</span><h1>{currentUser}</h1><p>{currentRole === "vendor" ? "厂商账号" : "采购商账号"}</p><button className="primary-btn" onClick={() => navigate(currentRole === "vendor" ? "/admin/vendor-workspace" : "/account/posts")}>进入我的中心</button><button className="outline-btn" onClick={() => void logoutSession().finally(() => { localStorage.clear(); window.location.reload(); })}>切换账号</button></section></main>;
  }

  return <main className="account-auth-page"><section className="account-auth-intro"><Link to="/">大陆农机配件</Link><h1>连接真实需求与源头厂家</h1><p>查产品、找厂家、发布供求信息，一套账号在网页和 App 中使用。</p></section><form className="account-auth-card" onSubmit={submit}>
    <div className="account-auth-tabs"><button className={mode === "login" ? "active" : ""} type="button" onClick={() => setMode("login")}>登录</button><button className={mode === "register" ? "active" : ""} type="button" onClick={() => setMode("register")}>注册</button></div>
    <h2>{mode === "login" ? "账号登录" : "创建平台账号"}</h2>
    {mode === "register" && <div className="account-role-tabs"><button className={role === "vendor" ? "active" : ""} type="button" onClick={() => setRole("vendor")}>我是厂商</button><button className={role === "buyer" ? "active" : ""} type="button" onClick={() => setRole("buyer")}>我是采购商</button></div>}
    {error && <p className="form-error" role="alert">{error}</p>}
    <label>账号<span className="account-input"><UserRound size={18} /><input autoComplete="username" required value={username} onChange={(event) => setUsername(event.target.value)} /></span></label>
    {mode === "register" && role === "vendor" && <label>公司全称<span className="account-input"><Building2 size={18} /><input required value={companyName} onChange={(event) => setCompanyName(event.target.value)} /></span></label>}
    {mode === "register" && role === "buyer" && <><label>称呼<input value={displayName} onChange={(event) => setDisplayName(event.target.value)} /></label><label>联系人<input value={contactName} onChange={(event) => setContactName(event.target.value)} /></label><label>联系电话<input inputMode="tel" value={phone} onChange={(event) => setPhone(event.target.value)} /></label></>}
    <label>密码<span className="account-input"><Lock size={18} /><input autoComplete={mode === "login" ? "current-password" : "new-password"} minLength={6} required type="password" value={password} onChange={(event) => setPassword(event.target.value)} /></span></label>
    {mode === "register" && <label>确认密码<input minLength={6} required type="password" value={confirmPassword} onChange={(event) => setConfirmPassword(event.target.value)} /></label>}
    <button className="primary-btn account-submit" disabled={submitting}>{submitting ? "正在提交…" : mode === "login" ? "登录" : role === "buyer" ? "注册采购商账号" : "提交厂商入驻"}</button>
    <p className="account-staff-link">平台员工请前往 <Link to="/admin/login">CMS 登录</Link></p>
  </form></main>;
}
