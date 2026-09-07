package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/alertrule"
)

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

// AlertRuleRepository provides data access for alert rules.
type AlertRuleRepository interface {
	Create(ctx context.Context, input CreateAlertRuleInput) (*ent.AlertRule, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ent.AlertRule, error)
	List(ctx context.Context, enabledOnly bool, offset, limit int) ([]*ent.AlertRule, int, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateAlertRuleInput) (*ent.AlertRule, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// EntAlertRuleRepository implements AlertRuleRepository using Ent.
type EntAlertRuleRepository struct {
	client *ent.Client
}

// NewEntAlertRuleRepository creates a new Ent-backed alert rule repository.
func NewEntAlertRuleRepository(client *ent.Client) *EntAlertRuleRepository {
	return &EntAlertRuleRepository{client: client}
}

// Create inserts a new alert rule.
func (r *EntAlertRuleRepository) Create(ctx context.Context, input CreateAlertRuleInput) (*ent.AlertRule, error) {
	b := r.client.AlertRule.Create().
		SetName(input.Name).
		SetMetric(alertrule.Metric(input.Metric)).
		SetThreshold(input.Threshold).
		SetEnabled(input.Enabled)
	if input.Description != nil {
		b.SetDescription(*input.Description)
	}

	ar, err := b.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create alert rule: %w", err)
	}
	return ar, nil
}

// GetByID returns an alert rule by ID.
func (r *EntAlertRuleRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.AlertRule, error) {
	ar, err := r.client.AlertRule.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get alert rule: %w", err)
	}
	return ar, nil
}

// List returns a paginated list of alert rules.
func (r *EntAlertRuleRepository) List(ctx context.Context, enabledOnly bool, offset, limit int) ([]*ent.AlertRule, int, error) {
	q := r.client.AlertRule.Query()
	if enabledOnly {
		q = q.Where(alertrule.EnabledEQ(true))
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count alert rules: %w", err)
	}

	list, err := q.Order(ent.Desc(alertrule.FieldCreatedAt)).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list alert rules: %w", err)
	}

	return list, total, nil
}

// Update modifies an alert rule.
func (r *EntAlertRuleRepository) Update(ctx context.Context, id uuid.UUID, input UpdateAlertRuleInput) (*ent.AlertRule, error) {
	b := r.client.AlertRule.UpdateOneID(id)
	if input.Name != nil {
		b.SetName(*input.Name)
	}
	if input.Metric != nil {
		b.SetMetric(alertrule.Metric(*input.Metric))
	}
	if input.Threshold != nil {
		b.SetThreshold(*input.Threshold)
	}
	if input.Enabled != nil {
		b.SetEnabled(*input.Enabled)
	}
	if input.Description != nil {
		b.SetDescription(*input.Description)
	}

	ar, err := b.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update alert rule: %w", err)
	}
	return ar, nil
}

// Delete removes an alert rule.
func (r *EntAlertRuleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.client.AlertRule.DeleteOneID(id).Exec(ctx); err != nil {
		return fmt.Errorf("delete alert rule: %w", err)
	}
	return nil
}

var _ AlertRuleRepository = (*EntAlertRuleRepository)(nil)
