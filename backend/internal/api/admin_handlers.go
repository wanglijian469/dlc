package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"dalu-nongji-parts/backend/internal/auth"
	"dalu-nongji-parts/backend/internal/config"
	"dalu-nongji-parts/backend/internal/database"
	"dalu-nongji-parts/backend/internal/model"
	"dalu-nongji-parts/backend/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AdminHandler struct {
	DB               *gorm.DB
	Config           config.Config
	StaticPages      *StaticPageService
	AccessProtection *AccessProtectionService
	Watermarks       *WatermarkService
}

const (
	loginInvalidRequestMessage     = "请求格式错误"
	loginInvalidCredentialsMessage = "用户名或密码错误"
	loginSessionFailureMessage     = "登录失败"
)

func (h AdminHandler) Login(c *gin.Context) {
	h.login(c, "vendor")
}

func (h AdminHandler) StaffLogin(c *gin.Context) {
	h.login(c, "staff")
}

func (h AdminHandler) login(c *gin.Context, audience string) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, 400, loginInvalidRequestMessage)
		return
	}
	key := c.ClientIP() + "|" + strings.ToLower(strings.TrimSpace(req.Username))
	if !loginAllowed(key, time.Now()) {
		Fail(c, http.StatusTooManyRequests, 429, "登录尝试过多，请稍后再试")
		return
	}
	var user model.AdminUser
	if err := h.DB.Where("username = ? AND is_enabled = ?", req.Username, true).First(&user).Error; err != nil || !auth.CheckPassword(user.PasswordHash, req.Password) {
		recordLoginFailure(key, time.Now())
		Fail(c, http.StatusUnauthorized, 401, loginInvalidCredentialsMessage)
		return
	}
	role := strings.TrimSpace(user.Role)
	if role == "" {
		role = "admin"
	}
	if !loginAudienceAllows(audience, role) {
		recordLoginFailure(key, time.Now())
		message := "该入口仅供厂商账号使用"
		if audience == "staff" {
			message = "该入口仅供 CMS 员工账号使用"
		}
		Fail(c, http.StatusForbidden, 403, message)
		return
	}
	csrf, err := h.startSession(c, user)
	if err != nil {
		Fail(c, http.StatusInternalServerError, 500, loginSessionFailureMessage)
		return
	}
	clearLoginFailures(key)
	OK(c, gin.H{"username": user.Username, "role": role, "vendorId": user.VendorID, "csrfToken": csrf})
}

func loginAudienceAllows(audience, role string) bool {
	if audience == "staff" {
		return role == "admin" || role == "editor" || role == "reviewer"
	}
	return audience == "vendor" && role == "vendor"
}

func (h AdminHandler) Profile(c *gin.Context) {
	OK(c, gin.H{"username": c.GetString("username"), "role": c.GetString("role"), "vendorId": c.GetUint("vendorId")})
}

func (h AdminHandler) DashboardStats(c *gin.Context) {
	var vendors, products, pending, pendingProducts, missingImages int64
	if err := h.DB.Model(&model.Vendor{}).Count(&vendors).Error; err != nil {
		Fail(c, 500, 500, "控制台统计加载失败")
		return
	}
	if err := h.DB.Model(&model.Product{}).Count(&products).Error; err != nil {
		Fail(c, 500, 500, "控制台统计加载失败")
		return
	}
	if err := h.DB.Model(&model.VendorSubmission{}).Where("status = ?", "pending").Count(&pending).Error; err != nil {
		Fail(c, 500, 500, "控制台统计加载失败")
		return
	}
	if err := h.DB.Model(&model.ProductSubmission{}).Where("status = ?", "pending").Count(&pendingProducts).Error; err != nil {
		Fail(c, 500, 500, "控制台统计加载失败")
		return
	}
	if err := h.DB.Model(&model.Vendor{}).Where("logo = '' OR cover_image = ''").Count(&missingImages).Error; err != nil {
		Fail(c, 500, 500, "控制台统计加载失败")
		return
	}
	var productMissing int64
	if err := h.DB.Model(&model.Product{}).Where("image = ''").Count(&productMissing).Error; err != nil {
		Fail(c, 500, 500, "控制台统计加载失败")
		return
	}
	OK(c, gin.H{"vendors": vendors, "products": products, "pendingReviews": pending, "pendingProductReviews": pendingProducts, "missingImages": missingImages + productMissing})
}

func (h AdminHandler) ListOperationLogs(c *gin.Context) {
	var rows []model.OperationLog
	query := h.DB.Model(&model.OperationLog{})
	if search := strings.TrimSpace(c.Query("search")); search != "" {
		like := "%" + search + "%"
		query = query.Where("username LIKE ? OR action LIKE ? OR resource LIKE ?", like, like, like)
	}
	page, pageSize := pageParams(c, 30)
	result, err := paginate(query.Order("created_at desc, id desc"), &rows, page, pageSize)
	if err != nil {
		Fail(c, 500, 500, "操作日志加载失败")
		return
	}
	OK(c, result)
}

func (h AdminHandler) ListMenus(c *gin.Context)  { list[model.Menu](c, h.DB, "sort_order asc, id asc") }
func (h AdminHandler) CreateMenu(c *gin.Context) { saveMenu(c, h.DB, 0) }
func (h AdminHandler) UpdateMenu(c *gin.Context) { saveMenu(c, h.DB, idParam(c)) }
func (h AdminHandler) DeleteMenu(c *gin.Context) { deleteMenu(c, h.DB, idParam(c)) }

