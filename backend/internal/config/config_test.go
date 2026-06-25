package config

import "testing"

func TestDefaultConfigUsesLocalMySQL(t *testing.T) {
	cfg := Load()
	if cfg.DBHost != "127.0.0.1" {
		t.Fatalf("DBHost = %q, want 127.0.0.1", cfg.DBHost)
	}
	if cfg.DBPort != "13306" {
		t.Fatalf("DBPort = %q, want 13306", cfg.DBPort)
	}
	if cfg.DBUser != "root" || cfg.DBPassword != "root" {
		t.Fatalf("default credentials = %q/%q, want root/root", cfg.DBUser, cfg.DBPassword)
	}
}
