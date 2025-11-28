package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// UserClaims represents user-specific JWT claims
type UserClaims struct {
	UserID    uint      `json:"user_id"`
	Username  string    `json:"username"`
	Email     string    `json:"email,omitempty"`
	Role      string    `json:"role,omitempty"`
	TokenType string    `json:"token_type"` // "access" or "refresh"
	DeviceID  string    `json:"device_id"`
	jwt.RegisteredClaims
}

// AuthToken signs and verifies JWT tokens with user claims support.
type AuthToken struct {
	secretKey []byte
	ttl       time.Duration
}

// NewAuthToken builds a token helper using the provided secret.
func NewAuthToken(secretKey string) *AuthToken {
	token := &AuthToken{
		secretKey: []byte(secretKey),
		ttl:       time.Hour,
	}
	if secretKey == "" {
		fmt.Println("auth token secret key cannot be empty")
	}
	return token
}

// WithTTL allows customising the expiration duration.
func (at *AuthToken) WithTTL(ttl time.Duration) *AuthToken {
	if ttl > 0 {
		at.ttl = ttl
	}
	return at
}

// GenerateToken issues a JWT for the provided device identifier (legacy support).
func (at *AuthToken) GenerateToken(deviceID string) (string, error) {
	if at == nil {
		return "", errors.New("auth token is nil")
	}
	if len(at.secretKey) == 0 {
		return "", errors.New("auth token secret is empty")
	}

	expireTime := time.Now().Add(at.ttl)
	claims := jwt.MapClaims{
		"device_id": deviceID,
		"exp":       expireTime.Unix(),
		"iat":       time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(at.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return tokenString, nil
}

// GenerateUserToken issues a JWT for a user with claims.
func (at *AuthToken) GenerateUserToken(userID uint, username, email, role, deviceID string) (string, error) {
	if at == nil {
		return "", errors.New("auth token is nil")
	}
	if len(at.secretKey) == 0 {
		return "", errors.New("auth token secret is empty")
	}

	now := time.Now()
	expireTime := now.Add(at.ttl)
	claims := UserClaims{
		UserID:    userID,
		Username:  username,
		Email:     email,
		Role:      role,
		TokenType: "access",
		DeviceID:  deviceID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "xiaozhi-flow",
			Subject:   fmt.Sprintf("user:%d", userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(at.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign user token: %w", err)
	}
	return tokenString, nil
}

// GenerateRefreshToken issues a refresh token for a user.
func (at *AuthToken) GenerateRefreshToken(userID uint, username string, deviceID string) (string, error) {
	if at == nil {
		return "", errors.New("auth token is nil")
	}
	if len(at.secretKey) == 0 {
		return "", errors.New("auth token secret is empty")
	}

	// Refresh tokens have longer TTL (30 days by default)
	refreshTTL := at.ttl * 30 // If access token is 7 days, refresh is 210 days (adjust as needed)
	if refreshTTL < 30*24*time.Hour {
		refreshTTL = 30 * 24 * time.Hour // Minimum 30 days
	}

	now := time.Now()
	expireTime := now.Add(refreshTTL)
	claims := UserClaims{
		UserID:    userID,
		Username:  username,
		TokenType: "refresh",
		DeviceID:  deviceID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "xiaozhi-flow",
			Subject:   fmt.Sprintf("user:%d:refresh", userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(at.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}
	return tokenString, nil
}

// VerifyToken validates the JWT and extracts the device identifier (legacy support).
func (at *AuthToken) VerifyToken(tokenString string) (bool, string, error) {
	if at == nil {
		return false, "", errors.New("auth token is nil")
	}
	if len(at.secretKey) == 0 {
		return false, "", errors.New("auth token secret is empty")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return at.secretKey, nil
	})
	if err != nil {
		return false, "", fmt.Errorf("failed to parse token: %w", err)
	}
	if !token.Valid {
		return false, "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false, "", errors.New("invalid claims")
	}

	// Check if it's a legacy token (device_id only) or new user token
	if deviceID, ok := claims["device_id"].(string); ok {
		return true, deviceID, nil
	}

	// For user tokens, extract device ID from claims
	if deviceID, ok := claims["device_id"].(string); ok {
		return true, deviceID, nil
	}

	return false, "", errors.New("device_id claim not found")
}

// VerifyUserToken validates a user JWT and returns the claims.
func (at *AuthToken) VerifyUserToken(tokenString string) (*UserClaims, error) {
	if at == nil {
		return nil, errors.New("auth token is nil")
	}
	if len(at.secretKey) == 0 {
		return nil, errors.New("auth token secret is empty")
	}

	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return at.secretKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse user token: %w", err)
	}
	if !token.Valid {
		return nil, errors.New("invalid user token")
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return nil, errors.New("invalid user claims")
	}

	return claims, nil
}

// IsAccessToken checks if the token is an access token.
func (claims *UserClaims) IsAccessToken() bool {
	return claims.TokenType == "access"
}

// IsRefreshToken checks if the token is a refresh token.
func (claims *UserClaims) IsRefreshToken() bool {
	return claims.TokenType == "refresh"
}

// IsExpired checks if the token is expired.
func (claims *UserClaims) IsExpired() bool {
	if claims.ExpiresAt == nil {
		return false
	}
	return time.Now().After(claims.ExpiresAt.Time)
}
