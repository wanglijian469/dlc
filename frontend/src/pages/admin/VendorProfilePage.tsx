import { FormEvent, useEffect, useRef, useState } from "react";
import { CheckCircle2, Clock3, ImageUp, Send } from "lucide-react";
import { getVendorProfile, submitVendorProfile, uploadFile } from "../../api/admin";
import { AdminLayout } from "../../components/admin/AdminLayout";
import type { Vendor } from "../../types/api";
import { VendorMediaEditor } from "../../components/admin/StructuredEditors";
import type { VendorMedia } from "../../types/api";
import { ProtectedMediaImage } from "../../components/admin/ProtectedMediaImage";
import { processingToggleDescription, vendorFieldGuidance } from "../../config/formGuidance";

type Field = { key: keyof Vendor; label: string; type?: "textarea" | "checkbox" | "image"; placeholder?: string };

const fields: Field[] = [
  { key: "name", label: "厂商全称" }, { key: "shortName", label: "厂商简称", placeholder: vendorFieldGuidance.shortName },
  { key: "logo", label: "厂商 Logo", type: "image" }, { key: "coverImage", label: "厂商封面图", type: "image" },
  { key: "province", label: "省份" }, { key: "city", label: "城市" }, { key: "county", label: "区县" },
  { key: "address", label: "详细地址" }, { key: "mainProducts", label: "主营产品", type: "textarea", placeholder: vendorFieldGuidance.mainProducts },
  { key: "serviceModels", label: "适配机型", type: "textarea", placeholder: vendorFieldGuidance.serviceModels }, { key: "serviceAdvantages", label: "服务优势", type: "textarea", placeholder: vendorFieldGuidance.serviceAdvantages },
  { key: "description", label: "公司简介", type: "textarea", placeholder: vendorFieldGuidance.description }, { key: "establishedYear", label: "成立年份" },
  { key: "factoryArea", label: "厂房面积" }, { key: "employeeCount", label: "员工规模" },
  { key: "annualCapacity", label: "年产能", type: "textarea", placeholder: vendorFieldGuidance.annualCapacity }, { key: "equipment", label: "主要设备", type: "textarea", placeholder: vendorFieldGuidance.equipment },
  { key: "certifications", label: "认证资质", type: "textarea", placeholder: vendorFieldGuidance.certifications },
  { key: "afterSalesService", label: "售后服务", type: "textarea", placeholder: vendorFieldGuidance.afterSalesService }, { key: "websiteUrl", label: "厂商官网" },
  { key: "contactName", label: "联系人" }, { key: "phone", label: "联系电话" }, { key: "wechat", label: "微信" },
  { key: "providesProcessing", label: "提供来图来样加工", type: "checkbox" },
  { key: "processingServices", label: "加工服务", type: "textarea", placeholder: vendorFieldGuidance.processingServices }, { key: "processingMaterials", label: "加工材料 / 配件类型", type: "textarea", placeholder: vendorFieldGuidance.processingMaterials },
  { key: "processingEquipment", label: "加工设备", type: "textarea", placeholder: vendorFieldGuidance.processingEquipment }, { key: "processingCapacity", label: "加工产能 / 交期", type: "textarea", placeholder: vendorFieldGuidance.processingCapacity },
  { key: "processingRegions", label: "加工服务区域", type: "textarea", placeholder: vendorFieldGuidance.processingRegions }, { key: "processingNotes", label: "加工接单说明", type: "textarea", placeholder: vendorFieldGuidance.processingNotes },
];

const groups: Array<{ title: string; keys: Array<keyof Vendor> }> = [
  { title: "基础资料", keys: ["name", "shortName", "province", "city", "county", "address", "mainProducts", "serviceModels", "description"] },
  { title: "展示素材", keys: ["logo", "coverImage", "serviceAdvantages"] },
  { title: "生产能力", keys: ["establishedYear", "factoryArea", "employeeCount", "annualCapacity", "equipment", "certifications"] },
  { title: "加工服务", keys: ["providesProcessing", "processingServices", "processingMaterials", "processingEquipment", "processingCapacity", "processingRegions", "processingNotes", "afterSalesService"] },
  { title: "联系方式", keys: ["websiteUrl", "contactName", "phone", "wechat"] },
];

