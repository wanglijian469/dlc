import type { StatItem } from "../../types/api";

export function StatsFooter({ stats, safeguards }: { stats: StatItem[]; safeguards: string[] }) {
  return (
    <footer className="stats-footer">
      <div className="safeguards">{safeguards.join(" · ")}</div>
      <div className="stats-row">
        {stats.map((item) => (
          <div key={item.label}>
            <strong>{item.value}</strong>
            <span>{item.label}</span>
          </div>
        ))}
      </div>
    </footer>
  );
}
