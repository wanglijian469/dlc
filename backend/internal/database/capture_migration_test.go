package database

import "testing"

func TestCaptureSchemaMigrationRunsFromVersion18(t *testing.T) {
	if CurrentSchemaVersion != 19 {
		t.Fatalf("CurrentSchemaVersion = %d, want 19 for capture workbench tables", CurrentSchemaVersion)
	}
	if !schemaMigrationRequired(18) {
		t.Fatal("version 18 database must run the capture workbench migration")
	}
	if schemaMigrationRequired(CurrentSchemaVersion) {
		t.Fatal("current database version must not rerun the migration")
	}
	if currentSchemaMigrationName != "capture-workbench-v1" {
		t.Fatalf("migration name = %q", currentSchemaMigrationName)
	}
}
