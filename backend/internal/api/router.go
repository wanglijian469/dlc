package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"dalu-nongji-parts/backend/internal/auth"
	"dalu-nongji-parts/backend/internal/config"
	"dalu-nongji-parts/backend/internal/model"
	"dalu-nongji-parts/backend/internal/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Deps struct {
	DB     *gorm.DB
	Config config.Config
}

func NewRouter(deps Deps) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     deps.Config.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))
	RegisterHealthRoute(router)
	RegisterPublicRoutesWithAuth(router, deps.DB, deps.Config.AuthSecret)
	RegisterAdminRoutes(router, deps.DB, deps.Config)
	mediaHandler := AdminHandler{DB: deps.DB, Config: deps.Config}
	router.GET("/api/media/:id", mediaHandler.PublicMedia)
	RegisterStaticRoutes(router, deps.DB, deps.Config.PublicDir)
	return router
}

func RegisterHealthRoute(router *gin.Engine) {
	router.GET("/api/health", func(c *gin.Context) { OK(c, gin.H{"status": "ok"}) })
}

func RegisterPublicRoutes(router *gin.Engine, db *gorm.DB) {
	RegisterPublicRoutesWithAuth(router, db, "")
}

func RegisterPublicRoutesWithAuth(router *gin.Engine, db *gorm.DB, secret string) {
	handler := PublicHandler{DB: db, HomeService: service.HomeService{DB: db}}
	api := router.Group("/api")
	api.Use(OptionalCMSAuth(db, secret))
	api.GET("/home", handler.Home)
	api.GET("/site-meta", handler.SiteMeta)
	api.GET("/layout-config", handler.LayoutConfig)
	api.GET("/menus", handler.Menus)
	api.GET("/pages/:slug", handler.Page)
	api.GET("/friend-links", handler.FriendLinks)
	api.GET("/processing-vendors", handler.ProcessingVendors)
	api.GET("/processing-filter-options", handler.ProcessingFilterOptions)
	api.GET("/vendors", handler.Vendors)
	api.GET("/vendors/recommended", handler.RecommendedVendors)
	api.GET("/vendors/:id", handler.VendorDetail)
	api.GET("/products", handler.Products)
	api.GET("/products/:id", handler.ProductDetail)
	api.GET("/products/:id/suppliers", handler.ProductSuppliers)
	api.GET("/search", handler.Search)
	api.GET("/filter-options", handler.FilterOptions)
}

// OptionalCMSAuth enriches public requests when a valid CMS account token is
// present. Invalid or missing tokens remain anonymous and never block browsing.
func OptionalCMSAuth(db *gorm.DB, secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if db == nil || secret == "" || !strings.HasPrefix(header, "Bearer ") {
			c.Next()
			return
		}
		username, err := auth.ParseToken(strings.TrimPrefix(header, "Bearer "), secret)
		if err == nil {
			var user model.AdminUser
			if db.Where("username = ? AND is_enabled = ?", username, true).First(&user).Error == nil {
				c.Set("authenticated", true)
				c.Set("username", username)
				c.Set("role", user.Role)
			}
		}
		c.Next()
	}
}

