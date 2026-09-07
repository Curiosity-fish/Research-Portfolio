package repository

import (
	"context"
	"fmt"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/setting"
)

// UpsertSettingInput is the data required to create or update a setting.
type UpsertSettingInput struct {
	Key         string
	Value       string
	ValueType   string
	Description string
	IsPublic    bool
}

// SettingRepository provides data access for key-value settings.
type SettingRepository interface {
	GetByKey(ctx context.Context, key string) (*ent.Setting, error)
	Upsert(ctx context.Context, input UpsertSettingInput) (*ent.Setting, error)
}

// EntSettingRepository implements SettingRepository using Ent.
type EntSettingRepository struct {
	client *ent.Client
}

// NewEntSettingRepository creates a new Ent-backed setting repository.
func NewEntSettingRepository(client *ent.Client) *EntSettingRepository {
	return &EntSettingRepository{client: client}
}

// GetByKey returns a setting by its unique key.
func (r *EntSettingRepository) GetByKey(ctx context.Context, key string) (*ent.Setting, error) {
	s, err := r.client.Setting.Query().Where(setting.KeyEQ(key)).Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get setting: %w", err)
	}
	return s, nil
}

// Upsert updates the setting with the given key or creates it. A concurrent
// insert that wins the race surfaces as a unique-constraint error; in that
// case the update is retried against the row created by the other writer.
func (r *EntSettingRepository) Upsert(ctx context.Context, input UpsertSettingInput) (*ent.Setting, error) {
	existing, err := r.client.Setting.Query().Where(setting.KeyEQ(input.Key)).Only(ctx)
	if err != nil {
		if !ent.IsNotFound(err) {
			return nil, fmt.Errorf("get setting: %w", err)
		}
		created, createErr := r.client.Setting.Create().
			SetKey(input.Key).
			SetValue(input.Value).
			SetType(setting.Type(input.ValueType)).
			SetDescription(input.Description).
			SetIsPublic(input.IsPublic).
			Save(ctx)
		if createErr == nil {
			return created, nil
		}
		if !ent.IsConstraintError(createErr) {
			return nil, fmt.Errorf("create setting: %w", createErr)
		}
		existing, err = r.client.Setting.Query().Where(setting.KeyEQ(input.Key)).Only(ctx)
		if err != nil {
			return nil, fmt.Errorf("get setting after constraint conflict: %w", err)
		}
	}

	updated, err := existing.Update().
		SetValue(input.Value).
		SetType(setting.Type(input.ValueType)).
		SetDescription(input.Description).
		SetIsPublic(input.IsPublic).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update setting: %w", err)
	}
	return updated, nil
}

var _ SettingRepository = (*EntSettingRepository)(nil)
