package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/repository"
)

type mockBillingRepository struct {
	preDeductFunc func(ctx context.Context, input repository.PreDeductInput) (*repository.BillingHold, error)
	settleFunc    func(ctx context.Context, hold *repository.BillingHold, actualTokens, actualCost int64, callLogID *uuid.UUID) error
	refundFunc    func(ctx context.Context, hold *repository.BillingHold, callLogID *uuid.UUID) error
}

func (m *mockBillingRepository) PreDeduct(ctx context.Context, input repository.PreDeductInput) (*repository.BillingHold, error) {
	if m.preDeductFunc != nil {
		return m.preDeductFunc(ctx, input)
	}
	return &repository.BillingHold{}, nil
}

func (m *mockBillingRepository) Settle(ctx context.Context, hold *repository.BillingHold, actualTokens, actualCost int64, callLogID *uuid.UUID) error {
	if m.settleFunc != nil {
		return m.settleFunc(ctx, hold, actualTokens, actualCost, callLogID)
	}
	return nil
}

func (m *mockBillingRepository) Refund(ctx context.Context, hold *repository.BillingHold, callLogID *uuid.UUID) error {
	if m.refundFunc != nil {
		return m.refundFunc(ctx, hold, callLogID)
	}
	return nil
}

func (m *mockBillingRepository) AdminAdjustBalance(ctx context.Context, userID, adminID uuid.UUID, amount int64, remark *string) (*ent.User, error) {
	return nil, nil
}

func TestBillingService_PreDeductUsesMaxTokens(t *testing.T) {
	repo := &mockBillingRepository{
		preDeductFunc: func(ctx context.Context, input repository.PreDeductInput) (*repository.BillingHold, error) {
			if input.EstimatedTokens != 256 {
				t.Fatalf("expected estimated tokens 256, got %d", input.EstimatedTokens)
			}
			return &repository.BillingHold{EstimatedTokens: input.EstimatedTokens}, nil
		},
	}
	svc := NewBillingService(repo, BillingConfig{DefaultMaxTokens: 8192})
	body := []byte(`{"model":"gpt-4o","max_tokens":256}`)
	model := &ent.AIModel{InputPrice: 1000, OutputPrice: 2000}

	_, err := svc.PreDeduct(context.Background(), uuid.New(), uuid.New(), model, body)
	if err != nil {
		t.Fatalf("pre-deduct: %v", err)
	}
}

func TestBillingService_PreDeductUsesDefaultMaxTokens(t *testing.T) {
	repo := &mockBillingRepository{
		preDeductFunc: func(ctx context.Context, input repository.PreDeductInput) (*repository.BillingHold, error) {
			if input.EstimatedTokens != 8192 {
				t.Fatalf("expected estimated tokens 8192, got %d", input.EstimatedTokens)
			}
			return &repository.BillingHold{}, nil
		},
	}
	svc := NewBillingService(repo, BillingConfig{DefaultMaxTokens: 8192})
	body := []byte(`{"model":"gpt-4o"}`)
	model := &ent.AIModel{InputPrice: 1000, OutputPrice: 2000}

	_, err := svc.PreDeduct(context.Background(), uuid.New(), uuid.New(), model, body)
	if err != nil {
		t.Fatalf("pre-deduct: %v", err)
	}
}

func TestBillingService_PreDeductPropagatesError(t *testing.T) {
	repo := &mockBillingRepository{
		preDeductFunc: func(ctx context.Context, input repository.PreDeductInput) (*repository.BillingHold, error) {
			return nil, domain.ErrInsufficientQuota
		},
	}
	svc := NewBillingService(repo, BillingConfig{})
	_, err := svc.PreDeduct(context.Background(), uuid.New(), uuid.New(), &ent.AIModel{}, []byte(`{}`))
	if !errors.Is(err, domain.ErrInsufficientQuota) {
		t.Fatalf("expected ErrInsufficientQuota, got %v", err)
	}
}

func TestBillingService_PreDeductEstimatedCost(t *testing.T) {
	repo := &mockBillingRepository{
		preDeductFunc: func(ctx context.Context, input repository.PreDeductInput) (*repository.BillingHold, error) {
			// max price = 3000 per 1k tokens; 100 tokens * 3000 / 1000 * 2 sides = 600
			if input.EstimatedCost != 600 {
				t.Fatalf("expected estimated cost 600, got %d", input.EstimatedCost)
			}
			return &repository.BillingHold{}, nil
		},
	}
	svc := NewBillingService(repo, BillingConfig{})
	body := []byte(`{"model":"gpt-4o","max_tokens":100}`)
	model := &ent.AIModel{InputPrice: 1000, OutputPrice: 3000}

	_, err := svc.PreDeduct(context.Background(), uuid.New(), uuid.New(), model, body)
	if err != nil {
		t.Fatalf("pre-deduct: %v", err)
	}
}

func TestBillingService_SettleComputesActualCost(t *testing.T) {
	repo := &mockBillingRepository{
		settleFunc: func(ctx context.Context, hold *repository.BillingHold, actualTokens, actualCost int64, callLogID *uuid.UUID) error {
			if actualTokens != 50 {
				t.Fatalf("expected actual tokens 50, got %d", actualTokens)
			}
			// prompt 30 * 1000 / 1000 + completion 20 * 3000 / 1000 = 30 + 60 = 90
			if actualCost != 90 {
				t.Fatalf("expected actual cost 90, got %d", actualCost)
			}
			return nil
		},
	}
	svc := NewBillingService(repo, BillingConfig{})
	model := &ent.AIModel{InputPrice: 1000, OutputPrice: 3000}
	hold := &repository.BillingHold{}
	usage := RelayUsage{PromptTokens: 30, CompletionTokens: 20, TotalTokens: 50}

	if err := svc.Settle(context.Background(), hold, model, usage, uuid.New()); err != nil {
		t.Fatalf("settle: %v", err)
	}
}