export function VendorProfilePage() {
  const [form, setForm] = useState<Partial<Vendor>>({});
  const [status, setStatus] = useState<string>("");
  const [reviewNote, setReviewNote] = useState("");
  const [message, setMessage] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const original = useRef("");

  const load = () => getVendorProfile().then((result) => {
    setForm(result.draft);
    original.current = JSON.stringify(result.draft);
    setStatus(result.submission?.status || (result.vendor.publicationStatus === "draft" ? "draft" : "published"));
    setReviewNote(result.submission?.reviewNote || "");
  }).catch(() => setMessage("厂商资料加载失败，请联系管理员确认账号绑定")).finally(() => setLoading(false));

  useEffect(() => { void load(); }, []);
  const dirty = !loading && JSON.stringify(form) !== original.current;
  useEffect(() => {
    const warn = (event: BeforeUnloadEvent) => { if (dirty) { event.preventDefault(); event.returnValue = ""; } };
    window.addEventListener("beforeunload", warn);
    return () => window.removeEventListener("beforeunload", warn);
  }, [dirty]);

  const submit = (event: FormEvent) => {
    event.preventDefault();
    setMessage("");
    setSaving(true);
    submitVendorProfile(form).then(() => {
      setStatus("pending");
      setReviewNote("");
      setMessage("资料已提交，管理员审核通过后将在前台更新");
      original.current = JSON.stringify(form);
    }).catch(() => setMessage("提交失败，请检查厂商名称、网址或图片状态")).finally(() => setSaving(false));
  };

  return (
    <AdminLayout title="我的厂商资料">
      <section className={`vendor-workflow-status ${status}`}>
        {status === "pending" ? <Clock3 size={20} /> : <CheckCircle2 size={20} />}
        <div><strong>{statusLabel(status)}</strong><p>{statusHint(status, reviewNote)}</p></div>
      </section>
      {message && <p className="admin-message">{message}</p>}
      {loading ? <div className="admin-panel">正在加载…</div> : (
        <form className="admin-form vendor-profile-form vendor-profile-grouped" onSubmit={submit}>
          {groups.map((group, groupIndex) => <details className="vendor-profile-section" key={group.title} open={groupIndex === 0}><summary>{group.title}</summary><div className="vendor-profile-fields">{fields.filter((field) => group.keys.includes(field.key)).map((field) => field.type === "checkbox" ? <label className="admin-toggle-field wide-field" key={field.key}><VendorField field={field} form={form} setForm={setForm} /><span><strong>{field.label}</strong><small>{processingToggleDescription}</small></span></label> : <label className={field.type === "textarea" ? "wide-field" : ""} key={field.key}>{field.label}<VendorField field={field} form={form} setForm={setForm} /></label>)}{group.title === "展示素材" && <div className="wide-field"><VendorMediaEditor value={(form.media as VendorMedia[] | undefined) || []} onChange={(media) => setForm({ ...form, media })} /></div>}</div></details>)}
          <div className="wide-field form-submit-row vendor-sticky-submit">
            <button className="primary-btn" disabled={saving || !dirty} type="submit"><Send size={16} />{saving ? "正在提交…" : "提交管理员审核"}</button>
            {dirty && <strong className="unsaved-indicator">有未保存修改</strong>}
            <span>提交不会立即覆盖前台正在展示的已审核资料。</span>
          </div>
        </form>
      )}
    </AdminLayout>
  );
}

function VendorField({ field, form, setForm }: { field: Field; form: Partial<Vendor>; setForm: (value: Partial<Vendor>) => void }) {
  const value = form[field.key];
  if (field.type === "checkbox") return <input type="checkbox" checked={Boolean(value)} onChange={(e) => setForm({ ...form, [field.key]: e.target.checked })} />;
  if (field.type === "textarea") return <textarea className={field.placeholder ? "writing-example" : undefined} placeholder={field.placeholder} value={String(value || "")} onChange={(e) => setForm({ ...form, [field.key]: e.target.value })} />;
  if (field.type === "image") return (
    <div className="image-field">
      <input value={String(value || "")} placeholder="图片 URL" onChange={(e) => setForm({ ...form, [field.key]: e.target.value })} />
      <label className="outline-btn small upload-button"><ImageUp size={15} />上传<input type="file" accept="image/jpeg,image/png,image/webp" onChange={(e) => {
        const file = e.target.files?.[0];
        const assetKey = field.key === "logo" ? "logoAssetId" : "coverAssetId";
        if (file) void uploadFile(file).then((result) => setForm({ ...form, [field.key]: result.url, [assetKey]: result.assetId })).catch(() => window.alert("图片上传失败，请使用有效的 JPEG、PNG 或 WebP 图片"));
      }} /></label>
      {value && <ProtectedMediaImage alt={`${field.label} 预览`} assetId={Number(form[field.key === "logo" ? "logoAssetId" : "coverAssetId"] || 0) || undefined} className="vendor-form-image" src={String(value)} />}
    </div>
  );
  return <input className={field.placeholder ? "writing-example" : undefined} placeholder={field.placeholder} value={String(value || "")} onChange={(e) => setForm({ ...form, [field.key]: e.target.value })} />;
}

function statusLabel(status: string) {
  return ({ draft: "公司资料待完善，当前尚未发布", pending: "资料审核中", rejected: "资料已驳回，请修改后重新提交", approved: "最近提交已审核通过", published: "当前展示的是已发布资料" } as Record<string, string>)[status] || "厂商资料";
}

function statusHint(status: string, note: string) {
  if (status === "draft") return "请完善公司资料并提交审核，审核通过后公司才会在前台展示。";
  if (status === "pending") return "管理员通过后，新资料会自动更新到前台厂商页面。";
  if (status === "rejected") return note ? `审核意见：${note}` : "请完善资料后再次提交。";
  return "修改资料后提交审核，审核期间不会影响当前前台内容。";
}
