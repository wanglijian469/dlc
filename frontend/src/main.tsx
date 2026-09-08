import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter } from "react-router-dom";
import { App } from "./App";
import { SiteProvider } from "./contexts/SiteContext";
import "./styles/global.css";
import "./styles/upgrade.css";
import "./styles/admin-upgrade.css";
import "./styles/mobile-public.css";
import "./styles/marketplace.css";
import "./styles/vendor-promotion.css";

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <BrowserRouter>
      <SiteProvider><App /></SiteProvider>
    </BrowserRouter>
  </React.StrictMode>,
);

// Gin emits an equivalent semantic document for crawlers and no-JavaScript
// visitors. Remove it only after React has mounted successfully so a failed
// bundle still leaves useful content on screen.
window.requestAnimationFrame(() => {
  window.requestAnimationFrame(() => document.getElementById("seo-fallback")?.remove());
});
