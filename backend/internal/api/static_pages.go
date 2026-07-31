package api

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"dalu-nongji-parts/backend/internal/config"
	"dalu-nongji-parts/backend/internal/model"
	"dalu-nongji-parts/backend/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const staticPageSettingKey = "site.staticPages"

type StaticPageService struct {
	db      *gorm.DB
	cfg     config.Config
	enabled atomic.Bool
	jobMu   sync.Mutex
}

type StaticPagePayload struct {
	Kind      string                  `json:"kind"`
	Slug      string                  `json:"slug"`
	Layout    service.LayoutConfig    `json:"layout"`
	Vendor    *model.Vendor           `json:"vendor,omitempty"`
	Products  []model.Product         `json:"products,omitempty"`
	Product   *model.Product          `json:"product,omitempty"`
	Suppliers []model.ProductSupplier `json:"suppliers,omitempty"`
	Related   []model.Product         `json:"related,omitempty"`
}

type staticBuildTarget struct {
	ResourceType string
	ResourceID   uint
}

type StaticPageResourceStatus struct {
	ResourceType string     `json:"resourceType"`
	ResourceID   uint       `json:"resourceId"`
	Status       string     `json:"status"`
	Slug         string     `json:"slug,omitempty"`
	Path         string     `json:"path,omitempty"`
	ErrorMessage string     `json:"errorMessage,omitempty"`
	GeneratedAt  *time.Time `json:"generatedAt,omitempty"`
	CanGenerate  bool       `json:"canGenerate"`
}

type StaticPageSummary struct {
	Available bool                  `json:"available"`
	Enabled   bool                  `json:"enabled"`
	Counts    map[string]int64      `json:"counts"`
	LatestJob *model.StaticBuildJob `json:"latestJob,omitempty"`
	OutputDir string                `json:"outputDir"`
}

func NewStaticPageService(db *gorm.DB, cfg config.Config) *StaticPageService {
	service := &StaticPageService{db: db, cfg: cfg}
	if db == nil {
		return service
	}
	db.Model(&model.StaticBuildJob{}).
		Where("status IN ?", []string{"queued", "running"}).
		Updates(map[string]any{"status": "failed", "completed_at": time.Now(), "errors": `["服务重启，任务已中断，请重新生成"]`})
	var setting model.SiteConfig
	if db.Where("config_key = ?", staticPageSettingKey).First(&setting).Error == nil {
		var enabled bool
		if json.Unmarshal([]byte(setting.ConfigValue), &enabled) == nil {
			service.enabled.Store(enabled)
		}
	}
	return service
}

func (s *StaticPageService) Available() bool {
	if s == nil || !s.cfg.StaticPagesEnabled || strings.TrimSpace(s.cfg.PublicDir) == "" || strings.TrimSpace(s.cfg.StaticPageDir) == "" {
		return false
	}
	_, err := os.Stat(filepath.Join(s.cfg.PublicDir, "index.html"))
	return err == nil
}

func (s *StaticPageService) Enabled() bool {
	return s != nil && s.Available() && s.enabled.Load()
}

func (s *StaticPageService) SetEnabled(enabled bool) error {
	if s == nil || s.db == nil {
		return errors.New("静态页面服务不可用")
	}
	if enabled && !s.Available() {
		return errors.New("服务器未启用静态页面能力，或前端生产目录尚未构建")
	}
	raw, _ := json.Marshal(enabled)
	row := model.SiteConfig{ConfigKey: staticPageSettingKey, ConfigValue: string(raw), Description: "厂商与产品静态页面开关"}
	if err := s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "config_key"}},
		DoUpdates: clause.AssignmentColumns([]string{"config_value", "description", "updated_at"}),
	}).Create(&row).Error; err != nil {
		return err
	}
	s.enabled.Store(enabled)
	return nil
}

