import { Link } from "react-router-dom";
import type { JoinConfig } from "../../types/api";

export function JoinBanner({ join }: { join: JoinConfig }) {
  return (
    <section className="join-banner">
      <div className="join-icon">企</div>
      <strong>{join.text}</strong>
      <Link className="primary-btn" to={join.path || "/join"}>
        {join.buttonText || "立即入驻"}
      </Link>
    </section>
  );
}
