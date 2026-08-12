# Processing Service Directory Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a `/service` processing-service directory that reuses vendor profiles and exposes processing capability fields, filters, and admin maintenance.

**Architecture:** Extend `Vendor` with processing fields, keep tag reuse through `vendor_tags`, and add processing-specific public list/filter endpoints. Frontend adds a dedicated route and card component for processing-enabled vendors while preserving the existing vendor directory.

**Tech Stack:** Go, Gin, GORM, MySQL, React, TypeScript, Vite, Vitest, Testing Library.

---

## Files

- Modify `backend/internal/model/vendor.go`: add processing fields to the vendor model.
- Modify `backend/internal/api/router.go`: register processing public routes.
- Modify `backend/internal/api/public_handlers.go`: add `ProcessingVendors` and `ProcessingFilterOptions`.
- Modify `backend/internal/api/public_handlers_test.go`: add serialization and API behavior tests.
- Modify `backend/internal/api/admin_handlers_vendor_test.go`: add vendor processing-field admin payload test.
- Modify `backend/internal/database/seed.go`: add processing tags and processing-capable seed vendors.
- Modify `backend/internal/database/seed_test.go`: verify processing tags and seeded vendor fields.
- Modify `frontend/src/types/api.ts`: add processing fields to `Vendor`, add processing filter type if needed.
- Modify `frontend/src/api/public.ts`: add `listProcessingVendors` and `getProcessingFilterOptions`.
- Create `frontend/src/pages/ProcessingServicesPage.tsx`: render `/service` directory.
- Create `frontend/src/pages/ProcessingServicesPage.test.tsx`: test page rendering and API calls.
- Modify `frontend/src/App.tsx`: route `/service` to the new page.
- Modify `frontend/src/pages/VendorDetailPage.tsx`: show processing service section when enabled.
- Modify `frontend/src/pages/VendorDetailPage.test.tsx`: test processing section.
- Modify `frontend/src/pages/admin/AdminResourcePage.tsx`: add processing fields and processing tag option.
- Modify `frontend/src/pages/admin/AdminResourcePage.test.tsx`: test processing fields and tag type.
- Modify `frontend/src/styles/global.css`: add focused styles only if existing classes cannot express the processing card.

---

### Task 1: Backend Model And Public API

**Files:**
- Modify `backend/internal/model/vendor.go`
- Modify `backend/internal/api/router.go`
- Modify `backend/internal/api/public_handlers.go`
- Modify `backend/internal/api/public_handlers_test.go`

- [ ] **Step 1: Write failing backend tests**

Add tests to `backend/internal/api/public_handlers_test.go`:

