package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"dalu-nongji-parts/backend/internal/config"
	"dalu-nongji-parts/backend/internal/database"
	"dalu-nongji-parts/backend/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Explicitly opt in: only a newly created, uniquely named disposable database is
// migrated and dropped. DB_NAME is deliberately never used as the target.
func promotionTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	if os.Getenv("RUN_PROMOTION_MYSQL_TESTS") != "1" {
		t.Skip("set RUN_PROMOTION_MYSQL_TESTS=1 to use an isolated MySQL test database")
	}
	cfg := config.Load()
	server, err := gorm.Open(mysql.Open(cfg.ServerDSN()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal("test MySQL server unavailable")
	}
	name := fmt.Sprintf("dlc_promotion_test_%d", time.Now().UnixNano())
	if err := server.Exec("CREATE DATABASE " + name + " CHARACTER SET utf8mb4").Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if regexp.MustCompile("^dlc_promotion_test_[0-9]+$").MatchString(name) {
			if err := server.Exec("DROP DATABASE " + name).Error; err != nil {
				t.Error(err)
			}
		}
		sqlDB, _ := server.DB()
		_ = sqlDB.Close()
	})
	cfg.DBName = name
	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sqlDB, _ := db.DB(); _ = sqlDB.Close() })
	if err := database.AutoMigrate(db); err != nil {
		t.Fatal(err)
	}
	return db
}

func promotionRequest(t *testing.T, handler gin.HandlerFunc, params gin.Params, vendorID, userID uint, body any, expected int) []byte {
	t.Helper()
	raw, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/", bytes.NewReader(raw))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = params
	c.Set("role", "vendor")
	c.Set("vendorId", vendorID)
	c.Set("userId", userID)
	c.Set("username", "qa-vendor")
	handler(c)
	if w.Code != expected {
		t.Fatalf("status %d, want %d: %s", w.Code, expected, w.Body.String())
	}
	return w.Body.Bytes()
}

