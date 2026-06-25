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
		Fail(c, 500, 500, "failed to load home data")
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
	query := h.DB.Preload("Tags").Where("is_visible = ?", true)
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("name LIKE ? OR main_products LIKE ?", like, like)
	}
	if province := c.Query("province"); province != "" {
		query = query.Where("province = ?", province)
	}
	if c.Query("sort") == "latest" {
		query = query.Order("created_at desc")
	} else {
		query = query.Order("is_recommended desc, sort_order asc, id asc")
	}
	pageSize := clamp(queryInt(c, "pageSize", 20), 1, 100)
	page := clamp(queryInt(c, "page", 1), 1, 10000)
	query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&vendors)
	OK(c, vendors)
}

func (h PublicHandler) RecommendedVendors(c *gin.Context) {
	var vendors []model.Vendor
	h.DB.Preload("Tags").Where("is_visible = ? AND is_recommended = ?", true, true).Order("sort_order asc, id asc").Limit(5).Find(&vendors)
	OK(c, vendors)
}

func (h PublicHandler) VendorDetail(c *gin.Context) {
	var vendor model.Vendor
	if err := h.DB.Preload("Tags").First(&vendor, c.Param("id")).Error; err != nil {
		Fail(c, 404, 404, "vendor not found")
		return
	}
	OK(c, vendor)
}

func (h PublicHandler) Products(c *gin.Context) {
	var products []model.Product
	query := h.DB.Preload("Category").Preload("Vendor").Where("status = ?", 1)
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		query = query.Where("name LIKE ? OR compatible_models LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	query.Order("sort_order asc, id asc").Limit(clamp(queryInt(c, "pageSize", 30), 1, 100)).Find(&products)
	OK(c, products)
}

func (h PublicHandler) Search(c *gin.Context) {
	keyword := strings.TrimSpace(c.Query("keyword"))
	if keyword == "" {
		OK(c, gin.H{"vendors": []model.Vendor{}, "products": []model.Product{}, "categories": []model.Category{}})
		return
	}
	like := "%" + keyword + "%"
	var vendors []model.Vendor
	var products []model.Product
	var categories []model.Category
	h.DB.Preload("Tags").Where("is_visible = ? AND (name LIKE ? OR main_products LIKE ?)", true, like, like).Limit(20).Find(&vendors)
	h.DB.Preload("Category").Preload("Vendor").Where("status = ? AND (name LIKE ? OR compatible_models LIKE ?)", 1, like, like).Limit(20).Find(&products)
	h.DB.Where("is_enabled = ? AND name LIKE ?", true, like).Limit(20).Find(&categories)
	OK(c, gin.H{"vendors": vendors, "products": products, "categories": categories})
}

func queryInt(c *gin.Context, key string, fallback int) int {
	value, err := strconv.Atoi(c.Query(key))
	if err != nil {
		return fallback
	}
	return value
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
