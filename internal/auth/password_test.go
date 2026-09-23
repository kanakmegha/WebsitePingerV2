package auth

import (
	"strings"
	"testing"
)

func TestHashPassword_Success(t *testing.T) {
	rawPassword := "SecureP@ssw0rd2026"

	hashStr, err := HashPassword(rawPassword)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !strings.HasPrefix(hashStr, "argon2id$1$65536$4$") {
		t.Errorf("unexpected hash prefix format: %s", hashStr)
	}

	parts := strings.Split(hashStr, "$")
	if len(parts) != 6 {
		t.Fatalf("expected 6 parts in encoded hash string, got %d", len(parts))
	}
}

func TestVerifyPassword_Success(t *testing.T) {
	rawPassword := "SecureP@ssw0rd2026"

	hashStr, err := HashPassword(rawPassword)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	match, err := VerifyPassword(hashStr, rawPassword)
	if err != nil {
		t.Fatalf("expected no error during verification, got: %v", err)
	}
	if !match {
		t.Error("expected password to match successfully, but got match=false")
	}
}

func TestVerifyPassword_WrongPassword(t *testing.T) {
	rawPassword := "SecureP@ssw0rd2026"
	wrongPassword := "WrongP@ssw0rd2026"

	hashStr, err := HashPassword(rawPassword)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	match, err := VerifyPassword(hashStr, wrongPassword)
	if err != nil {
		t.Fatalf("expected no error for wrong password verification, got: %v", err)
	}
	if match {
		t.Error("expected match=false for incorrect password, got match=true")
	}
}

func TestVerifyPassword_CorruptedHash(t *testing.T) {
	corruptedHashes := []string{
		"invalid$format",
		"bcrypt$10$somerandomhash",
		"argon2id$1$65536",
		"argon2id$invalid_time$65536$4$salt$hash",
		"argon2id$1$65536$4$invalid_b64!@#$hash",
	}

	for _, invalidHash := range corruptedHashes {
		match, err := VerifyPassword(invalidHash, "password123")
		if err == nil {
			t.Errorf("expected error for corrupted hash '%s', got nil error", invalidHash)
		}
		if match {
			t.Errorf("expected match=false for corrupted hash '%s', got match=true", invalidHash)
		}
	}
}

func TestHashPassword_PasswordTooShort(t *testing.T) {
	shortPassword := "short" // 5 characters

	_, err := HashPassword(shortPassword)
	if err == nil {
		t.Fatal("expected error for short password (< 6 chars), got nil")
	}

	if err != ErrPasswordTooShort {
		t.Errorf("expected ErrPasswordTooShort, got: %v", err)
	}
}