func TestVendorPromotionMySQL(t *testing.T) {
	db := promotionTestDB(t)
	gin.SetMode(gin.TestMode)
	vendor := model.Vendor{Name: "隔离验收厂商", Slug: "qa-vendor", PublicationStatus: "published", IsVisible: true, Phone: "13900000001", PhonePublic: false, ContentVersion: 1}
	if err := db.Create(&vendor).Error; err != nil {
		t.Fatal(err)
	}
	other := model.Vendor{Name: "另一测试厂商", Slug: "qa-other", PublicationStatus: "published", IsVisible: true}
	if err := db.Create(&other).Error; err != nil {
		t.Fatal(err)
	}
	account := model.AdminUser{Username: "qa-vendor", Role: "vendor", VendorID: &vendor.ID, IsEnabled: true}
	if err := db.Create(&account).Error; err != nil {
		t.Fatal(err)
	}
	cat := model.Category{Name: "测试配件", Slug: "qa-parts", IsEnabled: true}
	if err := db.Create(&cat).Error; err != nil {
		t.Fatal(err)
	}
	product := model.Product{Name: "公共型号", Slug: "qa-catalog", CategoryID: cat.ID, PublicationStatus: "published", Image: "/images/not-this-vendor.jpg"}
	if err := db.Create(&product).Error; err != nil {
		t.Fatal(err)
	}
	admin := AdminHandler{DB: db}
	payload := vendorProductDraft{VendorProductName: "本厂液压件", VendorModel: "QA-1", CategoryID: cat.ID, PriceUnit: "桶", MinOrderQuantity: 3, AvailableQuantity: 0, SupplyAbility: "历史月供100台"}
	save := func(key string, version uint, p vendorProductDraft, target string, id uint, status int) model.VendorWorkDraft {
		data := promotionRequest(t, admin.SaveWorkDraft, gin.Params{{Key: "key", Value: key}}, vendor.ID, account.ID, map[string]any{"kind": "product", "version": version, "payload": p, "targetType": target, "targetId": id}, status)
		var response struct{ Data model.VendorWorkDraft }
		_ = json.Unmarshal(data, &response)
		return response.Data
	}
	row := save("retryable-key", 0, payload, "", 0, 200)
	retry := save("retryable-key", 0, payload, "", 0, 200)
	if row.ID == 0 || row.ID != retry.ID || row.Version != retry.Version {
		t.Fatal("initial save retry must be idempotent")
	}
	changed := payload
	changed.VendorModel = "QA-2"
	save("retryable-key", 0, changed, "", 0, 409)
	params := gin.Params{{Key: "id", Value: fmt.Sprint(row.ID)}}
	promotionRequest(t, admin.CommitWorkDraft, params, other.ID, account.ID, map[string]any{"version": row.Version}, 404)
	promotionRequest(t, admin.CommitWorkDraft, params, vendor.ID, account.ID, map[string]any{"version": row.Version}, 200)
	promotionRequest(t, admin.CommitWorkDraft, params, vendor.ID, account.ID, map[string]any{"version": row.Version}, 200)
	var count int64
	db.Model(&model.ProductSubmission{}).Count(&count)
	if count != 1 {
		t.Fatalf("duplicate submissions: %d", count)
	}
	db.Model(&model.ProductSupplier{}).Count(&count)
	if count != 0 {
		t.Fatal("private submit published before review")
	}
	var submission model.ProductSubmission
	db.First(&submission)
	promotionRequest(t, admin.ReviewProductSubmission, gin.Params{{Key: "id", Value: fmt.Sprint(submission.ID)}}, vendor.ID, account.ID, map[string]any{"status": "approved", "resolution": "link_product", "targetProductId": product.ID}, 200)
	var supplier model.ProductSupplier
	db.First(&supplier)
	if supplier.AvailableQuantity != 0 || supplier.PriceUnit != "桶" || supplier.MinOrderQuantity != 3 {
		t.Fatal("historic supply values changed")
	}
	db.Model(&model.UserNotification{}).Where("user_id = ? AND business_type = ?", account.ID, "product_review").Count(&count)
	if count != 1 {
		t.Fatal("review notification missing")
	}
	public, err := loadShowroomSupplier(db, vendor.Slug, supplier.ID)
	if err != nil {
		t.Fatal(err)
	}
	if public.Image != "" || public.Vendor.Phone == vendor.Phone {
		t.Fatal("cross-product image or private phone leaked")
	}
	if _, err := loadShowroomSupplier(db, other.Slug, supplier.ID); err == nil {
		t.Fatal("wrong vendor detail accessible")
	}
	save("wrong-owner-key", 0, payload, "supplier", supplier.ID+999, 404)
	privateAsset := model.MediaAsset{Purpose: "capture_source", Status: "published", VendorID: &vendor.ID, OwnerUsername: "qa-vendor"}
	if err := db.Create(&privateAsset).Error; err != nil {
		t.Fatal(err)
	}
	changed.Image = fmt.Sprintf("/api/media/%d", privateAsset.ID)
	save("private-image-key", 0, changed, "", 0, 400)
	// Editing a published offer remains a pending submission; public name is unchanged.
	changed = payload
	changed.VendorProductName = "修改后名称"
	update := save("published-edit-key", 0, changed, "supplier", supplier.ID, 200)
	promotionRequest(t, admin.CommitWorkDraft, gin.Params{{Key: "id", Value: fmt.Sprint(update.ID)}}, vendor.ID, account.ID, map[string]any{"version": update.Version}, 200)
	public, _ = loadShowroomSupplier(db, vendor.Slug, supplier.ID)
	if public.VendorProductName != payload.VendorProductName {
		t.Fatal("pending edit overwrote public content")
	}
	// Public catalog heat is never backfilled to a vendor. Contact metrics are
	// separate events; visitors with both view and contact form the conversion rate.
	now := time.Now()
	events := []model.AnalyticsEvent{
		{EventType: "page_view", Path: showroomPath(public), VendorID: vendor.ID, SupplierID: supplier.ID, VisitorHash: "one", Source: "qr", EventWindow: 1, CreatedAt: now},
		{EventType: "contact_phone_click", Path: showroomPath(public), VendorID: vendor.ID, SupplierID: supplier.ID, VisitorHash: "one", Source: "qr", EventWindow: 1, CreatedAt: now},
		{EventType: "page_view", Path: "/v/qa-vendor", VendorID: vendor.ID, VisitorHash: "two", Source: "share", EventWindow: 1, CreatedAt: now},
		{EventType: "page_view", Path: "/products/qa-catalog", ContentType: "product", ContentID: product.ID, VisitorHash: "three", Source: "site", EventWindow: 1, CreatedAt: now},
	}
	if err := db.Create(&events).Error; err != nil {
		t.Fatal(err)
	}
	analytics := AnalyticsHandler{DB: db, Config: config.Config{AuthSecret: "test-only"}}
	metrics, metricsErr := analytics.showroomAnalytics(vendor.ID, now.Add(-time.Hour))
	if metricsErr != nil {
		t.Fatal(metricsErr)
	}
	if metrics["pv"] != int64(2) || metrics["productPV"] != int64(1) || metrics["conversionRate"] != 0.5 {
		t.Fatalf("incorrect attribution: %v", metrics)
	}
	promotionRequest(t, analytics.RecordEvent, nil, vendor.ID, account.ID, map[string]any{"eventType": "page_view", "path": showroomPath(public)}, 200)
	db.Model(&model.AnalyticsEvent{}).Count(&count)
	if count != 4 {
		t.Fatal("vendor self preview counted")
	}

	stale := save("stale-base-version", 0, changed, "supplier", supplier.ID, 200)
	if err := db.Model(&supplier).Update("content_version", supplier.ContentVersion+1).Error; err != nil {
		t.Fatal(err)
	}
	promotionRequest(t, admin.CommitWorkDraft, gin.Params{{Key: "id", Value: fmt.Sprint(stale.ID)}}, vendor.ID, account.ID, map[string]any{"version": stale.Version}, 409)
	static := StaticPageService{db: db, cfg: config.Config{StaticPagesEnabled: true, PublicDir: t.TempDir(), StaticPageDir: t.TempDir()}}
	static.enabled.Store(true)
	if err := os.WriteFile(filepath.Join(static.cfg.PublicDir, "index.html"), []byte("<html><head></head><body><div id=\"root\"></div></body></html>"), 0600); err != nil {
		t.Fatal(err)
	}
	build, err := static.generate(staticBuildTarget{ResourceType: "supplier", ResourceID: supplier.ID})
	if err != nil {
		t.Fatal(err)
	}
	rendered, err := os.ReadFile(filepath.Join(static.cfg.StaticPageDir, build.FilePath))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(rendered), vendor.Phone) {
		t.Fatal("private phone leaked in static page")
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", showroomPath(public), nil)
	if !static.Serve(c) {
		allowed, canonical := static.canGenerate("supplier", supplier.ID)
		var saved model.StaticPageBuild
		db.First(&saved, build.ID)
		t.Fatalf("generated static route missed: enabled=%v allowed=%v canonical=%s request=%s saved=%+v", static.Enabled(), allowed, canonical, c.Request.URL.Path, saved)
	}
	if err := db.Model(&supplier).Update("status", "disabled").Error; err != nil {
		t.Fatal(err)
	}

	if _, err := loadShowroomSupplier(db, vendor.Slug, supplier.ID); err == nil {
		t.Fatal("disabled offer still public")
	}
	if static.Serve(c) {
		t.Fatal("stale static page served after disabling")
	}
	var cached model.StaticPageBuild
	db.First(&cached, build.ID)
	if cached.Status != "stale" {
		t.Fatal("offer update failed cache invalidation")
	}
	oldPhoto := model.MediaAsset{StorageKey: "retained-only-in-private-draft", OwnerUsername: "qa-vendor", VendorID: &vendor.ID, Status: "staged", Purpose: "submission", CreatedAt: now.Add(-8 * 24 * time.Hour)}
	if err := db.Create(&oldPhoto).Error; err != nil {
		t.Fatal(err)
	}
	draftJSON := json.RawMessage(fmt.Sprintf("{\"logoAssetId\":%d}", oldPhoto.ID))
	if err := db.Create(&model.VendorWorkDraft{VendorID: vendor.ID, UserID: account.ID, ClientKey: "retained-profile", Kind: "profile", Payload: draftJSON}).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.CleanupOrphanedMedia(db, t.TempDir()); err != nil {
		t.Fatal(err)
	}
	db.First(&oldPhoto, oldPhoto.ID)
	if oldPhoto.Status != "staged" {
		t.Fatal("draft photo was cleaned before submission")
	}
	// Simulate the version 19 schema; additive upgrade is retryable and never
	// replays account purges or historical product/SEO backfills.
	legacyAccount := model.AdminUser{Username: "legacy-account", Role: "user", IsEnabled: true}
	if err := db.Create(&legacyAccount).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.SchemaMigration{Version: 19, Name: "previous", AppliedAt: now}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Migrator().DropColumn(&model.ProductSupplier{}, "ShowroomOrder"); err != nil {
		t.Fatal(err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatal(err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatal(err)
	}
	if !db.Migrator().HasColumn(&model.ProductSupplier{}, "ShowroomOrder") {
		t.Fatal("missing new column")
	}
	if err := db.First(&legacyAccount, legacyAccount.ID).Error; err != nil {
		t.Fatal("legacy account changed by incremental upgrade")
	}
	var latest uint
	db.Model(&model.SchemaMigration{}).Select("MAX(version)").Scan(&latest)
	if latest != 20 {
		t.Fatal("schema version not upgraded")
	}
	db.First(&supplier, supplier.ID)
	if supplier.PriceUnit != "桶" || supplier.AvailableQuantity != 0 {
		t.Fatal("upgrade changed historic supply")
	}

}
