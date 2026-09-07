package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/account"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/encryption"
	"github.com/school-api/school-api-v1/internal/repository"
)

// CreateAccountInput is the service-level input for creating an account.
type CreateAccountInput struct {
	PlatformID uuid.UUID
	Name       string
	APIKey     string
	Weight     int
	MaxRPM     int
	Status     account.Status
}

// UpdateAccountInput is the service-level input for updating an account.
type UpdateAccountInput struct {
	Name    *string
	APIKey  *string
	Weight  *int
	MaxRPM  *int
	Status  *account.Status
}

// AccountResponse is the public representation of an account.
// The encrypted API key is never returned.
type AccountResponse struct {
	ID              string `json:"id"`
	PlatformID      string `json:"platform_id"`
	Name            string `json:"name"`
	APIKeyEncrypted bool   `json:"api_key_encrypted"`
	Weight          int    `json:"weight"`
	MaxRPM          int    `json:"max_rpm"`
	Status          string `json:"status"`
	ErrorCount      int    `json:"error_count"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

// AccountService defines account management business logic.
type AccountService interface {
	CreateAccount(ctx context.Context, input CreateAccountInput) (*AccountResponse, error)
	ListAccounts(ctx context.Context) ([]AccountResponse, error)
	ListAccountsByPlatform(ctx context.Context, platformID uuid.UUID) ([]AccountResponse, error)
	GetAccount(ctx context.Context, id uuid.UUID) (*AccountResponse, error)
	UpdateAccount(ctx context.Context, id uuid.UUID, input UpdateAccountInput) (*AccountResponse, error)
	DeleteAccount(ctx context.Context, id uuid.UUID) error
	DecryptAPIKey(ctx context.Context, accountID uuid.UUID) (string, error)
}

type accountService struct {
	repo         repository.AccountRepository
	platformRepo repository.PlatformRepository
	encryptKey   []byte
}

// NewAccountService creates a new AccountService.
func NewAccountService(repo repository.AccountRepository, platformRepo repository.PlatformRepository, encryptKey []byte) AccountService {
	return &accountService{
		repo:         repo,
		platformRepo: platformRepo,
		encryptKey:   encryptKey,
	}
}

// CreateAccount creates a new account with an encrypted API key.
func (s *accountService) CreateAccount(ctx context.Context, input CreateAccountInput) (*AccountResponse, error) {
	if _, err := s.platformRepo.GetByID(ctx, input.PlatformID); err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrNotFound
		}
		return nil, domain.WrapInternal(err)
	}

	if input.Status == "" {
		input.Status = account.StatusActive
	}
	if input.Weight <= 0 {
		input.Weight = 1
	}
	if input.MaxRPM < 0 {
		input.MaxRPM = 0
	}

	encrypted, err := encryption.Encrypt(s.encryptKey, input.APIKey)
	if err != nil {
		return nil, domain.WrapInternal(fmt.Errorf("encrypt api key: %w", err))
	}

	a, err := s.repo.Create(ctx, repository.CreateAccountInput{
		PlatformID:      input.PlatformID,
		Name:            input.Name,
		APIKeyEncrypted: encrypted,
		Weight:          input.Weight,
		MaxRPM:          input.MaxRPM,
		Status:          input.Status,
	})
	if err != nil {
		return nil, domain.WrapInternal(err)
	}
	return toAccountResponse(a), nil
}

// ListAccounts returns all accounts.
func (s *accountService)ListAccounts(ctx context.Context) ([]AccountResponse, error) {
	accounts, err := s.repo.List(ctx)
	if err != nil {
		return nil, domain.WrapInternal(err)
	}

	resp := make([]AccountResponse, 0, len(accounts))
	for _, a := range accounts {
		resp = append(resp, *toAccountResponse(a))
	}
	return resp, nil
}

// ListAccountsByPlatform returns accounts for a specific platform.
func (s *accountService)ListAccountsByPlatform(ctx context.Context, platformID uuid.UUID) ([]AccountResponse, error) {
	accounts, err := s.repo.ListByPlatformID(ctx, platformID)
	if err != nil {
		return nil, domain.WrapInternal(err)
	}

	resp := make([]AccountResponse, 0, len(accounts))
	for _, a := range accounts {
		resp = append(resp, *toAccountResponse(a))
	}
	return resp, nil
}

// GetAccount returns an account by ID.
func (s *accountService)GetAccount(ctx context.Context, id uuid.UUID) (*AccountResponse, error) {
	a, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrNotFound
		}
		return nil, domain.WrapInternal(err)
	}
	return toAccountResponse(a), nil
}

// UpdateAccount updates an existing account.
func (s *accountService)UpdateAccount(ctx context.Context, id uuid.UUID, input UpdateAccountInput) (*AccountResponse, error) {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrNotFound
		}
		return nil, domain.WrapInternal(err)
	}

	repoInput := repository.UpdateAccountInput{
		Name:   input.Name,
		Weight: input.Weight,
		MaxRPM: input.MaxRPM,
		Status: input.Status,
	}

	if input.APIKey != nil {
		encrypted, err := encryption.Encrypt(s.encryptKey, *input.APIKey)
		if err != nil {
			return nil, domain.WrapInternal(fmt.Errorf("encrypt api key: %w", err))
		}
		repoInput.APIKeyEncrypted = &encrypted
	}

	a, err := s.repo.Update(ctx, id, repoInput)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrNotFound
		}
		return nil, domain.WrapInternal(err)
	}
	return toAccountResponse(a), nil
}

// DeleteAccount removes an account by ID.
func (s *accountService)DeleteAccount(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		if ent.IsNotFound(err) {
			return domain.ErrNotFound
		}
		return domain.WrapInternal(err)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return domain.WrapInternal(err)
	}
	return nil
}

// DecryptAPIKey decrypts the stored API key for a given account.
func (s *accountService)DecryptAPIKey(ctx context.Context, accountID uuid.UUID) (string, error) {
	a, err := s.repo.GetByID(ctx, accountID)
	if err != nil {
		if ent.IsNotFound(err) {
			return "", domain.ErrNotFound
		}
		return "", domain.WrapInternal(err)
	}
	plaintext, err := encryption.Decrypt(s.encryptKey, a.APIKeyEncrypted)
	if err != nil {
		return "", domain.WrapInternal(fmt.Errorf("decrypt api key: %w", err))
	}
	return plaintext, nil
}

func toAccountResponse(a *ent.Account) *AccountResponse {
	return &AccountResponse{
		ID:              a.ID.String(),
		PlatformID:      a.PlatformID.String(),
		Name:            a.Name,
		APIKeyEncrypted: a.APIKeyEncrypted != "",
		Weight:          a.Weight,
		MaxRPM:          a.MaxRpm,
		Status:          a.Status.String(),
		ErrorCount:      a.ErrorCount,
		CreatedAt:       a.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       a.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
