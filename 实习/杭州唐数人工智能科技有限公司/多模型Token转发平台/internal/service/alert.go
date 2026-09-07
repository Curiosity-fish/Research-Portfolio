package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/account"
	"github.com/school-api/school-api-v1/ent/alertrule"
	"github.com/school-api/school-api-v1/ent/balancerecord"
	"github.com/school-api/school-api-v1/ent/user"
	"github.com/school-api/school-api-v1/ent/usertoken"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/repository"
)

// AlertRuleResponse is the public representation of an alert rule.
type AlertRuleResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Metric      string  `json:"metric"`
	Threshold   int64   `json:"threshold"`
	Enabled     bool    `json:"enabled"`
	Description *string `json:"description,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// AlertRecordResponse is the public representation of an alert record.
type AlertRecordResponse struct {
	ID             string  `json:"id"`
	RuleID         *string `json:"rule_id,omitempty"`
	Metric         string  `json:"metric"`
	UserID         *string `json:"user_id,omitempty"`
	TokenID        *string `json:"token_id,omitempty"`
	AccountID      *string `json:"account_id,omitempty"`
	TriggeredValue int64   `json:"triggered_value"`
	Message        string  `json:"message"`
	IsResolved     bool    `json:"is_resolved"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

// CreateAlertRuleInput is the data required to create an alert rule.
type CreateAlertRuleInput struct {
	Name        string
	Metric      string
	Threshold   int64
	Enabled     bool
	Description *string
}

// UpdateAlertRuleInput is the data required to update an alert rule.
type UpdateAlertRuleInput struct {
	Name        *string
	Metric      *string
	Threshold   *int64
	Enabled     *bool
	Description *string
}

// AlertService handles alert rules and triggered alert records.
type AlertService struct {
	ruleRepo   repository.AlertRuleRepository
	recordRepo repository.AlertRecordRepository
	client     *ent.Client
}

// NewAlertService creates a new AlertService.
func NewAlertService(ruleRepo repository.AlertRuleRepository, recordRepo repository.AlertRecordRepository, client *ent.Client) *AlertService {
	return &AlertService{
		ruleRepo:   ruleRepo,
		recordRepo: recordRepo,
		client:     client,
	}
}

// CreateAlertRule creates a new alert rule.
func (s *AlertService) CreateAlertRule(ctx context.Context, input CreateAlertRuleInput) (*AlertRuleResponse, error) {
	if err := validateAlertMetric(input.Metric); err != nil {
		return nil, err
	}

	rule, err := s.ruleRepo.Create(ctx, repository.CreateAlertRuleInput{
		Name:        input.Name,
		Metric:      input.Metric,
		Threshold:   input.Threshold,
		Enabled:     input.Enabled,
		Description: input.Description,
	})
	if err != nil {
		return nil, domain.WrapInternal(err)
	}

	return toAlertRuleResponse(rule), nil
}

// GetAlertRule returns an alert rule by ID.
func (s *AlertService) GetAlertRule(ctx context.Context, id uuid.UUID) (*AlertRuleResponse, error) {
	rule, err := s.ruleRepo.GetByID(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrNotFound
		}
		return nil, domain.WrapInternal(err)
	}
	return toAlertRuleResponse(rule), nil
}

// ListAlertRules returns a paginated list of alert rules.
func (s *AlertService) ListAlertRules(ctx context.Context, enabledOnly bool, page, pageSize int) ([]AlertRuleResponse, int, error) {
	page, pageSize = normalizePage(page, pageSize)
	rules, total, err := s.ruleRepo.List(ctx, enabledOnly, (page-1)*pageSize, pageSize)
	if err != nil {
		return nil, 0, domain.WrapInternal(err)
	}

	resp := make([]AlertRuleResponse, 0, len(rules))
	for _, r := range rules {
		resp = append(resp, *toAlertRuleResponse(r))
	}
	return resp, total, nil
}

// UpdateAlertRule updates an alert rule.
func (s *AlertService) UpdateAlertRule(ctx context.Context, id uuid.UUID, input UpdateAlertRuleInput) (*AlertRuleResponse, error) {
	if input.Metric != nil {
		if err := validateAlertMetric(*input.Metric); err != nil {
			return nil, err
		}
	}

	rule, err := s.ruleRepo.Update(ctx, id, repository.UpdateAlertRuleInput{
		Name:        input.Name,
		Metric:      input.Metric,
		Threshold:   input.Threshold,
		Enabled:     input.Enabled,
		Description: input.Description,
	})
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrNotFound
		}
		return nil, domain.WrapInternal(err)
	}
	return toAlertRuleResponse(rule), nil
}

// DeleteAlertRule removes an alert rule.
func (s *AlertService) DeleteAlertRule(ctx context.Context, id uuid.UUID) error {
	if err := s.ruleRepo.Delete(ctx, id); err != nil {
		if ent.IsNotFound(err) {
			return domain.ErrNotFound
		}
		return domain.WrapInternal(err)
	}
	return nil
}

