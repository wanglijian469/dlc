package main

import (
	"context"
	"log"
	"net/url"
	"strings"
	"time"

	"dalu-nongji-parts/backend/internal/api"
	"dalu-nongji-parts/backend/internal/config"
	"dalu-nongji-parts/backend/internal/database"
	"dalu-nongji-parts/backend/internal/service"
)

func main() {
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	if cfg.RunMigrations {
		if err := database.Migrate(db); err != nil {
			log.Fatalf("migrate database: %v", err)
		}
	} else if err := database.CheckMigrations(db); err != nil {
		log.Fatalf("check database migrations: %v", err)
	}
	if err := database.SeedDefaults(db, cfg); err != nil {
		log.Fatalf("seed database: %v", err)
	}
	if strings.EqualFold(cfg.Environment, "production") {
		meta := (service.HomeService{DB: db}).SiteMeta(context.Background())
		parsed, parseErr := url.Parse(meta.SiteURL)
		if parseErr != nil || parsed.Scheme != "https" || parsed.Host == "" || strings.Contains(strings.ToLower(parsed.Host), "example") {
			log.Fatalf("production requires a formal HTTPS siteUrl in site.meta")
		}
	}
	if cfg.RunMigrations {
		if err := database.BackfillPlatformData(db); err != nil {
			log.Fatalf("backfill platform data: %v", err)
		}
	}
	if err := database.CleanupOrphanedMedia(db, cfg.MediaDir); err != nil {
		log.Printf("cleanup staged media: %v", err)
	}
	if err := database.CleanupAnalytics(db, 90*24*time.Hour); err != nil {
		log.Printf("cleanup analytics: %v", err)
	}
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			if err := database.CleanupAnalytics(db, 90*24*time.Hour); err != nil {
				log.Printf("cleanup analytics: %v", err)
			}
		}
	}()
	go func() {
		if err := api.PublishScheduledRevisions(db); err != nil {
			log.Printf("publish scheduled content: %v", err)
		}
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			if err := api.PublishScheduledRevisions(db); err != nil {
				log.Printf("publish scheduled content: %v", err)
			}
		}
	}()
	router := api.NewRouter(api.Deps{DB: db, Config: cfg})
	log.Printf("starting API on %s", cfg.HTTPAddr)
	if err := router.Run(cfg.HTTPAddr); err != nil {
		log.Fatal(err)
	}
}
