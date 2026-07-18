package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"dalu-nongji-parts/backend/internal/model"
	"dalu-nongji-parts/backend/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PublicHandler struct {
	DB          *gorm.DB
	HomeService service.HomeService
}

func (h PublicHandler) Home(c *gin.Context) {
	payload, err := h.HomeService.GetHome(c.Request.Context())
	if err != nil {
		Fail(c, 500, 500, "首页数据加载失败")
		return
	}
	if !c.GetBool("authenticated") {
		redactVendorSlice(payload.RecommendedVendors)
		redactVendorSlice(payload.MoreVendors)
		redactVendorSlice(payload.ProcessingVendors)
	}
	OK(c, payload)
}

func (h PublicHandler) SiteMeta(c *gin.Context) {
	OK(c, h.HomeService.SiteMeta(c.Request.Context()))
}

func (h PublicHandler) LayoutConfig(c *gin.Context) {
	// Revalidate on each page load so category-driven menus are visible after a
	// normal refresh while still allowing intermediary caches to store a copy.
	c.Header("Cache-Control", "public, no-cache")
	OK(c, h.HomeService.Layout(c.Request.Context()))
}

func (h PublicHandler) Menus(c *gin.Context) {
	if c.Query("type") == "sidebar" {
		OK(c, h.HomeService.SidebarMenus(c.Request.Context()))
		return
	}
	var menus []model.Menu
	query := h.DB.Where("is_enabled = ?", true)
	if typ := c.Query("type"); typ != "" {
		query = query.Where("menu_type = ?", typ)
	}
	query.Order("sort_order asc, id asc").Find(&menus)
	OK(c, menus)
}

func (h PublicHandler) Page(c *gin.Context) {
	var page model.ContentPage
	if err := publishedPageQuery(h.DB, "page").Where("slug = ?", c.Param("slug")).First(&page).Error; err != nil {
		Fail(c, http.StatusNotFound, 404, "页面不存在")
		return
	}
	OK(c, page)
}

func (h PublicHandler) Articles(c *gin.Context) {
	var pages []model.ContentPage
	page, pageSize := pageParams(c, 12)
	query := publishedPageQuery(h.DB, "article")
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("title LIKE ? OR summary LIKE ? OR seo_keywords LIKE ?", like, like, like)
	}
	result, err := paginate(query.Order("published_at desc, updated_at desc, id desc"), &pages, page, pageSize)
	if err != nil {
		Fail(c, http.StatusInternalServerError, 500, "行业文章加载失败")
		return
	}
	OK(c, result)
}

func (h PublicHandler) Article(c *gin.Context) {
	var page model.ContentPage
	if err := publishedPageQuery(h.DB, "article").Where("slug = ?", c.Param("slug")).First(&page).Error; err != nil {
		Fail(c, http.StatusNotFound, 404, "行业文章不存在")
		return
	}
	OK(c, page)
}

func (h PublicHandler) FriendLinks(c *gin.Context) {
	var links []model.FriendLink
	h.DB.Where("is_enabled = ?", true).Order("sort_order asc, id asc").Find(&links)
	OK(c, links)
}

func (h PublicHandler) Vendors(c *gin.Context) {
	var vendors []model.Vendor
	page, pageSize := pageParams(c, 12)
	query := publishedVendorQuery(h.DB).Preload("Tags").Preload("Media", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order asc, id asc") })
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("name LIKE ? OR short_name LIKE ? OR main_products LIKE ?", like, like, like)
	}
	if province := strings.TrimSpace(c.Query("province")); province != "" {
		query = query.Where("province = ?", province)
	}
	if tagID := queryUint(c, "tagId"); tagID > 0 {
		query = query.Joins("JOIN vendor_tags ON vendor_tags.vendor_id = vendors.id AND vendor_tags.tag_id = ?", tagID)
	}
	if categoryID := queryUint(c, "categoryId"); categoryID > 0 {
		var category model.Category
		if err := h.DB.First(&category, categoryID).Error; err == nil && category.Name != "" {
			query = query.Where("main_products LIKE ?", "%"+category.Name+"%")
		}
	}
	if c.Query("sort") == "latest" {
		query = query.Order("created_at desc")
	} else {
		query = query.Order("is_recommended desc, sort_order asc, id asc")
	}
	result, err := paginate(query, &vendors, page, pageSize)
	if err != nil {
		Fail(c, 500, 500, "厂商列表加载失败")
		return
	}
	if !c.GetBool("authenticated") {
		redactVendorSlice(vendors)
	}
	OK(c, result)
}

