import { FormEvent, useState } from "react";
import { Lock, ShieldCheck, UserRound } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { login } from "../../api/admin";

export function AdminLoginPage() {
  const [username, setUsername] = useState("admin");
  const [password, setPassword] = useState("admin123");
  const [error, setError] = useState("");
  const navigate = useNavigate();

  const submit = (event: FormEvent) => {
    event.preventDefault();
    setError("");
    login(username, password)
      .then((result) => {
        localStorage.setItem("admin_token", result.token);
        localStorage.setItem("cms_role", result.role);
        localStorage.setItem("cms_username", result.username);
        navigate(result.role === "vendor" ? "/admin/vendor-profile" : "/admin/dashboard");
      })
      .catch(() => setError("登录失败，请检查账号密码"));
  };

  return (
    <main className="admin-login">
      <section className="admin-login-visual" aria-label="后台品牌">
        <div className="admin-login-logo">农</div>
        <p>Gin-Vue-Admin 融合后台</p>
        <h1>大陆农机配件 CMS</h1>
        <span>面向厂商、产品、导航与内容配置的一体化管理台</span>
      </section>
      <form className="admin-login-card" onSubmit={submit}>
        <div className="admin-login-heading">
          <span>Welcome back</span>
          <h2>CMS 登录</h2>
        </div>
        {error && <p className="form-error">{error}</p>}
        <label>
          用户名
          <span className="admin-input-icon">
            <UserRound aria-hidden="true" size={17} />
            <input aria-label="用户名" value={username} onChange={(event) => setUsername(event.target.value)} />
          </span>
        </label>
        <label>
          密码
          <span className="admin-input-icon">
            <Lock aria-hidden="true" size={17} />
            <input aria-label="密码" type="password" value={password} onChange={(event) => setPassword(event.target.value)} />
          </span>
        </label>
        <button className="primary-btn" type="submit">
          <ShieldCheck aria-hidden="true" size={17} />
          登录
        </button>
      </form>
    </main>
  );
}
