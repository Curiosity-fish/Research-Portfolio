package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/adminuser"
)

// AdminRepository provides data access for admin_users.
type AdminRepository interface {
	GetByUsername(ctx context.Context, username string) (*ent.AdminUser, error)
	UpdateLastLoginAt(ctx context.Context, id uuid.UUID) error
	CreateAdmin(ctx context.Context, username, passwordHash string, role adminuser.Role) (*ent.AdminUser, error)
}

// EntAdminRepository implements AdminRepository using the generated Ent client.
type EntAdminRepository struct {
	client *ent.Client
}

// NewEntAdminRepository creates a new Ent-backed admin repository.
func NewEntAdminRepository(client *ent.Client) *EntAdminRepository {
	return &EntAdminRepository{client: client}
}

// GetByUsername looks up an admin user by username.
func (r *EntAdminRepository) GetByUsername(ctx context.Context, username string) (*ent.AdminUser, error) {
	admin, err := r.client.AdminUser.Query().
		Where(adminuser.Username(username)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("query admin by username: %w", err)
	}
	return admin, nil
}

// UpdateLastLoginAt sets the admin's last_login_at to the current UTC time.
func (r *EntAdminRepository) UpdateLastLoginAt(ctx context.Context, id uuid.UUID) error {
	if err := r.client.AdminUser.UpdateOneID(id).
		SetLastLoginAt(time.Now().UTC()).
		Exec(ctx); err != nil {
		return fmt.Errorf("update last login: %w", err)
	}
	return nil
}

// CreateAdmin creates a new admin user with the given credentials and role.
func (r *EntAdminRepository) CreateAdmin(ctx context.Context, username, passwordHash string, role adminuser.Role) (*ent.AdminUser, error) {
	admin, err := r.client.AdminUser.Create().
		SetUsername(username).
		SetPasswordHash(passwordHash).
		SetRole(role).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create admin: %w", err)
	}
	return admin, nil
}
