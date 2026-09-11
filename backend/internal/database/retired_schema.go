package database

import (
	"fmt"

	"gorm.io/gorm"
)

// These objects belonged to abandoned catalog/master-data prototypes. They
// have no model, API, job, frontend, mobile, or deployment-script references.
var retiredDatabaseTables = []string{
	"master_assets",
	"master_categories",
	"master_import_batches",
	"master_offerings",
	"master_products",
	"master_revisions",
	"master_vendors",
}

var retiredDatabaseColumns = map[string][]string{
	"categories":        {"catalog_code"},
	"media_assets":      {"original_storage_key"},
	"products":          {"catalog_code"},
	"product_suppliers": {"catalog_code"},
	"vendors": {
		"catalog_code",
		"quality_control",
		"supply_regions",
		"cooperation_terms",
		"source_url",
		"source_note",
		"service_models",
	},
}

func dropRetiredDatabaseObjects(db *gorm.DB) error {
	for _, table := range retiredDatabaseTables {
		if !db.Migrator().HasTable(table) {
			continue
		}
		if err := db.Migrator().DropTable(table); err != nil {
			return fmt.Errorf("drop retired table %s: %w", table, err)
		}
	}
	for table, columns := range retiredDatabaseColumns {
		for _, column := range columns {
			if !db.Migrator().HasColumn(table, column) {
				continue
			}
			if err := db.Migrator().DropColumn(table, column); err != nil {
				return fmt.Errorf("drop retired column %s.%s: %w", table, column, err)
			}
		}
	}
	return nil
}