func saveMenu(c *gin.Context, db *gorm.DB, id uint) {
	var item model.Menu
	var existing model.Menu
	if id > 0 {
		if err := db.First(&existing, id).Error; err != nil {
			Fail(c, http.StatusNotFound, 404, "菜单不存在")
			return
		}
		if existing.CategoryID > 0 {
			Fail(c, http.StatusConflict, 409, "该菜单由配件分类生成，请到“配件分类”中维护")
			return
		}
		item = existing
	}
	if c.ShouldBindJSON(&item) != nil {
		Fail(c, http.StatusBadRequest, 400, "请求参数格式不正确")
		return
	}
	item.ID = id
	// Category anchor rows are maintained by the category migration only; do
	// not allow generic menu requests to create or forge that relationship.
	item.CategoryID = 0
	item.Name, item.Path, item.MenuType = strings.TrimSpace(item.Name), strings.TrimSpace(item.Path), strings.TrimSpace(item.MenuType)
	if item.Name == "" || item.MenuType == "" {
		Fail(c, http.StatusBadRequest, 400, "菜单名称和类型不能为空")
		return
	}
	if item.ParentID == id && id > 0 {
		Fail(c, http.StatusBadRequest, 400, "菜单不能将自己设为父级")
		return
	}
	if item.ParentID > 0 {
		var parent model.Menu
		if db.First(&parent, item.ParentID).Error != nil {
			Fail(c, http.StatusBadRequest, 400, "父级菜单不存在")
			return
		}
		if parent.ParentID > 0 || parent.MenuType != item.MenuType {
			Fail(c, http.StatusBadRequest, 400, "菜单最多支持两级且父子类型必须一致")
			return
		}
	}
	if item.Path != "" {
		if !strings.HasPrefix(item.Path, "/") || strings.HasPrefix(item.Path, "//") {
			Fail(c, http.StatusBadRequest, 400, "菜单路径必须是站内绝对路径")
			return
		}
		var count int64
		db.Model(&model.Menu{}).Where("menu_type = ? AND path = ? AND id <> ?", item.MenuType, item.Path, id).Count(&count)
		if count > 0 {
			Fail(c, http.StatusConflict, 409, "同类型菜单路径已存在")
			return
		}
	}
	if categoryID := managedCategoryIDForMenu(db, item); categoryID > 0 {
		Fail(c, http.StatusConflict, 409, "分类导航由“配件分类”自动生成，请在分类管理中新增或调整")
		return
	}
	if db.Save(&item).Error != nil {
		Fail(c, http.StatusInternalServerError, 500, "菜单保存失败")
		return
	}
	logOperation(db, c.GetString("username"), upsertAction(id), "menus", item.ID)
	OK(c, item)
}

func deleteMenu(c *gin.Context, db *gorm.DB, id uint) {
	var item model.Menu
	if err := db.First(&item, id).Error; err != nil {
		Fail(c, http.StatusNotFound, 404, "菜单不存在")
		return
	}
	if item.CategoryID > 0 {
		Fail(c, http.StatusConflict, 409, "该菜单由配件分类生成，请到“配件分类”中维护")
		return
	}
	if err := db.Delete(&item).Error; err != nil {
		Fail(c, http.StatusInternalServerError, 500, "菜单删除失败")
		return
	}
	logOperation(db, c.GetString("username"), "delete", "menus", id)
	OK(c, gin.H{"deleted": true})
}

func managedCategoryIDForMenu(db *gorm.DB, menu model.Menu) uint {
	if menu.ParentID != 0 || (menu.MenuType != "sidebar" && menu.MenuType != "mobile") {
		return 0
	}
	var categories []model.Category
	if db.Where("parent_id = ?", 0).Find(&categories).Error != nil {
		return 0
	}
	if parsed, err := url.Parse(menu.Path); err == nil {
		if raw := parsed.Query().Get("categoryId"); raw != "" {
			if value, err := strconv.ParseUint(raw, 10, 64); err == nil {
				for _, category := range categories {
					if category.ID == uint(value) {
						return category.ID
					}
				}
			}
		}
	}
	for _, category := range categories {
		if strings.EqualFold(strings.TrimSpace(menu.Name), strings.TrimSpace(category.Name)) {
			return category.ID
		}
	}
	return 0
}

