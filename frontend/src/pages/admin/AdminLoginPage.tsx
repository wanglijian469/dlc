import { FormEvent, useState } from "react";
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
        navigate("/admin/dashboard");
      })
      .catch(() => setError("登录失败，请检查账号密码"));
  };

  return (
    <main className="admin-login">
      <form onSubmit={submit}>
        <h1>后台登录</h1>
        {error && <p className="form-error">{error}</p>}
        <label>
          用户名
          <input aria-label="用户名" value={username} onChange={(event) => setUsername(event.target.value)} />
        </label>
        <label>
          密码
          <input aria-label="密码" type="password" value={password} onChange={(event) => setPassword(event.target.value)} />
        </label>
        <button className="primary-btn" type="submit">
          登录
        </button>
      </form>
    </main>
  );
}
