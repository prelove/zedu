package auth_test

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

// ==================== P0.1: ValidateSecret ====================

func TestValidateSecretRejectsEmpty(t *testing.T) {
	if err := auth.ValidateSecret(""); err == nil {
		t.Fatalf("expected error for empty secret")
	}
}

func TestValidateSecretRejectsShort(t *testing.T) {
	if err := auth.ValidateSecret("short"); err == nil {
		t.Fatalf("expected error for short secret")
	}
}

func TestValidateSecretAcceptsLongEnough(t *testing.T) {
	if err := auth.ValidateSecret("this-is-a-secure-secret-at-least-32-chars"); err != nil {
		t.Fatalf("expected no error for valid secret, got %v", err)
	}
}

// ==================== P1.4: ValidatePassword ====================

func TestValidatePasswordRejectsShort(t *testing.T) {
	if err := auth.ValidatePassword("Ab1"); err == nil {
		t.Fatalf("expected error for short password")
	}
}

func TestValidatePasswordRejectsNoLetter(t *testing.T) {
	if err := auth.ValidatePassword("12345678"); err == nil {
		t.Fatalf("expected error for password with no letter")
	}
}

func TestValidatePasswordRejectsNoDigit(t *testing.T) {
	if err := auth.ValidatePassword("abcdefgh"); err == nil {
		t.Fatalf("expected error for password with no digit")
	}
}

func TestValidatePasswordAcceptsValid(t *testing.T) {
	if err := auth.ValidatePassword("Pass1234"); err != nil {
		t.Fatalf("expected no error for valid password, got %v", err)
	}
}

// ==================== P2.3: Strict HS256 ====================

func TestJWTRejectsHS384(t *testing.T) {
	secret := "test-secret-key-at-least-32-chars"
	// Sign with HS384.
	token := jwt.NewWithClaims(jwt.SigningMethodHS384, auth.Claims{
		UserID: 1,
		Role:   "OWNER",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(60 * time.Minute)),
		},
	})
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign HS384: %v", err)
	}

	_, err = auth.VerifyAccessToken(secret, tokenStr)
	if err == nil {
		t.Fatalf("expected error for HS384 token, got nil")
	}
}

func TestJWTRejectsHS512(t *testing.T) {
	secret := "test-secret-key-at-least-32-chars"
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, auth.Claims{
		UserID: 1,
		Role:   "OWNER",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(60 * time.Minute)),
		},
	})
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign HS512: %v", err)
	}

	_, err = auth.VerifyAccessToken(secret, tokenStr)
	if err == nil {
		t.Fatalf("expected error for HS512 token, got nil")
	}
}

// ==================== P2.3: VerifyPasswordDummy ====================

func TestVerifyPasswordDummyDoesNotPanic(t *testing.T) {
	// Just ensure it runs without panic.
	auth.VerifyPasswordDummy("somepassword")
}
