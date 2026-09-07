package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// APIKeyPrefix is the required prefix for external API keys.
	APIKeyPrefix = "sk-"

	// apiKeyRandomBytes is the number of random bytes after the prefix.
	apiKeyRandomBytes = 32

	apiKeyUserIDKey    = "school-api-v1/user_id"
	apiKeyTokenIDKey   = "school-api-v1/token_id"
	apiKeyGroupIDKey   = "school-api-v1/group_id"
	apiKeyGroupCodeKey = "school-api-v1/group_code"
)

// APIKeyInfo holds the authenticated identity for an API key request.
type APIKeyInfo struct {
	UserID    uuid.UUID
	TokenID   uuid.UUID
	GroupID   uuid.UUID
	GroupCode string
}

// APIKeyLookupResult is the raw data returned by the persistence layer for a
// token hash lookup.
type APIKeyLookupResult struct {
	UserID     uuid.UUID
	TokenID    uuid.UUID
	IsEnabled  bool
	ExpiresAt  *time.Time
	UserStatus string
	GroupID    uuid.UUID
	GroupCode  string
}

// APIKeyLookup finds a token by its hash and returns the data needed to
// validate it and build an APIKeyInfo.
type APIKeyLookup interface {
	GetByTokenHash(ctx context.Context, hash string) (*APIKeyLookupResult, error)
}

// APIKeyValidator validates a plaintext API key.
type APIKeyValidator interface {
	Validate(ctx context.Context, key string) (*APIKeyInfo, error)
}

// apiKeyValidator implements APIKeyValidator.
type apiKeyValidator struct {
	lookup APIKeyLookup
}

// NewAPIKeyValidator creates a new APIKeyValidator.
func NewAPIKeyValidator(lookup APIKeyLookup) APIKeyValidator {
	return &apiKeyValidator{lookup: lookup}
}

// Validate checks the key and returns the authenticated identity.
func (v *apiKeyValidator) Validate(ctx context.Context, key string) (*APIKeyInfo, error) {
	result, err := v.lookup.GetByTokenHash(ctx, HashAPIKey(key))
	if err != nil {
		return nil, err
	}

	if !result.IsEnabled {
		return nil, errors.New("token disabled")
	}
	if result.ExpiresAt != nil && result.ExpiresAt.Before(time.Now().UTC()) {
		return nil, errors.New("token expired")
	}
	if result.UserStatus != "active" {
		return nil, errors.New("user inactive")
	}

	return &APIKeyInfo{
		UserID:    result.UserID,
		TokenID:   result.TokenID,
		GroupID:   result.GroupID,
		GroupCode: result.GroupCode,
	}, nil
}

// HashAPIKey returns the stored hash for a plaintext API key.
func HashAPIKey(key string) string {
	sum := sha256.Sum256([]byte(key))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// GenerateAPIKey creates a new random API key and its stored hash.
func GenerateAPIKey() (plaintext, hash string, err error) {
	raw := make([]byte, apiKeyRandomBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", "", fmt.Errorf("generate random bytes: %w", err)
	}
	plaintext = APIKeyPrefix + base64.RawURLEncoding.EncodeToString(raw)
	return plaintext, HashAPIKey(plaintext), nil
}

// SetAPIKeyContext stores API key identity information in the gin context.
func SetAPIKeyContext(c *gin.Context, info *APIKeyInfo) {
	c.Set(apiKeyUserIDKey, info.UserID)
	c.Set(apiKeyTokenIDKey, info.TokenID)
	c.Set(apiKeyGroupIDKey, info.GroupID)
	c.Set(apiKeyGroupCodeKey, info.GroupCode)
}

// SetUserIDContext stores a user ID in the gin context under the same key
// used by API key auth, so downstream handlers can use UserID regardless of
// which authentication scheme was used.
func SetUserIDContext(c *gin.Context, userID uuid.UUID) {
	c.Set(apiKeyUserIDKey, userID)
}

// UserID returns the user ID stored in the context, if any.
func UserID(c *gin.Context) (uuid.UUID, bool) {
	v, ok := c.Get(apiKeyUserIDKey)
	if !ok {
		return uuid.UUID{}, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}

// TokenID returns the token ID stored in the context, if any.
func TokenID(c *gin.Context) (uuid.UUID, bool) {
	v, ok := c.Get(apiKeyTokenIDKey)
	if !ok {
		return uuid.UUID{}, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}

// GroupID returns the group ID stored in the context, if any.
func GroupID(c *gin.Context) (uuid.UUID, bool) {
	v, ok := c.Get(apiKeyGroupIDKey)
	if !ok {
		return uuid.UUID{}, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}

// GroupCode returns the group code stored in the context, if any.
func GroupCode(c *gin.Context) (string, bool) {
	v, ok := c.Get(apiKeyGroupCodeKey)
	if !ok {
		return "", false
	}
	code, ok := v.(string)
	return code, ok
}
