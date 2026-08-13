import { useEffect, useState, type FormEvent } from "react";
import { Link, useParams } from "react-router-dom";
import { getAuction, listAuctions, placeAuctionBid } from "../api/market";
import { getApiErrorMessage } from "../api/client";
import { PageFrame } from "../components/public/PageFrame";
import type { ProcurementAuction } from "../types/api";

const statusText: Record<string, string> = { draft: "草稿", open: "竞价中", awaiting_award: "已截止待定标", awarded: "已中标", unawarded: "未成交", cancelled: "已取消", suspended: "平台暂停" };
const money = (cents?: number) => cents ? "¥" + (cents / 100).toFixed(2) : "待报价";

export function AuctionsPage() {
  const [items, setItems] = useState<ProcurementAuction[]>([]);
  const [keyword, setKeyword] = useState("");
  const [message, setMessage] = useState("");
  const load = () => listAuctions({ keyword: keyword || undefined }).then(setItems).catch((error) => setMessage(getApiErrorMessage(error, "竞价大厅暂时不可用")));
  useEffect(() => { void load(); }, []);
  return <PageFrame breadcrumbs={[{ label: "首页", path: "/" }]} title="采购竞价大厅" subtitle="采购商发布需求，合格厂商逐次降价；采购商可综合价格、交期与说明定标">
    <form className="auction-search" onSubmit={(event) => { event.preventDefault(); void load(); }}><input placeholder="搜索采购名称或规格" value={keyword} onChange={(event) => setKeyword(event.target.value)} /><button className="primary-btn" type="submit">搜索</button><Link className="outline-btn" to="/account/auctions/new">发布竞价</Link></form>
    {message && <p className="inline-notice">{message}</p>}
    <div className="auction-grid">{items.map((item) => <article className="auction-card" key={item.id}><header><span className={"auction-status " + item.status}>{statusText[item.status]}</span><small>{item.category?.name || "农机配件采购"}</small></header><h2>{item.title}</h2><p>{item.specification || item.compatibleModels || "采购方暂未填写详细规格"}</p><dl><div><dt>采购数量</dt><dd>{item.quantity} {item.unit}</dd></div><div><dt>当前最低单价</dt><dd>{money(item.leadingPriceCents)}</dd></div><div><dt>有效报价</dt><dd>{item.bidCount} 家</dd></div><div><dt>收货地区</dt><dd>{item.deliveryProvince}{item.deliveryCity ? " · " + item.deliveryCity : ""}</dd></div></dl><footer><span>截止：{new Date(item.endAt).toLocaleString()}</span><Link className="primary-btn small" to={"/auctions/" + item.id}>查看竞价</Link></footer></article>)}</div>
    {!items.length && !message && <p className="structured-empty">暂无公开竞价。</p>}
  </PageFrame>;
}

export function AuctionDetailPage() {
  const { id = "" } = useParams();
  const [item, setItem] = useState<ProcurementAuction>();
  const [message, setMessage] = useState("");
  const [price, setPrice] = useState("");
  const [deliveryDays, setDeliveryDays] = useState("7");
  const [supplyNote, setSupplyNote] = useState("");
  const load = () => getAuction(id).then(setItem).catch((error) => setMessage(getApiErrorMessage(error, "竞价详情加载失败")));
  useEffect(() => { void load(); }, [id]);
  const submit = (event: FormEvent) => {
    event.preventDefault();
    if (!item) return;
    placeAuctionBid(item.id, { unitPriceCents: Math.round(Number(price) * 100), taxIncluded: true, deliveryDays: Number(deliveryDays), supplyNote, expectedVersion: item.version }).then((result) => {
      setMessage(result.extended ? "报价成功，竞价已自动延时 5 分钟" : "报价成功");
      setPrice("");
      void load();
    }).catch((error) => setMessage(getApiErrorMessage(error, "报价失败")));
  };
  if (!item) return <PageFrame breadcrumbs={[{ label: "竞价大厅", path: "/auctions" }]} title="竞价详情"><p className="state-page">{message || "正在加载…"}</p></PageFrame>;
  const isVendor = localStorage.getItem("cms_role") === "vendor";
  return <PageFrame breadcrumbs={[{ label: "竞价大厅", path: "/auctions" }]} title={item.title} subtitle={statusText[item.status] + " · 截止 " + new Date(item.endAt).toLocaleString()}>
    <section className="auction-detail-card"><div className="auction-detail-summary"><strong>{money(item.leadingPriceCents)}<small>当前匿名最低单价</small></strong><strong>{item.bidCount}<small>有效报价厂商</small></strong>{item.myRank ? <strong>第 {item.myRank} 名<small>本厂当前排名</small></strong> : null}</div><dl><div><dt>采购数量</dt><dd>{item.quantity} {item.unit}</dd></div><div><dt>规格要求</dt><dd>{item.specification || "未填写"}</dd></div><div><dt>适配信息</dt><dd>{item.compatibleModels || "未填写"}</dd></div><div><dt>收货地区</dt><dd>{item.deliveryProvince} {item.deliveryCity}</dd></div><div><dt>期望交期</dt><dd>{item.expectedDeliveryNote || "协商"}</dd></div><div><dt>最高预算</dt><dd>{item.maxBudgetCents ? money(item.maxBudgetCents) : "未公开"}</dd></div></dl>{item.description && <p>{item.description}</p>}{item.extensionMinutes > 0 && <p className="inline-notice">因截止前出现新低价，已累计延时 {item.extensionMinutes} 分钟。</p>}</section>
    {item.status === "open" && isVendor && <form className="auction-bid-form admin-panel" onSubmit={submit}><h2>提交本厂报价</h2><p>只能低于本厂上一次报价。其他厂商身份不会向您公开。</p><div className="auction-form-grid"><label>单价（元）<input min="0.01" required step="0.01" type="number" value={price} onChange={(event) => setPrice(event.target.value)} /></label><label>交期（天）<input min="0" required type="number" value={deliveryDays} onChange={(event) => setDeliveryDays(event.target.value)} /></label><label className="wide-field">供货说明<textarea value={supplyNote} onChange={(event) => setSupplyNote(event.target.value)} /></label></div><button className="primary-btn" type="submit">确认降价报价</button></form>}
    {item.status === "open" && !isVendor && <p className="inline-notice">只有已审核、已公开且账号正常的厂商可以参与报价。</p>}
    {message && <p className="admin-message">{message}</p>}
  </PageFrame>;
}