func RegisterAdminRoutes(router *gin.Engine, db *gorm.DB, cfg config.Config) {
	handler := AdminHandler{DB: db, Config: cfg}
	admin := router.Group("/api/admin")
	admin.POST("/login", handler.Login)
	admin.POST("/register", handler.Register)
	protected := admin.Group("")
	protected.Use(CMSAuth(db, cfg.AuthSecret))
	protected.GET("/profile", handler.Profile)

	cmsOnly := protected.Group("")
	cmsOnly.Use(RequireAnyRole("admin", "vendor"))
	cmsOnly.POST("/uploads", handler.SecureUpload)
	cmsOnly.POST("/remote-images", handler.DownloadRemoteImage)
	cmsOnly.GET("/media/:id", handler.PreviewMedia)
	cmsOnly.DELETE("/media/:id", handler.DeleteMedia)
	cmsOnly.GET("/vendor-profile", handler.GetVendorProfile)
	cmsOnly.PUT("/vendor-profile", handler.SubmitVendorProfile)
	cmsOnly.GET("/vendor-products", handler.ListOwnProducts)
	cmsOnly.GET("/vendor-product-catalog", handler.SearchVendorProductCatalog)
	cmsOnly.POST("/vendor-products", handler.CreateOwnProduct)
	cmsOnly.POST("/vendor-products/link", handler.LinkOwnProduct)
	cmsOnly.PUT("/vendor-products/:id", handler.UpdateOwnProduct)
	cmsOnly.DELETE("/vendor-products/:id", handler.DeleteOwnProduct)

	adminOnly := protected.Group("")
	adminOnly.Use(RequireRole("admin"))
	adminOnly.GET("/dashboard", handler.DashboardStats)
	adminOnly.GET("/operation-logs", handler.ListOperationLogs)
	adminOnly.POST("/imports/:resource", handler.ImportWorkbook)
	adminOnly.GET("/menus", handler.ListMenus)
	adminOnly.POST("/menus", handler.CreateMenu)
	adminOnly.PUT("/menus/:id", handler.UpdateMenu)
	adminOnly.DELETE("/menus/:id", handler.DeleteMenu)
	adminOnly.GET("/vendors", handler.ListVendors)
	adminOnly.POST("/vendors", handler.CreateVendor)
	adminOnly.PUT("/vendors/:id", handler.UpdateVendor)
	adminOnly.DELETE("/vendors/:id", handler.DeleteVendor)
	adminOnly.GET("/vendor-submissions", handler.ListVendorSubmissions)
	adminOnly.PUT("/vendor-submissions/:id/review", handler.ReviewVendorSubmission)
	adminOnly.GET("/product-submissions", handler.ListProductSubmissions)
	adminOnly.PUT("/product-submissions/:id/review", handler.ReviewProductSubmission)
	adminOnly.GET("/users", handler.ListCMSUsers)
	adminOnly.POST("/users", handler.CreateCMSUser)
	adminOnly.PUT("/users/:id", handler.UpdateCMSUser)
	adminOnly.DELETE("/users/:id", handler.DeleteCMSUser)
	adminOnly.GET("/tags", handler.ListTags)
	adminOnly.POST("/tags", handler.CreateTag)
	adminOnly.PUT("/tags/:id", handler.UpdateTag)
	adminOnly.DELETE("/tags/:id", handler.DeleteTag)
	adminOnly.GET("/categories", handler.ListCategories)
	adminOnly.POST("/categories", handler.CreateCategory)
	adminOnly.PUT("/categories/:id", handler.UpdateCategory)
	adminOnly.DELETE("/categories/:id", handler.DeleteCategory)
	adminOnly.GET("/products", handler.ListProducts)
	adminOnly.POST("/products", handler.CreateProduct)
	adminOnly.PUT("/products/:id", handler.UpdateProduct)
	adminOnly.DELETE("/products/:id", handler.DeleteProduct)
	adminOnly.GET("/products/:id/suppliers", handler.ListAdminProductSuppliers)
	adminOnly.POST("/products/:id/suppliers", handler.SaveAdminProductSupplier)
	adminOnly.DELETE("/products/:id/suppliers/:supplierId", handler.DisableAdminProductSupplier)
	adminOnly.PUT("/products/:id/merge", handler.MergeProducts)
	adminOnly.GET("/banners", handler.ListBanners)
	adminOnly.POST("/banners", handler.CreateBanner)
	adminOnly.PUT("/banners/:id", handler.UpdateBanner)
	adminOnly.DELETE("/banners/:id", handler.DeleteBanner)
	adminOnly.GET("/pages", handler.ListPages)
	adminOnly.POST("/pages", handler.CreatePage)
	adminOnly.PUT("/pages/:id", handler.UpdatePage)
	adminOnly.DELETE("/pages/:id", handler.DeletePage)
	adminOnly.GET("/friend-links", handler.ListFriendLinks)
	adminOnly.POST("/friend-links", handler.CreateFriendLink)
	adminOnly.PUT("/friend-links/:id", handler.UpdateFriendLink)
	adminOnly.DELETE("/friend-links/:id", handler.DeleteFriendLink)
	adminOnly.GET("/configs", handler.ListConfigs)
	adminOnly.PUT("/configs/:key", handler.UpdateConfig)
}

func AdminAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			Fail(c, 401, 401, "未登录或登录已过期")
			c.Abort()
			return
		}
		username, err := auth.ParseToken(strings.TrimPrefix(header, "Bearer "), secret)
		if err != nil {
			Fail(c, 401, 401, "登录已过期，请重新登录")
			c.Abort()
			return
		}
		c.Set("username", username)
		c.Next()
	}
}

