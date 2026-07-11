import { useEffect, useState } from "react";
import { getHome } from "../api/public";
import { HeroSearch } from "../components/public/HeroSearch";
import { MobileBottomNav } from "../components/public/MobileBottomNav";
import { MobileCategoryGrid } from "../components/public/MobileCategoryGrid";
import { MobileHeader } from "../components/public/MobileHeader";
import { MoreVendors } from "../components/public/MoreVendors";
import { PublicHeader } from "../components/public/PublicHeader";
import { RecommendedVendors } from "../components/public/RecommendedVendors";
import { SidebarNav } from "../components/public/SidebarNav";
import { StatsFooter } from "../components/public/StatsFooter";
import { ProcessingSection } from "../components/public/HomeMarketplaceSections";
import type { HomeModule, HomePayload } from "../types/api";

export function HomePage() {
  const [home, setHome] = useState<HomePayload | null>(null);
  const [error, setError] = useState("");
  const load = () => { setError(""); getHome().then(setHome).catch(() => setError("首页数据加载失败，请检查后端服务")); };
  useEffect(load, []);
  if (error) return <div className="state-page"><p>{error}</p><button className="primary-btn" onClick={load}>重试</button></div>;
  if (!home) return <HomeSkeleton />;
  return <HomeView home={home} />;
}

export function HomeView({ home }: { home: HomePayload }) {
  const configured = home.modules || [];
  const moduleOf = (type: HomeModule["type"], fallback: HomeModule) => configured.find((item) => item.type === type) || fallback;
  const recommended = moduleOf("recommendedVendors", { type: "recommendedVendors", title: "推荐厂商", visible: true, limit: home.homeSections?.recommendedLimit || 4, path: "/vendors", sortOrder: 20 });
  const regular = moduleOf("moreVendors", { type: "moreVendors", title: "农机配件厂商", visible: true, limit: home.homeSections?.moreLimit || 8, path: "/vendors", sortOrder: 30 });
  const processing = moduleOf("processingServices", { type: "processingServices", title: "加工服务厂商", subtitle: "按加工能力查找可承接来图、来样与批量加工的生产企业", visible: true, limit: 4, path: "/service", sortOrder: 40 });
  const hasVendors = home.recommendedVendors.length > 0 || home.moreVendors.length > 0 || Boolean(home.processingVendors?.length);

  return (
    <div className="site-shell">
      <PublicHeader menus={home.topMenus} siteMeta={home.siteMeta} />
      <MobileHeader auxiliaryMenus={home.auxiliaryMenus} menus={home.sidebarMenus} siteMeta={home.siteMeta} />
      <main className="site-body">
        <SidebarNav auxiliaryMenus={[]} menus={home.sidebarMenus} />
        <div className="content">
          <HeroSearch banner={home.banner} />
          <MobileCategoryGrid menus={home.mobileMenus} />
          {!hasVendors ? <section className="home-empty-directory"><h2>公开厂商资料正在完善</h2><p>厂商资料经核验或提交审核通过后，将在这里公开展示。</p><a className="primary-btn" href="/vendors">查看厂商目录</a></section> : <>
            {recommended.visible && home.recommendedVendors.length > 0 && <RecommendedVendors homeSections={{ recommendedTitle: recommended.title || "推荐厂商", recommendedLink: recommended.path }} vendors={home.recommendedVendors.slice(0, recommended.limit || 4)} />}
            {regular.visible && home.moreVendors.length > 0 && <MoreVendors homeSections={{ moreTitle: regular.title || "农机配件厂商", moreLink: regular.path }} vendors={home.moreVendors.slice(0, regular.limit || 8)} />}
            {processing.visible && <ProcessingSection module={processing} vendors={home.processingVendors || []} />}
          </>}
        </div>
      </main>
      <StatsFooter safeguards={home.safeguards} stats={home.stats} siteMeta={home.siteMeta} />
      <MobileBottomNav menus={home.mobileBottomMenus} />
    </div>
  );
}

function HomeSkeleton() {
  return <div className="site-shell"><div className="skeleton-header" /><main className="site-body"><div className="skeleton-sidebar" /><div className="content"><div className="skeleton-hero" /><div className="skeleton-grid">{Array.from({ length: 8 }).map((_, index) => <div className="skeleton-card" key={index} />)}</div></div></main></div>;
}
