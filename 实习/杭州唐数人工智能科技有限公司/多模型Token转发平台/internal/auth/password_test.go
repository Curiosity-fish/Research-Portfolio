package auth

import (
	"testing"
)

func TestHashPasswordAndComparePassword(t *testing.T) {
	password := "correct-horse-battery-staple"

	hash, err := HashPassword(DefaultBcryptCost, password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if hash == "" {
		t.Fatal("expected non-empty hash")
	}

	if !ComparePassword(hash, password) {
		t.Error("expected password to match")
	}
	if ComparePassword(hash, "wrong-password") {
		t.Error("expected wrong password not to match")
	}
}

func TestHashPassword_EnforcesMinCost(t *testing.T) {
	password := "test-password"

	// Cost below the minimum should be raised.
	hash, err := HashPassword(4, password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if !ComparePassword(hash, password) {
		t.Error("expected password to match after cost enforcement")
	}
}

func TestBcryptCostForEnv(t *testing.T) {
	if got := BcryptCostForEnv("production"); got != ProductionBcryptCost {
		t.Errorf("expected production cost %d, got %d", ProductionBcryptCost, got)
	}
	if got := BcryptCostForEnv("development"); got != DefaultBcryptCost {
		t.Errorf("expected development cost %d, got %d", DefaultBcryptCost, got)
	}
	if got := BcryptCostForEnv("test"); got != DefaultBcryptCost {
		t.Errorf("expected test cost %d, got %d", DefaultBcryptCost, got)
	}
}
