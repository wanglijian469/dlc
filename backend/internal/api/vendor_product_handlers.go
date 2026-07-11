package api

import (
	"net/http"
	"strings"

	"dalu-nongji-parts/backend/internal/model"
	"github.com/gin-gonic/gin"
)

func (h AdminHandler) ListOwnProducts(c *gin.Context) {
	vendorID := c.GetUint("vendorId")
	if c.GetString("role") != "vendor" || vendorID == 0 {
		Fail(c, http.StatusForbidden, 403, "账号未绑定厂商")
		return
	}
	var products []model.Product
	if err := h.DB.Preload("Category").Where("vendor_id = ?", vendorID).Order("sort_order asc, id desc").Find(&products).Error; err != nil {
		Fail(c, 500, 500, "产品资料加载失败")
		return
	}
	OK(c, products)
}

func (h AdminHandler) CreateOwnProduct(c *gin.Context) {
	vendorID := c.GetUint("vendorId")
	if c.GetString("role") != "vendor" || vendorID == 0 {
		Fail(c, http.StatusForbidden, 403, "账号未绑定厂商")
		return
	}
	var input model.Product
	if err := c.ShouldBindJSON(&input); err != nil || strings.TrimSpace(input.Name) == "" {
		Fail(c, 400, 400, "请填写产品名称")
		return
	}
	input.ID = 0
	input.VendorID = vendorID
	input.Status = 2 // Vendor-entered products stay offline until an administrator reviews them.
	if err := h.DB.Create(&input).Error; err != nil {
		Fail(c, 500, 500, "产品资料保存失败")
		return
	}
	logOperation(h.DB, c.GetString("username"), "create", "vendor-products", input.ID)
	OK(c, input)
}

func (h AdminHandler) UpdateOwnProduct(c *gin.Context) {
	vendorID := c.GetUint("vendorId")
	if c.GetString("role") != "vendor" || vendorID == 0 {
		Fail(c, http.StatusForbidden, 403, "账号未绑定厂商")
		return
	}
	var product model.Product
	if err := h.DB.Where("id = ? AND vendor_id = ?", c.Param("id"), vendorID).First(&product).Error; err != nil {
		Fail(c, 404, 404, "产品不存在")
		return
	}
	var input model.Product
	if err := c.ShouldBindJSON(&input); err != nil || strings.TrimSpace(input.Name) == "" {
		Fail(c, 400, 400, "请填写产品名称")
		return
	}
	updates := map[string]interface{}{"name": input.Name, "image": input.Image, "category_id": input.CategoryID, "compatible_models": input.CompatibleModels, "description": input.Description, "detail_content": input.DetailContent, "gallery": input.GalleryRaw, "specs": input.SpecsRaw, "price_note": input.PriceNote, "status": 2}
	if err := h.DB.Model(&product).Updates(updates).Error; err != nil {
		Fail(c, 500, 500, "产品资料保存失败")
		return
	}
	logOperation(h.DB, c.GetString("username"), "update", "vendor-products", product.ID)
	h.DB.Preload("Category").First(&product, product.ID)
	OK(c, product)
}

func (h AdminHandler) DeleteOwnProduct(c *gin.Context) {
	vendorID := c.GetUint("vendorId")
	if c.GetString("role") != "vendor" || vendorID == 0 {
		Fail(c, http.StatusForbidden, 403, "账号未绑定厂商")
		return
	}
	result := h.DB.Where("id = ? AND vendor_id = ?", c.Param("id"), vendorID).Delete(&model.Product{})
	if result.Error != nil {
		Fail(c, 500, 500, "产品删除失败")
		return
	}
	if result.RowsAffected == 0 {
		Fail(c, 404, 404, "产品不存在")
		return
	}
	OK(c, gin.H{"deleted": true})
}