```go
func TestVendorDetailPayloadIncludesProcessingFields(t *testing.T) {
	payload, err := json.Marshal(Response{
		Code:    0,
		Message: "ok",
		Data: model.Vendor{
			Name:                "测试加工厂商",
			ProvidesProcessing:  true,
			ProcessingServices:  "来图来样加工、数控车削",
			ProcessingMaterials: "钢件、铸铁件",
			ProcessingEquipment: "数控车床、加工中心",
			ProcessingCapacity:  "小批量 3 天交付，批量订单按图报价",
			ProcessingRegions:   "华北、华东",
			ProcessingNotes:     "支持图纸、样件和批量代工",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	body := string(payload)
	for _, want := range []string{
		`"providesProcessing":true`,
		`"processingServices":"来图来样加工、数控车削"`,
		`"processingMaterials":"钢件、铸铁件"`,
		`"processingEquipment":"数控车床、加工中心"`,
		`"processingCapacity":"小批量 3 天交付，批量订单按图报价"`,
		`"processingRegions":"华北、华东"`,
		`"processingNotes":"支持图纸、样件和批量代工"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("body = %s, want %s", body, want)
		}
	}
}
```

- [ ] **Step 2: Run the failing backend test**

Run:

```bash
go test ./internal/api -run TestVendorDetailPayloadIncludesProcessingFields
```

Expected: FAIL because `model.Vendor` does not yet define the processing fields.

- [ ] **Step 3: Add vendor model fields**

Add fields to `backend/internal/model/vendor.go` after `Certifications`:

```go
ProvidesProcessing  bool   `gorm:"default:false;index" json:"providesProcessing"`
ProcessingServices  string `gorm:"size:500" json:"processingServices"`
ProcessingMaterials string `gorm:"size:500" json:"processingMaterials"`
ProcessingEquipment string `gorm:"type:text" json:"processingEquipment"`
ProcessingCapacity  string `gorm:"size:500" json:"processingCapacity"`
ProcessingRegions   string `gorm:"size:500" json:"processingRegions"`
ProcessingNotes     string `gorm:"type:text" json:"processingNotes"`
```

- [ ] **Step 4: Verify processing field serialization passes**

Run:

```bash
go test ./internal/api -run TestVendorDetailPayloadIncludesProcessingFields
```

Expected: PASS.

- [ ] **Step 5: Write failing public endpoint tests**

Add a test using an in-memory SQLite DB or the existing test DB helper if present. The test should create visible and non-processing vendors, processing and vendor tags, then call:

```go
req := httptest.NewRequest(http.MethodGet, "/api/processing-vendors?keyword=数控&tagId=2", nil)
```

Assert the response includes only the processing-enabled vendor and excludes normal vendors. Add a second request to `/api/processing-filter-options` and assert it returns only `tagType = "processing"` tags.

- [ ] **Step 6: Run the failing public endpoint tests**

Run:

```bash
go test ./internal/api -run Processing
```

Expected: FAIL because the routes and handlers do not exist yet.

- [ ] **Step 7: Implement public processing endpoints**

Add routes in `RegisterPublicRoutes`:

```go
api.GET("/processing-vendors", handler.ProcessingVendors)
api.GET("/processing-filter-options", handler.ProcessingFilterOptions)
```

Add handlers to `PublicHandler`:

```go
func (h PublicHandler) ProcessingVendors(c *gin.Context) {
	var vendors []model.Vendor
	page, pageSize := pageParams(c, 12)
	query := h.DB.Model(&model.Vendor{}).Preload("Tags").Where("is_visible = ? AND provides_processing = ?", true, true)
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("name LIKE ? OR short_name LIKE ? OR main_products LIKE ? OR processing_services LIKE ? OR processing_materials LIKE ? OR processing_equipment LIKE ? OR processing_notes LIKE ?", like, like, like, like, like, like, like)
	}
	if province := strings.TrimSpace(c.Query("province")); province != "" {
		query = query.Where("province = ?", province)
	}
	if tagID := queryUint(c, "tagId"); tagID > 0 {
		query = query.Joins("JOIN vendor_tags ON vendor_tags.vendor_id = vendors.id AND vendor_tags.tag_id = ?", tagID)
	}
	if c.Query("sort") == "latest" {
		query = query.Order("created_at desc")
	} else {
		query = query.Order("is_recommended desc, sort_order asc, id asc")
	}
	OK(c, paginate(query, &vendors, page, pageSize))
}

func (h PublicHandler) ProcessingFilterOptions(c *gin.Context) {
	var provinces []string
	var tags []model.Tag
	h.DB.Model(&model.Vendor{}).Where("is_visible = ? AND provides_processing = ? AND province <> ''", true, true).Distinct().Order("province asc").Pluck("province", &provinces)
	h.DB.Where("tag_type = ?", "processing").Order("sort_order asc, id asc").Find(&tags)
	OK(c, gin.H{"provinces": provinces, "serviceTags": tags})
}
```

- [ ] **Step 8: Verify backend API tests**

Run:

```bash
go test ./internal/api
```

Expected: PASS.

---

### Task 2: Admin Persistence And Seed Data

**Files:**
- Modify `backend/internal/api/admin_handlers_vendor_test.go`
- Modify `backend/internal/database/seed.go`
- Modify `backend/internal/database/seed_test.go`

- [ ] **Step 1: Write failing admin persistence test**

Add a test that marshals or saves a vendor payload with:

```go
ProvidesProcessing: true,
ProcessingServices: "数控车削、焊接加工",
ProcessingMaterials: "钢件、轴套、齿轮坯",
ProcessingEquipment: "数控车床、焊接工位",
ProcessingCapacity: "支持小批量试制和批量代工",
ProcessingRegions: "全国发货",
ProcessingNotes: "来图来样均可",
```

Assert the response or saved row includes those values.

- [ ] **Step 2: Run failing admin test**

Run:

```bash
go test ./internal/api -run Processing
```

Expected: FAIL until the model fields and save flow are available together.

- [ ] **Step 3: Write failing seed tests**

Add to `backend/internal/database/seed_test.go`:

```go
func TestDefaultSeedContainsProcessingTagsAndVendors(t *testing.T) {
	seed := DefaultSeed()
	processingTags := 0
	for _, tag := range seed.Tags {
		if tag.TagType == "processing" {
			processingTags++
		}
	}
	if processingTags < 4 {
		t.Fatalf("processing tags = %d, want at least 4", processingTags)
	}
	processingVendors := 0
	for _, vendor := range seed.Vendors {
		if vendor.ProvidesProcessing {
			processingVendors++
			if strings.TrimSpace(vendor.ProcessingServices) == "" {
				t.Fatalf("processing vendor %q missing services", vendor.Name)
			}
		}
	}
	if processingVendors < 3 {
		t.Fatalf("processing vendors = %d, want at least 3", processingVendors)
	}
}
```

- [ ] **Step 4: Run failing seed test**

Run:

```bash
go test ./internal/database -run Processing
```

Expected: FAIL because no processing tags or processing-enabled vendors are seeded yet.

- [ ] **Step 5: Add seed data**

Extend `seedTags()` with processing tags:

```go
{Name: "来图来样加工", TagType: "processing", Color: "blue", SortOrder: 101},
{Name: "数控车削", TagType: "processing", Color: "blue", SortOrder: 102},
{Name: "焊接加工", TagType: "processing", Color: "orange", SortOrder: 103},
{Name: "热处理", TagType: "processing", Color: "green", SortOrder: 104},
{Name: "批量代工", TagType: "processing", Color: "green", SortOrder: 105},
```

Set several seeded vendors to `ProvidesProcessing: true` and populate the processing fields with varied text.

- [ ] **Step 6: Verify backend seed and admin tests**

Run:

```bash
go test ./internal/api ./internal/database
```

Expected: PASS.

---

### Task 3: Frontend Service Directory

**Files:**
- Modify `frontend/src/types/api.ts`
- Modify `frontend/src/api/public.ts`
- Create `frontend/src/pages/ProcessingServicesPage.tsx`
- Create `frontend/src/pages/ProcessingServicesPage.test.tsx`
- Modify `frontend/src/App.tsx`
- Modify `frontend/src/styles/global.css` only if needed

- [ ] **Step 1: Write failing page test**

Create `frontend/src/pages/ProcessingServicesPage.test.tsx`:

```tsx
import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { getHome, getProcessingFilterOptions, listProcessingVendors } from "../api/public";
import { ProcessingServicesPage } from "./ProcessingServicesPage";

