package service

import (
	"context"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/repository"
)

// GroupResponse is the public representation of a group.
type GroupResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description *string `json:"description,omitempty"`
	SortOrder   int    `json:"sort_order"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// GroupService handles read-only group queries for admin pickers.
type GroupService struct {
	repo repository.GroupRepository
}

// NewGroupService creates a new GroupService.
func NewGroupService(repo repository.GroupRepository) *GroupService {
	return &GroupService{repo: repo}
}

// ListGroups returns all groups ordered by sort_order.
func (s *GroupService) ListGroups(ctx context.Context) ([]GroupResponse, error) {
	groups, err := s.repo.List(ctx)
	if err != nil {
		return nil, domain.WrapInternal(err)
	}
	resp := make([]GroupResponse, 0, len(groups))
	for _, g := range groups {
		resp = append(resp, *toGroupResponse(g))
	}
	return resp, nil
}

func toGroupResponse(g *ent.Group) *GroupResponse {
	resp := &GroupResponse{
		ID:        g.ID.String(),
		Name:      g.Name,
		Code:      g.Code,
		SortOrder: g.SortOrder,
		Status:    g.Status.String(),
		CreatedAt: g.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: g.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if g.Description != nil {
		resp.Description = g.Description
	}
	return resp
}
