# 大陆农机配件 MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a runnable MySQL-backed “大陆农机配件” public website and CMS MVP matching the approved design and homepage mockup.

**Architecture:** The repository will contain a Gin/GORM backend in `backend/` and a Vite React/TypeScript frontend in `frontend/`. The backend exposes data-driven public APIs and authenticated admin CRUD APIs; the frontend consumes those APIs for both the public homepage and CMS.

**Tech Stack:** Go 1.25, Gin, GORM, MySQL 5.7, React, TypeScript, Vite, React Router, Axios, Vitest, Testing Library.

---

## File Structure

Create backend:

- `backend/go.mod`: Go module and dependencies.
- `backend/cmd/server/main.go`: API server entrypoint.
- `backend/internal/config/config.go`: environment-backed configuration.
- `backend/internal/database/database.go`: MySQL connection, AutoMigrate, seed orchestration.
- `backend/internal/database/seed.go`: default menu, vendor, product, banner, config, admin seed data.
- `backend/internal/model/*.go`: GORM models.
- `backend/internal/api/response.go`: unified JSON response helpers.
- `backend/internal/api/public_handlers.go`: health, home, menus, vendors, products, search.
- `backend/internal/api/admin_handlers.go`: login and admin CRUD handlers.
- `backend/internal/api/router.go`: Gin route registration.
- `backend/internal/auth/auth.go`: password hashing and token helpers.
- `backend/internal/service/home.go`: homepage aggregation and menu tree construction.
- `backend/internal/service/home_test.go`: menu tree and home aggregation tests.
- `backend/internal/auth/auth_test.go`: token and password tests.
- `backend/internal/api/admin_auth_test.go`: admin auth middleware tests.
- `backend/README.md`: backend run instructions.

Create frontend:

- `frontend/package.json`: Vite scripts and dependencies.
- `frontend/index.html`: app root.
- `frontend/vite.config.ts`: Vite/Vitest configuration.
- `frontend/tsconfig.json`, `frontend/tsconfig.node.json`: TypeScript config.
- `frontend/src/main.tsx`: React app entrypoint.
- `frontend/src/App.tsx`: route tree.
- `frontend/src/api/client.ts`: Axios client and auth interceptor.
- `frontend/src/api/public.ts`: public API calls.
- `frontend/src/api/admin.ts`: admin API calls.
- `frontend/src/types/api.ts`: shared API types.
- `frontend/src/utils/navigation.ts`: search and vendor navigation helpers.
- `frontend/src/utils/navigation.test.ts`: navigation behavior tests.
- `frontend/src/pages/HomePage.tsx`: public homepage.
- `frontend/src/pages/SearchPage.tsx`, `VendorDetailPage.tsx`, `PlaceholderPage.tsx`: minimal public routes.
- `frontend/src/pages/admin/*.tsx`: admin login, dashboard, CRUD pages.
- `frontend/src/components/public/*.tsx`: public homepage components.
- `frontend/src/components/admin/*.tsx`: admin layout, table, forms.
- `frontend/src/styles/global.css`: shared styles and responsive layout.
- `frontend/README.md`: frontend run instructions.

Modify:

- `README.md`: full-stack run instructions.
- `docs/superpowers/plans/2026-06-25-dalu-nongji-parts-mvp.md`: check off tasks as executed.

---

### Task 1: Repository Scaffold

**Files:**
- Create: `README.md`
- Create: `backend/go.mod`
- Create: `backend/cmd/server/main.go`
- Create: `frontend/package.json`
- Create: `frontend/index.html`
- Create: `frontend/src/main.tsx`
- Create: `frontend/src/App.tsx`

- [ ] **Step 1: Write scaffold smoke test**

Create `backend/internal/config/config_test.go`:

```go
package config

import "testing"

func TestDefaultConfigUsesLocalMySQL(t *testing.T) {
	cfg := Load()
	if cfg.DBHost != "127.0.0.1" {
		t.Fatalf("DBHost = %q, want 127.0.0.1", cfg.DBHost)
	}
	if cfg.DBPort != "13306" {
		t.Fatalf("DBPort = %q, want 13306", cfg.DBPort)
	}
	if cfg.DBUser != "root" || cfg.DBPassword != "root" {
		t.Fatalf("default credentials = %q/%q, want root/root", cfg.DBUser, cfg.DBPassword)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```powershell
