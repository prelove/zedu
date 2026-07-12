package auth_test

import (
	"strings"
	"testing"
	"time"

	"github.com/prelove/zedu/backend/internal/platform/auth"
)

func TestPasswordHashAndVerify(t *testing.T) {
	password := "MySecret123!"
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	if hash == "" {
		t.Fatalf("hash is empty")
	}
	if hash == password {
		t.Fatalf("hash must not equal plaintext password")
	}
	if strings.Contains(hash, password) {
		t.Fatalf("hash must not contain plaintext password")
	}

	if !auth.VerifyPassword(hash, password) {
		t.Fatalf("verify password should return true for correct password")
	}

	if auth.VerifyPassword(hash, "wrongpassword") {
		t.Fatalf("verify password should return false for wrong password")
	}
}

func TestPasswordHashDifferentEachTime(t *testing.T) {
	hash1, _ := auth.HashPassword("same123")
	hash2, _ := auth.HashPassword("same123")
	if hash1 == hash2 {
		t.Fatalf("bcrypt hashes must differ for same password (salt)")
	}
}

func TestJWTSignAndVerify(t *testing.T) {
	secret := "test-secret-key"
	userID := int64(42)
	role := "OWNER"

	token, err := auth.SignAccessToken(secret, userID, role, 60*time.Minute)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	if token == "" {
		t.Fatalf("token is empty")
	}

	claims, err := auth.VerifyAccessToken(secret, token)
	if err != nil {
		t.Fatalf("verify token: %v", err)
	}
	if claims.UserID != userID {
		t.Fatalf("expected user_id %d, got %d", userID, claims.UserID)
	}
	if claims.Role != role {
		t.Fatalf("expected role %s, got %s", role, claims.Role)
	}
}

func TestJWTExpiredTokenRejected(t *testing.T) {
	secret := "test-secret-key"
	token, err := auth.SignAccessToken(secret, 1, "OWNER", -1*time.Minute)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	_, err = auth.VerifyAccessToken(secret, token)
	if err == nil {
		t.Fatalf("expected error for expired token, got nil")
	}
}

func TestJWTWrongSecretRejected(t *testing.T) {
	token, err := auth.SignAccessToken("secret-a", 1, "OWNER", 60*time.Minute)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	_, err = auth.VerifyAccessToken("secret-b", token)
	if err == nil {
		t.Fatalf("expected error for wrong secret, got nil")
	}
}

func TestRefreshTokenGeneration(t *testing.T) {
	token1, err := auth.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	if token1 == "" {
		t.Fatalf("refresh token is empty")
	}

	token2, err := auth.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("generate second token: %v", err)
	}
	if token1 == token2 {
		t.Fatalf("two refresh tokens must differ")
	}
}

func TestHashRefreshToken(t *testing.T) {
	token := "some-refresh-token-value"
	hash := auth.HashRefreshToken(token)
	if hash == "" {
		t.Fatalf("hash is empty")
	}
	if hash == token {
		t.Fatalf("hash must not equal token")
	}

	// Same token produces same hash.
	hash2 := auth.HashRefreshToken(token)
	if hash != hash2 {
		t.Fatalf("same token must produce same hash")
	}

	// Different token produces different hash.
	hash3 := auth.HashRefreshToken("different-token")
	if hash == hash3 {
		t.Fatalf("different tokens must produce different hashes")
	}
}
