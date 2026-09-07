package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/setting"
	"github.com/school-api/school-api-v1/ent/user"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/repository"
)

// fakeSettingRepo is an in-memory SettingRepository for unit tests.
type fakeSettingRepo struct {
	byKey map[string]repository.UpsertSettingInput
}

func newFakeSettingRepo() *fakeSettingRepo {
	return &fakeSettingRepo{byKey: map[string]repository.UpsertSettingInput{}}
}

func (f *fakeSettingRepo) GetByKey(_ context.Context, key string) (*ent.Setting, error) {
	input, ok := f.byKey[key]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	return &ent.Setting{
		Key:         input.Key,
		Value:       input.Value,
		Type:        setting.Type(input.ValueType),
		Description: input.Description,
		IsPublic:    input.IsPublic,
	}, nil
}

func (f *fakeSettingRepo) Upsert(_ context.Context, input repository.UpsertSettingInput) (*ent.Setting, error) {
	f.byKey[input.Key] = input
	return &ent.Setting{Key: input.Key, Value: input.Value, Type: setting.Type(input.ValueType)}, nil
}

func TestSettingsService_FeatureModeDefaultsAndRoundTrip(t *testing.T) {
	ctx := context.Background()
	svc := NewSettingsService(newFakeSettingRepo())

	mode, err := svc.FeatureMode(ctx)
	if err != nil {
		t.Fatalf("default feature mode: %v", err)
	}
	if mode != FeatureModeBoth {
		t.Fatalf("expected default both, got %s", mode)
	}

	if err := svc.SetFeatureMode(ctx, FeatureModeQuotaOnly); err != nil {
		t.Fatalf("set feature mode: %v", err)
	}
	mode, err = svc.FeatureMode(ctx)
	if err != nil {
		t.Fatalf("get feature mode: %v", err)
	}
	if mode != FeatureModeQuotaOnly {
		t.Fatalf("expected quota_only, got %s", mode)
	}
}

func TestSettingsService_SetFeatureModeRejectsInvalidValue(t *testing.T) {
	svc := NewSettingsService(newFakeSettingRepo())
	err := svc.SetFeatureMode(context.Background(), "everything")
	if err == nil {
		t.Fatal("expected validation error for invalid mode")
	}
	var appErr *domain.AppError
	if !errors.As(err, &appErr) || appErr.Code != 400 {
		t.Fatalf("expected 400 AppError, got %v", err)
	}
}

func TestSettingsService_FeatureModeFallsBackOnCorruptValue(t *testing.T) {
	repo := newFakeSettingRepo()
	repo.byKey[SettingKeyFeatureMode] = repository.UpsertSettingInput{
		Key: SettingKeyFeatureMode, Value: "garbage", ValueType: "string",
	}
	svc := NewSettingsService(repo)
	mode, err := svc.FeatureMode(context.Background())
	if err != nil {
		t.Fatalf("feature mode with corrupt value: %v", err)
	}
	if mode != FeatureModeBoth {
		t.Fatalf("expected fallback both, got %s", mode)
	}
}

func TestSettingsService_DefaultQuotaLimit(t *testing.T) {
	ctx := context.Background()
	svc := NewSettingsService(newFakeSettingRepo())

	// Shrink the range first so small values are assignable in this test.
	if err := svc.SetQuotaRange(ctx, QuotaRange{Min: 1000, Max: 10_000}); err != nil {
		t.Fatalf("set quota range: %v", err)
	}

	quota, err := svc.DefaultQuotaLimit(ctx)
	if err != nil {
		t.Fatalf("default quota: %v", err)
	}
	if quota != nil {
		t.Fatalf("expected nil default quota, got %v", *quota)
	}

	if err := svc.SetDefaultQuotaLimit(ctx, 5000); err != nil {
		t.Fatalf("set default quota: %v", err)
	}
	quota, err = svc.DefaultQuotaLimit(ctx)
	if err != nil {
		t.Fatalf("get default quota: %v", err)
	}
	if quota == nil || *quota != 5000 {
		t.Fatalf("expected 5000, got %v", quota)
	}

	// Values outside the configured range are rejected.
	if err := svc.SetDefaultQuotaLimit(ctx, 999); err == nil {
		t.Fatal("expected validation error below range")
	}
	if err := svc.SetDefaultQuotaLimit(ctx, 10_001); err == nil {
		t.Fatal("expected validation error above range")
	}

	// Zero disables the default again and stays allowed.
	if err := svc.SetDefaultQuotaLimit(ctx, 0); err != nil {
		t.Fatalf("clear default quota: %v", err)
	}
	quota, err = svc.DefaultQuotaLimit(ctx)
	if err != nil {
		t.Fatalf("get cleared default quota: %v", err)
	}
	if quota != nil {
		t.Fatalf("expected nil after clearing, got %v", *quota)
	}

	if err := svc.SetDefaultQuotaLimit(ctx, -1); err == nil {
		t.Fatal("expected validation error for negative quota")
	}
}

