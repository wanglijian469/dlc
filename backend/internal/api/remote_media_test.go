package api

import "testing"

func TestValidateRemoteImageURLRejectsUnsafeAddresses(t *testing.T) {
	for _, value := range []string{
		"file:///tmp/image.png",
		"ftp://example.com/image.png",
		"http://127.0.0.1/image.png",
		"http://[::1]/image.png",
		"https://user:pass@example.com/image.png",
	} {
		if _, err := validateRemoteImageURL(value); err == nil {
			t.Fatalf("validateRemoteImageURL(%q) returned nil error", value)
		}
	}
	parsed, err := validateRemoteImageURL("https://images.example.com/catalog/logo.png")
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Hostname() != "images.example.com" {
		t.Fatalf("hostname = %q", parsed.Hostname())
	}
}
