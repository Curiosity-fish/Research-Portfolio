package handler

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/school-api/school-api-v1/internal/server/respond"
	"github.com/school-api/school-api-v1/internal/service"
)

// SettingsService is the subset of the settings service used by SettingsHandler.
type SettingsService interface {
	FeatureMode(ctx context.Context) (string, error)
	SetFeatureMode(ctx context.Context, mode string) error
	DefaultQuotaLimit(ctx context.Context) (*int64, error)
	SetDefaultQuotaLimit(ctx context.Context, value int64) error
	QuotaRange(ctx context.Context) (*service.QuotaRange, error)
	SetQuotaRange(ctx context.Context, rng service.QuotaRange) error
	RechargeLimits(ctx context.Context) (*service.RechargeLimits, error)
	SetRechargeLimits(ctx context.Context, limits service.RechargeLimits) error
	PublicRechargeConfig(ctx context.Context) (*service.PublicRechargeConfig, error)
}

// SetFeatureModeRequest is the request body for updating the feature mode.
type SetFeatureModeRequest struct {
	Mode string `json:"mode" binding:"required"`
}

// SetDefaultQuotaRequest is the request body for updating the default quota.
// Value uses gte so an explicit 0 (disable the default) is accepted.
type SetDefaultQuotaRequest struct {
	Value int64 `json:"value" binding:"gte=0"`
}

// SetQuotaRangeRequest is the request body for updating the quota range.
// Cross-field consistency (min ≤ max) is validated by the service, mirroring
// the recharge limits handler.
type SetQuotaRangeRequest struct {
	Min int64 `json:"min" binding:"required,gt=0"`
	Max int64 `json:"max" binding:"required,gt=0"`
}

// SetRechargeLimitsRequest is the request body for updating recharge config.
type SetRechargeLimitsRequest struct {
	MinAmount    int64   `json:"min_amount" binding:"required,gt=0"`
	MaxAmount    int64   `json:"max_amount" binding:"required,gt=0"`
	QuickAmounts []int64 `json:"quick_amounts" binding:"required,min=1,dive,gt=0"`
}

// SettingsHandler handles platform settings endpoints.
type SettingsHandler struct {
	svc SettingsService
}

// NewSettingsHandler creates a new SettingsHandler.
func NewSettingsHandler(svc SettingsService) *SettingsHandler {
	return &SettingsHandler{svc: svc}
}

// GetFeatureMode handles GET /api/v1/admin/settings/feature-mode.
func (h *SettingsHandler) GetFeatureMode(c *gin.Context) {
	mode, err := h.svc.FeatureMode(c.Request.Context())
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.OK(c, gin.H{"mode": mode})
}

// PutFeatureMode handles PUT /api/v1/admin/settings/feature-mode.
func (h *SettingsHandler) PutFeatureMode(c *gin.Context) {
	var req SetFeatureModeRequest
	if !bindAndValidate(c, &req) {
		return
	}
	if err := h.svc.SetFeatureMode(c.Request.Context(), req.Mode); err != nil {
		respond.Error(c, err)
		return
	}
	respond.OK(c, gin.H{"mode": req.Mode})
}

// GetDefaultQuota handles GET /api/v1/admin/settings/default-quota.
func (h *SettingsHandler) GetDefaultQuota(c *gin.Context) {
	quota, err := h.svc.DefaultQuotaLimit(c.Request.Context())
	if err != nil {
		respond.Error(c, err)
		return
	}
	value := int64(0)
	if quota != nil {
		value = *quota
	}
	respond.OK(c, gin.H{"value": value})
}

// PutDefaultQuota handles PUT /api/v1/admin/settings/default-quota.
func (h *SettingsHandler) PutDefaultQuota(c *gin.Context) {
	var req SetDefaultQuotaRequest
	if !bindAndValidate(c, &req) {
		return
	}
	if err := h.svc.SetDefaultQuotaLimit(c.Request.Context(), req.Value); err != nil {
		respond.Error(c, err)
		return
	}
	respond.OK(c, gin.H{"value": req.Value})
}

// GetQuotaRange handles GET /api/v1/admin/settings/quota-range.
func (h *SettingsHandler) GetQuotaRange(c *gin.Context) {
	rng, err := h.svc.QuotaRange(c.Request.Context())
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.OK(c, rng)
}

// PutQuotaRange handles PUT /api/v1/admin/settings/quota-range.
func (h *SettingsHandler) PutQuotaRange(c *gin.Context) {
	var req SetQuotaRangeRequest
	if !bindAndValidate(c, &req) {
		return
	}
	rng := service.QuotaRange{Min: req.Min, Max: req.Max}
	if err := h.svc.SetQuotaRange(c.Request.Context(), rng); err != nil {
		respond.Error(c, err)
		return
	}
	respond.OK(c, rng)
}

// GetRechargeLimits handles GET /api/v1/admin/settings/recharge.
func (h *SettingsHandler) GetRechargeLimits(c *gin.Context) {
	limits, err := h.svc.RechargeLimits(c.Request.Context())
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.OK(c, limits)
}

// PutRechargeLimits handles PUT /api/v1/admin/settings/recharge.
func (h *SettingsHandler) PutRechargeLimits(c *gin.Context) {
	var req SetRechargeLimitsRequest
	if !bindAndValidate(c, &req) {
		return
	}
	limits := service.RechargeLimits{
		MinAmount:    req.MinAmount,
		MaxAmount:    req.MaxAmount,
		QuickAmounts: req.QuickAmounts,
	}
	if err := h.svc.SetRechargeLimits(c.Request.Context(), limits); err != nil {
		respond.Error(c, err)
		return
	}
	respond.OK(c, limits)
}

// GetPublicRechargeConfig handles GET /api/v1/recharge/config.
func (h *SettingsHandler) GetPublicRechargeConfig(c *gin.Context) {
	config, err := h.svc.PublicRechargeConfig(c.Request.Context())
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.OK(c, config)
}
