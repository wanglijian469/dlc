package main

import (
	"log"

	"dalu-nongji-parts/backend/internal/api"
	"dalu-nongji-parts/backend/internal/config"
	"dalu-nongji-parts/backend/internal/database"
)

func main() {
	cfg := config.Load()
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
	router := api.NewRouter(api.Deps{DB: db, Config: cfg})
	log.Printf("starting API on %s", cfg.HTTPAddr)
	if err := router.Run(cfg.HTTPAddr); err != nil {
		log.Fatal(err)
	}
}