func (h AdminHandler) ListVendors(c *gin.Context) {
	var rows []model.Vendor
	query := h.DB.Model(&model.Vendor{}).Preload("Tags").Preload("Media", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order asc, id asc") })
	if search := strings.TrimSpace(c.Query("search")); search != "" {
		like := "%" + search + "%"
		query = query.Where("name LIKE ? OR short_name LIKE ? OR main_products LIKE ?", like, like, like)
	}
	if province := strings.TrimSpace(c.Query("province")); province != "" {
		query = query.Where("province = ?", province)
	}
	if visible := strings.TrimSpace(c.Query("visible")); visible != "" {
		query = query.Where("is_visible = ?", visible == "true" || visible == "1")
	}
	if publication := strings.TrimSpace(c.Query("publicationStatus")); publication != "" {
		query = query.Where("publication_status = ?", publication)
	}
	if adminPaginationRequested(c) {
		page, pageSize := pageParams(c, 20)
		result, err := paginate(query.Order("sort_order asc, id asc"), &rows, page, pageSize)
		if err != nil {
			Fail(c, 500, 500, "厂商列表加载失败")
			return
		}
		for i := range rows {
			rows[i].TagIDs = tagIDsFromTags(rows[i].Tags)
		}
		OK(c, result)
		return
	}
	if err := query.Order("sort_order asc, id asc").Find(&rows).Error; err != nil {
		Fail(c, http.StatusInternalServerError, 500, "璇诲彇澶辫触")
		return
	}
	for i := range rows {
		rows[i].TagIDs = tagIDsFromTags(rows[i].Tags)
	}
	OK(c, rows)
}

func (h AdminHandler) ListVendorOptions(c *gin.Context) {
	var rows []model.VendorOption
	query := h.DB.Model(&model.Vendor{})
	if search := strings.TrimSpace(c.Query("search")); search != "" {
		like := "%" + search + "%"
		query = query.Where(
			"name LIKE ? OR short_name LIKE ? OR province LIKE ? OR city LIKE ? OR main_products LIKE ?",
			like, like, like, like, like,
		)
	}
	page, pageSize := pageParams(c, 20)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		Fail(c, http.StatusInternalServerError, 500, "厂商搜索失败")
		return
	}
	if err := query.Select("id, name, short_name, province, city, main_products, publication_status, is_visible").
		Order("name asc, id asc").Offset((page - 1) * pageSize).Limit(pageSize).Scan(&rows).Error; err != nil {
		Fail(c, http.StatusInternalServerError, 500, "厂商搜索失败")
		return
	}
	OK(c, PageResult{Items: rows, Page: page, PageSize: pageSize, Total: total})
}

func (h AdminHandler) CreateVendor(c *gin.Context) { saveVendor(c, h.DB, 0) }
func (h AdminHandler) UpdateVendor(c *gin.Context) { saveVendor(c, h.DB, idParam(c)) }
func (h AdminHandler) DeleteVendor(c *gin.Context) { remove[model.Vendor](c, h.DB, "vendors") }

func (h AdminHandler) ListTags(c *gin.Context)  { list[model.Tag](c, h.DB, "sort_order asc, id asc") }
func (h AdminHandler) CreateTag(c *gin.Context) { create[model.Tag](c, h.DB, "tags") }
func (h AdminHandler) UpdateTag(c *gin.Context) { update[model.Tag](c, h.DB, "tags") }
func (h AdminHandler) DeleteTag(c *gin.Context) { remove[model.Tag](c, h.DB, "tags") }

func (h AdminHandler) ListCategories(c *gin.Context) {
	var categories []model.Category
	if err := h.DB.Order("parent_id asc, sort_order asc, id asc").Find(&categories).Error; err != nil {
		Fail(c, http.StatusInternalServerError, 500, "分类列表加载失败")
		return
	}
	OK(c, categories)
}
func (h AdminHandler) CreateCategory(c *gin.Context) { saveCategory(c, h.DB, 0) }
func (h AdminHandler) UpdateCategory(c *gin.Context) { saveCategory(c, h.DB, idParam(c)) }
func (h AdminHandler) DeleteCategory(c *gin.Context) { deleteCategory(c, h.DB, idParam(c)) }

func (h AdminHandler) ListProducts(c *gin.Context) {
	query := h.DB.Model(&model.Product{}).Preload("Category")
	if search := strings.TrimSpace(c.Query("search")); search != "" {
		like := "%" + search + "%"
		query = query.Where("products.name LIKE ? OR products.description LIKE ? OR products.compatible_models LIKE ?", like, like, like)
	}
	publicationStatus := strings.TrimSpace(c.Query("publicationStatus"))
	if publicationStatus == "" {
		publicationStatus = strings.TrimSpace(c.Query("status"))
	}
	if publicationStatus != "" {
		query = query.Where("products.publication_status = ?", publicationStatus)
	}
	if vendorID := strings.TrimSpace(c.Query("vendorId")); vendorID != "" {
		query = query.Where("EXISTS (SELECT 1 FROM product_suppliers ps WHERE ps.product_id = products.id AND ps.vendor_id = ? AND ps.status <> 'disabled' AND ps.deleted_at IS NULL)", vendorID)
	}
	if categoryID := strings.TrimSpace(c.Query("categoryId")); categoryID != "" {
		query = query.Where("products.category_id = ?", categoryID)
	}
	switch strings.TrimSpace(c.Query("associationStatus")) {
	case "linked":
		query = query.Where("EXISTS (SELECT 1 FROM product_suppliers ps WHERE ps.product_id = products.id AND ps.status <> 'disabled' AND ps.deleted_at IS NULL)")
	case "unlinked":
		query = query.Where("NOT EXISTS (SELECT 1 FROM product_suppliers ps WHERE ps.product_id = products.id AND ps.status <> 'disabled' AND ps.deleted_at IS NULL)")
	}
	if adminPaginationRequested(c) {
		var rows []model.Product
		page, pageSize := pageParams(c, 20)
		result, err := paginate(query.Order("sort_order asc, id asc"), &rows, page, pageSize)
		if err != nil {
			Fail(c, 500, 500, "产品列表加载失败")
			return
		}
		enrichProductSummaries(h.DB, rows, 0)
		enrichAdminProductAssociations(h.DB, rows)
		OK(c, result)
		return
	}
	var rows []model.Product
	if err := query.Order("sort_order asc, id asc").Find(&rows).Error; err != nil {
		Fail(c, http.StatusInternalServerError, 500, "读取失败")
		return
	}
	enrichProductSummaries(h.DB, rows, 0)
	enrichAdminProductAssociations(h.DB, rows)
	OK(c, rows)
}