func (h PublicHandler) ProcessingVendors(c *gin.Context) {
	var vendors []model.Vendor
	page, pageSize := pageParams(c, 12)
	query := publishedVendorQuery(h.DB).Preload("Tags").Preload("Media", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order asc, id asc") }).Where("provides_processing = ?", true)
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("name LIKE ? OR short_name LIKE ? OR main_products LIKE ? OR processing_services LIKE ? OR processing_materials LIKE ? OR processing_equipment LIKE ? OR processing_capacity LIKE ? OR processing_regions LIKE ? OR processing_notes LIKE ?", like, like, like, like, like, like, like, like, like)
	}
	if province := strings.TrimSpace(c.Query("province")); province != "" {
		query = query.Where("province = ?", province)
	}
	if tagID := queryUint(c, "tagId"); tagID > 0 {
		query = query.Joins("JOIN vendor_tags ON vendor_tags.vendor_id = vendors.id AND vendor_tags.tag_id = ?", tagID).
			Joins("JOIN tags processing_tags ON processing_tags.id = vendor_tags.tag_id AND processing_tags.tag_type = ?", "processing")
	}
	if c.Query("sort") == "latest" {
		query = query.Order("created_at desc")
	} else {
		query = query.Order("is_recommended desc, sort_order asc, id asc")
	}
	result, err := paginate(query, &vendors, page, pageSize)
	if err != nil {
		Fail(c, 500, 500, "加工厂商列表加载失败")
		return
	}
	if !c.GetBool("authenticated") {
		redactVendorSlice(vendors)
	}
	OK(c, result)
}

func (h PublicHandler) ProcessingFilterOptions(c *gin.Context) {
	var provinces []string
	var tags []model.Tag
	h.DB.Model(&model.Vendor{}).Where("is_visible = ? AND publication_status = ? AND provides_processing = ? AND province <> ''", true, "published", true).Distinct().Order("province asc").Pluck("province", &provinces)
	h.DB.Where("tag_type = ?", "processing").Order("sort_order asc, id asc").Find(&tags)
	OK(c, gin.H{"provinces": provinces, "categories": []model.Category{}, "serviceTags": tags})
}

func (h PublicHandler) RecommendedVendors(c *gin.Context) {
	var vendors []model.Vendor
	publishedVendorQuery(h.DB).Preload("Tags").Preload("Media", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order asc, id asc") }).Where("is_recommended = ?", true).Order("sort_order asc, id asc").Limit(5).Find(&vendors)
	if !c.GetBool("authenticated") {
		redactVendorSlice(vendors)
	}
	OK(c, vendors)
}

func (h PublicHandler) VendorDetail(c *gin.Context) {
	var vendor model.Vendor
	if err := publishedVendorQuery(h.DB).Preload("Tags").Preload("Media", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order asc, id asc") }).First(&vendor, "id = ?", c.Param("id")).Error; err != nil {
		Fail(c, 404, 404, "厂商不存在")
		return
	}
	for _, tag := range vendor.Tags {
		vendor.TagIDs = append(vendor.TagIDs, tag.ID)
	}
	if !c.GetBool("authenticated") {
		redactVendor(&vendor)
	}
	OK(c, vendor)
}

func (h PublicHandler) VendorBySlug(c *gin.Context) {
	var vendor model.Vendor
	if err := publishedVendorQuery(h.DB).Preload("Tags").Preload("Media", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order asc, id asc") }).First(&vendor, "slug = ?", c.Param("slug")).Error; err != nil {
		Fail(c, http.StatusNotFound, 404, "厂商不存在")
		return
	}
	for _, tag := range vendor.Tags {
		vendor.TagIDs = append(vendor.TagIDs, tag.ID)
	}
	if !c.GetBool("authenticated") {
		redactVendor(&vendor)
	}
	OK(c, vendor)
}

func (h PublicHandler) Products(c *gin.Context) {
	var products []model.Product
	page, pageSize := pageParams(c, 12)
	query := visibleProductQuery(h.DB).Preload("Category")
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("products.name LIKE ? OR products.compatible_models LIKE ? OR products.description LIKE ?", like, like, like)
	}
	if categoryID := queryUint(c, "categoryId"); categoryID > 0 {
		query = query.Where("products.category_id IN ?", productCategoryIDs(h.DB, categoryID))
	} else if categorySlug := strings.TrimSpace(c.Query("categorySlug")); categorySlug != "" {
		var category model.Category
		if publishedCategoryQuery(h.DB).First(&category, "slug = ?", categorySlug).Error != nil {
			Fail(c, http.StatusNotFound, 404, "产品分类不存在")
			return
		}
		query = query.Where("products.category_id IN ?", productCategoryIDs(h.DB, category.ID))
	}
	if vendorID := queryUint(c, "vendorId"); vendorID > 0 {
		query = query.Where("EXISTS (SELECT 1 FROM product_suppliers ps WHERE ps.product_id = products.id AND ps.vendor_id = ? AND ps.status = 'approved' AND ps.deleted_at IS NULL)", vendorID)
	} else if vendorSlug := strings.TrimSpace(c.Query("vendorId")); vendorSlug != "" {
		query = query.Where("EXISTS (SELECT 1 FROM product_suppliers ps JOIN vendors filter_vendor ON filter_vendor.id = ps.vendor_id WHERE ps.product_id = products.id AND filter_vendor.slug = ? AND ps.status = 'approved' AND ps.deleted_at IS NULL)", vendorSlug)
	}
	if c.Query("hot") == "true" {
		query = query.Where("products.is_hot = ?", true)
	}
	if c.Query("recommended") == "true" {
		query = query.Where("products.is_recommended = ?", true)
	}
	if c.Query("sort") == "latest" {
		query = query.Order("products.created_at desc")
	} else {
		query = query.Order("products.is_recommended desc, products.sort_order asc, products.id asc")
	}
	result, err := paginate(query, &products, page, pageSize)
	if err != nil {
		Fail(c, 500, 500, "产品列表加载失败")
		return
	}
	enrichProductSummaries(h.DB, products, queryUint(c, "vendorId"))
	if !c.GetBool("authenticated") {
		for i := range products {
			if products[i].Supplier != nil {
				redactVendor(&products[i].Supplier.Vendor)
			}
		}
	}
	OK(c, result)
}

