package api

import (
	"strconv"
	"strings"

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
	OK(c, payload)
}

func (h PublicHandler) Menus(c *gin.Context) {
	var menus []model.Menu
	query := h.DB.Where("is_enabled = ?", true)
	if typ := c.Query("type"); typ != "" {
		query = query.Where("menu_type = ?", typ)
	}
	query.Order("sort_order asc, id asc").Find(&menus)
	if c.Query("type") == "sidebar" {
		OK(c, service.BuildMenuTree(menus))
		return
	}
	OK(c, menus)
}

func (h PublicHandler) Vendors(c *gin.Context) {
	var vendors []model.Vendor
	page, pageSize := pageParams(c, 12)
	query := h.DB.Model(&model.Vendor{}).Preload("Tags").Where("is_visible = ?", true)
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
	OK(c, paginate(query, &vendors, page, pageSize))
}

func (h PublicHandler) RecommendedVendors(c *gin.Context) {
	var vendors []model.Vendor
	h.DB.Preload("Tags").Where("is_visible = ? AND is_recommended = ?", true, true).Order("sort_order asc, id asc").Limit(5).Find(&vendors)
	OK(c, vendors)
}

func (h PublicHandler) VendorDetail(c *gin.Context) {
	var vendor model.Vendor
	if err := h.DB.Preload("Tags").First(&vendor, c.Param("id")).Error; err != nil {
		Fail(c, 404, 404, "厂商不存在")
		return
	}
	for _, tag := range vendor.Tags {
		vendor.TagIDs = append(vendor.TagIDs, tag.ID)
	}
	OK(c, vendor)
}

func (h PublicHandler) Products(c *gin.Context) {
	var products []model.Product
	page, pageSize := pageParams(c, 12)
	query := h.DB.Model(&model.Product{}).Preload("Category").Preload("Vendor").Where("status = ?", 1)
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("name LIKE ? OR compatible_models LIKE ? OR description LIKE ?", like, like, like)
	}
	if categoryID := queryUint(c, "categoryId"); categoryID > 0 {
		query = query.Where("category_id = ?", categoryID)
	}
	if vendorID := queryUint(c, "vendorId"); vendorID > 0 {
		query = query.Where("vendor_id = ?", vendorID)
	}
	if c.Query("hot") == "true" {
		query = query.Where("is_hot = ?", true)
	}
	if c.Query("recommended") == "true" {
		query = query.Where("is_recommended = ?", true)
	}
	if c.Query("sort") == "latest" {
		query = query.Order("created_at desc")
	} else {
		query = query.Order("is_recommended desc, sort_order asc, id asc")
	}
	OK(c, paginate(query, &products, page, pageSize))
}

func (h PublicHandler) Search(c *gin.Context) {
	keyword := strings.TrimSpace(c.Query("keyword"))
	page, pageSize := pageParams(c, 10)
	if keyword == "" {
		OK(c, gin.H{
			"vendors":    PageResult{Items: []model.Vendor{}, Page: page, PageSize: pageSize, Total: 0},
			"products":   PageResult{Items: []model.Product{}, Page: page, PageSize: pageSize, Total: 0},
			"categories": PageResult{Items: []model.Category{}, Page: page, PageSize: pageSize, Total: 0},
		})
		return
	}
	like := "%" + keyword + "%"
	var vendors []model.Vendor
	var products []model.Product
	var categories []model.Category
	vendorQuery := h.DB.Model(&model.Vendor{}).Preload("Tags").Where("is_visible = ? AND (name LIKE ? OR short_name LIKE ? OR main_products LIKE ?)", true, like, like, like).Order("is_recommended desc, sort_order asc, id asc")
	productQuery := h.DB.Model(&model.Product{}).Preload("Category").Preload("Vendor").Where("status = ? AND (name LIKE ? OR compatible_models LIKE ? OR description LIKE ?)", 1, like, like, like).Order("is_recommended desc, sort_order asc, id asc")
	categoryQuery := h.DB.Model(&model.Category{}).Where("is_enabled = ? AND name LIKE ?", true, like).Order("sort_order asc, id asc")
	OK(c, gin.H{
		"vendors":    paginate(vendorQuery, &vendors, page, pageSize),
		"products":   paginate(productQuery, &products, page, pageSize),
		"categories": paginate(categoryQuery, &categories, page, pageSize),
	})
}

func (h PublicHandler) FilterOptions(c *gin.Context) {
	var provinces []string
	var categories []model.Category
	var tags []model.Tag
	h.DB.Model(&model.Vendor{}).Where("is_visible = ? AND province <> ''", true).Distinct().Order("province asc").Pluck("province", &provinces)
	h.DB.Where("is_enabled = ?", true).Order("sort_order asc, id asc").Find(&categories)
	h.DB.Where("tag_type = ?", "vendor").Order("sort_order asc, id asc").Find(&tags)
	OK(c, gin.H{"provinces": provinces, "categories": categories, "serviceTags": tags})
}

func paginate(query *gorm.DB, dest interface{}, page int, pageSize int) PageResult {
	var total int64
	query.Count(&total)
	query.Offset((page - 1) * pageSize).Limit(pageSize).Find(dest)
	return PageResult{Items: dest, Page: page, PageSize: pageSize, Total: total}
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
