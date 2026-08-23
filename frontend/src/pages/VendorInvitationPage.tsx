import { useEffect, useState, type FormEvent } from "react";
import { Building2, CheckCircle2, KeyRound, UserRound } from "lucide-react";
import { Link, useParams } from "react-router-dom";
import { acceptVendorInvitation, getVendorInvitation } from "../api/admin";
import { getApiErrorMessage } from "../api/client";

export function VendorInvitationPage() {
  const token = useParams().token || "";
  const [vendorName, setVendorName] = useState("");
  const [expiresAt, setExpiresAt] = useState("");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [message, setMessage] = useState("");
  const [accepted, setAccepted] = useState(false);
  const [busy, setBusy] = useState(false);
  useEffect(() => { getVendorInvitation(token).then((value) => { setVendorName(value.vendorName); setExpiresAt(value.expiresAt); }).catch((error) => setMessage(getApiErrorMessage(error, "邀请链接无效"))); }, [token]);
  const submit = (event: FormEvent) => {
    event.preventDefault(); setMessage("");
    if (password !== confirm) { setMessage("两次输入的密码不一致"); return; }
    setBusy(true); acceptVendorInvitation(token, { username, password }).then(() => setAccepted(true)).catch((error) => setMessage(getApiErrorMessage(error, "账号开通失败"))).finally(() => setBusy(false));
  };
  return <main className="account-auth-page vendor-invitation-page"><section className="account-auth-intro"><Link to="/">大陆农机配件</Link><h1>接管企业资料</h1><p>设置账号后即可维护企业资料、产品与供货信息。</p></section><section className="account-auth-card">
    {accepted ? <div className="invitation-success"><CheckCircle2 size={46} /><h2>账号已开通</h2><p>你已接管 {vendorName} 的企业资料。</p><Link className="primary-btn" to="/account/login">立即登录</Link></div> : <form onSubmit={submit}>
      <Building2 size={34} /><h2>{vendorName || "厂商接管邀请"}</h2>{expiresAt && <p>邀请有效期至 {new Date(expiresAt).toLocaleString("zh-CN")}</p>}
      {message && <p className="form-error" role="alert">{message}</p>}
      <label>登录账号<span className="account-input"><UserRound size={18} /><input autoComplete="username" required value={username} onChange={(event) => setUsername(event.target.value)} /></span></label>
      <label>设置密码<span className="account-input"><KeyRound size={18} /><input autoComplete="new-password" minLength={6} required type="password" value={password} onChange={(event) => setPassword(event.target.value)} /></span></label>
      <label>确认密码<input autoComplete="new-password" minLength={6} required type="password" value={confirm} onChange={(event) => setConfirm(event.target.value)} /></label>
      <button className="primary-btn account-submit" disabled={busy || !vendorName}>{busy ? "正在开通…" : "确认接管企业资料"}</button>
    </form>}
  </section></main>;
}
