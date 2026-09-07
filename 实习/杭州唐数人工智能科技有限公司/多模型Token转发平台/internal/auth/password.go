// Package auth provides password hashing and JWT utilities used by the admin
// authentication flow.
package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	// DefaultBcryptCost is used in development and test environments where
	// login throughput is not critical and fast test feedback is preferred.
	DefaultBcryptCost = 10

	// ProductionBcryptCost is used in production to increase the work factor
	// for offline brute-force resistance.
	ProductionBcryptCost = 12

	// MinBcryptCost is the lowest cost allowed. Values below this are raised
	// to prevent insecure configurations.
	MinBcryptCost = 10
)

// HashPassword returns a bcrypt hash of password using the supplied cost.
// If cost is below MinBcryptCost it is silently raised.
func HashPassword(cost int, password string) (string, error) {
	if cost < MinBcryptCost {
		cost = MinBcryptCost
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	return string(hash), nil
}

// ComparePassword checks a bcrypt hash against a plaintext password.
// It returns true only if the password matches the hash.
func ComparePassword(hashed, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password)) == nil
}

// BcryptCostForEnv returns the appropriate bcrypt cost for the given
// environment name.
func BcryptCostForEnv(env string) int {
	if env == "production" {
		return ProductionBcryptCost
	}
	return DefaultBcryptCost
}