cd backend
go test ./internal/config
```

Expected: FAIL because `backend` module and `config.Load` do not exist.

- [ ] **Step 3: Create minimal backend scaffold**

Create `backend/go.mod`:

```go
module dalu-nongji-parts/backend

go 1.25

require (
	github.com/gin-contrib/cors v1.7.6
	github.com/gin-gonic/gin v1.10.1
	golang.org/x/crypto v0.31.0
	gorm.io/driver/mysql v1.5.7
	gorm.io/gorm v1.25.12
)
```

Create `backend/internal/config/config.go`:

```go
package config

import (
	"fmt"
	"os"
)

type Config struct {
	HTTPAddr      string
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	AdminUsername string
	AdminPassword string
	AuthSecret    string
}

func Load() Config {
	cfg := Config{
		HTTPAddr:      env("HTTP_ADDR", ":8080"),
		DBHost:        env("DB_HOST", "127.0.0.1"),
		DBPort:        env("DB_PORT", "13306"),
		DBUser:        env("DB_USER", "root"),
		DBPassword:    env("DB_PASSWORD", "root"),
		DBName:        env("DB_NAME", "dl_nongji_parts"),
		AdminUsername: env("ADMIN_USERNAME", "admin"),
		AdminPassword: env("ADMIN_PASSWORD", "admin123"),
		AuthSecret:    env("AUTH_SECRET", "dev-secret-change-me"),
	}
	return cfg
}

func (c Config) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
```

Create `backend/cmd/server/main.go`:

```go
package main

import (
	"log"

	"dalu-nongji-parts/backend/internal/config"
)

func main() {
	cfg := config.Load()
	log.Printf("starting API on %s", cfg.HTTPAddr)
}
```

- [ ] **Step 4: Run backend test to verify it passes**

Run:

```powershell
cd backend
go mod tidy
go test ./internal/config
```

Expected: PASS.

- [ ] **Step 5: Create minimal frontend scaffold**

Create `frontend/package.json`:

```json
{
  "scripts": {
    "dev": "vite --host 127.0.0.1",
    "build": "tsc -b && vite build",
    "test": "vitest run"
  },
  "dependencies": {
    "@vitejs/plugin-react": "^5.0.0",
    "axios": "^1.7.9",
    "lucide-react": "^0.468.0",
    "react": "^19.0.0",
    "react-dom": "^19.0.0",
    "react-router-dom": "^7.0.0"
  },
  "devDependencies": {
    "@testing-library/jest-dom": "^6.6.3",
    "@testing-library/react": "^16.1.0",
    "@types/react": "^19.0.1",
    "@types/react-dom": "^19.0.2",
    "jsdom": "^25.0.1",
    "typescript": "^5.7.2",
    "vite": "^6.0.3",
    "vitest": "^2.1.8"
  }
}
```

Create a minimal React app that renders `大陆农机配件`.

- [ ] **Step 6: Verify frontend scaffold**

Run:

```powershell
cd frontend
npm install
npm run build
```

Expected: build succeeds.

- [ ] **Step 7: Commit**

```powershell
git add README.md backend frontend
git commit -m "chore: scaffold full-stack app"
```

---

### Task 2: Backend Models, Migration, and Seed Data

**Files:**
- Create: `backend/internal/model/menu.go`
- Create: `backend/internal/model/vendor.go`
- Create: `backend/internal/model/tag.go`
- Create: `backend/internal/model/category.go`
- Create: `backend/internal/model/product.go`
- Create: `backend/internal/model/banner.go`
- Create: `backend/internal/model/site_config.go`
- Create: `backend/internal/model/admin_user.go`
- Create: `backend/internal/database/database.go`
- Create: `backend/internal/database/seed.go`
- Test: `backend/internal/service/home_test.go`

- [ ] **Step 1: Write failing menu tree test**

Create `backend/internal/service/home_test.go`:

```go
package service

