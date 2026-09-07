package service

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/repository"
)

// BillingService handles quota/balance estimation and settlement around relay calls.
type BillingService interface {
	PreDeduct(ctx context.Context, userID, tokenID uuid.UUID, model *ent.AIModel, requestBody []byte) (*repository.BillingHold, error)
	Settle(ctx context.Context, hold *repository.BillingHold, model *ent.AIModel, usage RelayUsage, callLogID uuid.UUID) error
	Refund(ctx context.Context, hold *repository.BillingHold, callLogID *uuid.UUID) error
}

// BillingConfig carries tunables for the billing calculations.
type BillingConfig struct {
	DefaultMaxTokens int64
}

type billingService struct {
	repo             repository.BillingRepository
	defaultMaxTokens int64
}

// NewBillingService creates a BillingService with the given repository and defaults.
func NewBillingService(repo repository.BillingRepository, cfg BillingConfig) BillingService {
	if cfg.DefaultMaxTokens <= 0 {
		cfg.DefaultMaxTokens = 8192
	}
	return &billingService{
		repo:             repo,
		defaultMaxTokens: cfg.DefaultMaxTokens,
	}
}

// PreDeduct estimates the quota/cost for a request and reserves them atomically.
func (s *billingService) PreDeduct(ctx context.Context, userID, tokenID uuid.UUID, model *ent.AIModel, requestBody []byte) (*repository.BillingHold, error) {
	maxTokens, err := extractMaxTokens(requestBody)
	if err != nil {
		return nil, domain.NewAppError(400, "invalid_request_error", "无法解析 max_tokens")
	}
	if maxTokens == nil {
		maxTokens = &s.defaultMaxTokens
	}

	estimatedTokens := *maxTokens
	if estimatedTokens <= 0 {
		estimatedTokens = s.defaultMaxTokens
	}

	estimatedCost := s.estimateCost(model, estimatedTokens, estimatedTokens)

	hold, err := s.repo.PreDeduct(ctx, repository.PreDeductInput{
		UserID:          userID,
		TokenID:         tokenID,
		EstimatedTokens: estimatedTokens,
		EstimatedCost:   estimatedCost,
	})
	if err != nil {
		return nil, err
	}
	return hold, nil
}

// Settle reconciles the pre-deducted reservation with the actual upstream usage.
func (s *billingService) Settle(ctx context.Context, hold *repository.BillingHold, model *ent.AIModel, usage RelayUsage, callLogID uuid.UUID) error {
	actualCost := (usage.PromptTokens*model.InputPrice + usage.CompletionTokens*model.OutputPrice) / 1000
	return s.repo.Settle(ctx, hold, usage.TotalTokens, actualCost, &callLogID)
}

// Refund releases the reservation when the upstream call did not succeed.
func (s *billingService) Refund(ctx context.Context, hold *repository.BillingHold, callLogID *uuid.UUID) error {
	return s.repo.Refund(ctx, hold, callLogID)
}

// estimateCost computes a conservative pre-deduct cost using the higher of the
// input/output prices so that the held amount is likely to cover the final bill.
func (s *billingService) estimateCost(model *ent.AIModel, promptTokens, completionTokens int64) int64 {
	unitPrice := max(model.InputPrice, model.OutputPrice)
	return promptTokens*unitPrice/1000 + completionTokens*unitPrice/1000
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// extractMaxTokens parses the optional max_tokens field from the request body.
func extractMaxTokens(body []byte) (*int64, error) {
	var req struct {
		MaxTokens *int64 `json:"max_tokens"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, err
	}
	return req.MaxTokens, nil
}