vi.mock("../api/public", () => ({
  getHome: vi.fn(),
  getProcessingFilterOptions: vi.fn(),
  listProcessingVendors: vi.fn(),
}));

describe("ProcessingServicesPage", () => {
  beforeEach(() => {
    vi.mocked(getHome).mockResolvedValue({
      topMenus: [],
      sidebarMenus: [],
      auxiliaryMenus: [],
      mobileMenus: [],
      banner: { title: "" },
      recommendedVendors: [],
      moreVendors: [],
      stats: [],
      safeguards: [],
      join: { text: "", buttonText: "", path: "/" },
    });
    vi.mocked(getProcessingFilterOptions).mockResolvedValue({
      provinces: ["山东"],
      categories: [],
      serviceTags: [{ id: 21, name: "数控车削", tagType: "processing" }],
    });
    vi.mocked(listProcessingVendors).mockResolvedValue({
      items: [{
        id: 3,
        name: "山东精工加工有限公司",
        province: "山东",
        providesProcessing: true,
        processingServices: "数控车削、来图来样加工",
        processingEquipment: "数控车床、加工中心",
        tags: [{ id: 21, name: "数控车削", tagType: "processing" }],
      }],
      page: 1,
      pageSize: 12,
      total: 1,
    });
  });

  it("renders processing vendors with processing filters", async () => {
    render(<MemoryRouter><ProcessingServicesPage /></MemoryRouter>);
    expect(await screen.findByRole("heading", { name: "加工服务" })).toBeInTheDocument();
    expect(screen.getByText("山东精工加工有限公司")).toBeInTheDocument();
    expect(screen.getByText(/数控车削、来图来样加工/)).toBeInTheDocument();
    await waitFor(() => expect(vi.mocked(listProcessingVendors)).toHaveBeenCalledWith(expect.objectContaining({ page: 1, pageSize: 12 })));
  });
});
```

- [ ] **Step 2: Run failing frontend page test**

Run:

```bash
npm run test -- ProcessingServicesPage
```

Expected: FAIL because the page and API functions do not exist yet.

- [ ] **Step 3: Add frontend types and API functions**

Extend `Vendor` in `frontend/src/types/api.ts` with the processing fields from the backend model.

Add to `frontend/src/api/public.ts`:

```ts
export function listProcessingVendors(params: VendorListParams = {}) {
  return publicClient.get<never, PageResult<Vendor>>("/api/processing-vendors", { params });
}