// ListAlertRecords returns triggered alert records.
func (s *AlertService) ListAlertRecords(ctx context.Context, ruleID *uuid.UUID, isResolved *bool, page, pageSize int) ([]AlertRecordResponse, int, error) {
	page, pageSize = normalizePage(page, pageSize)
	records, total, err := s.recordRepo.List(ctx, repository.ListAlertRecordFilter{
		RuleID:     ruleID,
		IsResolved: isResolved,
		Offset:     (page - 1) * pageSize,
		Limit:      pageSize,
	})
	if err != nil {
		return nil, 0, domain.WrapInternal(err)
	}

	resp := make([]AlertRecordResponse, 0, len(records))
	for _, r := range records {
		resp = append(resp, *toAlertRecordResponse(r))
	}
	return resp, total, nil
}

// ResolveAlertRecord marks an alert record as resolved.
func (s *AlertService) ResolveAlertRecord(ctx context.Context, id, adminID uuid.UUID) (*AlertRecordResponse, error) {
	record, err := s.recordRepo.Resolve(ctx, id, repository.ResolveAlertRecordInput{
		ResolvedBy: adminID,
		ResolvedAt: time.Now(),
	})
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrNotFound
		}
		return nil, domain.WrapInternal(err)
	}
	return toAlertRecordResponse(record), nil
}

// EvaluateRules evaluates all enabled alert rules and creates alert records.
func (s *AlertService) EvaluateRules(ctx context.Context) (int, error) {
	rules, _, err := s.ruleRepo.List(ctx, true, 0, 1000)
	if err != nil {
		return 0, domain.WrapInternal(err)
	}

	created := 0
	for _, rule := range rules {
		n, err := s.evaluateRule(ctx, rule)
		if err != nil {
			return created, err
		}
		created += n
	}
	return created, nil
}

func (s *AlertService) evaluateRule(ctx context.Context, rule *ent.AlertRule) (int, error) {
	var candidates []repository.CreateAlertRecordInput
	var err error

	switch rule.Metric {
	case alertrule.MetricBalanceLow:
		candidates, err = s.evaluateBalanceLow(ctx, rule)
	case alertrule.MetricQuotaLow:
		candidates, err = s.evaluateQuotaLow(ctx, rule)
	case alertrule.MetricErrorRate:
		candidates, err = s.evaluateErrorRate(ctx, rule)
	case alertrule.MetricCostSpike:
		candidates, err = s.evaluateCostSpike(ctx, rule)
	default:
		return 0, domain.NewAppError(400, "INVALID_ALERT_METRIC", "不支持的告警指标")
	}
	if err != nil {
		return 0, err
	}

	// Deduplicate within this run, then skip targets that already have an
	// unresolved record for this rule.
	seen := make(map[string]struct{})
	toCreate := make([]repository.CreateAlertRecordInput, 0, len(candidates))
	for _, c := range candidates {
		key := alertTargetKey(c)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}

		exists, err := s.recordRepo.ExistsUnresolved(ctx, rule.ID, c.UserID, c.TokenID, c.AccountID)
		if err != nil {
			return 0, domain.WrapInternal(err)
		}
		if exists {
			continue
		}
		toCreate = append(toCreate, c)
	}

	if len(toCreate) == 0 {
		return 0, nil
	}

	// All records for one rule are written in a single transaction so a
	// failure cannot leave a partially evaluated rule behind.
	n, err := s.recordRepo.CreateBatch(ctx, toCreate)
	if err != nil {
		return 0, domain.WrapInternal(err)
	}
	return n, nil
}

func alertTargetKey(c repository.CreateAlertRecordInput) string {
	return fmt.Sprintf("%s|%s|%s|%s",
		optionalUUIDString(c.RuleID),
		optionalUUIDString(c.UserID),
		optionalUUIDString(c.TokenID),
		optionalUUIDString(c.AccountID),
	)
}

func optionalUUIDString(id *uuid.UUID) string {
	if id == nil {
		return ""
	}
	return id.String()
}

func (s *AlertService) evaluateBalanceLow(ctx context.Context, rule *ent.AlertRule) ([]repository.CreateAlertRecordInput, error) {
	users, err := s.client.User.Query().
		Where(user.StatusEQ(user.StatusActive), user.BalanceLTE(rule.Threshold)).
		All(ctx)
	if err != nil {
		return nil, domain.WrapInternal(err)
	}

	candidates := make([]repository.CreateAlertRecordInput, 0, len(users))
	for _, u := range users {
		candidates = append(candidates, repository.CreateAlertRecordInput{
			RuleID:         &rule.ID,
			Metric:         string(rule.Metric),
			UserID:         &u.ID,
			TriggeredValue: u.Balance,
			Message:        fmt.Sprintf("用户余额 %d 低于阈值 %d", u.Balance, rule.Threshold),
		})
	}
	return candidates, nil
}

