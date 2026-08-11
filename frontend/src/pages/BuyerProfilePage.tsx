import { FormEvent, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { getApiErrorMessage } from "../api/client";
import { getAccountProfile, updateBuyerProfile, type BuyerProfile } from "../api/market";
import { PageFrame } from "../components/public/PageFrame";

const emptyProfile: BuyerProfile = { displayName: "", contactName: "", phone: "", province: "", city: "" };

export function BuyerProfilePage() {
  const navigate = useNavigate();
  const [profile, setProfile] = useState<BuyerProfile>(emptyProfile);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");

  useEffect(() => {
    if (localStorage.getItem("cms_authenticated") !== "true" || localStorage.getItem("cms_role") !== "buyer") {
      navigate(`/account/login?returnTo=${encodeURIComponent("/account/profile")}`, { replace: true });
      return;
    }
    getAccountProfile()
      .then((result) => setProfile({ ...emptyProfile, ...result.profile }))
      .catch((reason) => setMessage(getApiErrorMessage(reason, "资料加载失败")))
      .finally(() => setLoading(false));
  }, [navigate]);

  const change = (key: keyof BuyerProfile, value: string) => setProfile((current) => ({ ...current, [key]: value }));
  const submit = (event: FormEvent) => {
    event.preventDefault();
    setSaving(true);
    setMessage("");
    updateBuyerProfile(profile)
      .then((result) => { setProfile({ ...emptyProfile, ...result.profile }); setMessage("采购商资料已保存"); })
      .catch((reason) => setMessage(getApiErrorMessage(reason, "资料保存失败")))
      .finally(() => setSaving(false));
  };

  return <PageFrame title="采购商资料" breadcrumbs={[{ label: "我的", path: "/account/posts" }]}>
    {loading ? <div className="state-page">正在加载…</div> : <form className="market-editor buyer-profile-form" onSubmit={submit}>
      {message && <p className="admin-message" role="status">{message}</p>}
      <div className="market-form-grid">
        <label>称呼<input maxLength={80} value={profile.displayName} onChange={(event) => change("displayName", event.target.value)} /></label>
        <label>联系人<input maxLength={80} value={profile.contactName} onChange={(event) => change("contactName", event.target.value)} /></label>
        <label>联系电话<input inputMode="tel" maxLength={40} value={profile.phone || ""} onChange={(event) => change("phone", event.target.value)} /></label>
        <label>省份<input maxLength={50} value={profile.province} onChange={(event) => change("province", event.target.value)} /></label>
        <label>城市<input maxLength={50} value={profile.city} onChange={(event) => change("city", event.target.value)} /></label>
      </div>
      <button className="primary-btn market-submit" disabled={saving}>{saving ? "正在保存…" : "保存资料"}</button>
    </form>}
  </PageFrame>;
}