func enrichAdminProductAssociations(db *gorm.DB, products []model.Product) {
	if len(products) == 0 {
		return
	}
	productIDs := make([]uint, 0, len(products))
	indexByID := make(map[uint]int, len(products))
	for i := range products {
		productIDs = append(productIDs, products[i].ID)
		indexByID[products[i].ID] = i
	}
	type associationRow struct {
		ProductID         uint
		ID                uint
		Name              string
		ShortName         string
		Province          string
		City              string
		MainProducts      string
		PublicationStatus string
		IsVisible         bool
	}
	var rows []associationRow
	if err := db.Table("product_suppliers ps").
		Select("ps.product_id, v.id, v.name, v.short_name, v.province, v.city, v.main_products, v.publication_status, v.is_visible").
		Joins("JOIN vendors v ON v.id = ps.vendor_id AND v.deleted_at IS NULL").
		Where("ps.product_id IN ? AND ps.status <> ? AND ps.deleted_at IS NULL", productIDs, "disabled").
		Order("v.name asc, v.id asc").
		Scan(&rows).Error; err != nil {
		return
	}
	for _, row := range rows {
		index, ok := indexByID[row.ProductID]
		if !ok {
			continue
		}
		products[index].AssociationCount++
		if len(products[index].AssociatedVendors) < 3 {
			products[index].AssociatedVendors = append(products[index].AssociatedVendors, model.VendorOption{
				ID: row.ID, Name: row.Name, ShortName: row.ShortName, Province: row.Province, City: row.City,
				MainProducts: row.MainProducts, PublicationStatus: row.PublicationStatus, IsVisible: row.IsVisible,
			})
		}
	}
}
func (h AdminHandler) CreateProduct(c *gin.Context) { saveProduct(c, h.DB, 0) }
func (h AdminHandler) UpdateProduct(c *gin.Context) { saveProduct(c, h.DB, idParam(c)) }
func (h AdminHandler) DeleteProduct(c *gin.Context) { remove[model.Product](c, h.DB, "products") }

func (h AdminHandler) ListBanners(c *gin.Context) {
	list[model.Banner](c, h.DB, "sort_order asc, id asc")
}
func (h AdminHandler) CreateBanner(c *gin.Context) { create[model.Banner](c, h.DB, "banners") }
func (h AdminHandler) UpdateBanner(c *gin.Context) { update[model.Banner](c, h.DB, "banners") }
func (h AdminHandler) DeleteBanner(c *gin.Context) { remove[model.Banner](c, h.DB, "banners") }

func (h AdminHandler) ListPages(c *gin.Context) {
	list[model.ContentPage](c, h.DB, "sort_order asc, id asc")
}
func (h AdminHandler) CreatePage(c *gin.Context) { savePage(c, h.DB, 0) }
func (h AdminHandler) UpdatePage(c *gin.Context) { savePage(c, h.DB, idParam(c)) }
func (h AdminHandler) DeletePage(c *gin.Context) { remove[model.ContentPage](c, h.DB, "pages") }

func (h AdminHandler) ListFriendLinks(c *gin.Context) {
	list[model.FriendLink](c, h.DB, "sort_order asc, id asc")
}
func (h AdminHandler) CreateFriendLink(c *gin.Context) { saveFriendLink(c, h.DB, 0) }
func (h AdminHandler) UpdateFriendLink(c *gin.Context) { saveFriendLink(c, h.DB, idParam(c)) }
func (h AdminHandler) DeleteFriendLink(c *gin.Context) {
	remove[model.FriendLink](c, h.DB, "friend-links")
}

func (h AdminHandler) ListConfigs(c *gin.Context) { list[model.SiteConfig](c, h.DB, "config_key asc") }

func (h AdminHandler) UpdateConfig(c *gin.Context) {
	var req model.SiteConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, 400, "璇锋眰鏍煎紡閿欒")
		return
	}
	key := c.Param("key")
	if err := validateConfigValue(key, req.ConfigValue); err != nil {
		Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	var item model.SiteConfig
	if err := h.DB.Where("config_key = ?", key).First(&item).Error; err != nil {
		Fail(c, http.StatusNotFound, 404, "配置不存在")
		return
	}
	item.ConfigValue = req.ConfigValue
	item.Description = req.Description
	if err := h.DB.Save(&item).Error; err != nil {
		Fail(c, http.StatusInternalServerError, 500, "淇濆瓨澶辫触")
		return
	}
	logOperation(h.DB, c.GetString("username"), "update", "configs", item.ID)
	OK(c, item)
}

func (h AdminHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		Fail(c, http.StatusBadRequest, 400, "璇烽€夋嫨涓婁紶鏂囦欢")
		return
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true, ".svg": true}
	if !allowed[ext] {
		Fail(c, http.StatusBadRequest, 400, "只支持图片文件上传")
		return
	}
	if file.Size > 5*1024*1024 {
		Fail(c, http.StatusBadRequest, 400, "鍥剧墖涓嶈兘瓒呰繃 5MB")
		return
	}
	publicDir := h.Config.PublicDir
	uploadDir := "uploads"
	if publicDir != "" {
		uploadDir = filepath.Join(publicDir, "uploads")
	}
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		Fail(c, http.StatusInternalServerError, 500, "鍒涘缓涓婁紶鐩綍澶辫触")
		return
	}
	name := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	path := filepath.Join(uploadDir, name)
	if err := c.SaveUploadedFile(file, path); err != nil {
		Fail(c, http.StatusInternalServerError, 500, "涓婁紶澶辫触")
		return
	}
	width, height := 0, 0
	if uploaded, openErr := os.Open(path); openErr == nil {
		if cfg, _, decodeErr := image.DecodeConfig(uploaded); decodeErr == nil {
			width, height = cfg.Width, cfg.Height
		}
		_ = uploaded.Close()
	}
	logOperation(h.DB, c.GetString("username"), "upload", "uploads", 0)
	OK(c, gin.H{"url": "/uploads/" + name, "width": width, "height": height, "size": file.Size, "mime": mime.TypeByExtension(ext)})
}

