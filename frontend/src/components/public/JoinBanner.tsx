import { Link } from "react-router-dom";
import type { JoinConfig } from "../../types/api";
import { trackAnalytics } from "../../analytics";

export function JoinBanner({ join }: { join: JoinConfig }) {
  return (
    <section className="join-banner">
      <div className="join-icon">企</div>
      <strong>{join.text}</strong>
      <Link className="primary-btn" to={join.path || "/join"} onClick={() => trackAnalytics({ eventType: "join_cta_click", path: "/join" })}>
        {join.buttonText || "立即入驻"}
      </Link>
    </section>
  );
}
