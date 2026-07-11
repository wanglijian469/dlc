import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter } from "react-router-dom";
import { App } from "./App";
import { SiteProvider } from "./contexts/SiteContext";
import "./styles/global.css";
import "./styles/upgrade.css";
import "./styles/admin-upgrade.css";

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <BrowserRouter>
      <SiteProvider><App /></SiteProvider>
    </BrowserRouter>
  </React.StrictMode>,
);