func TestSettingsService_QuotaRangeDefaultsAndRoundTrip(t *testing.T) {
	ctx := context.Background()
	svc := NewSettingsService(newFakeSettingRepo())

	// Unset range falls back to the customer-specified 50万-300万.
	rng, err := svc.QuotaRange(ctx)
	if err != nil {
		t.Fatalf("default quota range: %v", err)
	}
	if rng.Min != 500_000 || rng.Max != 3_000_000 {
		t.Fatalf("expected default range 500000-3000000, got %+v", rng)
	}

	custom := QuotaRange{Min: 600_000, Max: 2_000_000}
	if err := svc.SetQuotaRange(ctx, custom); err != nil {
		t.Fatalf("set quota range: %v", err)
	}
	rng, err = svc.QuotaRange(ctx)
	if err != nil {
		t.Fatalf("get quota range: %v", err)
	}
	if *rng != custom {
		t.Fatalf("round trip mismatch: %+v", rng)
	}
}

func TestSettingsService_QuotaRangeFallsBackOnCorruptValue(t *testing.T) {
	repo := newFakeSettingRepo()
	repo.byKey[SettingKeyQuotaRange] = repository.UpsertSettingInput{
		Key: SettingKeyQuotaRange, Value: "{not-json", ValueType: "json",
	}
	rng, err := NewSettingsService(repo).QuotaRange(context.Background())
	if err != nil {
		t.Fatalf("quota range with corrupt value: %v", err)
	}
	if rng.Min != 500_000 || rng.Max != 3_000_000 {
		t.Fatalf("expected fallback range, got %+v", rng)
	}

	// A stored but inconsistent range (min > max) also falls back.
	repo.byKey[SettingKeyQuotaRange] = repository.UpsertSettingInput{
		Key: SettingKeyQuotaRange, Value: `{"min":100,"max":50}`, ValueType: "json",
	}
	rng, err = NewSettingsService(repo).QuotaRange(context.Background())
	if err != nil {
		t.Fatalf("quota range with inverted bounds: %v", err)
	}
	if rng.Min != 500_000 || rng.Max != 3_000_000 {
		t.Fatalf("expected fallback range, got %+v", rng)
	}
}

func TestSettingsService_SetQuotaRangeValidation(t *testing.T) {
	ctx := context.Background()
	svc := NewSettingsService(newFakeSettingRepo())

	cases := []QuotaRange{
		{Min: 0, Max: 100},
		{Min: 100, Max: 0},
		{Min: 200, Max: 100},
	}
	for _, c := range cases {
		if err := svc.SetQuotaRange(ctx, c); err == nil {
			t.Fatalf("expected validation error for %+v", c)
		}
	}
}

func TestSettingsService_SetQuotaRangeRejectsRangeExcludingCurrentDefault(t *testing.T) {
	ctx := context.Background()
	svc := NewSettingsService(newFakeSettingRepo())

	if err := svc.SetQuotaRange(ctx, QuotaRange{Min: 1000, Max: 10_000}); err != nil {
		t.Fatalf("set quota range: %v", err)
	}
	if err := svc.SetDefaultQuotaLimit(ctx, 5000); err != nil {
		t.Fatalf("set default quota: %v", err)
	}

	// Narrowing the range below the stored default must be rejected so the
	// two settings never contradict each other.
	if err := svc.SetQuotaRange(ctx, QuotaRange{Min: 6000, Max: 10_000}); err == nil {
		t.Fatal("expected validation error for range excluding current default")
	}

	// Disabling the default lifts the constraint again.
	if err := svc.SetDefaultQuotaLimit(ctx, 0); err != nil {
		t.Fatalf("clear default quota: %v", err)
	}
	if err := svc.SetQuotaRange(ctx, QuotaRange{Min: 6000, Max: 10_000}); err != nil {
		t.Fatalf("set quota range after clearing default: %v", err)
	}
}

