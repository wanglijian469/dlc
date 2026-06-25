package config

import (
	"fmt"
	"os"
)

type Config struct {
	HTTPAddr      string
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	AdminUsername string
	AdminPassword string
	AuthSecret    string
	PublicDir     string
}

func Load() Config {
	return Config{
		HTTPAddr:      env("HTTP_ADDR", ":8080"),
		DBHost:        env("DB_HOST", "127.0.0.1"),
		DBPort:        env("DB_PORT", "13306"),
		DBUser:        env("DB_USER", "root"),
		DBPassword:    env("DB_PASSWORD", "root"),
		DBName:        env("DB_NAME", "dl_nongji_parts"),
		AdminUsername: env("ADMIN_USERNAME", "admin"),
		AdminPassword: env("ADMIN_PASSWORD", "admin123"),
		AuthSecret:    env("AUTH_SECRET", "dev-secret-change-me"),
		PublicDir:     env("PUBLIC_DIR", ""),
	}
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