func productCategoryIDs(db *gorm.DB, categoryID uint) []uint {
	ids := []uint{categoryID}
	var category model.Category
	if err := db.First(&category, categoryID).Error; err != nil || category.ParentID != 0 {
		return ids
	}
	var childIDs []uint
	db.Model(&model.Category{}).Where("parent_id = ?", categoryID).Pluck("id", &childIDs)
	return append(ids, childIDs...)
}

func (h PublicHandler) ProductDetail(c *gin.Context) {
	var product model.Product
	if err := visibleProductQuery(h.DB).Preload("Category").First(&product, "products.id = ?", c.Param("id")).Error; err != nil {
		Fail(c, http.StatusNotFound, 404, "产品不存在")
		return
	}
	rows := []model.Product{product}
	enrichProductSummaries(h.DB, rows, 0)
	product = rows[0]
	OK(c, product)
}

func (h PublicHandler) ProductBySlug(c *gin.Context) {
	var product model.Product
	if err := visibleProductQuery(h.DB).Preload("Category").First(&product, "products.slug = ?", c.Param("slug")).Error; err != nil {
		Fail(c, http.StatusNotFound, 404, "产品不存在")
		return
	}
	rows := []model.Product{product}
	enrichProductSummaries(h.DB, rows, 0)
	OK(c, rows[0])
}

func (h PublicHandler) CategoryBySlug(c *gin.Context) {
	var category model.Category
	if err := publishedCategoryQuery(h.DB).First(&category, "slug = ?", c.Param("slug")).Error; err != nil {
		Fail(c, http.StatusNotFound, 404, "分类不存在")
		return
	}
	OK(c, category)
}

func (h PublicHandler) ProductSuppliers(c *gin.Context) {
	var product model.Product
	productQuery := visibleProductQuery(h.DB)
	if _, err := strconv.ParseUint(c.Param("id"), 10, 64); err == nil {
		productQuery = productQuery.Where("products.id = ?", c.Param("id"))
	} else {
		productQuery = productQuery.Where("products.slug = ?", c.Param("id"))
	}
	if err := productQuery.First(&product).Error; err != nil {
		Fail(c, 404, 404, "产品不存在")
		return
	}
	var suppliers []model.ProductSupplier
	query := h.DB.Preload("Vendor").Where("product_id = ? AND status = ?", product.ID, "approved").Where("EXISTS (SELECT 1 FROM vendors v WHERE v.id = product_suppliers.vendor_id AND v.deleted_at IS NULL AND v.is_visible = 1 AND v.publication_status = 'published')").Order("id asc")
	if err := query.Find(&suppliers).Error; err != nil {
		Fail(c, 500, 500, "供应商列表加载失败")
		return
	}
	if !c.GetBool("authenticated") {
		for i := range suppliers {
			redactVendor(&suppliers[i].Vendor)
		}
	}
	OK(c, suppliers)
}

