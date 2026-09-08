import { useEffect, useState, type FormEvent } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { awardAuction, cancelAuction, createAuction, getOwnAuction, listNotifications, listOwnAuctions, publishAuction, readAllNotifications, readNotification, updateAuction, uploadMarketImage } from "../api/market";
import { getApiErrorMessage } from "../api/client";
import { PageFrame } from "../components/public/PageFrame";
import type { AuctionBid, ProcurementAuction, UserNotification } from "../types/api";

export function AuctionEditorPage() {
  const navigate = useNavigate();
  const { id } = useParams();
  const editID = Number(id || 0);
  const [message, setMessage] = useState("");
  const [form, setForm] = useState({ title: "", specification: "", compatibleModels: "", quantity: 1, unit: "件", deliveryProvince: "", deliveryCity: "", expectedDeliveryNote: "", description: "", image: "", imageAssetId: undefined as number | undefined, maxBudgetYuan: "", endAt: "" });
  const [version, setVersion] = useState(1);
  useEffect(() => {
    if (!editID) return;
    getOwnAuction(editID).then(({ auction }) => {
      setVersion(auction.version);
      setForm({ title: auction.title, specification: auction.specification || "", compatibleModels: auction.compatibleModels || "", quantity: auction.quantity, unit: auction.unit, deliveryProvince: auction.deliveryProvince, deliveryCity: auction.deliveryCity || "", expectedDeliveryNote: auction.expectedDeliveryNote || "", description: auction.description || "", image: auction.image || "", imageAssetId: auction.imageAssetId, maxBudgetYuan: auction.maxBudgetCents ? String(auction.maxBudgetCents / 100) : "", endAt: auction.endAt.slice(0, 16) });
    }).catch((error) => setMessage(getApiErrorMessage(error, "草稿加载失败")));
  }, [editID]);
  const submit = (event: FormEvent) => {
    event.preventDefault();
    const payload = { ...form, maxBudgetCents: Math.round(Number(form.maxBudgetYuan || 0) * 100), endAt: new Date(form.endAt).toISOString(), expectedVersion: version };
    const action = editID ? updateAuction(editID, payload) : createAuction(payload);
    action.then((row) => navigate("/account/auctions?created=" + row.id)).catch((error) => setMessage(getApiErrorMessage(error, "草稿保存失败")));
  };
  return <PageFrame breadcrumbs={[{ label: "竞价大厅", path: "/auctions" }]} title={editID ? "编辑采购竞价草稿" : "发布采购竞价"} subtitle="先保存草稿，确认联系人、地区和截止时间后再发布">
    <form className="auction-editor admin-panel" onSubmit={submit}><div className="auction-form-grid"><label>采购名称<input required value={form.title} onChange={(event) => setForm({ ...form, title: event.target.value })} /></label><label>数量<input min="1" required type="number" value={form.quantity} onChange={(event) => setForm({ ...form, quantity: Number(event.target.value) })} /></label><label>计价单位<input required value={form.unit} onChange={(event) => setForm({ ...form, unit: event.target.value })} /></label><label>最高总预算（元，可不填）<input min="0" step="0.01" type="number" value={form.maxBudgetYuan} onChange={(event) => setForm({ ...form, maxBudgetYuan: event.target.value })} /></label><label>收货省份<input value={form.deliveryProvince} onChange={(event) => setForm({ ...form, deliveryProvince: event.target.value })} /></label><label>收货城市<input value={form.deliveryCity} onChange={(event) => setForm({ ...form, deliveryCity: event.target.value })} /></label><label>截止时间<input required type="datetime-local" value={form.endAt} onChange={(event) => setForm({ ...form, endAt: event.target.value })} /></label><label>期望交期<input value={form.expectedDeliveryNote} onChange={(event) => setForm({ ...form, expectedDeliveryNote: event.target.value })} /></label><label className="wide-field">规格要求<textarea required value={form.specification} onChange={(event) => setForm({ ...form, specification: event.target.value })} /></label><label className="wide-field">适配信息<textarea value={form.compatibleModels} onChange={(event) => setForm({ ...form, compatibleModels: event.target.value })} /></label><label className="wide-field">需求图片<input accept="image/*" type="file" onChange={(event) => { const file = event.target.files?.[0]; if (!file) return; uploadMarketImage(file).then((result) => setForm((value) => ({ ...value, image: result.url, imageAssetId: result.assetId }))).catch((error) => setMessage(getApiErrorMessage(error, "图片上传失败"))); }} />{form.image && <img className="auction-upload-preview" alt="需求图片预览" src={form.image} />}</label><label className="wide-field">补充说明<textarea value={form.description} onChange={(event) => setForm({ ...form, description: event.target.value })} /></label></div>{message && <p className="admin-message">{message}</p>}<button className="primary-btn" type="submit">保存竞价草稿</button></form>
  </PageFrame>;
}

