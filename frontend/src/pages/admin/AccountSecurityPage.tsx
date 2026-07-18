import { FormEvent, useState } from "react";
import { KeyRound, ShieldCheck } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { changePassword } from "../../api/admin";
import { getApiErrorMessage } from "../../api/client";
import { AdminLayout } from "../../components/admin/AdminLayout";

export function AccountSecurityPage() {
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [error, setError] = useState("");
  const [saving, setSaving] = useState(false);
  const navigate = useNavigate();

  const submit = (event: FormEvent) => {
    event.preventDefault();
    setError("");
    const byteLength = new TextEncoder().encode(newPassword).length;
    if (byteLength < 8 || byteLength > 72) {
      setError("新密码长度需为 8–72 字节");
      return;
    }
    if (newPassword !== confirmPassword) {
      setError("两次输入的新密码不一致");
      return;
    }
    setSaving(true);
    changePassword(currentPassword, newPassword).then(() => {
      localStorage.removeItem("cms_authenticated");
      localStorage.removeItem("cms_role");
      localStorage.removeItem("cms_username");
      navigate("/account/login?passwordChanged=1", { replace: true });
    }).catch((reason) => setError(getApiErrorMessage(reason, "密码修改失败，请稍后重试"))).finally(() => setSaving(false));
  };

  return <AdminLayout title="账号安全">
    <section className="admin-panel account-security-panel">
      <div className="account-security-heading"><span><ShieldCheck size={22} /></span><div><h2>修改登录密码</h2><p>修改成功后，该账号在所有设备上的登录会话都会退出，请使用新密码重新登录。</p></div></div>
      <form className="account-security-form" onSubmit={submit}>
        {error && <p className="form-error" role="alert">{error}</p>}
        <label>当前密码<input autoComplete="current-password" required type="password" value={currentPassword} onChange={(event) => setCurrentPassword(event.target.value)} /></label>
        <label>新密码<input autoComplete="new-password" required type="password" value={newPassword} onChange={(event) => setNewPassword(event.target.value)} /><small>请输入 8–72 字节的新密码，不能与当前密码相同。</small></label>
        <label>确认新密码<input autoComplete="new-password" required type="password" value={confirmPassword} onChange={(event) => setConfirmPassword(event.target.value)} /></label>
        <button className="primary-btn" disabled={saving} type="submit"><KeyRound size={16} />{saving ? "正在修改…" : "修改密码并重新登录"}</button>
      </form>
      <p className="account-security-help">忘记当前密码？请联系平台管理员在 CMS 账号管理中重置。</p>
    </section>
  </AdminLayout>;
}
