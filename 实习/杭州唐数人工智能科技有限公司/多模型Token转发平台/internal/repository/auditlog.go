package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/auditlog"
)

// CreateAuditLogInput is the data required to create an audit log.
type CreateAuditLogInput struct {
	ActorType  string
	ActorID    uuid.UUID
	Action     string
	TargetType *string
	TargetID   *uuid.UUID
	Details    map[string]any
	IP         *string
	UserAgent  *string
}

// ListAuditLogFilter controls the audit log list query.
type ListAuditLogFilter struct {
	ActorType *string
	ActorID   *uuid.UUID
	Action    *string
	Offset    int
	Limit     int
}

// AuditLogRepository provides data access for audit logs.
type AuditLogRepository interface {
	Create(ctx context.Context, input CreateAuditLogInput) (*ent.AuditLog, error)
	List(ctx context.Context, filter ListAuditLogFilter) ([]*ent.AuditLog, int, error)
}

// EntAuditLogRepository implements AuditLogRepository using Ent.
type EntAuditLogRepository struct {
	client *ent.Client
}

// NewEntAuditLogRepository creates a new Ent-backed audit log repository.
func NewEntAuditLogRepository(client *ent.Client) *EntAuditLogRepository {
	return &EntAuditLogRepository{client: client}
}

// Create inserts a new audit log.
func (r *EntAuditLogRepository) Create(ctx context.Context, input CreateAuditLogInput) (*ent.AuditLog, error) {
	b := r.client.AuditLog.Create().
		SetActorType(auditlog.ActorType(input.ActorType)).
		SetActorID(input.ActorID).
		SetAction(input.Action)
	if input.TargetType != nil {
		b.SetTargetType(*input.TargetType)
	}
	if input.TargetID != nil {
		b.SetTargetID(*input.TargetID)
	}
	if input.Details != nil {
		b.SetDetails(input.Details)
	}
	if input.IP != nil {
		b.SetIP(*input.IP)
	}
	if input.UserAgent != nil {
		b.SetUserAgent(*input.UserAgent)
	}

	log, err := b.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create audit log: %w", err)
	}
	return log, nil
}

// List returns a paginated list of audit logs.
func (r *EntAuditLogRepository) List(ctx context.Context, filter ListAuditLogFilter) ([]*ent.AuditLog, int, error) {
	q := r.client.AuditLog.Query()
	if filter.ActorType != nil {
		q = q.Where(auditlog.ActorTypeEQ(auditlog.ActorType(*filter.ActorType)))
	}
	if filter.ActorID != nil {
		q = q.Where(auditlog.ActorIDEQ(*filter.ActorID))
	}
	if filter.Action != nil {
		q = q.Where(auditlog.ActionEQ(*filter.Action))
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count audit logs: %w", err)
	}

	list, err := q.Order(ent.Desc(auditlog.FieldCreatedAt)).
		Offset(filter.Offset).
		Limit(filter.Limit).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list audit logs: %w", err)
	}

	return list, total, nil
}

var _ AuditLogRepository = (*EntAuditLogRepository)(nil)
