package database

import "testing"

func TestCurrentSchemaMigrationRunsForRetiredDatabaseObjects(t *testing.T) {
	if CurrentSchemaVersion != 23 {
		t.Fatalf("CurrentSchemaVersion = %d, want 23 for retired database cleanup", CurrentSchemaVersion)
	}
	if !schemaMigrationRequired(22) || !schemaMigrationRequired(20) || !schemaMigrationRequired(21) {
		t.Fatal("version 20 and 21 databases must run the retired database cleanup")
	}
	if schemaMigrationRequired(CurrentSchemaVersion) {
		t.Fatal("current database version must not rerun the migration")
	}
	if currentSchemaMigrationName != "remove-supply-demand-v1" {
		t.Fatalf("migration name = %q", currentSchemaMigrationName)
	}
}
