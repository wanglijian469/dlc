package api

import (
	"net/http"

	"dalu-nongji-parts/backend/internal/auth"
	"dalu-nongji-parts/backend/internal/config"
	"dalu-nongji-parts/backend/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AdminHandler struct {
	DB     *gorm.DB
	Config config.Config
}

func (h AdminHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, 400, "请求格式错误")
		return
	}
	var user model.AdminUser
	if err := h.DB.Where("username = ? AND is_enabled = ?", req.Username, true).First(&user).Error; err != nil || !auth.CheckPassword(user.PasswordHash, req.Password) {
		Fail(c, http.StatusUnauthorized, 401, "用户名或密码错误")
		return
	}
	token, err := auth.IssueToken(user.Username, h.Config.AuthSecret)
	if err != nil {
		Fail(c, http.StatusInternalServerError, 500, "登录失败")
		return
	}
	OK(c, gin.H{"token": token, "username": user.Username})
}

func (h AdminHandler) Profile(c *gin.Context) {
	OK(c, gin.H{"username": c.GetString("username")})
}

func (h AdminHandler) ListMenus(c *gin.Context)  { list[model.Menu](c, h.DB, "sort_order asc, id asc") }
func (h AdminHandler) CreateMenu(c *gin.Context) { create[model.Menu](c, h.DB) }
func (h AdminHandler) UpdateMenu(c *gin.Context) { update[model.Menu](c, h.DB) }
func (h AdminHandler) DeleteMenu(c *gin.Context) { remove[model.Menu](c, h.DB) }
func (h AdminHandler) ListVendors(c *gin.Context) {
	listPreload[model.Vendor](c, h.DB, "Tags", "sort_order asc, id asc")
}
func (h AdminHandler) CreateVendor(c *gin.Context) { create[model.Vendor](c, h.DB) }
func (h AdminHandler) UpdateVendor(c *gin.Context) { update[model.Vendor](c, h.DB) }
func (h AdminHandler) DeleteVendor(c *gin.Context) { remove[model.Vendor](c, h.DB) }
func (h AdminHandler) ListTags(c *gin.Context)     { list[model.Tag](c, h.DB, "sort_order asc, id asc") }
func (h AdminHandler) CreateTag(c *gin.Context)    { create[model.Tag](c, h.DB) }
func (h AdminHandler) UpdateTag(c *gin.Context)    { update[model.Tag](c, h.DB) }
func (h AdminHandler) DeleteTag(c *gin.Context)    { remove[model.Tag](c, h.DB) }
func (h AdminHandler) ListCategories(c *gin.Context) {
	list[model.Category](c, h.DB, "sort_order asc, id asc")
}
func (h AdminHandler) CreateCategory(c *gin.Context) { create[model.Category](c, h.DB) }
func (h AdminHandler) UpdateCategory(c *gin.Context) { update[model.Category](c, h.DB) }
func (h AdminHandler) DeleteCategory(c *gin.Context) { remove[model.Category](c, h.DB) }
func (h AdminHandler) ListProducts(c *gin.Context) {
	listWithPreloads[model.Product](c, h.DB, []string{"Category", "Vendor"}, "sort_order asc, id asc")
}
func (h AdminHandler) CreateProduct(c *gin.Context) { create[model.Product](c, h.DB) }
func (h AdminHandler) UpdateProduct(c *gin.Context) { update[model.Product](c, h.DB) }
func (h AdminHandler) DeleteProduct(c *gin.Context) { remove[model.Product](c, h.DB) }
func (h AdminHandler) ListBanners(c *gin.Context) {
	list[model.Banner](c, h.DB, "sort_order asc, id asc")
}
func (h AdminHandler) CreateBanner(c *gin.Context) { create[model.Banner](c, h.DB) }
func (h AdminHandler) UpdateBanner(c *gin.Context) { update[model.Banner](c, h.DB) }
func (h AdminHandler) DeleteBanner(c *gin.Context) { remove[model.Banner](c, h.DB) }

func (h AdminHandler) ListConfigs(c *gin.Context) {
	list[model.SiteConfig](c, h.DB, "config_key asc")
}

func (h AdminHandler) UpdateConfig(c *gin.Context) {
	var req model.SiteConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, 400, "请求格式错误")
		return
	}
	key := c.Param("key")
	var item model.SiteConfig
	if err := h.DB.Where("config_key = ?", key).First(&item).Error; err != nil {
		Fail(c, http.StatusNotFound, 404, "配置不存在")
		return
	}
	item.ConfigValue = req.ConfigValue
	item.Description = req.Description
	if err := h.DB.Save(&item).Error; err != nil {
		Fail(c, http.StatusInternalServerError, 500, "保存失败")
		return
	}
	OK(c, item)
}

func list[T any](c *gin.Context, db *gorm.DB, order string) {
	var rows []T
	if err := db.Order(order).Find(&rows).Error; err != nil {
		Fail(c, http.StatusInternalServerError, 500, "读取失败")
		return
	}
	OK(c, rows)
}

func listPreload[T any](c *gin.Context, db *gorm.DB, preload string, order string) {
	listWithPreloads[T](c, db, []string{preload}, order)
}

func listWithPreloads[T any](c *gin.Context, db *gorm.DB, preloads []string, order string) {
	query := db
	for _, preload := range preloads {
		query = query.Preload(preload)
	}
	var rows []T
	if err := query.Order(order).Find(&rows).Error; err != nil {
		Fail(c, http.StatusInternalServerError, 500, "读取失败")
		return
	}
	OK(c, rows)
}

func create[T any](c *gin.Context, db *gorm.DB) {
	var item T
	if err := c.ShouldBindJSON(&item); err != nil {
		Fail(c, http.StatusBadRequest, 400, "请求格式错误")
		return
	}
	if err := db.Create(&item).Error; err != nil {
		Fail(c, http.StatusInternalServerError, 500, "创建失败")
		return
	}
	OK(c, item)
}

func update[T any](c *gin.Context, db *gorm.DB) {
	var item T
	if err := db.First(&item, c.Param("id")).Error; err != nil {
		Fail(c, http.StatusNotFound, 404, "记录不存在")
		return
	}
	if err := c.ShouldBindJSON(&item); err != nil {
		Fail(c, http.StatusBadRequest, 400, "请求格式错误")
		return
	}
	if err := db.Save(&item).Error; err != nil {
		Fail(c, http.StatusInternalServerError, 500, "保存失败")
		return
	}
	OK(c, item)
}

func remove[T any](c *gin.Context, db *gorm.DB) {
	var item T
	if err := db.First(&item, c.Param("id")).Error; err != nil {
		Fail(c, http.StatusNotFound, 404, "记录不存在")
		return
	}
	if err := db.Delete(&item).Error; err != nil {
		Fail(c, http.StatusInternalServerError, 500, "删除失败")
		return
	}
	OK(c, gin.H{"deleted": true})
}
