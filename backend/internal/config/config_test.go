package config

import "testing"

func TestProductionRejectsInsecureDefaults(t *testing.T) {
	cfg := Config{Environment: "production", AdminPassword: "admin123", AuthSecret: "dev-secret-change-me", AllowedOrigins: []string{"https://example.test"}}
	if err := cfg.Validate(); err == nil {
		t.Fatal("production should reject insecure defaults")
	}
}

func TestProductionRequiresCORSOrigins(t *testing.T) {
	cfg := Config{Environment: "production", AdminPassword: "strong-password", AuthSecret: "a-long-random-secret", AllowedOrigins: []string{""}}
	if err := cfg.Validate(); err == nil {
		t.Fatal("production should require an explicit CORS origin")
	}
}

func TestCaptureAIRequiresTencentCredentials(t *testing.T) {
	cfg := Config{CaptureAIEnabled: true}
	if err := cfg.Validate(); err == nil {
		t.Fatal("capture AI should reject missing Tencent credentials")
	}
	cfg.TencentSecretID, cfg.TencentSecretKey = "id", "key"
	if err := cfg.Validate(); err == nil {
		t.Fatal("capture AI should reject missing TokenHub API key")
	}
	cfg.TencentTokenHubKey = "tokenhub-key"
	if err := cfg.Validate(); err != nil {
		t.Fatalf("capture AI rejected configured credentials: %v", err)
	}
}