func (s *StaticPageService) Summary() StaticPageSummary {
	result := StaticPageSummary{
		Available: s != nil && s.Available(),
		Enabled:   s != nil && s.Enabled(),
		Counts: map[string]int64{
			"ready": 0, "unbuilt": 0, "stale": 0, "generating": 0, "failed": 0,
		},
	}
	if s == nil {
		return result
	}
	result.OutputDir = s.cfg.StaticPageDir
	if s.db == nil {
		return result
	}
	var publicVendors, publicProducts int64
	publishedVendorQuery(s.db).Count(&publicVendors)
	visibleProductQuery(s.db).Count(&publicProducts)
	type countRow struct {
		Status string
		Count  int64
	}
	var counts []countRow
	publicStaticBuildQuery(s.db).Select("status, COUNT(*) AS count").Group("status").Scan(&counts)
	var builtPublic int64
	publicStaticBuildQuery(s.db).Count(&builtPublic)
	for _, row := range counts {
		result.Counts[row.Status] = row.Count
	}
	result.Counts["unbuilt"] = publicVendors + publicProducts - builtPublic
	if result.Counts["unbuilt"] < 0 {
		result.Counts["unbuilt"] = 0
	}
	var latest model.StaticBuildJob
	if s.db.Order("id desc").First(&latest).Error == nil {
		decodeStaticJob(&latest)
		result.LatestJob = &latest
	}
	return result
}

func publicStaticBuildQuery(db *gorm.DB) *gorm.DB {
	return db.Model(&model.StaticPageBuild{}).
		Where("(resource_type = ? AND resource_id IN (?)) OR (resource_type = ? AND resource_id IN (?))",
			"vendor", publishedVendorQuery(db).Select("vendors.id"),
			"product", visibleProductQuery(db).Select("products.id"))
}

func (s *StaticPageService) ResourceStatuses(resourceType string, ids []uint) []StaticPageResourceStatus {
	if resourceType != "vendor" && resourceType != "product" {
		return nil
	}
	builds := make([]model.StaticPageBuild, 0)
	if s != nil && s.db != nil && len(ids) > 0 {
		s.db.Where("resource_type = ? AND resource_id IN ?", resourceType, ids).Find(&builds)
	}
	byID := make(map[uint]model.StaticPageBuild, len(builds))
	for _, build := range builds {
		byID[build.ResourceID] = build
	}
	results := make([]StaticPageResourceStatus, 0, len(ids))
	for _, id := range ids {
		build, ok := byID[id]
		status := StaticPageResourceStatus{ResourceType: resourceType, ResourceID: id, Status: "unbuilt"}
		if ok {
			status.Status = build.Status
			status.Slug = build.Slug
			status.GeneratedAt = build.GeneratedAt
			status.ErrorMessage = build.ErrorMessage
		}
		status.CanGenerate, status.Path = s.canGenerate(resourceType, id)
		results = append(results, status)
	}
	return results
}

func (s *StaticPageService) canGenerate(resourceType string, id uint) (bool, string) {
	if s == nil || s.db == nil || !s.Enabled() {
		return false, ""
	}
	switch resourceType {
	case "vendor":
		var vendor model.Vendor
		if publishedVendorQuery(s.db).First(&vendor, id).Error != nil || vendor.Slug == "" {
			return false, ""
		}
		return true, "/v/" + vendor.Slug
	case "product":
		var product model.Product
		if visibleProductQuery(s.db).First(&product, id).Error != nil || product.Slug == "" {
			return false, ""
		}
		return true, "/products/" + product.Slug
	default:
		return false, ""
	}
}

