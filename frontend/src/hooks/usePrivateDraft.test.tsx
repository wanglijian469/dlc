import { act, renderHook, waitFor, cleanup } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { saveWorkDraft } from "../api/workspace";
import { usePrivateDraft } from "./usePrivateDraft";
vi.mock("../api/workspace", () => ({ saveWorkDraft: vi.fn() }));
const save = vi.mocked(saveWorkDraft);
const row = { id: 3, clientKey: "restore-key", kind: "product" as const, targetType: "supplier", targetId: 9, version: 4, payload: { name: "历史产品", stock: 0 }, submissionId: 0, updatedAt: "" };
afterEach(() => {cleanup();vi.resetAllMocks();});
describe("private drafts", () => {
 it("restores the exact historic zero and reuses an unchanged saved version", async () => {
  const { result } = renderHook(() => usePrivateDraft("product", row.payload, false));
  act(() => result.current.activate(row));
  expect(result.current.status).toBe("已恢复草稿");
  await act(async () => expect(await result.current.save()).toEqual(row));
  expect(save).not.toHaveBeenCalled();
 });
 it("serializes saves and persists changes made during a slow request", async () => {
  let finish!: (value: unknown) => void;
  save.mockImplementationOnce(() => new Promise(resolve => { finish=resolve; }) as never).mockResolvedValueOnce({...row,version:6} as never);
  const {result,rerender}=renderHook(({value})=>usePrivateDraft("product",value,false),{initialProps:{value:{name:"A",stock:0}}});
  act(()=>result.current.activate(row));
  let first!: Promise<unknown>;act(()=>{first=result.current.save();});
  rerender({value:{name:"B",stock:0}});
  let second!: Promise<unknown>;act(()=>{second=result.current.save();});
  expect(save).toHaveBeenCalledTimes(1);
  await act(async()=>{finish({...row,version:5});await first;await second;});
  expect(save).toHaveBeenLastCalledWith("restore-key",expect.objectContaining({targetType:"supplier",targetId:9,version:5,payload:{name:"B",stock:0}}));
 });
 it("keeps the optimistic version on failure and allows explicit retry", async () => {
  save.mockRejectedValueOnce(new Error("网络中断")).mockResolvedValueOnce({...row,version:5} as never);
  const {result}=renderHook(()=>usePrivateDraft("product",{name:"修改",stock:0},false));
  act(()=>result.current.activate(row));
  await act(async()=>{await expect(result.current.save()).rejects.toThrow("网络中断");});
  expect(result.current.status).toContain("保存失败");
  await act(async()=>{await result.current.save();});
  expect(save.mock.calls.map(call=>call[1].version)).toEqual([4,4]);
  await waitFor(()=>expect(result.current.status).toBe("已保存"));
 });
});
