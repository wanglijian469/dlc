import { FormEvent, useState } from "react";
import { ArrowRight, Building2, Eye, EyeOff, Lock, LogOut, ShieldCheck, UserPlus, UserRound } from "lucide-react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { login, logoutSession, register, staffLogin, type AccountRole, type LoginResponse } from "../../api/admin";
import { getApiErrorMessage } from "../../api/client";
import { trackAnalytics } from "../../analytics";

type Mode = "login" | "register";
type ExistingSession = { username: string; role: AccountRole };

export function AdminLoginPage({ staffOnly = false }: { staffOnly?: boolean }) {
  const [mode, setMode] = useState<Mode>("login");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [companyName, setCompanyName] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [existingSession, setExistingSession] = useState<ExistingSession | null>(readExistingSession);
  const navigate = useNavigate();
  const location = useLocation();
  const isRegistration = !staffOnly && mode === "register";

  const switchMode = (next: Mode) => {
    setMode(next);
    setError("");
  };

  const completeLogin = (result: LoginResponse) => {
    localStorage.setItem("cms_authenticated", "true");
    localStorage.setItem("cms_role", result.role);
    localStorage.setItem("cms_username", result.username);
    setExistingSession({ username: result.username, role: result.role });
    navigate(destinationForRole(result.role));
  };

  const logout = () => {
    void logoutSession().catch(() => undefined);
    localStorage.removeItem("cms_authenticated");
    localStorage.removeItem("cms_role");
    localStorage.removeItem("cms_username");
    setExistingSession(null);
    setMode("login");
    setError("");
  };

  const submit = (event: FormEvent) => {
    event.preventDefault();
    setError("");
    if (isRegistration && password !== confirmPassword) {
      setError("两次输入的密码不一致");
      return;
    }
    if (isRegistration && !companyName.trim()) {
      setError("厂商入驻必须填写公司全称");
      return;
    }
    setSubmitting(true);
    const action = isRegistration
      ? register({ username, password, role: "vendor", companyName: companyName.trim() })
      : (staffOnly ? staffLogin(username, password) : login(username, password));
    action
      .then((result) => {
        if (isRegistration) trackAnalytics({ eventType: "vendor_register_success", path: "/join" });
        completeLogin(result);
      })
      .catch((reason) => setError(getApiErrorMessage(reason, isRegistration ? "入驻提交失败，请检查填写内容" : "登录失败，请检查账号密码")))
      .finally(() => setSubmitting(false));
  };

  const visualTitle = staffOnly ? "让内容管理更有秩序" : "连接采购需求与源头厂商";
  const visualCopy = staffOnly ? "仅限平台管理员、内容编辑和审核人员使用。" : "厂商可提交入驻资料、维护企业信息，并在审核通过后展示到前台。";

  return (
    <main className={`admin-login ${staffOnly ? "staff-login" : "vendor-login"}`}>
      <section className="admin-login-visual" aria-label="平台介绍">
        <div className="admin-login-brand"><img alt="" className="admin-login-logo" src="/favicon.svg?v=2" /><span>大陆农机配件平台</span></div>
        <h1>{visualTitle}</h1>
        <span>{visualCopy}</span>
        <div className="admin-login-points">
          <span>真实厂商资料</span><span>审核后公开展示</span><span>安全会话保护</span>
        </div>
      </section>

      {existingSession ? (
        <section className="admin-session-card" aria-label="当前登录状态">
          <div className="session-card-icon"><ShieldCheck size={24} /></div>
          <p>当前已登录</p>
          <h2>{existingSession.username}</h2>
          <span>{roleLabel(existingSession.role)}</span>
          <div className="session-card-actions">
            <button className="primary-btn" type="button" onClick={() => navigate(destinationForRole(existingSession.role))}><ArrowRight size={17} />{destinationLabel(existingSession.role)}</button>
            <button className="session-switch" type="button" onClick={logout}><LogOut size={16} />切换账号</button>
          </div>
        </section>
      ) : (
        <form className="admin-login-card" onSubmit={submit}>
          {new URLSearchParams(location.search).get("passwordChanged") === "1" && <p className="form-success" role="status">密码已修改，请使用新密码重新登录。</p>}
          {!staffOnly && <div className="admin-auth-tabs" role="tablist" aria-label="厂商账号入口">
            <button aria-selected={mode === "login"} className={mode === "login" ? "active" : ""} role="tab" type="button" onClick={() => switchMode("login")}>厂商登录</button>
            <button aria-selected={mode === "register"} className={mode === "register" ? "active" : ""} role="tab" type="button" onClick={() => switchMode("register")}>厂商入驻</button>
          </div>}
          <div className="admin-login-heading">
            <h2>{staffOnly ? "CMS 员工登录" : isRegistration ? "提交厂商入驻" : "厂商账号登录"}</h2>
            <p>{staffOnly ? "使用平台分配的员工账号登录。" : isRegistration ? "提交后可完善企业资料，审核通过后在前台公开展示。" : "登录后维护厂商资料与产品信息。"}</p>
          </div>
          {error && <p className="form-error" role="alert">{error}</p>}
          <label>
            账号名称
            <span className="admin-input-icon"><UserRound aria-hidden="true" size={18} /><input aria-label="用户名" autoComplete="username" placeholder="请输入账号名称" required value={username} onChange={(event) => setUsername(event.target.value)} /></span>
          </label>
          {isRegistration && <label>
            公司全称
            <span className="admin-input-icon"><Building2 aria-hidden="true" size={18} /><input aria-label="公司全称" autoComplete="organization" placeholder="请填写营业执照上的公司全称" required value={companyName} onChange={(event) => setCompanyName(event.target.value)} /></span>
          </label>}
          <label>
            密码{isRegistration && "（至少 6 位）"}
            <span className="admin-input-icon"><Lock aria-hidden="true" size={18} /><input aria-label="密码" autoComplete={isRegistration ? "new-password" : "current-password"} minLength={isRegistration ? 6 : undefined} placeholder="请输入密码" required type={showPassword ? "text" : "password"} value={password} onChange={(event) => setPassword(event.target.value)} /><button aria-label={showPassword ? "隐藏密码" : "显示密码"} className="password-toggle" type="button" onClick={() => setShowPassword((value) => !value)}>{showPassword ? <EyeOff size={17} /> : <Eye size={17} />}</button></span>
          </label>
          {isRegistration && <label>
            确认密码
            <span className="admin-input-icon"><Lock aria-hidden="true" size={18} /><input aria-label="确认密码" autoComplete="new-password" minLength={6} placeholder="请再次输入密码" required type={showPassword ? "text" : "password"} value={confirmPassword} onChange={(event) => setConfirmPassword(event.target.value)} /></span>
          </label>}
          <button className="primary-btn admin-login-submit" disabled={submitting} type="submit">
            {isRegistration ? <UserPlus aria-hidden="true" size={18} /> : <ShieldCheck aria-hidden="true" size={18} />}
            {submitting ? "正在提交…" : isRegistration ? "提交入驻并登录" : "登录厂商工作台"}
          </button>
          {staffOnly ? <p className="admin-login-switch">厂商账号请前往 <Link to="/account/login">厂商登录</Link></p> : <p className="admin-login-switch">CMS 员工请从 <Link to="/admin/login">后台入口登录</Link></p>}
        </form>
      )}
    </main>
  );
}

function readExistingSession(): ExistingSession | null {
  const username = localStorage.getItem("cms_username");
  const role = localStorage.getItem("cms_role") as AccountRole | null;
  if (localStorage.getItem("cms_authenticated") !== "true" || !username || !role) return null;
  if (role !== "vendor" && role !== "admin" && role !== "editor" && role !== "reviewer") return null;
  return { username, role };
}

function destinationForRole(role: AccountRole) {
  return role === "vendor" ? "/admin/vendor-workspace" : "/admin/dashboard";
}

function destinationLabel(role: AccountRole) {
  return role === "vendor" ? "进入厂商工作台" : "进入管理中心";
}

function roleLabel(role: AccountRole) {
  return ({ vendor: "厂商账号", admin: "管理员", editor: "内容编辑", reviewer: "内容审核" } as Record<AccountRole, string>)[role];
}