func (s *StaticPageService) StartJob(scope, resourceType string, resourceIDs []uint, username string) (model.StaticBuildJob, error) {
	if s == nil || s.db == nil || !s.Enabled() {
		return model.StaticBuildJob{}, errors.New("请先在平台配置中启用静态页面")
	}
	if scope != "all" && scope != "single" {
		return model.StaticBuildJob{}, errors.New("生成范围无效")
	}
	if scope == "single" && (resourceType != "vendor" && resourceType != "product" || len(resourceIDs) == 0) {
		return model.StaticBuildJob{}, errors.New("请选择需要生成的厂商或产品")
	}
	resourceIDs = uniqueUintIDs(resourceIDs)
	if scope == "single" {
		for _, id := range resourceIDs {
			if allowed, _ := s.canGenerate(resourceType, id); !allowed {
				return model.StaticBuildJob{}, fmt.Errorf("%s #%d 不是可生成的公开资源", resourceType, id)
			}
		}
	}
	s.jobMu.Lock()
	defer s.jobMu.Unlock()
	var running model.StaticBuildJob
	if s.db.Where("status IN ?", []string{"queued", "running"}).Order("id desc").First(&running).Error == nil {
		decodeStaticJob(&running)
		if running.Scope == "all" || (running.Scope == scope && running.ResourceType == resourceType && uintSlicesEqual(running.ResourceIDs, resourceIDs)) {
			return running, nil
		}
		return model.StaticBuildJob{}, errors.New("当前有静态页面生成任务正在运行，请稍后再试")
	}
	targets := s.targets(scope, resourceType, resourceIDs)
	rawIDs, _ := json.Marshal(resourceIDs)
	job := model.StaticBuildJob{
		Scope: scope, ResourceType: resourceType, ResourceIDsRaw: string(rawIDs),
		Status: "queued", Total: len(targets), RequestedBy: username,
	}
	if err := s.db.Create(&job).Error; err != nil {
		return model.StaticBuildJob{}, err
	}
	job.ResourceIDs = resourceIDs
	go s.runJob(job.ID, targets)
	return job, nil
}

func uintSlicesEqual(left, right []uint) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func (s *StaticPageService) targets(scope, resourceType string, resourceIDs []uint) []staticBuildTarget {
	if scope == "single" {
		unique := uniqueUintIDs(resourceIDs)
		result := make([]staticBuildTarget, 0, len(unique))
		for _, id := range unique {
			result = append(result, staticBuildTarget{ResourceType: resourceType, ResourceID: id})
		}
		return result
	}
	var vendorIDs, productIDs []uint
	publishedVendorQuery(s.db).Order("vendors.id asc").Pluck("vendors.id", &vendorIDs)
	visibleProductQuery(s.db).Order("products.id asc").Pluck("products.id", &productIDs)
	result := make([]staticBuildTarget, 0, len(vendorIDs)+len(productIDs))
	for _, id := range vendorIDs {
		result = append(result, staticBuildTarget{ResourceType: "vendor", ResourceID: id})
	}
	for _, id := range productIDs {
		result = append(result, staticBuildTarget{ResourceType: "product", ResourceID: id})
	}
	return result
}

func (s *StaticPageService) runJob(jobID uint, targets []staticBuildTarget) {
	now := time.Now()
	s.db.Model(&model.StaticBuildJob{}).Where("id = ?", jobID).Updates(map[string]any{"status": "running", "started_at": &now})
	errorsList := make([]string, 0)
	succeeded, failed := 0, 0
	for index, target := range targets {
		if _, err := s.generate(target); err != nil {
			failed++
			errorsList = append(errorsList, fmt.Sprintf("%s #%d：%v", target.ResourceType, target.ResourceID, err))
		} else {
			succeeded++
		}
		rawErrors, _ := json.Marshal(errorsList)
		s.db.Model(&model.StaticBuildJob{}).Where("id = ?", jobID).Updates(map[string]any{
			"processed": index + 1, "succeeded": succeeded, "failed": failed, "errors": string(rawErrors),
		})
	}
	completedAt := time.Now()
	status := "completed"
	if failed > 0 {
		status = "completed_with_errors"
	}
	s.db.Model(&model.StaticBuildJob{}).Where("id = ?", jobID).Updates(map[string]any{"status": status, "completed_at": &completedAt})
}

func (s *StaticPageService) Job(id uint) (model.StaticBuildJob, error) {
	var job model.StaticBuildJob
	if s == nil || s.db == nil {
		return job, gorm.ErrRecordNotFound
	}
	if err := s.db.First(&job, id).Error; err != nil {
		return job, err
	}
	decodeStaticJob(&job)
	return job, nil
}

