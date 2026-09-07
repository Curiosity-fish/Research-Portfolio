package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	entquotarequest "github.com/school-api/school-api-v1/ent/quotarequest"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/repository"
)

// QuotaRequestResponse is the public representation of a quota request.
type QuotaRequestResponse struct {
	ID              string     `json:"id"`
	UserID          string     `json:"user_id"`
	TokenID         string     `json:"token_id"`
	RequestedAmount int64      `json:"requested_amount"`
	Reason          *string    `json:"reason,omitempty"`
	Status          string     `json:"status"`
	ReviewedBy      *string    `json:"reviewed_by,omitempty"`
	ReviewedAt      *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt       string     `json:"created_at"`
	UpdatedAt       string     `json:"updated_at"`
}

// CreateQuotaRequestInput is the service-level input for creating a request.
type CreateQuotaRequestInput struct {
	UserID          uuid.UUID
	TokenID         uuid.UUID
	RequestedAmount int64
	Reason          *string
}

// ListQuotaRequestFilter is the service-level filter for listing requests.
type ListQuotaRequestFilter struct {
	Status   *string
	Page     int
	PageSize int
}

// QuotaRequestService handles quota request business logic.
type QuotaRequestService struct {
	requestRepo repository.QuotaRequestRepository
	tokenRepo   repository.UserTokenRepository
	settings    QuotaPolicy
}

// NewQuotaRequestService creates a new QuotaRequestService. settings may be
// nil, in which case the feature-mode gate on new requests is skipped.
func NewQuotaRequestService(requestRepo repository.QuotaRequestRepository, tokenRepo repository.UserTokenRepository, settings QuotaPolicy) *QuotaRequestService {
	return &QuotaRequestService{requestRepo: requestRepo, tokenRepo: tokenRepo, settings: settings}
}

// CreateQuotaRequest submits a new quota request after validating that the token
// belongs to the user and that no other pending request exists.
func (s *QuotaRequestService) CreateQuotaRequest(ctx context.Context, input CreateQuotaRequestInput) (*QuotaRequestResponse, error) {
	if input.RequestedAmount <= 0 {
		return nil, domain.NewAppError(400, "INVALID_REQUEST", "申请配额必须大于 0")
	}

	if s.settings != nil {
		mode, err := s.settings.FeatureMode(ctx)
		if err != nil {
			return nil, err
		}
		if mode == FeatureModeRechargeOnly {
			return nil, domain.ErrFeatureDisabled
		}
	}

	tok, err := s.tokenRepo.GetByID(ctx, input.TokenID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrNotFound
		}
		return nil, domain.WrapInternal(err)
	}
	if tok.UserID != input.UserID {
		return nil, domain.ErrNotFound
	}

	pending, err := s.requestRepo.GetPendingByUserID(ctx, input.UserID)
	if err != nil {
		return nil, domain.WrapInternal(err)
	}
	if pending != nil {
		return nil, domain.NewAppError(409, "PENDING_QUOTA_REQUEST_EXISTS", "已存在待审批的配额申请")
	}

	req, err := s.requestRepo.Create(ctx, repository.CreateQuotaRequestInput{
		UserID:          input.UserID,
		TokenID:         input.TokenID,
		RequestedAmount: input.RequestedAmount,
		Reason:          input.Reason,
	})
	if err != nil {
		return nil, domain.WrapInternal(err)
	}
	return toQuotaRequestResponse(req), nil
}

// ListMyQuotaRequests returns the quota requests submitted by the given user.
func (s *QuotaRequestService) ListMyQuotaRequests(ctx context.Context, userID uuid.UUID) ([]QuotaRequestResponse, error) {
	reqs, _, err := s.requestRepo.List(ctx, repository.ListQuotaRequestFilter{
		Page:     1,
		PageSize: 100,
	})
	if err != nil {
		return nil, domain.WrapInternal(err)
	}

	resp := make([]QuotaRequestResponse, 0, len(reqs))
	for _, req := range reqs {
		if req.UserID == userID {
			resp = append(resp, *toQuotaRequestResponse(req))
		}
	}
	return resp, nil
}

// ListQuotaRequests returns a paginated list of quota requests for admins.
func (s *QuotaRequestService) ListQuotaRequests(ctx context.Context, filter ListQuotaRequestFilter) ([]QuotaRequestResponse, int, error) {
	var status *repository.QuotaRequestStatus
	if filter.Status != nil {
		s := repository.QuotaRequestStatus(*filter.Status)
		status = &s
	}

	reqs, total, err := s.requestRepo.List(ctx, repository.ListQuotaRequestFilter{
		Status:   status,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	})
	if err != nil {
		return nil, 0, domain.WrapInternal(err)
	}

	resp := make([]QuotaRequestResponse, 0, len(reqs))
	for _, req := range reqs {
		resp = append(resp, *toQuotaRequestResponse(req))
	}
	return resp, total, nil
}

// ApproveQuotaRequest approves a pending quota request.
func (s *QuotaRequestService) ApproveQuotaRequest(ctx context.Context, requestID, adminID uuid.UUID) (*QuotaRequestResponse, error) {
	req, err := s.requestRepo.Approve(ctx, requestID, adminID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrNotFound
		}
		return nil, domain.WrapInternal(err)
	}
	return toQuotaRequestResponse(req), nil
}

// RejectQuotaRequest rejects a pending quota request.
func (s *QuotaRequestService) RejectQuotaRequest(ctx context.Context, requestID, adminID uuid.UUID) (*QuotaRequestResponse, error) {
	req, err := s.requestRepo.GetByID(ctx, requestID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrNotFound
		}
		return nil, domain.WrapInternal(err)
	}
	if req.Status != entquotarequest.StatusPending {
		return nil, domain.NewAppError(409, "QUOTA_REQUEST_NOT_PENDING", "该申请不处于待审批状态")
	}

	updated, err := s.requestRepo.UpdateStatus(ctx, requestID, repository.UpdateQuotaRequestStatusInput{
		Status:     repository.QuotaRequestStatusRejected,
		ReviewedBy: adminID,
		ReviewedAt: time.Now(),
	})
	if err != nil {
		return nil, domain.WrapInternal(err)
	}
	return toQuotaRequestResponse(updated), nil
}

func toQuotaRequestResponse(req *ent.QuotaRequest) *QuotaRequestResponse {
	resp := &QuotaRequestResponse{
		ID:              req.ID.String(),
		UserID:          req.UserID.String(),
		TokenID:         req.TokenID.String(),
		RequestedAmount: req.RequestedAmount,
		Status:          string(req.Status),
		CreatedAt:       req.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       req.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if req.Reason != nil && *req.Reason != "" {
		resp.Reason = req.Reason
	}
	if req.ReviewedBy != nil {
		s := req.ReviewedBy.String()
		resp.ReviewedBy = &s
	}
	if req.ReviewedAt != nil {
		resp.ReviewedAt = req.ReviewedAt
	}
	return resp
}
