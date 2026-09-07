package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/repository"
)

// captureQuotaBatchRepo records the last input it received.
type captureQuotaBatchRepo struct {
	input repository.QuotaBatchInput
	calls int
}

func (c *captureQuotaBatchRepo) Apply(_ context.Context, input repository.QuotaBatchInput) (*repository.QuotaBatchResult, error) {
	c.calls++
	c.input = input
	return &repository.QuotaBatchResult{MatchedUsers: 1, UpdatedTokens: 1}, nil
}

// stubQuotaRange implements QuotaRangeProvider with a fixed range.
type stubQuotaRange struct {
	rng QuotaRange
	err error
}

func (s *stubQuotaRange) QuotaRange(context.Context) (*QuotaRange, error) {
	return &s.rng, s.err
}

func TestQuotaBatchService_SetModeValidatesAgainstRange(t *testing.T) {
	ctx := context.Background()
	repo := &captureQuotaBatchRepo{}
	svc := NewQuotaBatchService(repo, &stubQuotaRange{rng: QuotaRange{Min: 500_000, Max: 3_000_000}})

	if _, err := svc.Apply(ctx, ApplyQuotaBatchInput{
		UserIDs: []uuid.UUID{uuid.New()}, Mode: QuotaBatchModeSet, Value: 499_999,
	}); err == nil {
		t.Fatal("expected validation error below range")
	}
	if _, err := svc.Apply(ctx, ApplyQuotaBatchInput{
		UserIDs: []uuid.UUID{uuid.New()}, Mode: QuotaBatchModeSet, Value: 3_000_001,
	}); err == nil {
		t.Fatal("expected validation error above range")
	}
	if _, err := svc.Apply(ctx, ApplyQuotaBatchInput{
		UserIDs: []uuid.UUID{uuid.New()}, Mode: QuotaBatchModeSet, Value: 0,
	}); err == nil {
		t.Fatal("expected validation error for zero value")
	}
	if repo.calls != 0 {
		t.Fatalf("repo must not be called for invalid values, got %d calls", repo.calls)
	}

	if _, err := svc.Apply(ctx, ApplyQuotaBatchInput{
		UserIDs: []uuid.UUID{uuid.New()}, Mode: QuotaBatchModeSet, Value: 1_000_000,
	}); err != nil {
		t.Fatalf("apply in-range set: %v", err)
	}
	if repo.calls != 1 {
		t.Fatalf("expected 1 repo call, got %d", repo.calls)
	}
	if repo.input.MaxQuota != 3_000_000 {
		t.Fatalf("expected MaxQuota forwarded to repo, got %d", repo.input.MaxQuota)
	}
}

func TestQuotaBatchService_TargetValidation(t *testing.T) {
	ctx := context.Background()
	svc := NewQuotaBatchService(&captureQuotaBatchRepo{}, &stubQuotaRange{rng: QuotaRange{Min: 1, Max: 10}})

	// Neither target.
	if _, err := svc.Apply(ctx, ApplyQuotaBatchInput{Mode: QuotaBatchModeSet, Value: 5}); err == nil {
		t.Fatal("expected validation error for missing target")
	}
	// Both targets.
	dep := uuid.New()
	if _, err := svc.Apply(ctx, ApplyQuotaBatchInput{
		UserIDs: []uuid.UUID{uuid.New()}, DepartmentID: &dep, Mode: QuotaBatchModeSet, Value: 5,
	}); err == nil {
		t.Fatal("expected validation error for dual target")
	}
	// Empty user list is treated as "no target".
	if _, err := svc.Apply(ctx, ApplyQuotaBatchInput{UserIDs: nil, Mode: QuotaBatchModeSet, Value: 5}); err == nil {
		t.Fatal("expected validation error for empty user list")
	}
	// Unknown mode.
	if _, err := svc.Apply(ctx, ApplyQuotaBatchInput{
		UserIDs: []uuid.UUID{uuid.New()}, Mode: "double", Value: 5,
	}); err == nil {
		t.Fatal("expected validation error for unknown mode")
	}
	// Add mode rejects zero.
	if _, err := svc.Apply(ctx, ApplyQuotaBatchInput{
		UserIDs: []uuid.UUID{uuid.New()}, Mode: QuotaBatchModeAdd, Value: 0,
	}); err == nil {
		t.Fatal("expected validation error for zero delta")
	}
	// Oversized user list.
	ids := make([]uuid.UUID, MaxQuotaBatchUsers+1)
	for i := range ids {
		ids[i] = uuid.New()
	}
	if _, err := svc.Apply(ctx, ApplyQuotaBatchInput{
		UserIDs: ids, Mode: QuotaBatchModeSet, Value: 5,
	}); err == nil {
		t.Fatal("expected validation error for oversized user list")
	}
}

func TestQuotaBatchService_AddModeSkipsRangeCheckAndForwardsDepartment(t *testing.T) {
	ctx := context.Background()
	repo := &captureQuotaBatchRepo{}
	svc := NewQuotaBatchService(repo, &stubQuotaRange{rng: QuotaRange{Min: 500_000, Max: 3_000_000}})

	dep := uuid.New()
	if _, err := svc.Apply(ctx, ApplyQuotaBatchInput{
		DepartmentID: &dep, Mode: QuotaBatchModeAdd, Value: -100,
	}); err != nil {
		t.Fatalf("apply add by department: %v", err)
	}
	if repo.calls != 1 || repo.input.DepartmentID == nil || *repo.input.DepartmentID != dep {
		t.Fatalf("department target not forwarded: %+v", repo.input)
	}
	if repo.input.Mode != repository.QuotaBatchModeAdd || repo.input.Value != -100 {
		t.Fatalf("unexpected forwarded input: %+v", repo.input)
	}
}

func TestQuotaBatchService_NilSettingsOnlyChecksPositivity(t *testing.T) {
	ctx := context.Background()
	repo := &captureQuotaBatchRepo{}
	svc := NewQuotaBatchService(repo, nil)

	// Without a settings provider the range is unknown, so any positive
	// set value passes and MaxQuota stays zero.
	if _, err := svc.Apply(ctx, ApplyQuotaBatchInput{
		UserIDs: []uuid.UUID{uuid.New()}, Mode: QuotaBatchModeSet, Value: 123,
	}); err != nil {
		t.Fatalf("apply without settings: %v", err)
	}
	if repo.input.MaxQuota != 0 {
		t.Fatalf("expected zero MaxQuota without settings, got %d", repo.input.MaxQuota)
	}
}

func TestQuotaBatchService_SettingsErrorPropagates(t *testing.T) {
	svc := NewQuotaBatchService(&captureQuotaBatchRepo{}, &stubQuotaRange{err: domain.WrapInternal(errors.New("boom"))})
	_, err := svc.Apply(context.Background(), ApplyQuotaBatchInput{
		UserIDs: []uuid.UUID{uuid.New()}, Mode: QuotaBatchModeSet, Value: 1,
	})
	var appErr *domain.AppError
	if !errors.As(err, &appErr) || appErr.Code != 500 {
		t.Fatalf("expected wrapped 500 error, got %v", err)
	}
	// Add mode also fails when the range cannot be loaded.
	_, err = svc.Apply(context.Background(), ApplyQuotaBatchInput{
		UserIDs: []uuid.UUID{uuid.New()}, Mode: QuotaBatchModeAdd, Value: 1,
	})
	if err == nil {
		t.Fatal("expected error for add mode with broken settings")
	}
}
