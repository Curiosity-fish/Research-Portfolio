package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/repository"
)

// AuditLogResponse is the public representation of an audit log.
type AuditLogResponse struct {
	ID        string         `json:"id"`
	ActorType string         `json:"actor_type"`
	ActorID   string         `json:"actor_id"`
	Action    string         `json:"action"`
	TargetType *string       `json:"target_type,omitempty"`
	TargetID   *string       `json:"target_id,omitempty"`
	Details    map[string]any `json:"details"`
	IP         *string       `json:"ip,omitempty"`
	UserAgent  *string       `json:"user_agent,omitempty"`
	CreatedAt  string         `json:"created_at"`
	UpdatedAt  string         `json:"updated_at"`
}

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

// AuditLogService handles audit log business logic.
type AuditLogService struct {
	repo repository.AuditLogRepository
}

// NewAuditLogService creates a new AuditLogService.
func NewAuditLogService(repo repository.AuditLogRepository) *AuditLogService {
	return &AuditLogService{repo: repo}
}

// Log writes an audit log entry.
func (s *AuditLogService) Log(ctx context.Context, input CreateAuditLogInput) error {
	_, err := s.repo.Create(ctx, repository.CreateAuditLogInput{
		ActorType:  input.ActorType,
		ActorID:    input.ActorID,
		Action:     input.Action,
		TargetType: input.TargetType,
		TargetID:   input.TargetID,
		Details:    input.Details,
		IP:         input.IP,
		UserAgent:  input.UserAgent,
	})
	if err != nil {
		return domain.WrapInternal(err)
	}
	return nil
}

// ListAuditLogsAdmin returns audit logs for the admin panel.
func (s *AuditLogService) ListAuditLogsAdmin(ctx context.Context, actorType *string, actorID *uuid.UUID, action *string, page, pageSize int) ([]AuditLogResponse, int, error) {
	page, pageSize = normalizePage(page, pageSize)
	logs, total, err := s.repo.List(ctx, repository.ListAuditLogFilter{
		ActorType: actorType,
		ActorID:   actorID,
		Action:    action,
		Offset:    (page - 1) * pageSize,
		Limit:     pageSize,
	})
	if err != nil {
		return nil, 0, domain.WrapInternal(err)
	}

	resp := make([]AuditLogResponse, 0, len(logs))
	for _, log := range logs {
		resp = append(resp, *toAuditLogResponse(log))
	}
	return resp, total, nil
}

func toAuditLogResponse(log *ent.AuditLog) *AuditLogResponse {
	resp := &AuditLogResponse{
		ID:        log.ID.String(),
		ActorType: string(log.ActorType),
		ActorID:   log.ActorID.String(),
		Action:    log.Action,
		Details:   log.Details,
		CreatedAt: log.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: log.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if log.TargetType != nil {
		resp.TargetType = log.TargetType
	}
	if log.TargetID != nil {
		s := log.TargetID.String()
		resp.TargetID = &s
	}
	if log.IP != nil {
		resp.IP = log.IP
	}
	if log.UserAgent != nil {
		resp.UserAgent = log.UserAgent
	}
	return resp
}
