package main

import (
	"flag"
	"log"

	"dalu-nongji-parts/backend/internal/config"
	"dalu-nongji-parts/backend/internal/database"
)

func main() {
	migrateOnly := flag.Bool("migrate-only", false, "apply schema migrations without seeding or legacy backfill")
	flag.Parse()
	cfg := config.Load()
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		log.Fatalf("migrate database: %v", err)
	}
	if *migrateOnly {
		log.Printf("database schema migration completed")
		return
	}
	if err := database.SeedDefaults(db, cfg); err != nil {
		log.Fatalf("seed database: %v", err)
	}
	if err := database.BackfillPlatformData(db); err != nil {
		log.Fatalf("backfill platform data: %v", err)
	}
	log.Printf("database initialized: %s:%s/%s", cfg.DBHost, cfg.DBPort, cfg.DBName)
}
