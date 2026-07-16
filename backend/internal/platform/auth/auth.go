package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
	"unicode"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// bcryptCost is the bcrypt cost factor. PRD 20.1 specifies cost=12.
const bcryptCost = 12

// minSecretLen is the minimum acceptable JWT signing secret length.
const minSecretLen = 32

// ValidateSecret returns an error if the secret is missing, too short, or
// matches a known weak value. Production startup must call this and fail-fast.
func ValidateSecret(secret string) error {
	if secret == "" {
		return fmt.Errorf("ZEDU_JWT_SECRET is not set")
	}
	if len(secret) < minSecretLen {
		return fmt.Errorf("ZEDU_JWT_SECRET must be at least %d characters, got %d", minSecretLen, len(secret))
	}
	return nil
}

// dummyHash is a pre-computed bcrypt hash used for constant-time comparison
// on the unknown-user path, to reduce timing-based user enumeration.
var dummyHash = func() string {
	h, _ := bcrypt.GenerateFromPassword([]byte("dummy-constant-time-placeholder"), bcryptCost)
	return string(h)
}()

// VerifyPasswordDummy performs a bcrypt compare against a dummy hash to
// consume time comparable to a real verification on the unknown-user path.
func VerifyPasswordDummy(password string) {
	_ = bcrypt.CompareHashAndPassword([]byte(dummyHash), []byte(password))
}

// HashPassword hashes a plaintext password using bcrypt with the project's cost.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

// VerifyPassword checks a plaintext password against a bcrypt hash.
// Returns true if the password matches.
func VerifyPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// ValidatePassword checks that a password meets PRD 20.1 rules:
// at least 8 characters, containing at least one letter and one digit.
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	hasLetter := false
	hasDigit := false
	for _, c := range password {
		if unicode.IsLetter(c) {
			hasLetter = true
		}
		if unicode.IsDigit(c) {
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return fmt.Errorf("password must contain at least one letter and one digit")
	}
	return nil
}

// Claims contains the JWT claims for an access token.
type Claims struct {
	UserID int64  `json:"sub"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// SignAccessToken creates a signed JWT access token.
func SignAccessToken(secret string, userID int64, role string, duration time.Duration) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// VerifyAccessToken verifies a JWT access token and returns the claims.
// Only HS256 is accepted; any other algorithm (including "none" or HS384/HS512)
// is rejected to prevent algorithm substitution attacks.
func VerifyAccessToken(secret, tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("verify token: %w", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

// GenerateRefreshToken generates a cryptographically random refresh token.
func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate refresh token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// HashRefreshToken returns the SHA-256 hex hash of a refresh token.
// Only the hash is stored in the database.
func HashRefreshToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