func (s *AlertService) evaluateQuotaLow(ctx context.Context, rule *ent.AlertRule) ([]repository.CreateAlertRecordInput, error) {
	tokens, err := s.client.UserToken.Query().
		Where(
			usertoken.IsEnabledEQ(true),
			usertoken.QuotaLimitNotNil(),
		).
		All(ctx)
	if err != nil {
		return nil, domain.WrapInternal(err)
	}

	candidates := make([]repository.CreateAlertRecordInput, 0, len(tokens))
	for _, tok := range tokens {
		remaining := int64(0)
		if tok.QuotaLimit != nil {
			remaining = *tok.QuotaLimit - tok.QuotaUsed
		}
		if remaining < 0 {
			remaining = 0
		}
		if remaining > rule.Threshold {
			continue
		}
		candidates = append(candidates, repository.CreateAlertRecordInput{
			RuleID:         &rule.ID,
			Metric:         string(rule.Metric),
			UserID:         &tok.UserID,
			TokenID:        &tok.ID,
			TriggeredValue: remaining,
			Message:        fmt.Sprintf("Token 剩余配额 %d 低于阈值 %d", remaining, rule.Threshold),
		})
	}
	return candidates, nil
}

func (s *AlertService) evaluateErrorRate(ctx context.Context, rule *ent.AlertRule) ([]repository.CreateAlertRecordInput, error) {
	accounts, err := s.client.Account.Query().
		Where(account.StatusEQ(account.StatusActive), account.ErrorCountGTE(int(rule.Threshold))).
		All(ctx)
	if err != nil {
		return nil, domain.WrapInternal(err)
	}

	candidates := make([]repository.CreateAlertRecordInput, 0, len(accounts))
	for _, a := range accounts {
		candidates = append(candidates, repository.CreateAlertRecordInput{
			RuleID:         &rule.ID,
			Metric:         string(rule.Metric),
			AccountID:      &a.ID,
			TriggeredValue: int64(a.ErrorCount),
			Message:        fmt.Sprintf("账号连续错误次数 %d 达到阈值 %d", a.ErrorCount, rule.Threshold),
		})
	}
	return candidates, nil
}

func (s *AlertService) evaluateCostSpike(ctx context.Context, rule *ent.AlertRule) ([]repository.CreateAlertRecordInput, error) {
	start := time.Now().UTC().Truncate(24 * time.Hour)
	records, err := s.client.BalanceRecord.Query().
		Where(
			balancerecord.TypeEQ(balancerecord.TypeConsume),
			balancerecord.CreatedAtGTE(start),
		).
		All(ctx)
	if err != nil {
		return nil, domain.WrapInternal(err)
	}

	byUser := make(map[uuid.UUID]int64)
	for _, rec := range records {
		byUser[rec.UserID] += rec.Amount
	}

	candidates := make([]repository.CreateAlertRecordInput, 0, len(byUser))
	for uid, total := range byUser {
		if total >= rule.Threshold {
			candidates = append(candidates, repository.CreateAlertRecordInput{
				RuleID:         &rule.ID,
				Metric:         string(rule.Metric),
				UserID:         &uid,
				TriggeredValue: total,
				Message:        fmt.Sprintf("用户今日消费 %d 达到阈值 %d", total, rule.Threshold),
			})
		}
	}
	return candidates, nil
}

func validateAlertMetric(metric string) error {
	switch alertrule.Metric(metric) {
	case alertrule.MetricBalanceLow, alertrule.MetricQuotaLow, alertrule.MetricErrorRate, alertrule.MetricCostSpike:
		return nil
	default:
		return domain.NewAppError(400, "INVALID_ALERT_METRIC", "不支持的告警指标")
	}
}

func toAlertRuleResponse(r *ent.AlertRule) *AlertRuleResponse {
	return &AlertRuleResponse{
		ID:          r.ID.String(),
		Name:        r.Name,
		Metric:      string(r.Metric),
		Threshold:   r.Threshold,
		Enabled:     r.Enabled,
		Description: r.Description,
		CreatedAt:   r.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   r.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toAlertRecordResponse(r *ent.AlertRecord) *AlertRecordResponse {
	resp := &AlertRecordResponse{
		ID:             r.ID.String(),
		Metric:         r.Metric,
		TriggeredValue: r.TriggeredValue,
		Message:        r.Message,
		IsResolved:     r.IsResolved,
		CreatedAt:      r.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:      r.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if r.RuleID != nil {
		s := r.RuleID.String()
		resp.RuleID = &s
	}
	if r.UserID != nil {
		s := r.UserID.String()
		resp.UserID = &s
	}
	if r.TokenID != nil {
		s := r.TokenID.String()
		resp.TokenID = &s
	}
	if r.AccountID != nil {
		s := r.AccountID.String()
		resp.AccountID = &s
	}
	return resp
}