func (h PublicHandler) Search(c *gin.Context) {
	keyword := strings.TrimSpace(c.Query("keyword"))
	page, pageSize := pageParams(c, 10)
	if keyword == "" {
		OK(c, gin.H{"vendors": PageResult{Items: []model.Vendor{}, Page: page, PageSize: pageSize, Total: 0}, "products": PageResult{Items: []model.Product{}, Page: page, PageSize: pageSize, Total: 0}, "categories": PageResult{Items: []model.Category{}, Page: page, PageSize: pageSize, Total: 0}})
		return
	}
	like := "%" + keyword + "%"
	var vendors []model.Vendor
	var products []model.Product
	var categories []model.Category
	vendorQuery := h.DB.Model(&model.Vendor{}).Preload("Tags").Preload("Media", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order asc, id asc") }).Where("is_visible = ? AND publication_status = ? AND (name LIKE ? OR short_name LIKE ? OR main_products LIKE ?)", true, "published", like, like, like).Order("is_recommended desc, sort_order asc, id asc")
	productQuery := visibleProductQuery(h.DB).Preload("Category").Where("products.name LIKE ? OR products.compatible_models LIKE ? OR products.description LIKE ?", like, like, like).Order("products.is_recommended desc, products.sort_order asc, products.id asc")
	categoryQuery := publishedCategoryQuery(h.DB).Where("name LIKE ?", like).Order("sort_order asc, id asc")
	vendorResult, err := paginate(vendorQuery, &vendors, page, pageSize)
	if err != nil {
		Fail(c, 500, 500, "搜索失败")
		return
	}
	productResult, err := paginate(productQuery, &products, page, pageSize)
	if err != nil {
		Fail(c, 500, 500, "搜索失败")
		return
	}
	categoryResult, err := paginate(categoryQuery, &categories, page, pageSize)
	if err != nil {
		Fail(c, 500, 500, "搜索失败")
		return
	}
	enrichProductSummaries(h.DB, products, 0)
	if !c.GetBool("authenticated") {
		redactVendorSlice(vendors)
	}
	OK(c, gin.H{"vendors": vendorResult, "products": productResult, "categories": categoryResult})
}

func redactVendorSlice(vendors []model.Vendor) {
	for i := range vendors {
		redactVendor(&vendors[i])
	}
}

func redactProductVendors(products []model.Product) {
	for i := range products {
		if products[i].Vendor != nil {
			redactVendor(products[i].Vendor)
		}
	}
}

func redactVendor(vendor *model.Vendor) {
	if vendor.Phone != "" {
		vendor.Phone = maskPhone(vendor.Phone)
	}
	vendor.Wechat = ""
	vendor.ContactName = ""
}

func maskPhone(value string) string {
	runes := []rune(value)
	digitPositions := make([]int, 0, len(runes))
	for index, char := range runes {
		if char >= '0' && char <= '9' {
			digitPositions = append(digitPositions, index)
		}
	}
	if len(digitPositions) <= 5 {
		return "****"
	}
	for _, position := range digitPositions[3 : len(digitPositions)-2] {
		runes[position] = '*'
	}
	return string(runes)
}

func (h PublicHandler) FilterOptions(c *gin.Context) {
	var provinces []string
	var categories []model.Category
	var tags []model.Tag
	h.DB.Model(&model.Vendor{}).Where("is_visible = ? AND publication_status = ? AND province <> ''", true, "published").Distinct().Order("province asc").Pluck("province", &provinces)
	publishedCategoryQuery(h.DB).Order("sort_order asc, id asc").Find(&categories)
	h.DB.Where("tag_type = ?", "vendor").Order("sort_order asc, id asc").Find(&tags)
	OK(c, gin.H{"provinces": provinces, "categories": categories, "serviceTags": tags})
}

func publishedPageQuery(db *gorm.DB, pageType string) *gorm.DB {
	return db.Model(&model.ContentPage{}).Where("page_type = ? AND is_enabled = ? AND publication_status = ? AND (published_at IS NULL OR published_at <= ?)", pageType, true, "published", time.Now())
}

func publishedVendorQuery(db *gorm.DB) *gorm.DB {
	return db.Model(&model.Vendor{}).Where("is_visible = ? AND publication_status = ? AND (published_at IS NULL OR published_at <= ?)", true, "published", time.Now())
}

func publishedCategoryQuery(db *gorm.DB) *gorm.DB {
	return db.Model(&model.Category{}).Where("is_enabled = ? AND publication_status = ? AND (published_at IS NULL OR published_at <= ?)", true, "published", time.Now())
}

func paginate(query *gorm.DB, dest interface{}, page int, pageSize int) (PageResult, error) {
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return PageResult{}, err
	}
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(dest).Error; err != nil {
		return PageResult{}, err
	}
	return PageResult{Items: dest, Page: page, PageSize: pageSize, Total: total}, nil
}

func pageParams(c *gin.Context, defaultSize int) (int, int) {
	pageSize := clamp(queryInt(c, "pageSize", defaultSize), 1, 100)
	page := clamp(queryInt(c, "page", 1), 1, 10000)
	return page, pageSize
}

func queryInt(c *gin.Context, key string, fallback int) int {
	value, err := strconv.Atoi(c.Query(key))
	if err != nil {
		return fallback
	}
	return value
}

func queryUint(c *gin.Context, key string) uint {
	value, err := strconv.ParseUint(c.Query(key), 10, 64)
	if err != nil {
		return 0
	}
	return uint(value)
}

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