func decodeStaticJob(job *model.StaticBuildJob) {
	_ = json.Unmarshal([]byte(job.ResourceIDsRaw), &job.ResourceIDs)
	_ = json.Unmarshal([]byte(job.ErrorsRaw), &job.Errors)
}

func (s *StaticPageService) generate(target staticBuildTarget) (model.StaticPageBuild, error) {
	var build model.StaticPageBuild
	s.db.Where("resource_type = ? AND resource_id = ?", target.ResourceType, target.ResourceID).First(&build)
	oldPath := build.FilePath
	build.ResourceType, build.ResourceID, build.Status = target.ResourceType, target.ResourceID, model.StaticPageStatusGenerating
	build.ErrorMessage = ""
	if build.ID == 0 {
		if err := s.db.Create(&build).Error; err != nil {
			return build, err
		}
	} else if err := s.db.Save(&build).Error; err != nil {
		return build, err
	}

	payload, slug, version, path, err := s.buildPayload(target)
	if err != nil {
		s.failBuild(&build, err)
		return build, err
	}
	index, err := os.ReadFile(filepath.Join(s.cfg.PublicDir, "index.html"))
	if err != nil {
		s.failBuild(&build, err)
		return build, fmt.Errorf("读取前端生产文件失败: %w", err)
	}
	doc, ok := s.seoDocument(path)
	if !ok {
		err = errors.New("无法生成页面 SEO 信息")
		s.failBuild(&build, err)
		return build, err
	}
	page := injectSEOHTML(string(index), doc, service.HomeService{DB: s.db}.SiteMeta(context.Background()))
	rawPayload, _ := json.Marshal(payload)
	page = strings.Replace(page, semanticFallback(doc), staticSemanticFallback(payload, doc), 1)
	page = strings.Replace(page, "</body>", `<script id="static-page-data" type="application/json">`+string(rawPayload)+`</script></body>`, 1)
	hashBytes := sha256.Sum256([]byte(page))
	hash := hex.EncodeToString(hashBytes[:])
	relativePath, absolutePath, err := s.writePage(target.ResourceType, target.ResourceID, hash, []byte(page))
	if err != nil {
		s.failBuild(&build, err)
		return build, err
	}
	generatedAt := time.Now()
	build.Slug, build.Status, build.ContentVersion = slug, model.StaticPageStatusReady, version
	build.FilePath, build.ContentHash, build.ErrorMessage, build.GeneratedAt = relativePath, hash, "", &generatedAt
	if err := s.db.Save(&build).Error; err != nil {
		_ = os.Remove(absolutePath)
		return build, err
	}
	if oldPath != "" && oldPath != relativePath {
		if oldAbsolute, safe := s.safeAbsolutePath(oldPath); safe {
			_ = os.Remove(oldAbsolute)
		}
	}
	return build, nil
}

func (s *StaticPageService) failBuild(build *model.StaticPageBuild, err error) {
	build.Status, build.ErrorMessage = model.StaticPageStatusFailed, trimError(err)
	_ = s.db.Save(build).Error
}

func trimError(err error) string {
	if err == nil {
		return ""
	}
	value := err.Error()
	if len([]rune(value)) > 500 {
		return string([]rune(value)[:500])
	}
	return value
}