export function getProcessingFilterOptions() {
  return publicClient.get<never, FilterOptions>("/api/processing-filter-options");
}
```

- [ ] **Step 4: Implement `ProcessingServicesPage` and route**

Build the page from the `VendorsPage` pattern, but call processing APIs, use title `加工服务`, and display processing service/equipment/capacity text. Update `App.tsx` so `/service` renders `<ProcessingServicesPage />`.

- [ ] **Step 5: Verify frontend page test**

Run:

```bash
npm run test -- ProcessingServicesPage
```

Expected: PASS.

---

### Task 4: Frontend Admin And Vendor Detail

**Files:**
- Modify `frontend/src/pages/VendorDetailPage.tsx`
- Modify `frontend/src/pages/VendorDetailPage.test.tsx`
- Modify `frontend/src/pages/admin/AdminResourcePage.tsx`
- Modify `frontend/src/pages/admin/AdminResourcePage.test.tsx`

- [ ] **Step 1: Write failing vendor detail test**

Add to `VendorDetailPage.test.tsx` a processing-enabled vendor and assert the detail page includes:

```tsx
expect(screen.getByText("加工服务能力")).toBeInTheDocument();
expect(screen.getByText(/数控车削、焊接加工/)).toBeInTheDocument();
expect(screen.getByText(/数控车床、焊接工位/)).toBeInTheDocument();
```

- [ ] **Step 2: Write failing admin form test**

Add to `AdminResourcePage.test.tsx`:

```tsx
it("submits processing service fields for vendors", async () => {
  renderAdmin("/admin/vendors");
  fireEvent.change(await screen.findByLabelText("厂商名称"), { target: { value: "山东精工加工有限公司" } });
  fireEvent.click(screen.getByLabelText("是否提供加工服务"));
  fireEvent.change(screen.getByLabelText("加工服务能力"), { target: { value: "数控车削、焊接加工" } });
  fireEvent.change(screen.getByLabelText("加工设备"), { target: { value: "数控车床、焊接工位" } });
  const form = screen.getByLabelText("厂商名称").closest("form") as HTMLFormElement;
  fireEvent.click(within(form).getByRole("button", { name: "新增" }));
  await waitFor(() =>
    expect(mockedCreateResource).toHaveBeenCalledWith(
      "vendors",
      expect.objectContaining({
        providesProcessing: true,
        processingServices: "数控车削、焊接加工",
        processingEquipment: "数控车床、焊接工位",
      }),
    ),
  );
});
```

Add a tag type assertion that the tags form includes `加工服务` with value `processing`.

- [ ] **Step 3: Run failing frontend admin/detail tests**

Run:

```bash
npm run test -- VendorDetailPage AdminResourcePage
```

Expected: FAIL because UI fields and detail section do not exist.

- [ ] **Step 4: Implement vendor detail processing section**

Show a section only when `vendor.providesProcessing` is true. Include processing services, materials, equipment, capacity, regions, and notes only when values are present.

- [ ] **Step 5: Implement admin form fields**

Add processing fields to the vendor schema and add `{ label: "加工服务", value: "processing" }` to the tag type select options.

- [ ] **Step 6: Verify frontend admin/detail tests**

Run:

```bash
npm run test -- VendorDetailPage AdminResourcePage
```

Expected: PASS.

---

### Task 5: Full Verification

**Files:**
- No new files unless test output reveals focused fixes.

- [ ] **Step 1: Format Go files**

Run:

```bash
gofmt -w backend/internal/model/vendor.go backend/internal/api/public_handlers.go backend/internal/api/router.go backend/internal/api/public_handlers_test.go backend/internal/api/admin_handlers_vendor_test.go backend/internal/database/seed.go backend/internal/database/seed_test.go
```

- [ ] **Step 2: Run backend tests**

Run:

```bash
go test ./...
```

Expected: PASS.

- [ ] **Step 3: Run frontend tests**

Run:

```bash
npm run test
```

Expected: PASS.

- [ ] **Step 4: Run frontend build**

Run:

```bash
npm run build
```

Expected: PASS.

- [ ] **Step 5: Inspect diff**

Run:

```bash
git diff --stat
git diff --check
```

Expected: no whitespace errors, and changed files match the implementation scope.

---

## Self-Review

- Spec coverage: data fields, public API, admin maintenance, `/service` page, vendor detail section, seed data, and tests are covered by Tasks 1-5.
- Scope check: order workflow and independent service listings remain out of scope.
- Type consistency: backend JSON fields and frontend `Vendor` fields use `providesProcessing`, `processingServices`, `processingMaterials`, `processingEquipment`, `processingCapacity`, `processingRegions`, and `processingNotes`.
