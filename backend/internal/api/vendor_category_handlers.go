package api

import (
	"net/http"
	"strings"
	"time"

	"dalu-nongji-parts/backend/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h PublicHandler) VendorCategories(c *gin.Context) {
	var categories []model.VendorCategory
	if err := h.DB.Where("is_enabled = ?", true).Order("parent_id asc, sort_order asc, id asc").Find(&categories).Error; err != nil {
		Fail(c, http.StatusInternalServerError, 500, "厂商分类加载失败")
		return
	}
	if len(categories) == 0 {
		OK(c, []model.VendorCategory{})
		return
	}

	var counts []struct {
		CategoryID  uint
		VendorCount int64
	}
	err := h.DB.Table("vendor_category_assignments AS vca").
		Select("CASE WHEN vc.parent_id = 0 THEN vc.id ELSE vc.parent_id END AS category_id, COUNT(DISTINCT vca.vendor_id) AS vendor_count").
		Joins("JOIN vendor_categories vc ON vc.id = vca.vendor_category_id AND vc.deleted_at IS NULL AND vc.is_enabled = ?", true).
		Joins("JOIN vendors ON vendors.id = vca.vendor_id AND vendors.deleted_at IS NULL").
		Where("vendors.is_visible = ? AND vendors.publication_status = ? AND (vendors.published_at IS NULL OR vendors.published_at <= ?)", true, "published", time.Now()).
		Group("CASE WHEN vc.parent_id = 0 THEN vc.id ELSE vc.parent_id END").Scan(&counts).Error
	if err != nil {
		Fail(c, http.StatusInternalServerError, 500, "厂商分类统计加载失败")
		return
	}
	countByRoot := make(map[uint]int64, len(counts))
	for _, count := range counts {
		countByRoot[count.CategoryID] = count.VendorCount
	}

	var directCounts []struct {
		CategoryID  uint
		VendorCount int64
	}
	err = h.DB.Table("vendor_category_assignments AS vca").
		Select("vca.vendor_category_id AS category_id, COUNT(DISTINCT vca.vendor_id) AS vendor_count").
		Joins("JOIN vendors ON vendors.id = vca.vendor_id AND vendors.deleted_at IS NULL").
		Where("vendors.is_visible = ? AND vendors.publication_status = ? AND (vendors.published_at IS NULL OR vendors.published_at <= ?)", true, "published", time.Now()).
		Group("vca.vendor_category_id").Scan(&directCounts).Error
	if err != nil {
		Fail(c, http.StatusInternalServerError, 500, "厂商分类统计加载失败")
		return
	}
	countByCategory := make(map[uint]int64, len(directCounts))
	for _, count := range directCounts {
		countByCategory[count.CategoryID] = count.VendorCount
	}
	OK(c, buildVendorCategoryTree(categories, countByRoot, countByCategory))
}

func buildVendorCategoryTree(categories []model.VendorCategory, rootCounts, directCounts map[uint]int64) []model.VendorCategory {
	children := make(map[uint][]model.VendorCategory)
	for _, category := range categories {
		if category.ParentID > 0 {
			category.VendorCount = directCounts[category.ID]
			children[category.ParentID] = append(children[category.ParentID], category)
		}
	}
	result := make([]model.VendorCategory, 0)
	for _, category := range categories {
		if category.ParentID != 0 {
			continue
		}
		category.VendorCount = rootCounts[category.ID]
		category.Children = children[category.ID]
		result = append(result, category)
	}
	return result
}

func vendorCategoryFilterIDs(db *gorm.DB, categoryID uint) []uint {
	if categoryID == 0 {
		return nil
	}
	var category model.VendorCategory
	if err := db.Where("is_enabled = ?", true).First(&category, categoryID).Error; err != nil {
		// Disabled and deleted categories are not public filters. Keep a concrete
		// impossible ID so callers still apply the EXISTS clause and return no rows.
		return []uint{0}
	}
	ids := []uint{category.ID}
	if category.ParentID == 0 {
		var childIDs []uint
		db.Model(&model.VendorCategory{}).Where("parent_id = ? AND is_enabled = ?", category.ID, true).Pluck("id", &childIDs)
		ids = append(ids, childIDs...)
	}
	return ids
}