func TestSettingsService_RechargeLimitsRoundTripAndDefaults(t *testing.T) {
	ctx := context.Background()
	svc := NewSettingsService(newFakeSettingRepo())

	limits, err := svc.RechargeLimits(ctx)
	if err != nil {
		t.Fatalf("default limits: %v", err)
	}
	if limits.MinAmount <= 0 || limits.MaxAmount < limits.MinAmount || len(limits.QuickAmounts) == 0 {
		t.Fatalf("expected sane default limits, got %+v", limits)
	}

	custom := RechargeLimits{MinAmount: 2_000_000, MaxAmount: 20_000_000, QuickAmounts: []int64{2_000_000, 10_000_000}}
	if err := svc.SetRechargeLimits(ctx, custom); err != nil {
		t.Fatalf("set limits: %v", err)
	}
	limits, err = svc.RechargeLimits(ctx)
	if err != nil {
		t.Fatalf("get limits: %v", err)
	}
	if limits.MinAmount != custom.MinAmount || limits.MaxAmount != custom.MaxAmount ||
		len(limits.QuickAmounts) != len(custom.QuickAmounts) {
		t.Fatalf("round trip mismatch: %+v", limits)
	}

	// Corrupt stored JSON falls back to defaults instead of failing requests.
	repo := newFakeSettingRepo()
	repo.byKey[SettingKeyRechargeLimits] = repository.UpsertSettingInput{
		Key: SettingKeyRechargeLimits, Value: "{not-json", ValueType: "json",
	}
	fallback, err := NewSettingsService(repo).RechargeLimits(ctx)
	if err != nil {
		t.Fatalf("limits with corrupt value: %v", err)
	}
	if fallback.MinAmount <= 0 {
		t.Fatalf("expected fallback limits, got %+v", fallback)
	}
}

func TestSettingsService_SetRechargeLimitsValidation(t *testing.T) {
	svc := NewSettingsService(newFakeSettingRepo())
	cases := []RechargeLimits{
		{MinAmount: 0, MaxAmount: 10, QuickAmounts: []int64{1}},
		{MinAmount: 20, MaxAmount: 10, QuickAmounts: []int64{1}},
		{MinAmount: 1, MaxAmount: 10, QuickAmounts: nil},
		{MinAmount: 1, MaxAmount: 10, QuickAmounts: []int64{0}},
	}
	for _, c := range cases {
		if err := svc.SetRechargeLimits(context.Background(), c); err == nil {
			t.Fatalf("expected validation error for %+v", c)
		}
	}
}

func TestSettingsService_PublicRechargeConfig(t *testing.T) {
	svc := NewSettingsService(newFakeSettingRepo())
	config, err := svc.PublicRechargeConfig(context.Background())
	if err != nil {
		t.Fatalf("public config: %v", err)
	}
	if len(config.Channels) != 1 || config.Channels[0] != "mock" {
		t.Fatalf("expected mock channel, got %v", config.Channels)
	}
	if config.MinAmount <= 0 || config.MaxAmount < config.MinAmount {
		t.Fatalf("expected sane limits, got %+v", config)
	}
}

// stubRechargePolicy is a hand-rolled RechargePolicy for policy tests.
type stubRechargePolicy struct {
	mode   string
	limits *RechargeLimits
	err    error
}

func (s *stubRechargePolicy) FeatureMode(_ context.Context) (string, error) {
	return s.mode, s.err
}

func (s *stubRechargePolicy) RechargeLimits(_ context.Context) (*RechargeLimits, error) {
	return s.limits, s.err
}

type stubQuotaPolicy struct {
	mode string
	err  error
}