export function MyAuctionsPage() {
  const [items, setItems] = useState<ProcurementAuction[]>([]);
  const [message, setMessage] = useState("");
  const load = () => listOwnAuctions().then(setItems).catch((error) => setMessage(getApiErrorMessage(error, "我的竞价加载失败")));
  useEffect(() => { void load(); }, []);
  return <PageFrame breadcrumbs={[{ label: "竞价大厅", path: "/auctions" }]} title="我的采购竞价"><div className="section-title"><p>管理草稿、发布、取消和截止后的定标。</p><Link className="primary-btn small" to="/account/auctions/new">新建竞价</Link></div>{message && <p className="admin-message">{message}</p>}<div className="auction-grid">{items.map((item) => <article className="auction-card" key={item.id}><h2>{item.title}</h2><p>{item.quantity} {item.unit} · {item.bidCount || 0} 家报价</p><footer><Link className="outline-btn small" to={"/account/auctions/" + item.id}>管理</Link>{item.status === "draft" && <Link className="outline-btn small" to={"/account/auctions/" + item.id + "/edit"}>编辑</Link>}{item.status === "draft" && <button className="primary-btn small" type="button" onClick={() => publishAuction(item.id).then(load).catch((error) => setMessage(getApiErrorMessage(error, "发布失败")))}>发布竞价</button>}{(item.status === "draft" || item.status === "open") && <button className="outline-btn small" type="button" onClick={() => { const reason = item.bidCount ? window.prompt("已有报价，请填写取消原因") : ""; if (reason === null) return; cancelAuction(item.id, reason || "").then(load).catch((error) => setMessage(getApiErrorMessage(error, "取消失败"))); }}>取消</button>}</footer></article>)}</div></PageFrame>;
}

export function BuyerAuctionDetailPage() {
  const { id = "" } = useParams();
  const [item, setItem] = useState<ProcurementAuction>();
  const [bids, setBids] = useState<AuctionBid[]>([]);
  const [message, setMessage] = useState("");
  const load = () => getOwnAuction(id).then((result) => { setItem(result.auction); setBids(result.bids); }).catch((error) => setMessage(getApiErrorMessage(error, "竞价管理信息加载失败")));
  useEffect(() => { void load(); }, [id]);
  if (!item) return <PageFrame breadcrumbs={[{ label: "我的采购竞价", path: "/account/auctions" }]} title="竞价管理"><p>{message || "正在加载…"}</p></PageFrame>;
  return <PageFrame breadcrumbs={[{ label: "我的采购竞价", path: "/account/auctions" }]} title={item.title} subtitle={"状态：" + item.status}><section className="admin-panel"><h2>厂商报价明细</h2><p>可综合价格、交期和供货说明选择任一厂商的最新有效报价，不强制最低价。</p><div className="vendor-post-list">{bids.map((bid) => <article key={bid.id}><div><h3>{bid.vendorName || "厂商 #" + bid.vendorId}</h3><p>单价 ¥{(bid.unitPriceCents / 100).toFixed(2)} · 总价 ¥{(bid.totalPriceCents / 100).toFixed(2)} · 交期 {bid.deliveryDays} 天{!bid.isLatest && " · 历史报价"}</p><small>{bid.supplyNote || "无补充说明"}</small></div>{item.status === "awaiting_award" && bid.isLatest && <button className="primary-btn small" type="button" onClick={() => { if (window.confirm("确认选择该报价中标？")) awardAuction(item.id, bid.id).then(load).catch((error) => setMessage(getApiErrorMessage(error, "定标失败"))); }}>选择中标</button>}</article>)}</div>{!bids.length && <p className="structured-empty">暂无报价。</p>}{message && <p className="admin-message">{message}</p>}</section></PageFrame>;
}

export function NotificationsPage() {
  const [items, setItems] = useState<UserNotification[]>([]);
  const [unread, setUnread] = useState(0);
  const load = () => listNotifications().then((result) => { setItems(result.items); setUnread(result.unreadCount); });
  useEffect(() => { void load(); }, []);
  return <PageFrame breadcrumbs={[{ label: "首页", path: "/" }]} title="站内消息" subtitle={"未读 " + unread + " 条"}><div className="section-title"><span>企业及产品审核结果、竞价动态都会在这里保留。</span><button className="outline-btn small" type="button" onClick={() => readAllNotifications().then(load)}>全部已读</button></div><div className="notification-list">{items.map((item) => <Link onClick={() => { if (!item.readAt) void readNotification(item.id).catch(() => undefined); }} className={item.readAt ? "" : "unread"} key={item.id} to={item.businessType === "auction" ? "/auctions/" + item.businessId : item.businessType === "vendor_review" ? "/admin/vendor-profile" : item.businessType === "product_review" ? "/admin/vendor-products?review=" + item.businessId : "/"}><strong>{item.title}</strong><p>{item.content}</p><time>{new Date(item.createdAt).toLocaleString()}</time></Link>)}</div></PageFrame>;
}
