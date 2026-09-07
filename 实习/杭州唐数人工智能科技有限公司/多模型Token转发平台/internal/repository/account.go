package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/account"
)

// CreateAccountInput is the data required to create an account.
type CreateAccountInput struct {
	PlatformID      uuid.UUID
	Name            string
	APIKeyEncrypted string
	Weight          int
	MaxRPM          int
	Status          account.Status
}

// UpdateAccountInput is the data that can be updated for an account.
// Pointer fields that are nil are left unchanged.
type UpdateAccountInput struct {
	Name            *string
	APIKeyEncrypted *string
	Weight          *int
	MaxRPM          *int
	Status          *account.Status
	ErrorCount      *int
}

// AccountRepository provides data access for accounts.
type AccountRepository interface {
	Create(ctx context.Context, input CreateAccountInput) (*ent.Account, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ent.Account, error)
	ListByPlatformID(ctx context.Context, platformID uuid.UUID) ([]*ent.Account, error)
	List(ctx context.Context) ([]*ent.Account, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateAccountInput) (*ent.Account, error)
	Delete(ctx context.Context, id uuid.UUID) error
	IncrementErrorCount(ctx context.Context, id uuid.UUID) error
}

// EntAccountRepository implements AccountRepository using Ent.
type EntAccountRepository struct {
	client *ent.Client
}

// NewEntAccountRepository creates a new Ent-backed account repository.
func NewEntAccountRepository(client *ent.Client) *EntAccountRepository {
	return &EntAccountRepository{client: client}
}

// Create inserts a new account.
func (r *EntAccountRepository) Create(ctx context.Context, input CreateAccountInput) (*ent.Account, error) {
	a, err := r.client.Account.Create().
		SetPlatformID(input.PlatformID).
		SetName(input.Name).
		SetAPIKeyEncrypted(input.APIKeyEncrypted).
		SetWeight(input.Weight).
		SetMaxRpm(input.MaxRPM).
		SetStatus(input.Status).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create account: %w", err)
	}
	return a, nil
}

// GetByID returns an account by ID.
func (r *EntAccountRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.Account, error) {
	a, err := r.client.Account.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get account: %w", err)
	}
	return a, nil
}

// ListByPlatformID returns accounts for a specific platform.
func (r *EntAccountRepository) ListByPlatformID(ctx context.Context, platformID uuid.UUID) ([]*ent.Account, error) {
	accounts, err := r.client.Account.Query().
		Where(account.PlatformIDEQ(platformID)).
		Order(ent.Asc(account.FieldName)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list accounts by platform: %w", err)
	}
	return accounts, nil
}

// List returns all accounts ordered by name.
func (r *EntAccountRepository) List(ctx context.Context) ([]*ent.Account, error) {
	accounts, err := r.client.Account.Query().
		Order(ent.Asc(account.FieldName)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}
	return accounts, nil
}

// Update modifies an account.
func (r *EntAccountRepository) Update(ctx context.Context, id uuid.UUID, input UpdateAccountInput) (*ent.Account, error) {
	b := r.client.Account.UpdateOneID(id)
	if input.Name != nil {
		b.SetName(*input.Name)
	}
	if input.APIKeyEncrypted != nil {
		b.SetAPIKeyEncrypted(*input.APIKeyEncrypted)
	}
	if input.Weight != nil {
		b.SetWeight(*input.Weight)
	}
	if input.MaxRPM != nil {
		b.SetMaxRpm(*input.MaxRPM)
	}
	if input.Status != nil {
		b.SetStatus(*input.Status)
	}
	if input.ErrorCount != nil {
		b.SetErrorCount(*input.ErrorCount)
	}

	a, err := b.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update account: %w", err)
	}
	return a, nil
}

// Delete removes an account.
func (r *EntAccountRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.client.Account.DeleteOneID(id).Exec(ctx); err != nil {
		return fmt.Errorf("delete account: %w", err)
	}
	return nil
}

// IncrementErrorCount atomically increases the account's consecutive failure counter by one.
func (r *EntAccountRepository) IncrementErrorCount(ctx context.Context, id uuid.UUID) error {
	if err := r.client.Account.UpdateOneID(id).AddErrorCount(1).Exec(ctx); err != nil {
		return fmt.Errorf("increment account error count: %w", err)
	}
	return nil
}
