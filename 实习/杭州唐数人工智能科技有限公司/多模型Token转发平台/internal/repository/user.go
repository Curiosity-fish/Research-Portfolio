package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/predicate"
	"github.com/school-api/school-api-v1/ent/user"
)

// CreateUserInput is the data required to create a user.
type CreateUserInput struct {
	Username     string
	PasswordHash string
	Name         string
	Email        string
	Phone        string
	Role         user.Role
	Gender       string
	Status       user.Status
	DepartmentID *uuid.UUID
	GroupID      *uuid.UUID
}

// UpdateUserInput is the data that can be updated for an existing user.
// Pointer fields that are nil are left unchanged; string fields that are empty
// and pointers to empty strings are cleared where the schema allows null.
type UpdateUserInput struct {
	Name         *string
	PasswordHash *string
	Email        *string
	Phone        *string
	Role         *user.Role
	Gender       *string
	DepartmentID *uuid.UUID
	GroupID      *uuid.UUID
}

// ListUserFilter controls the query for the user list endpoint.
type ListUserFilter struct {
	Keyword      string
	Status       *user.Status
	Role         *user.Role
	DepartmentID *uuid.UUID
	Page         int
	PageSize     int
}

// UserRepository provides data access for users.
type UserRepository interface {
	Create(ctx context.Context, input CreateUserInput) (*ent.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ent.User, error)
	GetByUsername(ctx context.Context, username string) (*ent.User, error)
	List(ctx context.Context, filter ListUserFilter) ([]*ent.User, int, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateUserInput) (*ent.User, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status user.Status) (*ent.User, error)
}

// EntUserRepository implements UserRepository using the generated Ent client.
type EntUserRepository struct {
	client *ent.Client
}

// NewEntUserRepository creates a new Ent-backed user repository.
func NewEntUserRepository(client *ent.Client) *EntUserRepository {
	return &EntUserRepository{client: client}
}

// Create inserts a new user.
func (r *EntUserRepository) Create(ctx context.Context, input CreateUserInput) (*ent.User, error) {
	b := r.client.User.Create().
		SetUsername(input.Username).
		SetPasswordHash(input.PasswordHash).
		SetName(input.Name).
		SetRole(input.Role).
		SetStatus(input.Status)

	if input.Email != "" {
		b.SetEmail(input.Email)
	}
	if input.Phone != "" {
		b.SetPhone(input.Phone)
	}
	if input.Gender != "" {
		b.SetGender(input.Gender)
	}
	if input.DepartmentID != nil {
		b.SetDepartmentID(*input.DepartmentID)
	}
	if input.GroupID != nil {
		b.SetGroupID(*input.GroupID)
	}

	u, err := b.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}

// GetByID looks up a user by ID.
func (r *EntUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.User, error) {
	u, err := r.client.User.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return u, nil
}

// GetByUsername looks up a user by username.
func (r *EntUserRepository) GetByUsername(ctx context.Context, username string) (*ent.User, error) {
	u, err := r.client.User.Query().
		Where(user.UsernameEQ(username)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	return u, nil
}

// List returns a paginated list of users matching the filter and the total count.
func (r *EntUserRepository) List(ctx context.Context, filter ListUserFilter) ([]*ent.User, int, error) {
	preds := r.buildPredicates(filter)

	page, pageSize := normalizePagination(filter.Page, filter.PageSize)

	users, err := r.client.User.Query().
		Where(preds...).
		// Tokens are eager-loaded so the service can aggregate quota fields
		// per user without a second query per row (class-view requirement).
		WithTokens().
		Order(ent.Desc(user.FieldCreatedAt)).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}

	total, err := r.client.User.Query().Where(preds...).Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	return users, total, nil
}

// Update modifies the fields of an existing user.
func (r *EntUserRepository) Update(ctx context.Context, id uuid.UUID, input UpdateUserInput) (*ent.User, error) {
	b := r.client.User.UpdateOneID(id)

	if input.Name != nil {
		b.SetName(*input.Name)
	}
	if input.PasswordHash != nil {
		b.SetPasswordHash(*input.PasswordHash)
	}
	if input.Role != nil {
		b.SetRole(*input.Role)
	}
	if input.Gender != nil {
		if *input.Gender == "" {
			b.ClearGender()
		} else {
			b.SetGender(*input.Gender)
		}
	}
	if input.Email != nil {
		if *input.Email == "" {
			b.ClearEmail()
		} else {
			b.SetEmail(*input.Email)
		}
	}
	if input.Phone != nil {
		if *input.Phone == "" {
			b.ClearPhone()
		} else {
			b.SetPhone(*input.Phone)
		}
	}
	if input.DepartmentID != nil {
		if *input.DepartmentID == (uuid.UUID{}) {
			b.ClearDepartmentID()
		} else {
			b.SetDepartmentID(*input.DepartmentID)
		}
	}
	if input.GroupID != nil {
		if *input.GroupID == (uuid.UUID{}) {
			b.ClearGroupID()
		} else {
			b.SetGroupID(*input.GroupID)
		}
	}

	u, err := b.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	return u, nil
}

// UpdateStatus changes a user's status.
func (r *EntUserRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status user.Status) (*ent.User, error) {
	u, err := r.client.User.UpdateOneID(id).
		SetStatus(status).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update user status: %w", err)
	}
	return u, nil
}

func (r *EntUserRepository) buildPredicates(filter ListUserFilter) []predicate.User {
	var preds []predicate.User

	if filter.Status != nil {
		preds = append(preds, user.StatusEQ(*filter.Status))
	}
	if filter.Role != nil {
		preds = append(preds, user.RoleEQ(*filter.Role))
	}
	if filter.DepartmentID != nil {
		preds = append(preds, user.DepartmentID(*filter.DepartmentID))
	}
	if filter.Keyword != "" {
		kw := filter.Keyword
		preds = append(preds, user.Or(
			user.UsernameContainsFold(kw),
			user.NameContainsFold(kw),
			user.EmailContainsFold(kw),
			user.PhoneContainsFold(kw),
		))
	}

	return preds
}

func normalizePagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}
