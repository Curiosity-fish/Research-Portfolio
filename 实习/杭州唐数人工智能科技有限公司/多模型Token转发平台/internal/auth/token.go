package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
)

const (
	// TokenType is the authorization scheme used in the Authorization header.
	TokenType = "Bearer"

	// Issuer identifies the token issuer.
	Issuer = "school-api-v1"

	// AudienceAdmin marks tokens issued to admin users (management console).
	AudienceAdmin = "admin"

	// AudienceUser marks tokens issued to regular users (user portal).
	AudienceUser = "user"
)

// Claims carries the custom and registered claims for an access token.
// AdminID/Role are only meaningful for AudienceAdmin tokens; UserID only for
// AudienceUser tokens.
type Claims struct {
	AdminID uuid.UUID `json:"admin_id,omitempty"`
	Role    string    `json:"role,omitempty"`
	UserID  uuid.UUID `json:"user_id,omitempty"`
	Status  string    `json:"status"`
	jwt.RegisteredClaims
}

// TokenManager issues and validates access tokens for admins and users.
type TokenManager interface {
	Generate(ctx context.Context, admin *ent.AdminUser) (string, error)
	GenerateUser(ctx context.Context, user *ent.User) (string, error)
	Parse(ctx context.Context, tokenStr string) (*Claims, error)
}

// jwtManager implements TokenManager with HMAC-SHA256 signed JWTs.
type jwtManager struct {
	secret []byte
	expiry time.Duration
}

// NewTokenManager creates a TokenManager that signs tokens with secret and
// expires them after expiry.
func NewTokenManager(secret string, expiry time.Duration) TokenManager {
	return &jwtManager{
		secret: []byte(secret),
		expiry: expiry,
	}
}

// Generate creates a new admin JWT.
func (m *jwtManager) Generate(ctx context.Context, admin *ent.AdminUser) (string, error) {
	now := time.Now().UTC()

	claims := &Claims{
		AdminID: admin.ID,
		Role:    string(admin.Role),
		Status:  string(admin.Status),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   admin.ID.String(),
			Issuer:    Issuer,
			Audience:  jwt.ClaimStrings{AudienceAdmin},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.expiry)),
		},
	}

	return m.sign(claims)
}

// GenerateUser creates a new user-portal JWT.
func (m *jwtManager) GenerateUser(ctx context.Context, user *ent.User) (string, error) {
	now := time.Now().UTC()

	claims := &Claims{
		UserID: user.ID,
		Status: string(user.Status),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			Issuer:    Issuer,
			Audience:  jwt.ClaimStrings{AudienceUser},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.expiry)),
		},
	}

	return m.sign(claims)
}

func (m *jwtManager) sign(claims *Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return tokenStr, nil
}

// Parse validates a token string and returns its claims. Callers must check
// the audience themselves (see middleware.JWTAuth / middleware.UserJWTAuth).
func (m *jwtManager) Parse(ctx context.Context, tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.secret, nil
	}, jwt.WithIssuer(Issuer))
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
