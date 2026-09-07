package service

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/repository"
)

// Well-known setting keys persisted in the settings table.
const (
	SettingKeyFeatureMode       = "feature_mode"
	SettingKeyDefaultQuotaLimit = "default_quota_limit"
	SettingKeyQuotaRange        = "default_quota_range"
	SettingKeyRechargeLimits    = "recharge_limits"
)

// Supported feature modes for the platform.
const (
	// FeatureModeBoth enables recharge orders and quota requests.
	FeatureModeBoth = "both"
	// FeatureModeQuotaOnly disables recharge orders (quota requests only).
	FeatureModeQuotaOnly = "quota_only"
	// FeatureModeRechargeOnly disables quota requests (recharge only).
	FeatureModeRechargeOnly = "recharge_only"
)

// Default values used when a setting has never been written. All amounts are
// micro-currency (1 元 = 1e6).
var (
	// rechargeChannels lists the enabled payment providers, matching the
	// provider enum on recharge orders.
	rechargeChannels = []string{"mock"}
)

// defaultRechargeLimits returns a fresh copy so callers cannot mutate shared
// state through the returned slice.
func defaultRechargeLimits() RechargeLimits {
	return RechargeLimits{
		MinAmount:    1_000_000,
		MaxAmount:    100_000_000,
		QuickAmounts: []int64{1_000_000, 5_000_000, 10_000_000, 50_000_000},
	}
}

// defaultQuotaRange is the allowed range for default and batch-assigned token
// quotas, in tokens (customer requirement: 50万-300万). It applies when no
// range has been stored yet.
func defaultQuotaRange() QuotaRange {
	return QuotaRange{Min: 500_000, Max: 3_000_000}
}

// QuotaRange bounds the quota values an admin may assign through the default
// quota setting and batch quota operations. Explicit per-token quotas set at
// token creation are deliberately not bounded by it.
type QuotaRange struct {
	Min int64 `json:"min"`
	Max int64 `json:"max"`
}

func (r QuotaRange) valid() bool {
	return r.Min > 0 && r.Max >= r.Min
}

// RechargeLimits is the validated recharge configuration.
type RechargeLimits struct {
	MinAmount    int64   `json:"min_amount"`
	MaxAmount    int64   `json:"max_amount"`
	QuickAmounts []int64 `json:"quick_amounts"`
}

// PublicRechargeConfig is the recharge configuration exposed to the user
// portal so it can render the recharge form without hardcoding values.
type PublicRechargeConfig struct {
	MinAmount    int64    `json:"min_amount"`
	MaxAmount    int64    `json:"max_amount"`
	QuickAmounts []int64  `json:"quick_amounts"`
	Channels     []string `json:"channels"`
}

// DefaultQuotaProvider supplies the configured default quota for new tokens.
// Implemented by SettingsService; nil implementations mean "no default".
type DefaultQuotaProvider interface {
	DefaultQuotaLimit(ctx context.Context) (*int64, error)
}

// RechargePolicy supplies the settings needed to gate recharge order creation.
// Implemented by SettingsService.
type RechargePolicy interface {
	FeatureMode(ctx context.Context) (string, error)
	RechargeLimits(ctx context.Context) (*RechargeLimits, error)
}

// QuotaPolicy supplies the feature mode needed to gate quota request creation.
// Implemented by SettingsService.
type QuotaPolicy interface {
	FeatureMode(ctx context.Context) (string, error)
}

// SettingsService exposes typed access to the well-known platform settings.
// Missing settings fall back to documented defaults so the platform works
// before any admin writes configuration; manually corrupted stored values
// also fall back instead of breaking request handling.
type SettingsService struct {
	repo repository.SettingRepository
}

// NewSettingsService creates a new SettingsService.
func NewSettingsService(repo repository.SettingRepository) *SettingsService {
	return &SettingsService{repo: repo}
}

