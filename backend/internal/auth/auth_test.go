package auth

import "testing"

func TestPasswordHashVerifies(t *testing.T) {
	hash, err := HashPassword("admin123")
	if err != nil {
		t.Fatal(err)
	}
	if !CheckPassword(hash, "admin123") {
		t.Fatal("expected password to verify")
	}
	if CheckPassword(hash, "wrong") {
		t.Fatal("expected wrong password to fail")
	}
}

func TestTokenRoundTrip(t *testing.T) {
	token, err := IssueToken("admin", "secret")
	if err != nil {
		t.Fatal(err)
	}
	username, err := ParseToken(token, "secret")
	if err != nil {
		t.Fatal(err)
	}
	if username != "admin" {
		t.Fatalf("username = %q, want admin", username)
	}
}
