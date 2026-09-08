package api

import (
	"strings"
	"testing"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestPublishableSupplierDeletionQueriesAreIndependent(t *testing.T) {
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       "user:password@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local",
		SkipInitializeWithVersion: true,
	}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
	if err != nil {
		t.Fatalf("open dry-run database: %v", err)
	}

	now := time.Date(2026, 8, 24, 0, 0, 0, 0, time.UTC)
	currentSQL := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		var count int64
		return publishableProductSuppliers(tx, 42, now).
			Where("product_suppliers.id = ?", 9).
			Count(&count)
	})
	remainingSQL := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		var count int64
		return publishableProductSuppliers(tx, 42, now).
			Where("product_suppliers.id <> ?", 9).
			Count(&count)
	})

	if !strings.Contains(currentSQL, "product_suppliers.id = 9") {
		t.Fatalf("current supplier query is missing its id filter: %s", currentSQL)
	}
	if !strings.Contains(remainingSQL, "product_suppliers.id <> 9") {
		t.Fatalf("remaining supplier query is missing its exclusion filter: %s", remainingSQL)
	}
	if strings.Contains(remainingSQL, "product_suppliers.id = 9") {
		t.Fatalf("remaining supplier query inherited the current supplier filter: %s", remainingSQL)
	}
}