// FeatureMode returns the current feature mode (default: both).
func (s *SettingsService) FeatureMode(ctx context.Context) (string, error) {
	setting, err := s.repo.GetByKey(ctx, SettingKeyFeatureMode)
	if err != nil {
		if ent.IsNotFound(err) {
			return FeatureModeBoth, nil
		}
		return "", domain.WrapInternal(err)
	}
	switch setting.Value {
	case FeatureModeBoth, FeatureModeQuotaOnly, FeatureModeRechargeOnly:
		return setting.Value, nil
	}
	return FeatureModeBoth, nil
}

// SetFeatureMode persists a new feature mode after validating the value.
func (s *SettingsService) SetFeatureMode(ctx context.Context, mode string) error {
	switch mode {
	case FeatureModeBoth, FeatureModeQuotaOnly, FeatureModeRechargeOnly:
	default:
		return domain.NewValidationError(map[string]string{
			"mode": "功能模式必须是 both、quota_only 或 recharge_only",
		})
	}
	_, err := s.repo.Upsert(ctx, repository.UpsertSettingInput{
		Key:       SettingKeyFeatureMode,
		Value:     mode,
		ValueType: "string",
		Description: "平台功能模式：both=充值+配额申请，" +
			"quota_only=仅配额申请（禁用充值），recharge_only=仅充值（禁用配额申请）",
	})
	if err != nil {
		return domain.WrapInternal(err)
	}
	return nil
}

// DefaultQuotaLimit returns the configured default quota for new tokens, or
// nil when unset or non-positive (meaning "no default").
func (s *SettingsService) DefaultQuotaLimit(ctx context.Context) (*int64, error) {
	setting, err := s.repo.GetByKey(ctx, SettingKeyDefaultQuotaLimit)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, domain.WrapInternal(err)
	}
	v, err := strconv.ParseInt(setting.Value, 10, 64)
	if err != nil || v <= 0 {
		return nil, nil
	}
	return &v, nil
}

// SetDefaultQuotaLimit persists the default quota for new tokens; zero
// disables the default. A non-zero value must fall inside the configured
// quota range.
func (s *SettingsService) SetDefaultQuotaLimit(ctx context.Context, value int64) error {
	if value < 0 {
		return domain.NewValidationError(map[string]string{"value": "默认配额不能为负数"})
	}
	if value > 0 {
		rng, err := s.QuotaRange(ctx)
		if err != nil {
			return err
		}
		if value < rng.Min || value > rng.Max {
			return domain.NewValidationError(map[string]string{
				"value": "默认配额必须在额度区间 [" + strconv.FormatInt(rng.Min, 10) +
					", " + strconv.FormatInt(rng.Max, 10) + "] 内",
			})
		}
	}
	_, err := s.repo.Upsert(ctx, repository.UpsertSettingInput{
		Key:         SettingKeyDefaultQuotaLimit,
		Value:       strconv.FormatInt(value, 10),
		ValueType:   "int",
		Description: "新建用户 Token 的默认配额上限（0 表示不设置默认配额）",
	})
	if err != nil {
		return domain.WrapInternal(err)
	}
	return nil
}

// QuotaRange returns the allowed quota range (defaults when unset, when the
// stored JSON cannot be parsed, or when the stored range is inconsistent).
func (s *SettingsService) QuotaRange(ctx context.Context) (*QuotaRange, error) {
	setting, err := s.repo.GetByKey(ctx, SettingKeyQuotaRange)
	if err != nil {
		if ent.IsNotFound(err) {
			rng := defaultQuotaRange()
			return &rng, nil
		}
		return nil, domain.WrapInternal(err)
	}
	var rng QuotaRange
	if err := json.Unmarshal([]byte(setting.Value), &rng); err != nil || !rng.valid() {
		fallback := defaultQuotaRange()
		return &fallback, nil
	}
	return &rng, nil
}

