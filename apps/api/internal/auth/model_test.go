package auth

import (
	"errors"
	"strings"
	"testing"
)

func TestNormalizeEmailCanonicalizesAddress(t *testing.T) {
	got, err := NormalizeEmail("  User@Example.COM ")
	if err != nil {
		t.Fatalf("normalize email: %v", err)
	}
	if got != "user@example.com" {
		t.Fatalf("expected normalized email, got %q", got)
	}
}

func TestNormalizeEmailRejectsDisplayName(t *testing.T) {
	_, err := NormalizeEmail("User <user@example.com>")
	if !errors.Is(err, ErrEmailInvalid) {
		t.Fatalf("expected invalid email error, got %v", err)
	}
}

func TestPasswordHashAndComparison(t *testing.T) {
	password := "correct horse battery staple"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if hash == password || hash == "" {
		t.Fatal("expected encoded password hash")
	}
	if err := ComparePassword(password, hash); err != nil {
		t.Fatalf("compare password: %v", err)
	}
	if !errors.Is(ComparePassword("wrong password", hash), ErrPasswordMismatch) {
		t.Fatal("expected password mismatch")
	}
}

func TestPasswordValidation(t *testing.T) {
	for _, password := range []string{"", "short", strings.Repeat("a", 73)} {
		if _, err := HashPassword(password); err == nil {
			t.Fatalf("expected password %q to be rejected", password)
		}
	}
	if err := ComparePassword("valid password", "not-a-bcrypt-hash"); !errors.Is(err, ErrPasswordHashInvalid) {
		t.Fatalf("expected invalid hash error, got %v", err)
	}
}