func list[T any](c *gin.Context, db *gorm.DB, order string) {
	var rows []T
	if err := db.Order(order).Find(&rows).Error; err != nil {
		Fail(c, http.StatusInternalServerError, 500, "璇诲彇澶辫触")
		return
	}
	OK(c, rows)
}

func listWithPreloads[T any](c *gin.Context, db *gorm.DB, preloads []string, order string) {
	query := db
	for _, preload := range preloads {
		query = query.Preload(preload)
	}
	var rows []T
	if err := query.Order(order).Find(&rows).Error; err != nil {
		Fail(c, http.StatusInternalServerError, 500, "璇诲彇澶辫触")
		return
	}
	OK(c, rows)
}

func adminPaginationRequested(c *gin.Context) bool {
	return c.Query("page") != "" || c.Query("pageSize") != ""
}

func create[T any](c *gin.Context, db *gorm.DB, resource string) {
	var item T
	if err := c.ShouldBindJSON(&item); err != nil {
		Fail(c, http.StatusBadRequest, 400, "璇锋眰鏍煎紡閿欒")
		return
	}
	if err := db.Create(&item).Error; err != nil {
		Fail(c, http.StatusInternalServerError, 500, "鍒涘缓澶辫触")
		return
	}
	logOperation(db, c.GetString("username"), "create", resource, 0)
	OK(c, item)
}

func update[T any](c *gin.Context, db *gorm.DB, resource string) {
	var item T
	if err := db.First(&item, c.Param("id")).Error; err != nil {
		Fail(c, http.StatusNotFound, 404, "记录不存在")
		return
	}
	if err := c.ShouldBindJSON(&item); err != nil {
		Fail(c, http.StatusBadRequest, 400, "璇锋眰鏍煎紡閿欒")
		return
	}
	if err := db.Save(&item).Error; err != nil {
		Fail(c, http.StatusInternalServerError, 500, "淇濆瓨澶辫触")
		return
	}
	logOperation(db, c.GetString("username"), "update", resource, idParam(c))
	OK(c, item)
}

func remove[T any](c *gin.Context, db *gorm.DB, resource string) {
	var item T
	if err := db.First(&item, c.Param("id")).Error; err != nil {
		Fail(c, http.StatusNotFound, 404, "记录不存在")
		return
	}
	if err := db.Delete(&item).Error; err != nil {
		Fail(c, http.StatusInternalServerError, 500, "鍒犻櫎澶辫触")
		return
	}
	logOperation(db, c.GetString("username"), "delete", resource, idParam(c))
	OK(c, gin.H{"deleted": true})
}

