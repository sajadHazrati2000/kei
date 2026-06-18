package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	accessTokenDuration  = 15 * time.Minute
	refreshTokenDuration = 7 * 24 * time.Hour
)

type accessClaims struct {
	UserID string `json:"user_id"`
	OrgID  string `json:"org_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type refreshClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func newJTI() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("auth.newJTI: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func generateAccessToken(secret, userID, orgID, role string) (string, error) {
	claims := accessClaims{
		UserID: userID,
		OrgID:  orgID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("auth.generateAccessToken: %w", err)
	}
	return signed, nil
}

// generateRefreshToken returns the signed JWT, its JTI (for DB storage), and its expiry.
func generateRefreshToken(secret, userID string) (tokenStr, jti string, expiresAt time.Time, err error) {
	jti, err = newJTI()
	if err != nil {
		return
	}
	expiresAt = time.Now().Add(refreshTokenDuration)
	claims := refreshClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err = token.SignedString([]byte(secret))
	if err != nil {
		err = fmt.Errorf("auth.generateRefreshToken: %w", err)
	}
	return
}

func validateRefreshToken(secret, tokenStr string) (*refreshClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &refreshClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("auth.validateRefreshToken: %w", err)
	}
	claims, ok := token.Claims.(*refreshClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("auth.validateRefreshToken: invalid token")
	}
	return claims, nil
}
