import { useEffect, useRef, useState } from "react";
import { saveWorkDraft, type WorkDraft } from "../api/workspace";
import { getApiErrorMessage } from "../api/client";

export function usePrivateDraft<T>(kind: "product" | "profile", value: T, enabled: boolean) {
 const [status, setStatus] = useState("尚未保存");
 const [revision, setRevision] = useState(0);
 const active = useRef<{ key: string; type: string; targetId: number; row?: WorkDraft<T>; saved: string } | undefined>(undefined);
 const flight = useRef<Promise<WorkDraft<T>> | null>(null);
 const current = useRef(value); current.current = value;
 const snapshot = JSON.stringify(value);
 const activate = (row?: WorkDraft<T>, type = "", targetId = 0) => {
  active.current = { key: row?.clientKey || ("draft_" + Date.now().toString(36) + "_" + Math.random().toString(36).slice(2)), type: row?.targetType || type, targetId: row?.targetId || targetId, row, saved: row ? JSON.stringify(row.payload) : "" };
  setStatus(row ? "已恢复草稿" : "尚未保存"); setRevision(v => v + 1);
 };
 const save = async (): Promise<WorkDraft<T>> => {
  if (flight.current) { await flight.current; return save(); }
  const session = active.current;
  if (!session) throw new Error("草稿尚未准备完成");
  const payload = current.current; const serialized = JSON.stringify(payload);
  if (session.row && session.saved === serialized) return session.row;
  setStatus("保存中…");
  const request = saveWorkDraft(session.key, { kind, targetType: session.type, targetId: session.targetId, version: session.row?.version || 0, payload });
  flight.current = request;
  try {
   const row = await request;
   session.row = row; session.saved = serialized;
   if (active.current === session) { setStatus("已保存"); setRevision(v => v + 1); }
   return row;
  } catch (error) {
   if (active.current === session) setStatus("保存失败：" + getApiErrorMessage(error, "请检查网络后重试"));
   throw error;
  } finally { flight.current = null; }
 };
 const dirty = enabled && !!active.current && snapshot !== active.current.saved;
 useEffect(() => {
  if (!dirty || status.startsWith("保存失败")) return;
  const timer = window.setTimeout(() => { void save().catch(() => undefined); }, 1500);
  return () => window.clearTimeout(timer);
 }, [snapshot, enabled, revision, status]);
 useEffect(() => {
  const warn = (e: BeforeUnloadEvent) => { if (dirty) { e.preventDefault(); e.returnValue = ""; } };
  const navigation = (e: MouseEvent) => {
   if (!dirty || !(e.target instanceof Element) || !e.target.closest("a[href]")) return;
   if (!window.confirm("还有未保存修改，确定离开？")) { e.preventDefault(); e.stopPropagation(); }
  };
  window.addEventListener("beforeunload", warn); document.addEventListener("click", navigation, true);
  return () => { window.removeEventListener("beforeunload", warn); document.removeEventListener("click", navigation, true); };
 }, [dirty]);
 return { activate, save, dirty, status, row: active.current?.row };
}
