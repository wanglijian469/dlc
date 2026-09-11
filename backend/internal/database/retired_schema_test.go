package database

import (
	"reflect"
	"testing"
)

func TestRetiredDatabaseObjectsAreExplicit(t *testing.T) {
	wantTables := []string{
		"master_assets", "master_categories", "master_import_batches",
		"master_offerings", "master_products", "master_revisions", "master_vendors",
	}
	if !reflect.DeepEqual(retiredDatabaseTables, wantTables) {
		t.Fatalf("retired tables = %v, want %v", retiredDatabaseTables, wantTables)
	}
	wantColumns := map[string][]string{
		"categories": {"catalog_code"}, "media_assets": {"original_storage_key"},
		"products": {"catalog_code"}, "product_suppliers": {"catalog_code"},
		"vendors": {"catalog_code", "quality_control", "supply_regions", "cooperation_terms", "source_url", "source_note", "service_models"},
	}
	if !reflect.DeepEqual(retiredDatabaseColumns, wantColumns) {
		t.Fatalf("retired columns = %v, want %v", retiredDatabaseColumns, wantColumns)
	}
}
