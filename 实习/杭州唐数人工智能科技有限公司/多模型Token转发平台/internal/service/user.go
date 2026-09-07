package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/user"
	"github.com/school-api/school-api-v1/internal/auth"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/repository"
)

// CreateUserInput is the service-level input for creating a user.
type CreateUserInput struct {
	Username     string
	Password     string
	Name         string
	Email        string
	Phone        string
	Role         user.Role
	Gender       string
	Status       user.Status
	DepartmentID *uuid.UUID
	GroupID      *uuid.UUID
}

// UpdateUserInput is the service-level input for updating a user.
type UpdateUserInput struct {
	Name         *string
	Password     *string
	Email        *string
	Phone        *string
	Role         *user.Role
	Gender       *string
	DepartmentID *uuid.UUID
	GroupID      *uuid.UUID
}

// ListUsersFilter controls user list queries.
type ListUsersFilter struct {
	Keyword      string
	Status       *user.Status
	Role         *user.Role
	DepartmentID *uuid.UUID
	Page         int
	PageSize     int
}

// UserResponse is the public representation of a user. The quota aggregate
// fields summarize the user's *enabled* tokens only (class-view requirement):
// TokenCount is the number of enabled tokens, QuotaLimitTotal is the sum of
// their finite limits, and a nil QuotaLimitTotal means at least one enabled
// token is unlimited.
type UserResponse struct {
	ID               string     `json:"id"`
	Username         string     `json:"username"`
	Name             string     `json:"name"`
	Email            *string    `json:"email,omitempty"`
	Phone            *string    `json:"phone,omitempty"`
	Role             string     `json:"role"`
	Gender           *string    `json:"gender,omitempty"`
	Status           string     `json:"status"`
	DepartmentID     *uuid.UUID `json:"department_id,omitempty"`
	GroupID          *uuid.UUID `json:"group_id,omitempty"`
	TokenCount       int        `json:"token_count"`
	QuotaLimitTotal  *int64     `json:"quota_limit_total"`
	QuotaUsedTotal   int64      `json:"quota_used_total"`
	CreatedAt        string     `json:"created_at"`
	UpdatedAt        string     `json:"updated_at"`
}