func (s *StaticPageService) buildPayload(target staticBuildTarget) (StaticPagePayload, string, uint, string, error) {
	payload := StaticPagePayload{Kind: target.ResourceType, Layout: service.HomeService{DB: s.db}.Layout(context.Background())}
	switch target.ResourceType {
	case "vendor":
		var vendor model.Vendor
		if err := publishedVendorQuery(s.db).Preload("Tags").Preload("Media", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order asc, id asc") }).First(&vendor, target.ResourceID).Error; err != nil {
			return payload, "", 0, "", errors.New("厂商未发布或前台不可见")
		}
		for _, tag := range vendor.Tags {
			vendor.TagIDs = append(vendor.TagIDs, tag.ID)
		}
		redactVendor(&vendor)
		var products []model.Product
		visibleProductQuery(s.db).Preload("Category").
			Where("EXISTS (SELECT 1 FROM product_suppliers ps WHERE ps.product_id = products.id AND ps.vendor_id = ? AND ps.status = 'approved' AND ps.deleted_at IS NULL)", vendor.ID).
			Order("products.is_recommended desc, products.sort_order asc, products.id asc").Limit(6).Find(&products)
		enrichProductSummaries(s.db, products, vendor.ID)
		payload.Slug, payload.Vendor, payload.Products = vendor.Slug, &vendor, products
		return payload, vendor.Slug, vendor.ContentVersion, "/v/" + vendor.Slug, nil
	case "product":
		var product model.Product
		if err := visibleProductQuery(s.db).Preload("Category").First(&product, target.ResourceID).Error; err != nil {
			return payload, "", 0, "", errors.New("产品未发布或没有有效厂商关联")
		}
		rows := []model.Product{product}
		enrichProductSummaries(s.db, rows, 0)
		product = rows[0]
		var suppliers []model.ProductSupplier
		s.db.Preload("Vendor").Where("product_id = ? AND status = ?", product.ID, "approved").
			Where("EXISTS (SELECT 1 FROM vendors v WHERE v.id = product_suppliers.vendor_id AND v.deleted_at IS NULL AND v.is_visible = 1 AND v.publication_status = 'published')").
			Order("id asc").Find(&suppliers)
		for i := range suppliers {
			redactVendor(&suppliers[i].Vendor)
		}
		var related []model.Product
		visibleProductQuery(s.db).Preload("Category").Where("products.category_id = ? AND products.id <> ?", product.CategoryID, product.ID).
			Order("products.is_recommended desc, products.sort_order asc, products.id asc").Limit(3).Find(&related)
		enrichProductSummaries(s.db, related, 0)
		payload.Slug, payload.Product, payload.Suppliers, payload.Related = product.Slug, &product, suppliers, related
		return payload, product.Slug, product.ContentVersion, "/products/" + product.Slug, nil
	default:
		return payload, "", 0, "", errors.New("不支持的静态页面类型")
	}
}

func (s *StaticPageService) seoDocument(path string) (seoDocument, bool) {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	base := service.HomeService{DB: s.db}.SiteMeta(context.Background()).SiteURL
	parsed, _ := url.Parse(base)
	host := "localhost"
	scheme := "http"
	if parsed != nil {
		if parsed.Host != "" {
			host = parsed.Host
		}
		if parsed.Scheme != "" {
			scheme = parsed.Scheme
		}
	}
	ctx.Request = httptest.NewRequest(http.MethodGet, scheme+"://"+host+path, nil)
	handler := SEOHandler{DB: s.db, Config: s.cfg, HomeService: service.HomeService{DB: s.db}}
	return handler.document(ctx)
}

