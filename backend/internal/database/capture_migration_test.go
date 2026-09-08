package database

import "testing"

func TestCaptureSchemaMigrationRunsFromVersion18(t *testing.T) {
	if CurrentSchemaVersion != 20 {
		t.Fatalf("CurrentSchemaVersion = %d, want 20 for vendor workspace tables", CurrentSchemaVersion)
	}
	if !schemaMigrationRequired(18) {
		t.Fatal("version 18 database must run the capture workbench migration")
	}
	if schemaMigrationRequired(CurrentSchemaVersion) {
		t.Fatal("current database version must not rerun the migration")
	}
	if currentSchemaMigrationName != "vendor-showroom-workspace-v1" {
		t.Fatalf("migration name = %q", currentSchemaMigrationName)
	}
}