func (s *stubQuotaPolicy) FeatureMode(_ context.Context) (string, error) {
	return s.mode, s.err
}

// unusedRechargeRepo fails the test if any persistence method is called;
// policy checks must run before the order row is written.
type unusedRechargeRepo struct {
	t *testing.T
}

func (unusedRechargeRepo) Create(context.Context, repository.CreateRechargeOrderInput) (*ent.RechargeOrder, error) {
	return nil, errors.New("Create must not be called")
}
func (unusedRechargeRepo) GetByID(context.Context, uuid.UUID) (*ent.RechargeOrder, error) {
	return nil, errors.New("GetByID must not be called")
}
func (unusedRechargeRepo) ListByUserID(context.Context, uuid.UUID, int, int) ([]*ent.RechargeOrder, int, error) {
	return nil, 0, errors.New("ListByUserID must not be called")
}
func (unusedRechargeRepo) List(context.Context, *uuid.UUID, int, int) ([]*ent.RechargeOrder, int, error) {
	return nil, 0, errors.New("List must not be called")
}
func (unusedRechargeRepo) MarkPaid(context.Context, uuid.UUID, string) (*ent.RechargeOrder, error) {
	return nil, errors.New("MarkPaid must not be called")
}
func (unusedRechargeRepo) MarkFailed(context.Context, uuid.UUID) (*ent.RechargeOrder, error) {
	return nil, errors.New("MarkFailed must not be called")
}

func TestRechargeOrderService_FeatureModeQuotaOnlyRejectsCreation(t *testing.T) {
	svc := NewRechargeOrderService(unusedRechargeRepo{}, &stubRechargePolicy{mode: FeatureModeQuotaOnly})
	_, err := svc.CreateRechargeOrder(context.Background(), uuid.New(), 1_000_000)
	appErr := domain.AsAppError(err)
	if appErr == nil || appErr.Code != 403 || appErr.BizCode != "FEATURE_DISABLED" {
		t.Fatalf("expected 403 FEATURE_DISABLED, got %v", err)
	}
}

func TestRechargeOrderService_AmountLimitsEnforced(t *testing.T) {
	limits := &RechargeLimits{MinAmount: 2_000_000, MaxAmount: 10_000_000, QuickAmounts: []int64{5_000_000}}
	policy := &stubRechargePolicy{mode: FeatureModeBoth, limits: limits}
	svc := NewRechargeOrderService(unusedRechargeRepo{}, policy)

	_, err := svc.CreateRechargeOrder(context.Background(), uuid.New(), 1_000_000)
	if appErr := domain.AsAppError(err); appErr == nil || appErr.Code != 400 {
		t.Fatalf("expected 400 below min, got %v", err)
	}

	_, err = svc.CreateRechargeOrder(context.Background(), uuid.New(), 11_000_000)
	if appErr := domain.AsAppError(err); appErr == nil || appErr.Code != 400 {
		t.Fatalf("expected 400 above max, got %v", err)
	}
}

func TestRechargeOrderService_NilPolicySkipsChecks(t *testing.T) {
	repo := &fakeRechargeRepo{order: &ent.RechargeOrder{ID: uuid.New(), Amount: 1}}
	svc := NewRechargeOrderService(repo, nil)
	resp, err := svc.CreateRechargeOrder(context.Background(), uuid.New(), 1)
	if err != nil {
		t.Fatalf("create with nil policy: %v", err)
	}
	if resp.ID != repo.order.ID.String() {
		t.Fatalf("expected order %s, got %s", repo.order.ID, resp.ID)
	}
}

type fakeRechargeRepo struct {
	order *ent.RechargeOrder
}

