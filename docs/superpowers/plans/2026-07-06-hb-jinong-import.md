# 河北冀农半自动导入 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add 河北冀农 as a real vendor imported from public website information, with source tracking and manual review state.

**Architecture:** Extend the existing `Vendor` model with source/review metadata, seed 河北冀农 and representative products through the existing default seed path, and expose the new fields through the existing admin and public vendor flows. Keep the import deliberately small and reviewable for MVP.

**Tech Stack:** Go, Gin, GORM, MySQL, React, TypeScript, Vitest.

---

### Task 1: Source Metadata

**Files:**
- Modify: `backend/internal/model/vendor.go`
- Modify: `frontend/src/types/api.ts`
- Modify: `frontend/src/pages/admin/AdminResourcePage.tsx`
- Test: `backend/internal/api/public_handlers_test.go`
- Test: `backend/internal/api/admin_handlers_vendor_test.go`
- Test: `frontend/src/pages/admin/AdminResourcePage.test.tsx`

- [ ] Add failing tests that expect `sourceUrl`, `sourceNote`, and `reviewStatus` to bind and serialize.
- [ ] Run the focused backend and frontend tests and confirm the new assertions fail before implementation.
- [ ] Add `SourceURL`, `SourceNote`, and `ReviewStatus` fields to the Go and TypeScript vendor types.
- [ ] Add backend validation that defaults blank `reviewStatus` to `pending` and rejects values other than `pending`, `verified`, and `rejected`.
- [ ] Add the fields to the admin vendor form.
- [ ] Run the focused tests and confirm they pass.

### Task 2: 河北冀农 Seed Data

**Files:**
- Modify: `backend/internal/database/seed.go`
- Test: `backend/internal/database/seed_test.go`

- [ ] Add failing seed tests that locate 河北冀农 by website URL, assert source metadata, assert review status is `pending`, and assert representative products are linked to the seeded vendor order.
- [ ] Run `go test ./internal/database` and confirm the test fails before implementation.
- [ ] Add 河北冀农 to `defaultVendors()` with publicly sourced company facts and a stable `SortOrder`.
- [ ] Add representative products for 液压翻转犁、旋耕机、驱动耙 associated with that vendor.
- [ ] Run `go test ./internal/database` and confirm it passes.

### Task 3: Public Detail Review Hint

**Files:**
- Modify: `frontend/src/pages/VendorDetailPage.tsx`
- Test: `frontend/src/pages/VendorDetailPage.test.tsx`

- [ ] Add a failing frontend test that expects a pending review badge and source link when `reviewStatus` is `pending`.
- [ ] Run the focused Vitest file and confirm the test fails before implementation.
- [ ] Render a small review badge and a source information card on the vendor detail page.
- [ ] Run the focused Vitest file and confirm it passes.

### Task 4: Verification

**Files:**
- No new files.

- [ ] Run `go test ./...` from `backend`.
- [ ] Run `npm run test -- --run` from `frontend`.
- [ ] Run `npm run build` from `frontend`.
- [ ] Review `git diff --check`.