func staticSemanticFallback(payload StaticPagePayload, doc seoDocument) string {
	var body bytes.Buffer
	body.WriteString(`<main id="seo-fallback" data-server-rendered="true">`)
	body.WriteString(`<nav aria-label="面包屑">`)
	for _, item := range doc.Breadcrumbs {
		body.WriteString(`<a href="` + template.HTMLEscapeString(item.URL) + `">` + template.HTMLEscapeString(item.Name) + `</a> / `)
	}
	body.WriteString(`</nav><h1>` + template.HTMLEscapeString(doc.BodyTitle) + `</h1><p>` + template.HTMLEscapeString(doc.BodyText) + `</p>`)
	if payload.Vendor != nil {
		vendor := payload.Vendor
		body.WriteString(`<section><h2>主营产品</h2><p>` + template.HTMLEscapeString(vendor.MainProducts) + `</p>`)
		if vendor.CoverImage != "" {
			body.WriteString(`<img alt="` + template.HTMLEscapeString(vendor.Name) + `" src="` + template.HTMLEscapeString(vendor.CoverImage) + `">`)
		}
		if vendor.Phone != "" || vendor.Wechat != "" || vendor.WechatQRCode != "" {
			body.WriteString(`<section><h2>联系方式</h2>`)
			if vendor.Phone != "" {
				body.WriteString(`<p>电话：` + template.HTMLEscapeString(vendor.Phone) + `</p>`)
			}
			if vendor.Wechat != "" {
				body.WriteString(`<p>微信：` + template.HTMLEscapeString(vendor.Wechat) + `</p>`)
			}
			if vendor.WechatQRCode != "" {
				body.WriteString(`<img alt="` + template.HTMLEscapeString(vendor.Name) + ` 微信二维码" src="` + template.HTMLEscapeString(vendor.WechatQRCode) + `">`)
			}
			body.WriteString(`</section>`)
		}
		body.WriteString(`</section>`)
		if len(payload.Products) > 0 {
			body.WriteString(`<section><h2>关联产品</h2><ul>`)
			for _, product := range payload.Products {
				body.WriteString(`<li><a href="/products/` + template.HTMLEscapeString(product.Slug) + `">` + template.HTMLEscapeString(product.Name) + `</a></li>`)
			}
			body.WriteString(`</ul></section>`)
		}
	}
	if payload.Product != nil {
		product := payload.Product
		if product.Image != "" {
			body.WriteString(`<img alt="` + template.HTMLEscapeString(product.Name) + `" src="` + template.HTMLEscapeString(product.Image) + `">`)
		}
		if len(product.Specs()) > 0 {
			body.WriteString(`<section><h2>规格参数</h2><dl>`)
			for _, spec := range product.Specs() {
				body.WriteString(`<dt>` + template.HTMLEscapeString(spec.Name) + `</dt><dd>` + template.HTMLEscapeString(spec.Value) + `</dd>`)
			}
			body.WriteString(`</dl></section>`)
		}
		if len(payload.Suppliers) > 0 {
			body.WriteString(`<section><h2>供应厂商</h2><ul>`)
			for _, supplier := range payload.Suppliers {
				body.WriteString(`<li><a href="/v/` + template.HTMLEscapeString(supplier.Vendor.Slug) + `">` + template.HTMLEscapeString(supplier.Vendor.Name) + `</a></li>`)
			}
			body.WriteString(`</ul></section>`)
		}
	}
	body.WriteString(`<nav aria-label="相关页面"><a href="/">首页</a> · <a href="/products">配件产品</a> · <a href="/vendors">厂商目录</a></nav></main>`)
	return body.String()
}

func (s *StaticPageService) writePage(resourceType string, resourceID uint, hash string, content []byte) (string, string, error) {
	root, err := filepath.Abs(s.cfg.StaticPageDir)
	if err != nil {
		return "", "", err
	}
	dir := filepath.Join(root, resourceType)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", "", fmt.Errorf("创建静态页面目录失败: %w", err)
	}
	name := fmt.Sprintf("%d-%s.html", resourceID, hash[:16])
	target := filepath.Join(dir, name)
	temp, err := os.CreateTemp(dir, ".static-page-*.tmp")
	if err != nil {
		return "", "", err
	}
	tempName := temp.Name()
	defer os.Remove(tempName)
	if _, err = temp.Write(content); err == nil {
		err = temp.Sync()
	}
	closeErr := temp.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return "", "", err
	}
	if err := os.Rename(tempName, target); err != nil {
		return "", "", fmt.Errorf("发布静态页面失败: %w", err)
	}
	relative, err := filepath.Rel(root, target)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || filepath.IsAbs(relative) {
		_ = os.Remove(target)
		return "", "", errors.New("静态页面输出路径无效")
	}
	return filepath.ToSlash(relative), target, nil
}

func (s *StaticPageService) safeAbsolutePath(relative string) (string, bool) {
	root, err := filepath.Abs(s.cfg.StaticPageDir)
	if err != nil {
		return "", false
	}
	target, err := filepath.Abs(filepath.Join(root, filepath.FromSlash(relative)))
	if err != nil {
		return "", false
	}
	rel, err := filepath.Rel(root, target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", false
	}
	return target, true
}

