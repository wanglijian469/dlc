package api

import (
	"errors"
	"net/http"
	"strings"

	"dalu-nongji-parts/backend/internal/database"
	"dalu-nongji-parts/backend/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"time"
)

func saveCategory(c *gin.Context, db *gorm.DB, id uint) {
	var category model.Category
	if id > 0 {
		if err := db.First(&category, id).Error; err != nil {
			Fail(c, http.StatusNotFound, 404, "分类不存在")
			return
		}
	}
	if err := c.ShouldBindJSON(&category); err != nil {
		Fail(c, http.StatusBadRequest, 400, "请求格式错误")
		return
	}
	category.ID = id
	category.Name = strings.TrimSpace(category.Name)
	if category.Name == "" {
		Fail(c, http.StatusBadRequest, 400, "分类名称不能为空")
		return
	}
	if message, conflict := validateCategoryHierarchy(db, category); message != "" {
		status := http.StatusBadRequest
		if conflict {
			status = http.StatusConflict
		}
		Fail(c, status, status, message)
		return
	}

	disabled := !category.IsEnabled
	if category.ContentVersion == 0 {
		category.ContentVersion = 1
	} else if id > 0 {
		category.ContentVersion++
	}
	if category.IsEnabled {
		category.PublicationStatus = "published"
		if category.PublishedAt == nil {
			now := time.Now()
			category.PublishedAt = &now
		}
	} else {
		category.PublicationStatus = "archived"
	}
	if err := database.EnsureCategorySlug(db, &category); err != nil {
		Fail(c, http.StatusConflict, 409, "分类页面标识冲突")
		return
	}
	if err := db.Save(&category).Error; err != nil {
		Fail(c, http.StatusInternalServerError, 500, "分类保存失败")
		return
	}
	// GORM applies the model's default:true tag to false on insert. Respect the
	// value submitted by the category form after the row has an ID.
	if id == 0 && disabled {
		if err := db.Model(&category).UpdateColumn("is_enabled", false).Error; err != nil {
			Fail(c, http.StatusInternalServerError, 500, "分类保存失败")
			return
		}
		category.IsEnabled = false
	}
	logOperation(db, c.GetString("username"), upsertAction(id), "categories", category.ID)
	OK(c, category)
}

func validateCategoryHierarchy(db *gorm.DB, category model.Category) (message string, conflict bool) {
	if category.ID > 0 && category.ParentID == category.ID {
		return "不能将分类自身设为父级", false
	}
	if category.ParentID > 0 {
		var parent model.Category
		if err := db.First(&parent, category.ParentID).Error; err != nil {
			return "请选择有效的一级分类", false
		}
		if parent.ParentID != 0 {
			return "配件分类最多支持两级，二级分类不能作为父级", false
		}
		if category.ID > 0 {
			var childCount int64
			db.Model(&model.Category{}).Where("parent_id = ?", category.ID).Count(&childCount)
			if childCount > 0 {
				return "该一级分类存在子分类，不能改为二级分类", false
			}
		}
	}

	query := db.Model(&model.Category{}).
		Where("parent_id = ? AND LOWER(TRIM(name)) = LOWER(?)", category.ParentID, category.Name)
	if category.ID > 0 {
		query = query.Where("id <> ?", category.ID)
	}
	var duplicateCount int64
	query.Count(&duplicateCount)
	if duplicateCount > 0 {
		return "同一父级下已存在同名分类", true
	}
	return "", false
}

func deleteCategory(c *gin.Context, db *gorm.DB, id uint) {
	var category model.Category
	if err := db.First(&category, id).Error; err != nil {
		Fail(c, http.StatusNotFound, 404, "分类不存在")
		return
	}
	var childCount int64
	if err := db.Model(&model.Category{}).Where("parent_id = ?", id).Count(&childCount).Error; err != nil {
		Fail(c, http.StatusInternalServerError, 500, "分类依赖检查失败")
		return
	}
	if childCount > 0 {
		Fail(c, http.StatusConflict, 409, "该分类存在子分类，请先迁移或删除子分类；也可以关闭启用状态临时下线")
		return
	}
	var productCount int64
	if err := db.Model(&model.Product{}).Where("category_id = ?", id).Count(&productCount).Error; err != nil {
		Fail(c, http.StatusInternalServerError, 500, "分类依赖检查失败")
		return
	}
	if productCount > 0 {
		Fail(c, http.StatusConflict, 409, "该分类已关联产品，请先迁移产品；也可以关闭启用状态临时下线")
		return
	}
	if err := db.Delete(&category).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		Fail(c, http.StatusInternalServerError, 500, "分类删除失败")
		return
	}
	logOperation(db, c.GetString("username"), "delete", "categories", id)
	OK(c, gin.H{"deleted": true})
}