import (
	"testing"

	"dalu-nongji-parts/backend/internal/model"
)

func TestBuildMenuTreeSortsAndNestsMenus(t *testing.T) {
	menus := []model.Menu{
		{ID: 3, Name: "变速箱齿轮", ParentID: 2, SortOrder: 1, IsEnabled: true},
		{ID: 1, Name: "首页", ParentID: 0, SortOrder: 1, IsEnabled: true},
		{ID: 2, Name: "传动配件", ParentID: 0, SortOrder: 2, IsEnabled: true, IsDefaultOpen: true},
		{ID: 4, Name: "禁用", ParentID: 0, SortOrder: 0, IsEnabled: false},
	}

	tree := BuildMenuTree(menus)
	if len(tree) != 2 {
		t.Fatalf("len(tree) = %d, want 2", len(tree))
	}
	if tree[0].Name != "首页" {
		t.Fatalf("first menu = %q, want 首页", tree[0].Name)
	}
	if len(tree[1].Children) != 1 || tree[1].Children[0].Name != "变速箱齿轮" {
		t.Fatalf("children = %#v, want nested child", tree[1].Children)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```powershell
cd backend
go test ./internal/service
```

Expected: FAIL because `model.Menu` and `BuildMenuTree` do not exist.

- [ ] **Step 3: Implement models and menu tree**

Implement GORM models using the fields from the design. Implement:

```go
func BuildMenuTree(menus []model.Menu) []model.Menu
```

Rules:

- Ignore disabled menus.
- Sort by `SortOrder` ascending, then `ID` ascending.
- Attach children to matching `ParentID`.
- Preserve `IsDefaultOpen`.

- [ ] **Step 4: Run test to verify it passes**

Run:

```powershell
cd backend
go test ./internal/service
```

Expected: PASS.

- [ ] **Step 5: Add migration and seed tests**

Create a test that uses SQLite is not allowed because runtime DB is MySQL and GORM dialect differences matter. Instead write pure seed builder tests:

```go
func TestDefaultSeedContainsTransmissionChildren(t *testing.T) {
	seed := DefaultSeed()
	children := 0
	for _, menu := range seed.Menus {
		if menu.ParentKey == "transmission" {
			children++
		}
	}
	if children != 7 {
		t.Fatalf("transmission children = %d, want 7", children)
	}
}
```

- [ ] **Step 6: Implement database connection and seed data**

Implement:

- `Connect(cfg config.Config) (*gorm.DB, error)`
- `AutoMigrate(db *gorm.DB) error`
- `SeedDefaults(db *gorm.DB, cfg config.Config) error`

Seed must include:

- Top nav entries: 首页、配件产品、厂商目录、加工服务、采购信息.
- Sidebar entries from the requirements, with `传动配件` default open and 7 children.
- Auxiliary links: 提交厂商、友情链接、关于平台.
- Mobile quick entries from the requirements.
- At least 12 vendors, 5 recommended.
- Tags: 源头厂商、支持定制、可开发票、现货充足、实地认证.
- At least 8 categories and 10 products.
- One active Banner.
- Stats configs and join banner copy.
- Default admin user.

- [ ] **Step 7: Run backend tests**

Run:

```powershell
cd backend
go test ./...
```

Expected: PASS.

- [ ] **Step 8: Commit**

```powershell
git add backend
git commit -m "feat: add backend models and seed data"
```

---

### Task 3: Backend Public API

**Files:**
- Create: `backend/internal/api/response.go`
- Create: `backend/internal/api/public_handlers.go`
- Create: `backend/internal/api/router.go`
- Modify: `backend/cmd/server/main.go`
- Test: `backend/internal/api/public_handlers_test.go`

- [ ] **Step 1: Write failing `/api/home` handler test**

Create a Gin test using an in-memory service stub:

```go
func TestHomeHandlerReturnsHomePayload(t *testing.T) {
	router := gin.New()
	RegisterPublicRoutes(router, PublicDeps{Home: fakeHomeService{}})

	req := httptest.NewRequest(http.MethodGet, "/api/home", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "recommendedVendors") {
		t.Fatalf("body = %s, want recommendedVendors", rec.Body.String())
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```powershell
cd backend
go test ./internal/api
```

Expected: FAIL because API router does not exist.

- [ ] **Step 3: Implement public API**

Implement:

- Unified `OK(c, data)` and `Fail(c, status, code, message)`.
- Public routes listed in the design.
- `HomeService.GetHome(ctx)` using GORM.
- Vendor list filters: `keyword`, `province`, `tag`, `page`, `pageSize`, `sort`.
- Search over vendor name/main products, product name, category name.

- [ ] **Step 4: Run tests to verify green**

Run:

```powershell
cd backend
go test ./internal/api ./internal/service
```

Expected: PASS.

- [ ] **Step 5: Start backend against MySQL**

Run:

```powershell
cd backend
go run ./cmd/server
```

Expected: server starts on `:8080`, migrates tables, and seeds data. If the database does not exist, create `dl_nongji_parts` in MySQL and rerun.

- [ ] **Step 6: Smoke public API**

Run:

```powershell
curl.exe -s http://127.0.0.1:8080/api/health
curl.exe -s http://127.0.0.1:8080/api/home
```

Expected: health returns ok; home contains menus, banner, vendors, configs.

- [ ] **Step 7: Commit**

```powershell
git add backend
git commit -m "feat: expose public API"
```

---

### Task 4: Backend Admin Auth and CRUD API

**Files:**
- Create: `backend/internal/auth/auth.go`
- Create: `backend/internal/auth/auth_test.go`
- Create: `backend/internal/api/admin_handlers.go`
- Create: `backend/internal/api/admin_auth_test.go`
- Modify: `backend/internal/api/router.go`

- [ ] **Step 1: Write failing auth tests**

Create `backend/internal/auth/auth_test.go`:

```go
package auth

import "testing"

func TestPasswordHashVerifies(t *testing.T) {
	hash, err := HashPassword("admin123")
	if err != nil {
		t.Fatal(err)
	}
	if !CheckPassword(hash, "admin123") {
		t.Fatal("expected password to verify")
	}
	if CheckPassword(hash, "wrong") {
		t.Fatal("expected wrong password to fail")
	}
}

func TestTokenRoundTrip(t *testing.T) {
	token, err := IssueToken("admin", "secret")
	if err != nil {
		t.Fatal(err)
	}
	username, err := ParseToken(token, "secret")
	if err != nil {
		t.Fatal(err)
	}
	if username != "admin" {
		t.Fatalf("username = %q, want admin", username)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```powershell
cd backend
go test ./internal/auth
```

Expected: FAIL because auth helpers do not exist.

- [ ] **Step 3: Implement auth helpers**

Use bcrypt for passwords. Use HMAC SHA-256 token format:

```text
base64(username|expiryUnix|signature)
```

Token expires after 24 hours.

- [ ] **Step 4: Run auth tests**

Run:

```powershell
cd backend
go test ./internal/auth
```

Expected: PASS.

- [ ] **Step 5: Write failing admin middleware test**

Test that `GET /api/admin/profile` without token returns 401 and with a valid token returns current username.

- [ ] **Step 6: Implement admin handlers**

Implement:

- `POST /api/admin/login`
- `GET /api/admin/profile`
- CRUD for menus, vendors, tags, categories, products, banners.
- Config list and update by key.

Validation:

- Required names cannot be blank.
- Sort order defaults to 0.
- Booleans default to sensible values.
- Deleting uses GORM soft delete where model has `DeletedAt`.

- [ ] **Step 7: Run tests**

Run:

```powershell
cd backend
go test ./...
```

Expected: PASS.

- [ ] **Step 8: Smoke admin API**

Run:

```powershell
curl.exe -s -X POST http://127.0.0.1:8080/api/admin/login -H "Content-Type: application/json" -d "{\"username\":\"admin\",\"password\":\"admin123\"}"
```

Expected: JSON response includes a token.

- [ ] **Step 9: Commit**

```powershell
git add backend
git commit -m "feat: add admin API"
```

---

### Task 5: Frontend API Layer and Navigation Tests

**Files:**
- Create: `frontend/src/types/api.ts`
- Create: `frontend/src/api/client.ts`
- Create: `frontend/src/api/public.ts`
- Create: `frontend/src/api/admin.ts`
- Create: `frontend/src/utils/navigation.ts`
- Create: `frontend/src/utils/navigation.test.ts`
- Modify: `frontend/vite.config.ts`

- [ ] **Step 1: Write failing navigation tests**

Create `frontend/src/utils/navigation.test.ts`:

```ts
import { describe, expect, it } from "vitest";
import { getSearchPath, getVendorEntryTarget } from "./navigation";

describe("navigation helpers", () => {
  it("creates encoded search paths", () => {
    expect(getSearchPath(" 液压油泵 ")).toBe("/search?keyword=%E6%B6%B2%E5%8E%8B%E6%B2%B9%E6%B3%B5");
  });

  it("does not create a search path for empty keywords", () => {
    expect(getSearchPath("   ")).toBe(null);
  });

  it("prefers vendor website when present", () => {
    expect(getVendorEntryTarget({ id: 7, websiteUrl: "https://example.com" })).toEqual({
      type: "external",
      href: "https://example.com",
    });
  });

  it("falls back to internal vendor detail", () => {
    expect(getVendorEntryTarget({ id: 7, websiteUrl: "" })).toEqual({
      type: "internal",
      href: "/vendors/7",
    });
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```powershell
cd frontend
npm run test -- src/utils/navigation.test.ts
```

Expected: FAIL because helper file does not exist.

- [ ] **Step 3: Implement API types and helpers**

Implement types for:

- `ApiResponse<T>`
- `Menu`
- `Vendor`
- `Tag`
- `Category`
- `Product`
- `Banner`
- `HomePayload`
- `SiteConfig`

Implement helpers exactly matching the tests.

- [ ] **Step 4: Run tests**

Run:

```powershell
cd frontend
npm run test -- src/utils/navigation.test.ts
```

Expected: PASS.

- [ ] **Step 5: Implement Axios clients**

Public client:

- Base URL from `VITE_API_BASE_URL`, fallback `http://127.0.0.1:8080`.
- Unwrap backend `{ code, message, data }`.

Admin client:

- Adds bearer token from `localStorage`.
- Clears token and redirects to `/admin/login` on 401.

- [ ] **Step 6: Build frontend**

Run:

```powershell
cd frontend
npm run build
```

Expected: PASS.

- [ ] **Step 7: Commit**

```powershell
git add frontend
git commit -m "feat: add frontend API layer"
```

---

### Task 6: Public Homepage

**Files:**
- Create: `frontend/src/pages/HomePage.tsx`
- Create: `frontend/src/components/public/PublicHeader.tsx`
- Create: `frontend/src/components/public/SidebarNav.tsx`
- Create: `frontend/src/components/public/HeroSearch.tsx`
- Create: `frontend/src/components/public/VendorCard.tsx`
- Create: `frontend/src/components/public/RecommendedVendors.tsx`
- Create: `frontend/src/components/public/MoreVendors.tsx`
- Create: `frontend/src/components/public/JoinBanner.tsx`
- Create: `frontend/src/components/public/StatsFooter.tsx`
- Create: `frontend/src/components/public/MobileHeader.tsx`
- Create: `frontend/src/components/public/MobileCategoryGrid.tsx`
- Create: `frontend/src/components/public/MobileBottomNav.tsx`
- Modify: `frontend/src/App.tsx`
- Modify: `frontend/src/styles/global.css`

- [ ] **Step 1: Write failing homepage render test**

Create `frontend/src/pages/HomePage.test.tsx`:

```tsx
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";
import { HomeView } from "./HomePage";

describe("HomeView", () => {
  it("renders the required homepage sections", () => {
    render(
      <MemoryRouter>
        <HomeView
          home={{
            topMenus: [{ id: 1, name: "首页", path: "/" }],
            sidebarMenus: [{ id: 2, name: "传动配件", isDefaultOpen: true, children: [{ id: 3, name: "变速箱齿轮" }] }],
            auxiliaryMenus: [{ id: 4, name: "提交厂商", path: "/join" }],
            mobileMenus: [{ id: 5, name: "农机易损件", icon: "wrench" }],
            banner: { title: "找农机配件，查源头厂商", subtitle: "原厂品质", searchPlaceholder: "搜索配件名称", hotKeywords: ["收割机"] },
            recommendedVendors: [{ id: 1, name: "山东沃得农机配件有限公司", mainProducts: "变速箱、链条", tags: [{ id: 1, name: "源头厂商" }] }],
            moreVendors: [],
            stats: [{ label: "入驻厂商", value: "2000+" }],
            safeguards: ["平台审核", "安心认证"],
            join: { text: "入驻成为厂商，展示您的产品与实力，获取更多采购商机", buttonText: "立即入驻", path: "/join" },
          }}
        />
      </MemoryRouter>
    );

    expect(screen.getByText("找农机配件，查源头厂商")).toBeInTheDocument();
    expect(screen.getByText("推荐厂商")).toBeInTheDocument();
    expect(screen.getByText("更多厂商")).toBeInTheDocument();
    expect(screen.queryByText("最新求购")).not.toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```powershell
cd frontend
npm run test -- src/pages/HomePage.test.tsx
```

Expected: FAIL because HomePage components do not exist.

- [ ] **Step 3: Implement homepage components**

Implement the sections from the approved design. Important layout rules:

- PC grid uses 5 columns for recommended vendors and more vendors at wide widths.
- Sidebar default-open menus show children.
- No homepage element contains “最新求购”.
- Mobile CSS shows top mobile header, grid shortcuts, list cards, and fixed bottom nav.
- Use lucide icons for menu, search, category, settings, user, home, factory, grid, wrench.

- [ ] **Step 4: Run homepage test**

Run:

```powershell
cd frontend
npm run test -- src/pages/HomePage.test.tsx
```

Expected: PASS.

- [ ] **Step 5: Integrate with `/api/home`**

`HomePage` should load:

```ts
const home = await getHome();
```

Show loading, error with retry, and empty states.

- [ ] **Step 6: Build frontend**

Run:

```powershell
cd frontend
npm run build
```

Expected: PASS.

- [ ] **Step 7: Commit**

```powershell
git add frontend
git commit -m "feat: build responsive public homepage"
```

---

### Task 7: Public Secondary Pages

**Files:**
- Create: `frontend/src/pages/SearchPage.tsx`
- Create: `frontend/src/pages/VendorDetailPage.tsx`
- Create: `frontend/src/pages/PlaceholderPage.tsx`
- Modify: `frontend/src/App.tsx`

- [ ] **Step 1: Write failing search helper test**

Add a test that rendering `/search?keyword=链条` displays the keyword and calls the search API.

- [ ] **Step 2: Implement pages**

Implement:

- `SearchPage`: keyword display and result sections for vendors/products/categories.
- `VendorDetailPage`: vendor info, tags, main products, contact, website button.
- `PlaceholderPage`: clean placeholder for products, join, about, purchase info.

- [ ] **Step 3: Run tests and build**

Run:

```powershell
cd frontend
npm run test
npm run build
```

Expected: PASS.

- [ ] **Step 4: Commit**

```powershell
git add frontend
git commit -m "feat: add public secondary pages"
```

---

### Task 8: CMS Login, Layout, and CRUD Screens

**Files:**
- Create: `frontend/src/pages/admin/AdminLoginPage.tsx`
- Create: `frontend/src/pages/admin/AdminDashboardPage.tsx`
- Create: `frontend/src/pages/admin/AdminResourcePage.tsx`
- Create: `frontend/src/components/admin/AdminLayout.tsx`
- Create: `frontend/src/components/admin/AdminTable.tsx`
- Create: `frontend/src/components/admin/ResourceForm.tsx`
- Modify: `frontend/src/App.tsx`
- Modify: `frontend/src/styles/global.css`

- [ ] **Step 1: Write failing admin route test**

Test that `AdminLoginPage` renders username and password inputs and a login button.

- [ ] **Step 2: Implement admin login**

Login flow:

- POST `/api/admin/login`.
- Store token in `localStorage`.
- Navigate to `/admin/dashboard`.
- Display API error message on failure.

- [ ] **Step 3: Implement admin layout**

Admin nav links:

- 控制台
- 导航菜单
- 厂商信息
- 厂商标签
- 配件分类
- 配件产品
- Banner 管理
- 平台配置

- [ ] **Step 4: Implement generic CRUD page**

Use a schema-driven `AdminResourcePage` for:

- menus
- vendors
- tags
- categories
- products
- banners

Each page supports:

- list
- create
- edit
- delete
- enabled/visible boolean fields
- text URL fields for images

Configs page supports editing key-value rows.

- [ ] **Step 5: Run tests and build**

Run:

```powershell
cd frontend
npm run test
npm run build
```

Expected: PASS.

- [ ] **Step 6: Manual CMS smoke**

With backend running:

- Open `/admin/login`.
- Login as `admin / admin123`.
- Create a test menu.
- Refresh public homepage and confirm new enabled sidebar menu appears.
- Edit a vendor website URL and confirm vendor card uses external target.

- [ ] **Step 7: Commit**

```powershell
git add frontend
git commit -m "feat: add CMS interface"
```

---

### Task 9: Full-Stack Run Scripts and Documentation

**Files:**
- Modify: `README.md`
- Create: `backend/README.md`
- Create: `frontend/README.md`
- Optional create: `scripts/dev.ps1`

- [ ] **Step 1: Write docs**

Root README must include:

```markdown
# 大陆农机配件

## Prerequisites

- Go 1.25+
- Node 24+
- MySQL 5.7 on `127.0.0.1:13306`

## Database

Create database:

```sql
CREATE DATABASE IF NOT EXISTS dl_nongji_parts DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

Default connection:

`root:root@127.0.0.1:13306/dl_nongji_parts`

## Run Backend

```powershell
cd backend
go run ./cmd/server
```

## Run Frontend

```powershell
cd frontend
npm install
npm run dev
```

## Admin

Open `/admin/login`.

Default account: `admin / admin123`
```
```

- [ ] **Step 2: Verify docs commands**

Run:

```powershell
cd backend
go test ./...
cd ..\frontend
npm run test
npm run build
```

Expected: PASS.

- [ ] **Step 3: Commit**

```powershell
git add README.md backend frontend
git commit -m "docs: add run instructions"
```

---

### Task 10: Visual and Integration Verification

**Files:**
- Modify CSS or components only if verification reveals layout defects.

- [ ] **Step 1: Start backend**

Run:

```powershell
cd backend
go run ./cmd/server
```

Expected: API listens on `http://127.0.0.1:8080`.

- [ ] **Step 2: Start frontend**

Run:

```powershell
cd frontend
npm run dev
```

Expected: Vite serves app, usually at `http://127.0.0.1:5173`.

- [ ] **Step 3: Browser verification**

Check:

- Desktop width 1440: top nav, left nav, banner, 5 recommended cards, 2 more-vendor rows.
- Mobile width 390: mobile top bar, search, 10 shortcut icons, vertical vendor cards, fixed bottom nav.
- No visible “最新求购”.
- Text does not overflow buttons or cards.
- Vendor entry button opens external website when URL exists.
- CMS login and one CRUD edit work.

- [ ] **Step 4: Final automated verification**

Run:

```powershell
cd backend
go test ./...
cd ..\frontend
npm run test
npm run build
```

Expected: PASS.

- [ ] **Step 5: Commit final polish**

```powershell
git add backend frontend README.md
git commit -m "chore: verify MVP"
```

---

## Self-Review Notes

- Spec coverage: This plan covers the public homepage, responsive mobile view, MySQL-backed API, CMS login, CMS CRUD, vendor website jump logic, search route, and no latest-purchase module.
- Placeholder scan: No unresolved placeholder values remain. Optional dev script is explicitly optional and not needed for acceptance.
- Type consistency: Backend JSON fields use lower camelCase; frontend types must match backend JSON tags.