// CMSAuth enriches a valid token with the current user's role and vendor scope.
// Keeping AdminAuth separate preserves the small stateless middleware used by tests.
func CMSAuth(db *gorm.DB, secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			Fail(c, http.StatusUnauthorized, 401, "未登录或登录已过期")
			c.Abort()
			return
		}
		username, err := auth.ParseToken(strings.TrimPrefix(header, "Bearer "), secret)
		if err != nil {
			Fail(c, http.StatusUnauthorized, 401, "登录已过期，请重新登录")
			c.Abort()
			return
		}
		c.Set("username", username)
		if db == nil {
			c.Set("role", "admin")
			c.Next()
			return
		}
		var user model.AdminUser
		if err := db.Where("username = ? AND is_enabled = ?", c.GetString("username"), true).First(&user).Error; err != nil {
			Fail(c, http.StatusUnauthorized, 401, "账号不存在或已停用")
			c.Abort()
			return
		}
		role := strings.TrimSpace(user.Role)
		if role == "" {
			role = "admin"
		}
		c.Set("role", role)
		if user.VendorID != nil {
			c.Set("vendorId", *user.VendorID)
		}
		c.Next()
	}
}

func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetString("role") != role {
			Fail(c, http.StatusForbidden, 403, "没有权限执行此操作")
			c.Abort()
			return
		}
		c.Next()
	}
}

func RequireAnyRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}
	return func(c *gin.Context) {
		if _, ok := allowed[c.GetString("role")]; !ok {
			Fail(c, http.StatusForbidden, 403, "没有权限执行此操作")
			c.Abort()
			return
		}
		c.Next()
	}
}

func RegisterStaticRoutes(router *gin.Engine, db *gorm.DB, publicDir string) {
	uploadsDir := "uploads"
	if publicDir != "" {
		uploadsDir = filepath.Join(publicDir, "uploads")
	}
	if _, err := os.Stat(uploadsDir); err == nil {
		router.GET("/uploads/:name", func(c *gin.Context) {
			name := filepath.Base(c.Param("name"))
			if name == "." || name == "" || name != c.Param("name") {
				c.Status(http.StatusNotFound)
				return
			}
			url := "/uploads/" + name
			var count int64
			db.Model(&model.Vendor{}).Where("is_visible = ? AND publication_status = ? AND (logo = ? OR cover_image = ?)", true, "published", url, url).Count(&count)
			if count == 0 {
				db.Model(&model.VendorMedia{}).Joins("JOIN vendors ON vendors.id = vendor_media.vendor_id").Where("vendor_media.url = ? AND vendors.is_visible = ? AND vendors.publication_status = ?", url, true, "published").Count(&count)
			}
			if count == 0 {
				db.Model(&model.Product{}).Where("products.publication_status = ? AND (products.image = ? OR products.gallery LIKE ?) AND EXISTS (SELECT 1 FROM product_suppliers ps JOIN vendors v ON v.id = ps.vendor_id WHERE ps.product_id = products.id AND ps.status = 'approved' AND v.is_visible = 1 AND v.publication_status = 'published')", "published", url, "%"+url+"%").Count(&count)
				if count == 0 {
					db.Model(&model.ProductSupplier{}).Joins("JOIN vendors ON vendors.id = product_suppliers.vendor_id").Where("product_suppliers.status = ? AND vendors.is_visible = ? AND vendors.publication_status = ? AND (product_suppliers.image = ? OR product_suppliers.gallery LIKE ?)", "approved", true, "published", url, "%"+url+"%").Count(&count)
				}
			}
			if count == 0 {
				db.Model(&model.Banner{}).Where("is_enabled = ? AND background_image = ?", true, url).Count(&count)
			}
			if count == 0 {
				c.Status(http.StatusNotFound)
				return
			}
			c.File(filepath.Join(uploadsDir, name))
		})
	}
	if publicDir == "" {
		return
	}
	if _, err := os.Stat(publicDir); err != nil {
		return
	}
	assetsDir := filepath.Join(publicDir, "assets")
	if _, err := os.Stat(assetsDir); err == nil {
		router.Static("/assets", assetsDir)
	}
	imagesDir := filepath.Join(publicDir, "images")
	if _, err := os.Stat(imagesDir); err == nil {
		router.Static("/images", imagesDir)
	}
	router.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			Fail(c, http.StatusNotFound, 404, "接口不存在")
			return
		}
		c.File(filepath.Join(publicDir, "index.html"))
	})
}
