export function LoadingState({ text = "数据加载中..." }: { text?: string }) {
  return <div className="state-panel">{text}</div>;
}

export function EmptyState({ text = "暂无匹配数据" }: { text?: string }) {
  return <div className="state-panel empty-state">{text}</div>;
}

export function ErrorState({ text, onRetry }: { text: string; onRetry: () => void }) {
  return (
    <div className="state-panel error-state">
      <p>{text}</p>
      <button className="primary-btn" type="button" onClick={onRetry}>
        重试
      </button>
    </div>
  );
}