// SetQuotaRange persists the allowed quota range. It is rejected when the
// currently stored default quota would fall outside the new range, so the two
// settings can never contradict each other.
func (s *SettingsService) SetQuotaRange(ctx context.Context, rng QuotaRange) error {
	if !rng.valid() {
		return domain.NewValidationError(map[string]string{
			"min": "额度区间要求 min > 0 且 max ≥ min",
		})
	}
	if current, err := s.DefaultQuotaLimit(ctx); err != nil {
		return err
	} else if current != nil && (*current < rng.Min || *current > rng.Max) {
		return domain.NewValidationError(map[string]string{
			"min": "当前默认配额 " + strconv.FormatInt(*current, 10) +
				" 不在新区间内，请先调整默认配额",
		})
	}
	payload, err := json.Marshal(rng)
	if err != nil {
		return domain.WrapInternal(err)
	}
	_, err = s.repo.Upsert(ctx, repository.UpsertSettingInput{
		Key:         SettingKeyQuotaRange,
		Value:       string(payload),
		ValueType:   "json",
		Description: "额度区间（单位：token）：默认额度与批量设置配额的取值范围",
	})
	if err != nil {
		return domain.WrapInternal(err)
	}
	return nil
}

// RechargeLimits returns the recharge configuration (defaults when unset or
// when the stored JSON cannot be parsed).
func (s *SettingsService) RechargeLimits(ctx context.Context) (*RechargeLimits, error) {
	setting, err := s.repo.GetByKey(ctx, SettingKeyRechargeLimits)
	if err != nil {
		if ent.IsNotFound(err) {
			limits := defaultRechargeLimits()
			return &limits, nil
		}
		return nil, domain.WrapInternal(err)
	}
	var limits RechargeLimits
	if err := json.Unmarshal([]byte(setting.Value), &limits); err != nil {
		fallback := defaultRechargeLimits()
		return &fallback, nil
	}
	return &limits, nil
}

// SetRechargeLimits persists the recharge configuration after validation.
func (s *SettingsService) SetRechargeLimits(ctx context.Context, limits RechargeLimits) error {
	if err := validateRechargeLimits(limits); err != nil {
		return err
	}
	payload, err := json.Marshal(limits)
	if err != nil {
		return domain.WrapInternal(err)
	}
	_, err = s.repo.Upsert(ctx, repository.UpsertSettingInput{
		Key:         SettingKeyRechargeLimits,
		Value:       string(payload),
		ValueType:   "json",
		Description: "充值配置：单笔最小/最大金额与快捷金额列表（单位：微货币）",
	})
	if err != nil {
		return domain.WrapInternal(err)
	}
	return nil
}

// PublicRechargeConfig returns the recharge configuration for the user portal
// together with the list of enabled payment channels.
func (s *SettingsService) PublicRechargeConfig(ctx context.Context) (*PublicRechargeConfig, error) {
	limits, err := s.RechargeLimits(ctx)
	if err != nil {
		return nil, err
	}
	quick := limits.QuickAmounts
	if quick == nil {
		quick = []int64{}
	}
	return &PublicRechargeConfig{
		MinAmount:    limits.MinAmount,
		MaxAmount:    limits.MaxAmount,
		QuickAmounts: quick,
		Channels:     rechargeChannels,
	}, nil
}

func validateRechargeLimits(limits RechargeLimits) error {
	if limits.MinAmount <= 0 {
		return domain.NewValidationError(map[string]string{"min_amount": "最小充值金额必须大于 0"})
	}
	if limits.MaxAmount <= 0 {
		return domain.NewValidationError(map[string]string{"max_amount": "最大充值金额必须大于 0"})
	}
	if limits.MinAmount > limits.MaxAmount {
		return domain.NewValidationError(map[string]string{"min_amount": "最小充值金额不能大于最大充值金额"})
	}
	if len(limits.QuickAmounts) == 0 {
		return domain.NewValidationError(map[string]string{"quick_amounts": "快捷金额列表不能为空"})
	}
	for i, amount := range limits.QuickAmounts {
		if amount <= 0 {
			return domain.NewValidationError(map[string]string{
				"quick_amounts": "快捷金额必须大于 0（第 " + strconv.Itoa(i+1) + " 项）",
			})
		}
	}
	return nil
}