func saveVendor(c *gin.Context, db *gorm.DB, id uint) {
	var input model.Vendor
	var previousVersion uint
	if id > 0 {
		if err := db.Preload("Tags").Preload("Media").First(&input, id).Error; err != nil {
			Fail(c, http.StatusNotFound, 404, "厂商不存在")
			return
		}
		previousVersion = input.ContentVersion
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		Fail(c, http.StatusBadRequest, 400, "璇锋眰鏍煎紡閿欒")
		return
	}
	if id > 0 {
		input.ID = id
	}
	if strings.TrimSpace(input.Name) == "" {
		Fail(c, http.StatusBadRequest, 400, "鍘傚晢鍚嶇О涓嶈兘涓虹┖")
		return
	}
	if input.WebsiteURL != "" && !validURL(input.WebsiteURL) {
		Fail(c, http.StatusBadRequest, 400, "官网地址格式不正确")
		return
	}
	if err := normalizeVendorReviewStatus(&input); err != nil {
		Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	service.ApplyVendorSEO(&input)
	if input.PublicationStatus == "" {
		if input.IsVisible {
			input.PublicationStatus = "published"
		} else {
			input.PublicationStatus = "hidden"
		}
	}
	if input.PublicationStatus != "draft" && input.PublicationStatus != "published" && input.PublicationStatus != "hidden" {
		Fail(c, http.StatusBadRequest, 400, "发布状态只能是 draft、published 或 hidden")
		return
	}
	input.IsVisible = input.PublicationStatus == "published"
	if input.DataOrigin == "" {
		input.DataOrigin = "admin"
	}
	if err := bindVendorImageAssets(db, &input); err != nil {
		Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	if id > 0 {
		input.ContentVersion = previousVersion + 1
	} else {
		input.ContentVersion = 1
	}
	if err := database.EnsureVendorSlug(db, &input); err != nil {
		Fail(c, http.StatusConflict, 409, "厂商页面标识冲突")
		return
	}
	if input.PublicationStatus == "published" && input.PublishedAt == nil {
		now := time.Now()
		input.PublishedAt = &now
	}
	tagIDs := uniqueUintIDs(input.TagIDs)
	media := input.Media
	input.TagIDs = tagIDs
	input.Tags = nil
	input.Media = nil
	tx := db.Begin()
	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()
	if err := tx.Omit("Tags", "Media").Save(&input).Error; err != nil {
		Fail(c, http.StatusInternalServerError, 500, "淇濆瓨澶辫触")
		return
	}
	if media != nil {
		if err := replaceVendorMedia(tx, input.ID, media); err != nil {
			Fail(c, http.StatusBadRequest, 400, err.Error())
			return
		}
	}
	if tagIDs != nil {
		var tags []model.Tag
		if len(tagIDs) > 0 {
			if err := tx.Where("id IN ?", tagIDs).Find(&tags).Error; err != nil {
				Fail(c, http.StatusInternalServerError, 500, "鏍囩璇诲彇澶辫触")
				return
			}
			if len(tags) != len(tagIDs) {
				Fail(c, http.StatusBadRequest, 400, "标签不存在")
				return
			}
		}
		if err := tx.Model(&input).Association("Tags").Replace(tags); err != nil {
			Fail(c, http.StatusInternalServerError, 500, "鏍囩淇濆瓨澶辫触")
			return
		}
	}
	if err := tx.Commit().Error; err != nil {
		Fail(c, 500, 500, "厂商资料保存失败")
		return
	}
	committed = true
	db.Preload("Tags").Preload("Media", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order asc, id asc") }).First(&input, input.ID)
	input.TagIDs = tagIDsFromTags(input.Tags)
	logOperation(db, c.GetString("username"), upsertAction(id), "vendors", input.ID)
	OK(c, input)
}

func bindVendorImageAssets(db *gorm.DB, vendor *model.Vendor) error {
	for _, item := range []struct {
		url    string
		target **uint
	}{
		{vendor.Logo, &vendor.LogoAssetID},
		{vendor.CoverImage, &vendor.CoverAssetID},
		{vendor.WechatQRCode, &vendor.WechatQRCodeAssetID},
	} {
		assetID := mediaAssetIDFromURL(item.url)
		if assetID == 0 {
			*item.target = nil
			continue
		}
		var asset model.MediaAsset
		if err := db.First(&asset, assetID).Error; err != nil {
			return fmt.Errorf("图片资源不存在")
		}
		*item.target = &asset.ID
	}
	return nil
}

func mediaAssetIDFromURL(value string) uint {
	if !strings.HasPrefix(value, "/api/media/") {
		return 0
	}
	id, err := strconv.ParseUint(strings.TrimPrefix(value, "/api/media/"), 10, 64)
	if err != nil {
		return 0
	}
	return uint(id)
}

type adminProductInput struct {
	model.Product
	VendorIDs []uint `json:"vendorIds"`
}

func saveProduct(c *gin.Context, db *gorm.DB, id uint) {
	var request adminProductInput
	if id > 0 {
		if err := db.First(&request.Product, id).Error; err != nil {
			Fail(c, http.StatusNotFound, 404, "产品不存在")
			return
		}
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		Fail(c, http.StatusBadRequest, 400, "璇锋眰鏍煎紡閿欒")
		return
	}
	input := request.Product
	if id > 0 {
		input.ID = id
	}
	if strings.TrimSpace(input.Name) == "" {
		Fail(c, http.StatusBadRequest, 400, "浜у搧鍚嶇О涓嶈兘涓虹┖")
		return
	}
	if input.Status == 0 {
		input.Status = 1
	}
	if input.Status != 1 && input.Status != 2 {
		Fail(c, http.StatusBadRequest, 400, "浜у搧鐘舵€佸彧鑳戒负 1 鎴?2")
		return
	}
	if input.PublicationStatus == "" {
		if input.Status == 1 {
			input.PublicationStatus = "published"
		} else {
			input.PublicationStatus = "hidden"
		}
	}
	if input.PublicationStatus != "published" && input.PublicationStatus != "draft" && input.PublicationStatus != "hidden" {
		Fail(c, http.StatusBadRequest, 400, "产品发布状态无效")
		return
	}
	if input.PublicationStatus == "published" {
		input.Status = 1
	} else {
		input.Status = 2
	}
	if input.ContentVersion == 0 {
		input.ContentVersion = 1
	}
	if id > 0 {
		input.ContentVersion++
	}
	if err := database.EnsureProductSlug(db, &input); err != nil {
		Fail(c, http.StatusConflict, 409, "产品页面标识冲突")
		return
	}
	if input.PublicationStatus == "published" && input.PublishedAt == nil {
		now := time.Now()
		input.PublishedAt = &now
	}
	if err := validateJSONStringArray(input.GalleryRaw, "浜у搧鍥惧簱"); err != nil {
		Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	if err := validateJSON(input.SpecsRaw, "瑙勬牸鍙傛暟"); err != nil {
		Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	if input.CategoryID > 0 && !recordExists[model.Category](db, input.CategoryID) {
		Fail(c, http.StatusBadRequest, 400, "分类不存在")
		return
	}
	vendorIDs := uniqueUintIDs(request.VendorIDs)
	if input.VendorID != nil {
		if *input.VendorID == 0 {
			input.VendorID = nil
		} else {
			vendorIDs = uniqueUintIDs(append(vendorIDs, *input.VendorID))
		}
	}
	if len(vendorIDs) > 0 {
		var vendorCount int64
		if err := db.Model(&model.Vendor{}).Where("id IN ?", vendorIDs).Count(&vendorCount).Error; err != nil {
			Fail(c, http.StatusInternalServerError, 500, "厂商校验失败")
			return
		}
		if vendorCount != int64(len(vendorIDs)) {
			Fail(c, http.StatusBadRequest, 400, "所选厂商不存在或已删除")
			return
		}
	}
	if input.PublicationStatus == "published" && !hasPublishableSupplier(db, id, vendorIDs) {
		Fail(c, http.StatusBadRequest, 400, "产品发布前必须关联至少一家已发布且前台可见的厂商")
		return
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&input).Error; err != nil {
			return err
		}
		for _, vendorID := range vendorIDs {
			if _, err := upsertAdminProductSupplier(tx, input, vendorID, c.GetString("username")); err != nil {
				return err
			}
		}
		if input.PublicationStatus == "published" && !hasPublishableSupplier(tx, input.ID, nil) {
			return errProductRequiresVendor
		}
		if input.Status == 1 {
			publishProductMedia(tx, input)
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, errProductRequiresVendor) {
			Fail(c, http.StatusBadRequest, 400, "产品发布前必须关联至少一家已发布且前台可见的厂商")
			return
		}
		Fail(c, http.StatusInternalServerError, 500, "保存失败")
		return
	}
	db.Preload("Category").First(&input, input.ID)
	enriched := []model.Product{input}
	enrichAdminProductAssociations(db, enriched)
	input = enriched[0]
	logOperation(db, c.GetString("username"), upsertAction(id), "products", input.ID)
	OK(c, input)
}

func hasPublishableSupplier(db *gorm.DB, productID uint, additionalVendorIDs []uint) bool {
	var count int64
	if productID > 0 {
		db.Model(&model.ProductSupplier{}).
			Joins("JOIN vendors ON vendors.id = product_suppliers.vendor_id AND vendors.deleted_at IS NULL").
			Where("product_suppliers.product_id = ? AND product_suppliers.status = ? AND vendors.is_visible = ? AND vendors.publication_status = ? AND (vendors.published_at IS NULL OR vendors.published_at <= ?)",
				productID, "approved", true, "published", time.Now()).
			Count(&count)
		if count > 0 {
			return true
		}
	}
	if len(additionalVendorIDs) == 0 {
		return false
	}
	db.Model(&model.Vendor{}).
		Where("id IN ? AND is_visible = ? AND publication_status = ? AND (published_at IS NULL OR published_at <= ?)",
			additionalVendorIDs, true, "published", time.Now()).
		Count(&count)
	return count > 0
}

func upsertAdminProductSupplier(db *gorm.DB, product model.Product, vendorID uint, username string) (model.ProductSupplier, error) {
	var supplier model.ProductSupplier
	err := db.Unscoped().Where("product_id = ? AND vendor_id = ?", product.ID, vendorID).First(&supplier).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return supplier, err
	}
	if err == gorm.ErrRecordNotFound {
		supplier = model.ProductSupplier{ProductID: product.ID, VendorID: vendorID, ContentVersion: 1}
		applySupplierDraft(&supplier, supplierFromProduct(product, product.ID, vendorID))
	} else {
		supplier.ContentVersion++
	}
	supplier.DeletedAt = gorm.DeletedAt{}
	supplier.Status = "approved"
	supplier.SourceType = "admin"
	supplier.ReviewedBy = username
	now := time.Now()
	supplier.ReviewedAt = &now
	if err := db.Unscoped().Save(&supplier).Error; err != nil {
		return supplier, err
	}
	return supplier, nil
}

func publishProductMedia(db *gorm.DB, product model.Product) {
	urls := productMediaURLs(product)
	ids := make([]uint, 0, len(urls))
	for _, value := range urls {
		if !strings.HasPrefix(value, "/api/media/") {
			continue
		}
		id, err := strconv.ParseUint(strings.TrimPrefix(value, "/api/media/"), 10, 64)
		if err == nil && id > 0 {
			ids = append(ids, uint(id))
		}
	}
	if len(ids) > 0 {
		now := time.Now()
		db.Model(&model.MediaAsset{}).Where("id IN ?", uniqueUintIDs(ids)).Updates(map[string]interface{}{"status": "published", "published_at": &now})
	}
}

func productMediaURLs(product model.Product) []string {
	urls := []string{product.Image}
	var gallery []string
	_ = json.Unmarshal([]byte(product.GalleryRaw), &gallery)
	urls = append(urls, gallery...)
	for _, spec := range product.Specs() {
		if strings.TrimSpace(spec.Image) != "" {
			urls = append(urls, spec.Image)
		}
	}
	return urls
}

func savePage(c *gin.Context, db *gorm.DB, id uint) {
	var input model.ContentPage
	if id > 0 {
		if err := db.First(&input, id).Error; err != nil {
			Fail(c, http.StatusNotFound, 404, "页面不存在")
			return
		}
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		Fail(c, http.StatusBadRequest, 400, "璇锋眰鏍煎紡閿欒")
		return
	}
	if id > 0 {
		input.ID = id
	}
	if strings.TrimSpace(input.PageType) == "" {
		input.PageType = "page"
	}
	if input.ContentVersion == 0 {
		input.ContentVersion = 1
	} else if id > 0 {
		input.ContentVersion++
	}
	if input.PublishedAt != nil && input.PublishedAt.After(time.Now()) {
		input.PublicationStatus = "scheduled"
		input.IsEnabled = true
	} else if input.IsEnabled {
		input.PublicationStatus = "published"
	} else {
		input.PublicationStatus = "archived"
	}
	if err := database.EnsurePageSlug(db, &input); err != nil {
		Fail(c, http.StatusConflict, 409, "页面标识已存在")
		return
	}
	if input.PageType != "page" && input.PageType != "article" {
		Fail(c, http.StatusBadRequest, 400, "内容类型仅支持平台页面或行业文章")
		return
	}
	if input.PageType == "article" && input.PublishedAt == nil {
		now := time.Now()
		input.PublishedAt = &now
	}
	if strings.TrimSpace(input.Slug) == "" || strings.TrimSpace(input.Title) == "" {
		Fail(c, http.StatusBadRequest, 400, "页面标识和标题不能为空")
		return
	}
	if err := validateContentBlocks(input.BlocksRaw); err != nil {
		Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	if err := db.Save(&input).Error; err != nil {
		Fail(c, http.StatusInternalServerError, 500, "淇濆瓨澶辫触")
		return
	}
	logOperation(db, c.GetString("username"), upsertAction(id), "pages", input.ID)
	OK(c, input)
}

func validateContentBlocks(value string) error {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	var blocks []model.ContentBlock
	if err := json.Unmarshal([]byte(value), &blocks); err != nil {
		return fmt.Errorf("结构化内容格式不正确")
	}
	allowed := map[string]bool{"hero": true, "text": true, "steps": true, "cta": true, "contact": true, "faq": true}
	for _, block := range blocks {
		if !allowed[block.Type] {
			return fmt.Errorf("不支持的内容区块类型：%s", block.Type)
		}
	}
	return nil
}

func saveFriendLink(c *gin.Context, db *gorm.DB, id uint) {
	var input model.FriendLink
	if id > 0 {
		if err := db.First(&input, id).Error; err != nil {
			Fail(c, http.StatusNotFound, 404, "友情链接不存在")
			return
		}
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		Fail(c, http.StatusBadRequest, 400, "璇锋眰鏍煎紡閿欒")
		return
	}
	if id > 0 {
		input.ID = id
	}
	if strings.TrimSpace(input.Name) == "" || !validURL(input.URL) {
		Fail(c, http.StatusBadRequest, 400, "閾炬帴鍚嶇О鍜?URL 蹇呴』姝ｇ‘濉啓")
		return
	}
	if err := db.Save(&input).Error; err != nil {
		Fail(c, http.StatusInternalServerError, 500, "淇濆瓨澶辫触")
		return
	}
	logOperation(db, c.GetString("username"), upsertAction(id), "friend-links", input.ID)
	OK(c, input)
}

func idParam(c *gin.Context) uint {
	value, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	return uint(value)
}

func upsertAction(id uint) string {
	if id > 0 {
		return "update"
	}
	return "create"
}

func normalizeVendorReviewStatus(vendor *model.Vendor) error {
	status := strings.TrimSpace(vendor.ReviewStatus)
	if status == "" {
		vendor.ReviewStatus = "pending"
		return nil
	}
	switch status {
	case "pending", "verified", "rejected":
		vendor.ReviewStatus = status
		return nil
	default:
		return fmt.Errorf("复核状态只能是 pending、verified 或 rejected")
	}
}

func tagIDsFromTags(tags []model.Tag) []uint {
	ids := make([]uint, 0, len(tags))
	for _, tag := range tags {
		ids = append(ids, tag.ID)
	}
	return uniqueUintIDs(ids)
}

func uniqueUintIDs(ids []uint) []uint {
	if ids == nil {
		return nil
	}
	seen := make(map[uint]bool, len(ids))
	unique := make([]uint, 0, len(ids))
	for _, id := range ids {
		if id == 0 || seen[id] {
			continue
		}
		seen[id] = true
		unique = append(unique, id)
	}
	return unique
}

func validURL(value string) bool {
	if strings.HasPrefix(value, "/") {
		return !strings.HasPrefix(value, "//")
	}
	parsed, err := url.ParseRequestURI(value)
	return err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != ""
}

func recordExists[T any](db *gorm.DB, id uint) bool {
	var item T
	return db.First(&item, id).Error == nil
}

func validateConfigValue(key string, value string) error {
	switch key {
	case "site.meta", "site.theme", "home.modules":
		return validateJSON(value, "閰嶇疆")
	case "home.stats":
		var rows []struct {
			Label string `json:"label"`
			Value string `json:"value"`
		}
		if err := json.Unmarshal([]byte(value), &rows); err != nil || len(rows) == 0 {
			return fmt.Errorf("缁熻閰嶇疆蹇呴』鏄寘鍚?label/value 鐨?JSON 鏁扮粍")
		}
	case "home.safeguards":
		var rows []string
		if err := json.Unmarshal([]byte(value), &rows); err != nil {
			return fmt.Errorf("保障文案必须是 JSON 字符串数组")
		}
	}
	return nil
}

func validateJSON(value string, label string) error {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	if !json.Valid([]byte(value)) {
		return fmt.Errorf("%s蹇呴』鏄悎娉?JSON", label)
	}
	return nil
}

func validateJSONStringArray(value string, label string) error {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	var rows []string
	if err := json.Unmarshal([]byte(value), &rows); err != nil {
		return fmt.Errorf("%s必须是 JSON 字符串数组", label)
	}
	return nil
}

func logOperation(db *gorm.DB, username string, action string, resource string, recordID uint) {
	if db == nil {
		return
	}
	if username == "" {
		username = "admin"
	}
	_ = db.Create(&model.OperationLog{Username: username, Action: action, Resource: resource, RecordID: recordID}).Error
}
