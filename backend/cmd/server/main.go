package main

import (
	"log"
	"time"

	"dalu-nongji-parts/backend/internal/api"
	"dalu-nongji-parts/backend/internal/config"
	"dalu-nongji-parts/backend/internal/database"
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
	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("migrate database: %v", err)
	}
	if err := database.SeedDefaults(db, cfg); err != nil {
		log.Fatalf("seed database: %v", err)
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
	router := api.NewRouter(api.Deps{DB: db, Config: cfg})
	log.Printf("starting API on %s", cfg.HTTPAddr)
	if err := router.Run(cfg.HTTPAddr); err != nil {
		log.Fatal(err)
	}
}
