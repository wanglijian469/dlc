package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	HTTPAddr           string
	DBHost             string
	DBPort             string
	DBUser             string
	DBPassword         string
	DBName             string
	AdminUsername      string
	AdminPassword      string
	AuthSecret         string
	PublicDir          string
	MediaDir           string
	Environment        string
	AllowedOrigins     []string
	TrustedProxyCIDRs  []string
	GeoIPDBPath        string
	SeedDemoData       bool
	RunMigrations      bool
	BaiduPushToken     string
	StaticPagesEnabled bool
	StaticPageDir      string
	CaptureAIEnabled   bool
	TencentSecretID    string
	TencentSecretKey   string
	TencentRegion      string
	TencentTokenHubKey string
	TokenHubBaseURL    string
	CaptureVisionModel string
}

func Load() Config {
	seedDemo, _ := strconv.ParseBool(env("SEED_DEMO_DATA", "false"))
	runMigrations, _ := strconv.ParseBool(env("RUN_MIGRATIONS", ""))
	staticPagesEnabled, _ := strconv.ParseBool(env("STATIC_PAGES_ENABLED", "true"))
	captureAIEnabled, _ := strconv.ParseBool(env("CAPTURE_AI_ENABLED", "false"))
	environment := env("APP_ENV", "development")
	if os.Getenv("RUN_MIGRATIONS") == "" {
		runMigrations = !strings.EqualFold(environment, "production")
	}
	origins := strings.Split(env("CORS_ALLOWED_ORIGINS", "http://127.0.0.1:5173,http://localhost:5173"), ",")
	trustedProxies := strings.Split(env("TRUSTED_PROXY_CIDRS", "127.0.0.1,::1"), ",")
	publicDir := env("PUBLIC_DIR", "")
	staticPageDir := env("STATIC_PAGE_DIR", "")
	if staticPageDir == "" {
		if publicDir != "" {
			staticPageDir = filepath.Join(publicDir, "static-pages")
		} else {
			staticPageDir = "static_pages"
		}
	}
	return Config{
		HTTPAddr:           env("HTTP_ADDR", ":8080"),
		DBHost:             env("DB_HOST", "127.0.0.1"),
		DBPort:             env("DB_PORT", "13306"),
		DBUser:             env("DB_USER", "root"),
		DBPassword:         env("DB_PASSWORD", "root"),
		DBName:             env("DB_NAME", "dl_nongji_parts"),
		AdminUsername:      env("ADMIN_USERNAME", "admin"),
		AdminPassword:      env("ADMIN_PASSWORD", "admin123"),
		AuthSecret:         env("AUTH_SECRET", "dev-secret-change-me"),
		PublicDir:          publicDir,
		MediaDir:           env("MEDIA_DIR", "media_storage"),
		Environment:        environment,
		AllowedOrigins:     origins,
		TrustedProxyCIDRs:  trustedProxies,
		GeoIPDBPath:        env("GEOIP_DB_PATH", ""),
		SeedDemoData:       seedDemo,
		RunMigrations:      runMigrations,
		BaiduPushToken:     env("BAIDU_PUSH_TOKEN", ""),
		StaticPagesEnabled: staticPagesEnabled,
		StaticPageDir:      staticPageDir,
		CaptureAIEnabled:   captureAIEnabled,
		TencentSecretID:    env("TENCENT_CLOUD_SECRET_ID", ""),
		TencentSecretKey:   env("TENCENT_CLOUD_SECRET_KEY", ""),
		TencentRegion:      env("TENCENT_CLOUD_REGION", "ap-guangzhou"),
		TencentTokenHubKey: env("TENCENT_TOKENHUB_API_KEY", ""),
		TokenHubBaseURL:    env("TENCENT_TOKENHUB_BASE_URL", "https://tokenhub.tencentmaas.com/v1"),
		CaptureVisionModel: env("CAPTURE_VISION_MODEL", "hunyuan-t1-vision-20250916"),
	}
}

func (c Config) Validate() error {
	if c.CaptureAIEnabled && (strings.TrimSpace(c.TencentSecretID) == "" || strings.TrimSpace(c.TencentSecretKey) == "" || strings.TrimSpace(c.TencentTokenHubKey) == "") {
		return fmt.Errorf("CAPTURE_AI_ENABLED requires Tencent OCR credentials and TENCENT_TOKENHUB_API_KEY")
	}
	if strings.EqualFold(c.Environment, "production") {
		if c.AdminPassword == "admin123" || c.AuthSecret == "dev-secret-change-me" {
			return fmt.Errorf("production refuses default administrator password or signing secret")
		}
		if len(c.AllowedOrigins) == 0 || (len(c.AllowedOrigins) == 1 && strings.TrimSpace(c.AllowedOrigins[0]) == "") {
			return fmt.Errorf("production requires CORS_ALLOWED_ORIGINS")
		}
		if strings.TrimSpace(c.PublicDir) == "" {
			return fmt.Errorf("production requires PUBLIC_DIR")
		}
	}
	return nil
}

func (c Config) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

func (c Config) ServerDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local", c.DBUser, c.DBPassword, c.DBHost, c.DBPort)
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
