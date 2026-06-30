package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// tokenManager issues and validates JWT access tokens and generates opaque
// refresh tokens.
type tokenManager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func newTokenManager(secret string, accessTTL, refreshTTL time.Duration) *tokenManager {
	return &tokenManager{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

// accessClaims are the registered claims for an access token.
type accessClaims struct {
	jwt.RegisteredClaims
}

// IssueAccessToken signs a short-lived HS256 access token for the user.
func (t *tokenManager) IssueAccessToken(userID string) (string, time.Duration, error) {
	now := time.Now()
	claims := accessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(t.accessTTL)),
			ID:        uuid.NewString(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(t.secret)
	if err != nil {
		return "", 0, fmt.Errorf("sign access token: %w", err)
	}
	return signed, t.accessTTL, nil
}

// ParseAccessToken validates an access token and returns the subject (user id).
func (t *tokenManager) ParseAccessToken(raw string) (string, error) {
	parsed, err := jwt.ParseWithClaims(raw, &accessClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return t.secret, nil
	})
	if err != nil {
		return "", err
	}
	claims, ok := parsed.Claims.(*accessClaims)
	if !ok || !parsed.Valid {
		return "", errors.New("invalid token")
	}
	if claims.Subject == "" {
		return "", errors.New("token missing subject")
	}
	return claims.Subject, nil
}

// generateRefreshToken returns a new opaque refresh token and its SHA-256 hash.
func generateRefreshToken() (token, hash string, err error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}
	token = base64.RawURLEncoding.EncodeToString(buf)
	hash = hashToken(token)
	return token, hash, nil
}

// hashToken returns the hex SHA-256 of a refresh token for storage/lookup.
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
