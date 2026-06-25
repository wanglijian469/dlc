import { useEffect, useState } from "react";
import { getHome } from "../api/public";
import { HeroSearch } from "../components/public/HeroSearch";
import { JoinBanner } from "../components/public/JoinBanner";
import { MobileBottomNav } from "../components/public/MobileBottomNav";
import { MobileCategoryGrid } from "../components/public/MobileCategoryGrid";
import { MobileHeader } from "../components/public/MobileHeader";
import { MoreVendors } from "../components/public/MoreVendors";
import { PublicHeader } from "../components/public/PublicHeader";
import { RecommendedVendors } from "../components/public/RecommendedVendors";
import { SidebarNav } from "../components/public/SidebarNav";
import { StatsFooter } from "../components/public/StatsFooter";
import type { HomePayload } from "../types/api";

export function HomePage() {
  const [home, setHome] = useState<HomePayload | null>(null);
  const [error, setError] = useState("");

  const load = () => {
    setError("");
    getHome()
      .then(setHome)
      .catch(() => setError("首页数据加载失败，请检查后端服务"));
  };

  useEffect(load, []);

  if (error) {
    return (
      <div className="state-page">
        <p>{error}</p>
        <button className="primary-btn" onClick={load}>
          重试
        </button>
      </div>
    );
  }
  if (!home) return <HomeSkeleton />;
  return <HomeView home={home} />;
}

export function HomeView({ home }: { home: HomePayload }) {
  return (
    <div className="site-shell">
      <PublicHeader menus={home.topMenus} />
      <MobileHeader />
      <main className="site-body">
        <SidebarNav auxiliaryMenus={home.auxiliaryMenus} menus={home.sidebarMenus} />
        <div className="content">
          <HeroSearch banner={home.banner} />
          <MobileCategoryGrid menus={home.mobileMenus} />
          <RecommendedVendors vendors={home.recommendedVendors} />
          <MoreVendors vendors={home.moreVendors} />
          <JoinBanner join={home.join} />
        </div>
      </main>
      <StatsFooter safeguards={home.safeguards} stats={home.stats} />
      <MobileBottomNav />
    </div>
  );
}

function HomeSkeleton() {
  return (
    <div className="site-shell">
      <div className="skeleton-header" />
      <main className="site-body">
        <div className="skeleton-sidebar" />
        <div className="content">
          <div className="skeleton-hero" />
          <div className="skeleton-grid">
            {Array.from({ length: 10 }).map((_, index) => (
              <div className="skeleton-card" key={index} />
            ))}
          </div>
        </div>
      </main>
    </div>
  );
}