// ListUsersResponse is the paginated list response.
type ListUsersResponse struct {
	Total    int            `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
	List     []UserResponse `json:"list"`
}

// UserRepository defines the persistence operations needed by UserService.
type UserRepository interface {
	Create(ctx context.Context, input repository.CreateUserInput) (*ent.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ent.User, error)
	GetByUsername(ctx context.Context, username string) (*ent.User, error)
	List(ctx context.Context, filter repository.ListUserFilter) ([]*ent.User, int, error)
	Update(ctx context.Context, id uuid.UUID, input repository.UpdateUserInput) (*ent.User, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status user.Status) (*ent.User, error)
}

// UserService handles user management business logic.
type UserService struct {
	repo       UserRepository
	bcryptCost int
}

// NewUserService creates a new UserService.
func NewUserService(repo UserRepository, bcryptCost int) *UserService {
	if bcryptCost < auth.MinBcryptCost {
		bcryptCost = auth.DefaultBcryptCost
	}
	return &UserService{repo: repo, bcryptCost: bcryptCost}
}

// CreateUser creates a new user with a hashed password.
func (s *UserService) CreateUser(ctx context.Context, input CreateUserInput) (*UserResponse, error) {
	if input.Status == "" {
		input.Status = user.StatusActive
	}
	if input.Role == "" {
		input.Role = user.RoleStudent
	}

	hash, err := auth.HashPassword(s.bcryptCost, input.Password)
	if err != nil {
		return nil, domain.WrapInternal(err)
	}

	u, err := s.repo.Create(ctx, repository.CreateUserInput{
		Username:     input.Username,
		PasswordHash: hash,
		Name:         input.Name,
		Email:        input.Email,
		Phone:        input.Phone,
		Role:         input.Role,
		Gender:       input.Gender,
		Status:       input.Status,
		DepartmentID: input.DepartmentID,
		GroupID:      input.GroupID,
	})
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, domain.ErrConflict
		}
		return nil, domain.WrapInternal(err)
	}

	return toUserResponse(u), nil
}

// ListUsers returns a paginated list of users.
func (s *UserService) ListUsers(ctx context.Context, filter ListUsersFilter) (*ListUsersResponse, error) {
	users, total, err := s.repo.List(ctx, repository.ListUserFilter{
		Keyword:      filter.Keyword,
		Status:       filter.Status,
		Role:         filter.Role,
		DepartmentID: filter.DepartmentID,
		Page:         filter.Page,
		PageSize:     filter.PageSize,
	})
	if err != nil {
		return nil, domain.WrapInternal(err)
	}

	page, pageSize := normalizePage(filter.Page, filter.PageSize)
	resp := &ListUsersResponse{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		List:     make([]UserResponse, 0, len(users)),
	}
	for _, u := range users {
		resp.List = append(resp.List, *toUserResponse(u))
	}
	return resp, nil
}

// GetUser returns a single user by ID.
func (s *UserService) GetUser(ctx context.Context, id uuid.UUID) (*UserResponse, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrNotFound
		}
		return nil, domain.WrapInternal(err)
	}
	return toUserResponse(u), nil
}

// UpdateUser updates an existing user.
func (s *UserService) UpdateUser(ctx context.Context, id uuid.UUID, input UpdateUserInput) (*UserResponse, error) {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrNotFound
		}
		return nil, domain.WrapInternal(err)
	}

	repoInput := repository.UpdateUserInput{
		Name:         input.Name,
		Email:        input.Email,
		Phone:        input.Phone,
		Role:         input.Role,
		Gender:       input.Gender,
		DepartmentID: input.DepartmentID,
		GroupID:      input.GroupID,
	}

	if input.Password != nil {
		hash, err := auth.HashPassword(s.bcryptCost, *input.Password)
		if err != nil {
			return nil, domain.WrapInternal(err)
		}
		repoInput.PasswordHash = &hash
	}

	u, err := s.repo.Update(ctx, id, repoInput)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, domain.ErrConflict
		}
		return nil, domain.WrapInternal(err)
	}

	return toUserResponse(u), nil
}

// UpdateUserStatus changes a user's status.
func (s *UserService) UpdateUserStatus(ctx context.Context, id uuid.UUID, status user.Status) (*UserResponse, error) {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrNotFound
		}
		return nil, domain.WrapInternal(err)
	}

	u, err := s.repo.UpdateStatus(ctx, id, status)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrNotFound
		}
		return nil, domain.WrapInternal(err)
	}
	return toUserResponse(u), nil
}

func toUserResponse(u *ent.User) *UserResponse {
	resp := &UserResponse{
		ID:        u.ID.String(),
		Username:  u.Username,
		Name:      u.Name,
		Role:      u.Role.String(),
		Status:    u.Status.String(),
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: u.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if u.Email != nil {
		resp.Email = u.Email
	}
	if u.Phone != nil {
		resp.Phone = u.Phone
	}
	if u.Gender != nil {
		resp.Gender = u.Gender
	}
	if u.DepartmentID != nil {
		resp.DepartmentID = u.DepartmentID
	}
	if u.GroupID != nil {
		resp.GroupID = u.GroupID
	}
	applyQuotaAggregate(resp, u.Edges.Tokens)
	return resp
}

// applyQuotaAggregate fills the enabled-token quota summary on resp. A nil
// limit marker wins over the finite sum: once an unlimited token is seen the
// total is reported as nil (unlimited) regardless of other tokens.
func applyQuotaAggregate(resp *UserResponse, tokens []*ent.UserToken) {
	var (
		count      int
		limitSum   int64
		hasFinite  bool
		unlimited  bool
		usedSum    int64
	)
	for _, t := range tokens {
		if !t.IsEnabled {
			continue
		}
		count++
		usedSum += t.QuotaUsed
		if t.QuotaLimit == nil {
			unlimited = true
		} else {
			limitSum += *t.QuotaLimit
			hasFinite = true
		}
	}
	resp.TokenCount = count
	resp.QuotaUsedTotal = usedSum
	if hasFinite && !unlimited {
		resp.QuotaLimitTotal = &limitSum
	}
}

func normalizePage(page, pageSize int) (int, int) {
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

// Compile-time interface check.
var _ repository.UserRepository = (*repository.EntUserRepository)(nil)
