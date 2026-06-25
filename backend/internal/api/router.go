package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"dalu-nongji-parts/backend/internal/auth"
	"dalu-nongji-parts/backend/internal/config"
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
		AllowOrigins:     []string{"http://127.0.0.1:5173", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))
	RegisterHealthRoute(router)
	RegisterPublicRoutes(router, deps.DB)
	RegisterAdminRoutes(router, deps.DB, deps.Config)
	RegisterStaticRoutes(router, deps.Config.PublicDir)
	return router
}

func RegisterHealthRoute(router *gin.Engine) {
	router.GET("/api/health", func(c *gin.Context) { OK(c, gin.H{"status": "ok"}) })
}

func RegisterPublicRoutes(router *gin.Engine, db *gorm.DB) {
	handler := PublicHandler{DB: db, HomeService: service.HomeService{DB: db}}
	api := router.Group("/api")
	api.GET("/home", handler.Home)
	api.GET("/menus", handler.Menus)
	api.GET("/vendors", handler.Vendors)
	api.GET("/vendors/recommended", handler.RecommendedVendors)
	api.GET("/vendors/:id", handler.VendorDetail)
	api.GET("/products", handler.Products)
	api.GET("/search", handler.Search)
}

func RegisterAdminRoutes(router *gin.Engine, db *gorm.DB, cfg config.Config) {
	handler := AdminHandler{DB: db, Config: cfg}
	admin := router.Group("/api/admin")
	admin.POST("/login", handler.Login)
	protected := admin.Group("")
	protected.Use(AdminAuth(cfg.AuthSecret))
	protected.GET("/profile", handler.Profile)
	protected.GET("/menus", handler.ListMenus)
	protected.POST("/menus", handler.CreateMenu)
	protected.PUT("/menus/:id", handler.UpdateMenu)
	protected.DELETE("/menus/:id", handler.DeleteMenu)
	protected.GET("/vendors", handler.ListVendors)
	protected.POST("/vendors", handler.CreateVendor)
	protected.PUT("/vendors/:id", handler.UpdateVendor)
	protected.DELETE("/vendors/:id", handler.DeleteVendor)
	protected.GET("/tags", handler.ListTags)
	protected.POST("/tags", handler.CreateTag)
	protected.PUT("/tags/:id", handler.UpdateTag)
	protected.DELETE("/tags/:id", handler.DeleteTag)
	protected.GET("/categories", handler.ListCategories)
	protected.POST("/categories", handler.CreateCategory)
	protected.PUT("/categories/:id", handler.UpdateCategory)
	protected.DELETE("/categories/:id", handler.DeleteCategory)
	protected.GET("/products", handler.ListProducts)
	protected.POST("/products", handler.CreateProduct)
	protected.PUT("/products/:id", handler.UpdateProduct)
	protected.DELETE("/products/:id", handler.DeleteProduct)
	protected.GET("/banners", handler.ListBanners)
	protected.POST("/banners", handler.CreateBanner)
	protected.PUT("/banners/:id", handler.UpdateBanner)
	protected.DELETE("/banners/:id", handler.DeleteBanner)
	protected.GET("/configs", handler.ListConfigs)
	protected.PUT("/configs/:key", handler.UpdateConfig)
}

func AdminAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			Fail(c, 401, 401, "unauthorized")
			c.Abort()
			return
		}
		username, err := auth.ParseToken(strings.TrimPrefix(header, "Bearer "), secret)
		if err != nil {
			Fail(c, 401, 401, "token expired")
			c.Abort()
			return
		}
		c.Set("username", username)
		c.Next()
	}
}

func RegisterStaticRoutes(router *gin.Engine, publicDir string) {
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
	router.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			Fail(c, http.StatusNotFound, 404, "not found")
			return
		}
		c.File(filepath.Join(publicDir, "index.html"))
	})
}
