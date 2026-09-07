package service

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/account"
	"github.com/school-api/school-api-v1/internal/encryption"
	"github.com/school-api/school-api-v1/internal/repository"
)

type mockAccountRepo struct {
	createFunc func(ctx context.Context, input repository.CreateAccountInput) (*ent.Account, error)
	getFunc    func(ctx context.Context, id uuid.UUID) (*ent.Account, error)
	listFunc   func(ctx context.Context) ([]*ent.Account, error)
	updateFunc func(ctx context.Context, id uuid.UUID, input repository.UpdateAccountInput) (*ent.Account, error)
	deleteFunc func(ctx context.Context, id uuid.UUID) error
}

func (m *mockAccountRepo) Create(ctx context.Context, input repository.CreateAccountInput) (*ent.Account, error) {
	return m.createFunc(ctx, input)
}

func (m *mockAccountRepo) GetByID(ctx context.Context, id uuid.UUID) (*ent.Account, error) {
	return m.getFunc(ctx, id)
}

func (m *mockAccountRepo) ListByPlatformID(ctx context.Context, platformID uuid.UUID) ([]*ent.Account, error) {
	return nil, nil
}

func (m *mockAccountRepo) List(ctx context.Context) ([]*ent.Account, error) {
	return m.listFunc(ctx)
}

func (m *mockAccountRepo) Update(ctx context.Context, id uuid.UUID, input repository.UpdateAccountInput) (*ent.Account, error) {
	return m.updateFunc(ctx, id, input)
}

func (m *mockAccountRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.deleteFunc(ctx, id)
}

func (m *mockAccountRepo) IncrementErrorCount(ctx context.Context, id uuid.UUID) error {
	return nil
}

func TestAccountService_CreateEncryptsKey(t *testing.T) {
	key := make([]byte, 32)
	platformID := uuid.New()

	repo := &mockAccountRepo{
		createFunc: func(ctx context.Context, input repository.CreateAccountInput) (*ent.Account, error) {
			// Verify the stored value is encrypted and can be decrypted.
			plaintext, err := encryption.Decrypt(key, input.APIKeyEncrypted)
			if err != nil {
				t.Errorf("decrypt stored key: %v", err)
			}
			if plaintext != "sk-test-key" {
				t.Errorf("expected decrypted key sk-test-key, got %s", plaintext)
			}
			return &ent.Account{
				ID:              uuid.New(),
				PlatformID:      input.PlatformID,
				Name:            input.Name,
				APIKeyEncrypted: input.APIKeyEncrypted,
				Weight:          input.Weight,
				MaxRpm:          input.MaxRPM,
				Status:          input.Status,
			}, nil
		},
	}

	platformRepo := &mockPlatformRepo{
		getFunc: func(ctx context.Context, id uuid.UUID) (*ent.Platform, error) {
			return &ent.Platform{ID: platformID}, nil
		},
	}

	svc := NewAccountService(repo, platformRepo, key)
	resp, err := svc.CreateAccount(context.Background(), CreateAccountInput{
		PlatformID: platformID,
		Name:       "Primary",
		APIKey:     "sk-test-key",
		Weight:     2,
		MaxRPM:     60,
		Status:     account.StatusActive,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if !resp.APIKeyEncrypted {
		t.Error("expected APIKeyEncrypted to be true")
	}
}

func TestAccountService_DefaultWeightAndRPM(t *testing.T) {
	key := make([]byte, 32)
	platformID := uuid.New()

	repo := &mockAccountRepo{
		createFunc: func(ctx context.Context, input repository.CreateAccountInput) (*ent.Account, error) {
			if input.Weight != 1 {
				t.Errorf("expected default weight 1, got %d", input.Weight)
			}
			if input.MaxRPM != 0 {
				t.Errorf("expected default max_rpm 0, got %d", input.MaxRPM)
			}
			return &ent.Account{ID: uuid.New(), PlatformID: input.PlatformID, Weight: input.Weight, MaxRpm: input.MaxRPM, Status: input.Status}, nil
		},
	}

	platformRepo := &mockPlatformRepo{
		getFunc: func(ctx context.Context, id uuid.UUID) (*ent.Platform, error) {
			return &ent.Platform{ID: platformID}, nil
		},
	}

	svc := NewAccountService(repo, platformRepo, key)
	_, err := svc.CreateAccount(context.Background(), CreateAccountInput{
		PlatformID: platformID,
		Name:       "Default",
		APIKey:     "sk-test",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
}