func (s *StaticPageService) Serve(c *gin.Context) bool {
	if s == nil || !s.Enabled() || (c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead) {
		return false
	}
	resourceType, slug, ok := staticRoute(c.Request.URL.Path)
	if !ok {
		return false
	}
	var build model.StaticPageBuild
	if s.db.Where("resource_type = ? AND slug = ? AND status = ?", resourceType, slug, model.StaticPageStatusReady).First(&build).Error != nil {
		return false
	}
	path, safe := s.safeAbsolutePath(build.FilePath)
	if !safe {
		return false
	}
	if _, err := os.Stat(path); err != nil {
		return false
	}
	etag := `"` + build.ContentHash + `"`
	c.Header("ETag", etag)
	c.Header("Cache-Control", "public, max-age=60, s-maxage=86400, stale-while-revalidate=604800")
	c.Header("X-Static-Page", "HIT")
	if c.GetHeader("If-None-Match") == etag {
		c.Status(http.StatusNotModified)
		return true
	}
	if c.Request.Method == http.MethodHead {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.Status(http.StatusOK)
		return true
	}
	c.File(path)
	return true
}

func staticRoute(path string) (string, string, bool) {
	for prefix, resourceType := range map[string]string{"/v/": "vendor", "/products/": "product"} {
		if strings.HasPrefix(path, prefix) {
			slug := strings.Trim(strings.TrimPrefix(path, prefix), "/")
			if slug != "" && !strings.Contains(slug, "/") {
				return resourceType, slug, true
			}
		}
	}
	return "", "", false
}

func parseStaticIDs(value string) []uint {
	parts := strings.Split(value, ",")
	ids := make([]uint, 0, len(parts))
	for _, part := range parts {
		id, err := strconv.ParseUint(strings.TrimSpace(part), 10, 64)
		if err == nil && id > 0 {
			ids = append(ids, uint(id))
		}
	}
	return uniqueUintIDs(ids)
}

func (h AdminHandler) StaticPageSummary(c *gin.Context) {
	if h.StaticPages == nil {
		Fail(c, http.StatusServiceUnavailable, 503, "静态页面服务不可用")
		return
	}
	OK(c, h.StaticPages.Summary())
}

func (h AdminHandler) UpdateStaticPageSettings(c *gin.Context) {
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if c.ShouldBindJSON(&req) != nil {
		Fail(c, http.StatusBadRequest, 400, "配置格式无效")
		return
	}
	if err := h.StaticPages.SetEnabled(req.Enabled); err != nil {
		Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	logOperation(h.DB, c.GetString("username"), "update", "static-pages", 0)
	OK(c, h.StaticPages.Summary())
}

func (h AdminHandler) StaticPageResourceStatuses(c *gin.Context) {
	resourceType := strings.TrimSpace(c.Query("resourceType"))
	ids := parseStaticIDs(c.Query("ids"))
	if (resourceType != "vendor" && resourceType != "product") || len(ids) == 0 {
		Fail(c, http.StatusBadRequest, 400, "资源类型或 ID 无效")
		return
	}
	OK(c, h.StaticPages.ResourceStatuses(resourceType, ids))
}

func (h AdminHandler) CreateStaticBuildJob(c *gin.Context) {
	var req struct {
		Scope        string `json:"scope"`
		ResourceType string `json:"resourceType"`
		ResourceIDs  []uint `json:"resourceIds"`
	}
	if c.ShouldBindJSON(&req) != nil {
		Fail(c, http.StatusBadRequest, 400, "任务参数无效")
		return
	}
	job, err := h.StaticPages.StartJob(req.Scope, req.ResourceType, req.ResourceIDs, c.GetString("username"))
	if err != nil {
		Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	logOperation(h.DB, c.GetString("username"), "generate", "static-pages", job.ID)
	OK(c, job)
}

func (h AdminHandler) StaticBuildJob(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	job, err := h.StaticPages.Job(uint(id))
	if err != nil {
		Fail(c, http.StatusNotFound, 404, "生成任务不存在")
		return
	}
	OK(c, job)
}
