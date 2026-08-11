import { FormEvent, useEffect, useState } from "react";
import { ImagePlus, Send } from "lucide-react";
import { useNavigate, useParams } from "react-router-dom";
import { getApiErrorMessage } from "../api/client";
import { createMarketPost, getOwnMarketPost, updateMarketPost, uploadMarketImage } from "../api/market";
import { getFilterOptions } from "../api/public";
import { PageFrame } from "../components/public/PageFrame";
import type { Category, MarketPostInput, MarketPostType } from "../types/api";

const emptyInput: MarketPostInput = { type: "demand", title: "", description: "", contactName: "", contactPhone: "", expiresInDays: 30, assetIds: [] };

export function MarketPostEditorPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const role = localStorage.getItem("cms_role");
  const [form, setForm] = useState<MarketPostInput>({ ...emptyInput, type: role === "vendor" ? "supply" : "demand" });
  const [categories, setCategories] = useState<Category[]>([]);
  const [previews, setPreviews] = useState<string[]>([]);
  const [error, setError] = useState("");
  const [saving, setSaving] = useState(false);
  useEffect(() => {
    if (localStorage.getItem("cms_authenticated") !== "true" || (role !== "buyer" && role !== "vendor")) {
      navigate(`/account/login?returnTo=${encodeURIComponent(id ? `/publish/${id}` : "/publish")}`, { replace: true });
      return;
    }
    getFilterOptions().then((result) => setCategories(result.categories));
    if (id) getOwnMarketPost(id).then(({ post, contactName, contactPhone, assetIds }) => { setForm({ type: post.type, title: post.title, categoryId: post.categoryId, compatibleModels: post.compatibleModels, province: post.province, city: post.city, quantity: post.quantity, deliveryNote: post.deliveryNote, description: post.description, contactName, contactPhone, expiresInDays: 30, assetIds }); setPreviews(post.images); }).catch(() => setError("供求信息加载失败"));
  }, [id, navigate, role]);
  const change = <K extends keyof MarketPostInput>(key: K, value: MarketPostInput[K]) => setForm((current) => ({ ...current, [key]: value }));
  const upload = async (files: FileList | null) => {
    if (!files) return;
    if ((form.assetIds?.length || 0) + files.length > 6) return setError("最多上传 6 张图片");
    setError("");
    try {
      for (const file of Array.from(files)) {
        const result = await uploadMarketImage(file);
        setForm((current) => ({ ...current, assetIds: [...(current.assetIds || []), result.assetId] }));
        setPreviews((current) => [...current, URL.createObjectURL(file)]);
      }
    } catch (reason) { setError(getApiErrorMessage(reason, "图片上传失败")); }
  };
  const submit = (event: FormEvent) => {
    event.preventDefault(); setError(""); setSaving(true);
    const action = id ? updateMarketPost(id, form) : createMarketPost(form);
    action.then((post) => navigate(`/purchase/${post.id}`, { replace: true })).catch((reason) => setError(getApiErrorMessage(reason, "供求信息保存失败"))).finally(() => setSaving(false));
  };
  return <PageFrame title={id ? "编辑供求信息" : role === "vendor" ? "发布供应" : "发布求购"} breadcrumbs={[{ label: "供求信息", path: "/purchase" }]}>
    <form className="market-editor" onSubmit={submit}>
      {error && <p className="form-error" role="alert">{error}</p>}
      <div className="market-form-grid"><label>类型<select disabled value={form.type} onChange={(event) => change("type", event.target.value as MarketPostType)}><option value="demand">求购</option><option value="supply">供应</option></select></label><label>配件分类<select value={form.categoryId || ""} onChange={(event) => change("categoryId", event.target.value ? Number(event.target.value) : undefined)}><option value="">请选择</option>{categories.map((category) => <option key={category.id} value={category.id}>{category.name}</option>)}</select></label><label className="wide">标题<input maxLength={160} minLength={4} required value={form.title} onChange={(event) => change("title", event.target.value)} /></label><label>省份<input value={form.province || ""} onChange={(event) => change("province", event.target.value)} /></label><label>城市<input value={form.city || ""} onChange={(event) => change("city", event.target.value)} /></label><label>数量<input value={form.quantity || ""} onChange={(event) => change("quantity", event.target.value)} /></label><label>适配机型<input value={form.compatibleModels || ""} onChange={(event) => change("compatibleModels", event.target.value)} /></label><label className="wide">交期说明<input value={form.deliveryNote || ""} onChange={(event) => change("deliveryNote", event.target.value)} /></label><label className="wide">详细说明<textarea minLength={10} required rows={7} value={form.description} onChange={(event) => change("description", event.target.value)} /></label><label>联系人<input required value={form.contactName} onChange={(event) => change("contactName", event.target.value)} /></label><label>联系电话<input inputMode="tel" minLength={6} required value={form.contactPhone} onChange={(event) => change("contactPhone", event.target.value)} /></label><label>有效期<select value={form.expiresInDays} onChange={(event) => change("expiresInDays", Number(event.target.value))}><option value={7}>7 天</option><option value={30}>30 天</option><option value={90}>90 天</option></select></label></div>
      <section className="market-image-upload"><label><ImagePlus size={20} />添加图片<input accept="image/jpeg,image/png,image/webp" multiple type="file" onChange={(event) => void upload(event.target.files)} /></label><div>{previews.map((preview, index) => <img alt={`已上传图片 ${index + 1}`} key={`${preview}-${index}`} src={preview} />)}</div></section>
      <button className="primary-btn market-submit" disabled={saving}><Send size={17} />{saving ? "正在保存…" : "立即发布"}</button>
    </form>
  </PageFrame>;
}
