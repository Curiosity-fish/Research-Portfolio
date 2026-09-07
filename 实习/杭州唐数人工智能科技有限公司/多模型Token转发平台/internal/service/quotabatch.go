package service

import (
	"context"
	"strconv"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/repository"
)

// MaxQuotaBatchUsers caps the explicit user list of one batch operation. The
// platform targets hundreds of users per class, so 1000 leaves headroom while
// keeping the single transaction bounded.
const MaxQuotaBatchUsers = 1000

// QuotaBatch modes accepted from the admin API.
const (
	QuotaBatchModeSet = string(repository.QuotaBatchModeSet)
	QuotaBatchModeAdd = string(repository.QuotaBatchModeAdd)
)

// QuotaRangeProvider supplies the configured quota range. Implemented by
// SettingsService.
type QuotaRangeProvider interface {
	QuotaRange(ctx context.Context) (*QuotaRange, error)
}

// QuotaBatchRepository applies the batch operation after validation.
type QuotaBatchRepository interface {
	Apply(ctx context.Context, input repository.QuotaBatchInput) (*repository.QuotaBatchResult, error)
}

// ApplyQuotaBatchInput is the service-level input for a batch quota change.
// Exactly one of UserIDs and DepartmentID must be set.
type ApplyQuotaBatchInput struct {
	UserIDs      []uuid.UUID
	DepartmentID *uuid.UUID
	Mode         string
	Value        int64
}

// QuotaBatchService validates and applies batch quota operations.
type QuotaBatchService struct {
	repo     QuotaBatchRepository
	settings QuotaRangeProvider
}

// NewQuotaBatchService creates a new QuotaBatchService. settings may be nil,
// in which case no range validation runs and set-mode values are only checked
// to be positive (used by tests and future embedding scenarios).
func NewQuotaBatchService(repo QuotaBatchRepository, settings QuotaRangeProvider) *QuotaBatchService {
	return &QuotaBatchService{repo: repo, settings: settings}
}

// Apply validates the request against the configured quota range and applies
// it to every enabled token of the target users.
func (s *QuotaBatchService) Apply(ctx context.Context, input ApplyQuotaBatchInput) (*repository.QuotaBatchResult, error) {
	if input.Mode != QuotaBatchModeSet && input.Mode != QuotaBatchModeAdd {
		return nil, domain.NewValidationError(map[string]string{
			"mode": "批量模式必须是 set 或 add",
		})
	}

	hasUsers := len(input.UserIDs) > 0
	hasDepartment := input.DepartmentID != nil
	if hasUsers == hasDepartment {
		return nil, domain.NewValidationError(map[string]string{
			"user_ids": "user_ids 与 department_id 必须二选一",
		})
	}
	if hasUsers && len(input.UserIDs) > MaxQuotaBatchUsers {
		return nil, domain.NewValidationError(map[string]string{
			"user_ids": "单次批量最多支持 " + strconv.Itoa(MaxQuotaBatchUsers) + " 个用户",
		})
	}

	// Set mode overwrites every target with the same value, so the range can
	// be checked once up front; add mode is bounded per token by the repo.
	quotaRange := QuotaRange{}
	if s.settings != nil {
		rng, err := s.settings.QuotaRange(ctx)
		if err != nil {
			return nil, err
		}
		quotaRange = *rng
	}
	if input.Mode == QuotaBatchModeSet {
		if input.Value <= 0 {
			return nil, domain.NewValidationError(map[string]string{
				"value": "设置的配额必须大于 0",
			})
		}
		if s.settings != nil && (input.Value < quotaRange.Min || input.Value > quotaRange.Max) {
			return nil, domain.NewValidationError(map[string]string{
				"value": "设置的配额必须在额度区间 [" + strconv.FormatInt(quotaRange.Min, 10) + ", " +
					strconv.FormatInt(quotaRange.Max, 10) + "] 内",
			})
		}
	} else if input.Value == 0 {
		return nil, domain.NewValidationError(map[string]string{
			"value": "增减量不能为 0",
		})
	}

	result, err := s.repo.Apply(ctx, repository.QuotaBatchInput{
		UserIDs:      input.UserIDs,
		DepartmentID: input.DepartmentID,
		Mode:         repository.QuotaBatchMode(input.Mode),
		Value:        input.Value,
		MaxQuota:     quotaRange.Max,
	})
	if err != nil {
		return nil, domain.WrapInternal(err)
	}
	return result, nil
}