func (f *fakeRechargeRepo) Create(_ context.Context, _ repository.CreateRechargeOrderInput) (*ent.RechargeOrder, error) {
	return f.order, nil
}
func (fakeRechargeRepo) GetByID(context.Context, uuid.UUID) (*ent.RechargeOrder, error) {
	return nil, errors.New("not implemented")
}
func (fakeRechargeRepo) ListByUserID(context.Context, uuid.UUID, int, int) ([]*ent.RechargeOrder, int, error) {
	return nil, 0, errors.New("not implemented")
}
func (fakeRechargeRepo) List(context.Context, *uuid.UUID, int, int) ([]*ent.RechargeOrder, int, error) {
	return nil, 0, errors.New("not implemented")
}
func (fakeRechargeRepo) MarkPaid(context.Context, uuid.UUID, string) (*ent.RechargeOrder, error) {
	return nil, errors.New("not implemented")
}
func (fakeRechargeRepo) MarkFailed(context.Context, uuid.UUID) (*ent.RechargeOrder, error) {
	return nil, errors.New("not implemented")
}

// unusedQuotaRepos fails the test if the quota request is persisted while the
// feature mode forbids it; the settings check runs before any repo access.
type unusedQuotaRequestRepo struct{}

func (unusedQuotaRequestRepo) Create(context.Context, repository.CreateQuotaRequestInput) (*ent.QuotaRequest, error) {
	return nil, errors.New("Create must not be called")
}
func (unusedQuotaRequestRepo) GetByID(context.Context, uuid.UUID) (*ent.QuotaRequest, error) {
	return nil, errors.New("GetByID must not be called")
}
func (unusedQuotaRequestRepo) GetPendingByUserID(context.Context, uuid.UUID) (*ent.QuotaRequest, error) {
	return nil, errors.New("GetPendingByUserID must not be called")
}
func (unusedQuotaRequestRepo) List(context.Context, repository.ListQuotaRequestFilter) ([]*ent.QuotaRequest, int, error) {
	return nil, 0, errors.New("List must not be called")
}
func (unusedQuotaRequestRepo) UpdateStatus(context.Context, uuid.UUID, repository.UpdateQuotaRequestStatusInput) (*ent.QuotaRequest, error) {
	return nil, errors.New("UpdateStatus must not be called")
}
func (unusedQuotaRequestRepo) Approve(context.Context, uuid.UUID, uuid.UUID) (*ent.QuotaRequest, error) {
	return nil, errors.New("Approve must not be called")
}

type unusedUserTokenRepo struct{}

func (unusedUserTokenRepo) Create(context.Context, repository.CreateUserTokenInput) (*ent.UserToken, error) {
	return nil, errors.New("Create must not be called")
}
func (unusedUserTokenRepo) ListByUserID(context.Context, uuid.UUID) ([]*ent.UserToken, error) {
	return nil, errors.New("ListByUserID must not be called")
}
func (unusedUserTokenRepo) GetByID(context.Context, uuid.UUID) (*ent.UserToken, error) {
	return nil, errors.New("GetByID must not be called")
}
func (unusedUserTokenRepo) Delete(context.Context, uuid.UUID) error {
	return errors.New("Delete must not be called")
}
func (unusedUserTokenRepo) UpdateStatus(context.Context, uuid.UUID, bool) (*ent.UserToken, error) {
	return nil, errors.New("UpdateStatus must not be called")
}
func (unusedUserTokenRepo) AddQuotaLimit(context.Context, uuid.UUID, int64) (*ent.UserToken, error) {
	return nil, errors.New("AddQuotaLimit must not be called")
}

func TestQuotaRequestService_FeatureModeRechargeOnlyRejectsCreation(t *testing.T) {
	svc := NewQuotaRequestService(unusedQuotaRequestRepo{}, unusedUserTokenRepo{}, &stubQuotaPolicy{mode: FeatureModeRechargeOnly})
	_, err := svc.CreateQuotaRequest(context.Background(), CreateQuotaRequestInput{
		UserID:          uuid.New(),
		TokenID:         uuid.New(),
		RequestedAmount: 100,
	})
	appErr := domain.AsAppError(err)
	if appErr == nil || appErr.Code != 403 || appErr.BizCode != "FEATURE_DISABLED" {
		t.Fatalf("expected 403 FEATURE_DISABLED, got %v", err)
	}
}

// stubDefaultQuota implements DefaultQuotaProvider with a fixed value.
type stubDefaultQuota struct {
	value *int64
}

func (s *stubDefaultQuota) DefaultQuotaLimit(context.Context) (*int64, error) {
	return s.value, nil
}

type captureTokenRepo struct {
	created repository.CreateUserTokenInput
}