func (h AdminHandler) ListVendorCategories(c *gin.Context) {
	var categories []model.VendorCategory
	if err := h.DB.Order("parent_id asc, sort_order asc, id asc").Find(&categories).Error; err != nil {
		Fail(c, http.StatusInternalServerError, 500, "厂商分类列表加载失败")
		return
	}
	OK(c, categories)
}

func (h AdminHandler) CreateVendorCategory(c *gin.Context) { saveVendorCategory(c, h.DB, 0) }
func (h AdminHandler) UpdateVendorCategory(c *gin.Context) { saveVendorCategory(c, h.DB, idParam(c)) }
func (h AdminHandler) DeleteVendorCategory(c *gin.Context) { deleteVendorCategory(c, h.DB, idParam(c)) }

func saveVendorCategory(c *gin.Context, db *gorm.DB, id uint) {
	var item model.VendorCategory
	if id > 0 {
		if err := db.First(&item, id).Error; err != nil {
			Fail(c, http.StatusNotFound, 404, "厂商分类不存在")
			return
		}
	}
	if err := c.ShouldBindJSON(&item); err != nil {
		Fail(c, http.StatusBadRequest, 400, "请求参数格式不正确")
		return
	}
	item.ID = id
	item.Name = strings.TrimSpace(item.Name)
	if item.Name == "" {
		Fail(c, http.StatusBadRequest, 400, "分类名称不能为空")
		return
	}
	if item.ParentID == id && id > 0 {
		Fail(c, http.StatusBadRequest, 400, "分类不能将自己设为父级")
		return
	}
	if item.ParentID > 0 {
		var parent model.VendorCategory
		if err := db.First(&parent, item.ParentID).Error; err != nil || parent.ParentID > 0 {
			Fail(c, http.StatusBadRequest, 400, "厂商分类最多支持两级，父级必须是一级分类")
			return
		}
	}
	if id > 0 && item.ParentID > 0 {
		var childCount int64
		db.Model(&model.VendorCategory{}).Where("parent_id = ?", id).Count(&childCount)
		if childCount > 0 {
			Fail(c, http.StatusConflict, 409, "包含二级分类的一级分类不能移动到其他分类下")
			return
		}
	}
	var duplicate int64
	db.Model(&model.VendorCategory{}).Where("parent_id = ? AND name = ? AND id <> ?", item.ParentID, item.Name, id).Count(&duplicate)
	if duplicate > 0 {
		Fail(c, http.StatusConflict, 409, "同级厂商分类名称已存在")
		return
	}
	if err := db.Save(&item).Error; err != nil {
		Fail(c, http.StatusInternalServerError, 500, "厂商分类保存失败")
		return
	}
	logOperation(db, c.GetString("username"), upsertAction(id), "vendor-categories", item.ID)
	OK(c, item)
}

func deleteVendorCategory(c *gin.Context, db *gorm.DB, id uint) {
	var item model.VendorCategory
	if err := db.First(&item, id).Error; err != nil {
		Fail(c, http.StatusNotFound, 404, "厂商分类不存在")
		return
	}
	var childCount, assignmentCount int64
	db.Model(&model.VendorCategory{}).Where("parent_id = ?", id).Count(&childCount)
	db.Model(&model.VendorCategoryAssignment{}).Where("vendor_category_id = ?", id).Count(&assignmentCount)
	if childCount > 0 || assignmentCount > 0 {
		Fail(c, http.StatusConflict, 409, "该分类仍有子分类或关联厂商，请先停用或解除关联")
		return
	}
	if err := db.Delete(&item).Error; err != nil {
		Fail(c, http.StatusInternalServerError, 500, "厂商分类删除失败")
		return
	}
	logOperation(db, c.GetString("username"), "delete", "vendor-categories", id)
	OK(c, gin.H{"deleted": true})
}

func vendorCategoryIDsFromRows(rows []model.VendorCategory) []uint {
	ids := make([]uint, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	return uniqueUintIDs(ids)
}