func (c *captureTokenRepo) Create(_ context.Context, input repository.CreateUserTokenInput) (*ent.UserToken, error) {
	c.created = input
	return &ent.UserToken{ID: uuid.New()}, nil
}
func (captureTokenRepo) ListByUserID(context.Context, uuid.UUID) ([]*ent.UserToken, error) {
	return nil, errors.New("not implemented")
}
func (captureTokenRepo) GetByID(context.Context, uuid.UUID) (*ent.UserToken, error) {
	return nil, errors.New("not implemented")
}
func (captureTokenRepo) Delete(context.Context, uuid.UUID) error {
	return errors.New("not implemented")
}
func (captureTokenRepo) UpdateStatus(context.Context, uuid.UUID, bool) (*ent.UserToken, error) {
	return nil, errors.New("not implemented")
}
func (captureTokenRepo) AddQuotaLimit(context.Context, uuid.UUID, int64) (*ent.UserToken, error) {
	return nil, errors.New("not implemented")
}

type fakeUserRepoForToken struct{}

func (fakeUserRepoForToken) Create(context.Context, repository.CreateUserInput) (*ent.User, error) {
	return nil, errors.New("not implemented")
}
func (fakeUserRepoForToken) GetByID(context.Context, uuid.UUID) (*ent.User, error) {
	return &ent.User{ID: uuid.New()}, nil
}
func (fakeUserRepoForToken) List(context.Context, repository.ListUserFilter) ([]*ent.User, int, error) {
	return nil, 0, errors.New("not implemented")
}
func (fakeUserRepoForToken) Update(context.Context, uuid.UUID, repository.UpdateUserInput) (*ent.User, error) {
	return nil, errors.New("not implemented")
}
func (fakeUserRepoForToken) UpdateStatus(context.Context, uuid.UUID, user.Status) (*ent.User, error) {
	return nil, errors.New("not implemented")
}
func (fakeUserRepoForToken) GetByUsername(context.Context, string) (*ent.User, error) {
	return nil, errors.New("not implemented")
}

func TestUserTokenService_DefaultQuotaAppliedWhenNotSpecified(t *testing.T) {
	defaultQuota := int64(7777)
	repo := &captureTokenRepo{}
	svc := NewUserTokenService(repo, fakeUserRepoForToken{}, nil, &stubDefaultQuota{value: &defaultQuota})

	_, _, err := svc.CreateUserToken(context.Background(), CreateUserTokenInput{UserID: uuid.New(), Name: "t"})
	if err != nil {
		t.Fatalf("create token: %v", err)
	}
	if repo.created.QuotaLimit == nil || *repo.created.QuotaLimit != defaultQuota {
		t.Fatalf("expected default quota %d, got %v", defaultQuota, repo.created.QuotaLimit)
	}
}

func TestUserTokenService_ExplicitQuotaNotOverridden(t *testing.T) {
	defaultQuota := int64(7777)
	repo := &captureTokenRepo{}
	svc := NewUserTokenService(repo, fakeUserRepoForToken{}, nil, &stubDefaultQuota{value: &defaultQuota})

	explicit := int64(42)
	_, _, err := svc.CreateUserToken(context.Background(), CreateUserTokenInput{
		UserID: uuid.New(), Name: "t", QuotaLimit: &explicit,
	})
	if err != nil {
		t.Fatalf("create token: %v", err)
	}
	if repo.created.QuotaLimit == nil || *repo.created.QuotaLimit != explicit {
		t.Fatalf("expected explicit quota %d, got %v", explicit, repo.created.QuotaLimit)
	}
}

func TestUserTokenService_NilSettingsLeavesQuotaUnset(t *testing.T) {
	repo := &captureTokenRepo{}
	svc := NewUserTokenService(repo, fakeUserRepoForToken{}, nil, nil)

	_, _, err := svc.CreateUserToken(context.Background(), CreateUserTokenInput{UserID: uuid.New(), Name: "t"})
	if err != nil {
		t.Fatalf("create token: %v", err)
	}
	if repo.created.QuotaLimit != nil {
		t.Fatalf("expected nil quota, got %v", *repo.created.QuotaLimit)
	}
}
