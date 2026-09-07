package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/account"
	"github.com/school-api/school-api-v1/ent/platform"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/encryption"
	"github.com/school-api/school-api-v1/internal/relay"
	"github.com/school-api/school-api-v1/internal/repository"
)

type mockRelayAIModelRepo struct {
	getByNameFunc func(ctx context.Context, name string) (*ent.AIModel, error)
	listFunc      func(ctx context.Context) ([]*ent.AIModel, error)
}

func (m *mockRelayAIModelRepo) Create(ctx context.Context, input repository.CreateAIModelInput) (*ent.AIModel, error) {
	return nil, nil
}
func (m *mockRelayAIModelRepo) GetByID(ctx context.Context, id uuid.UUID) (*ent.AIModel, error) {
	return nil, nil
}
func (m *mockRelayAIModelRepo) GetByName(ctx context.Context, name string) (*ent.AIModel, error) {
	return m.getByNameFunc(ctx, name)
}
func (m *mockRelayAIModelRepo) List(ctx context.Context) ([]*ent.AIModel, error) {
	return m.listFunc(ctx)
}
func (m *mockRelayAIModelRepo) Update(ctx context.Context, id uuid.UUID, input repository.UpdateAIModelInput) (*ent.AIModel, error) {
	return nil, nil
}
func (m *mockRelayAIModelRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }

type mockRelayGroupPlatformRepo struct {
	listPlatformsByGroupIDFunc func(ctx context.Context, groupID uuid.UUID) ([]uuid.UUID, error)
}

func (m *mockRelayGroupPlatformRepo) Bind(ctx context.Context, groupID, platformID uuid.UUID) (*ent.GroupPlatform, error) {
	return nil, nil
}
func (m *mockRelayGroupPlatformRepo) Unbind(ctx context.Context, groupID, platformID uuid.UUID) error {
	return nil
}
func (m *mockRelayGroupPlatformRepo) ListPlatformsByGroupID(ctx context.Context, groupID uuid.UUID) ([]uuid.UUID, error) {
	return m.listPlatformsByGroupIDFunc(ctx, groupID)
}
func (m *mockRelayGroupPlatformRepo) ListGroupsByPlatformID(ctx context.Context, platformID uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}

type mockRelayPlatformRepo struct {
	getByIDFunc func(ctx context.Context, id uuid.UUID) (*ent.Platform, error)
}

func (m *mockRelayPlatformRepo) Create(ctx context.Context, input repository.CreatePlatformInput) (*ent.Platform, error) {
	return nil, nil
}
func (m *mockRelayPlatformRepo) GetByID(ctx context.Context, id uuid.UUID) (*ent.Platform, error) {
	return m.getByIDFunc(ctx, id)
}
func (m *mockRelayPlatformRepo) GetByCode(ctx context.Context, code string) (*ent.Platform, error) {
	return nil, nil
}
func (m *mockRelayPlatformRepo) List(ctx context.Context) ([]*ent.Platform, error) { return nil, nil }
func (m *mockRelayPlatformRepo) Update(ctx context.Context, id uuid.UUID, input repository.UpdatePlatformInput) (*ent.Platform, error) {
	return nil, nil
}
func (m *mockRelayPlatformRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }

type mockRelayAccountRepo struct {
	listByPlatformIDFunc    func(ctx context.Context, platformID uuid.UUID) ([]*ent.Account, error)
	incrementErrorCountFunc func(ctx context.Context, id uuid.UUID) error
}

func (m *mockRelayAccountRepo) Create(ctx context.Context, input repository.CreateAccountInput) (*ent.Account, error) {
	return nil, nil
}
func (m *mockRelayAccountRepo) GetByID(ctx context.Context, id uuid.UUID) (*ent.Account, error) {
	return nil, nil
}
func (m *mockRelayAccountRepo) ListByPlatformID(ctx context.Context, platformID uuid.UUID) ([]*ent.Account, error) {
	return m.listByPlatformIDFunc(ctx, platformID)
}
func (m *mockRelayAccountRepo) List(ctx context.Context) ([]*ent.Account, error) { return nil, nil }
func (m *mockRelayAccountRepo) Update(ctx context.Context, id uuid.UUID, input repository.UpdateAccountInput) (*ent.Account, error) {
	return nil, nil
}
func (m *mockRelayAccountRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockRelayAccountRepo) IncrementErrorCount(ctx context.Context, id uuid.UUID) error {
	if m.incrementErrorCountFunc != nil {
		return m.incrementErrorCountFunc(ctx, id)
	}
	return nil
}

type mockRelayCallLogRepo struct {
	createFunc func(ctx context.Context, input repository.CreateCallLogInput) (*ent.CallLog, error)
}

func (m *mockRelayCallLogRepo) Create(ctx context.Context, input repository.CreateCallLogInput) (*ent.CallLog, error) {
	return m.createFunc(ctx, input)
}

// DeleteOlderThan satisfies the widened CallLogRepository interface; the relay
// path never prunes logs, so the stub just records nothing.
func (m *mockRelayCallLogRepo) DeleteOlderThan(context.Context, time.Time) (int, error) {
	return 0, nil
}

type mockRelayAccountService struct {
	decryptAPIKeyFunc func(ctx context.Context, accountID uuid.UUID) (string, error)
}

func (m *mockRelayAccountService) CreateAccount(ctx context.Context, input CreateAccountInput) (*AccountResponse, error) {
	return nil, nil
}
func (m *mockRelayAccountService) ListAccounts(ctx context.Context) ([]AccountResponse, error) {
	return nil, nil
}
func (m *mockRelayAccountService) ListAccountsByPlatform(ctx context.Context, platformID uuid.UUID) ([]AccountResponse, error) {
	return nil, nil
}
func (m *mockRelayAccountService) GetAccount(ctx context.Context, id uuid.UUID) (*AccountResponse, error) {
	return nil, nil
}
func (m *mockRelayAccountService) UpdateAccount(ctx context.Context, id uuid.UUID, input UpdateAccountInput) (*AccountResponse, error) {
	return nil, nil
}
func (m *mockRelayAccountService) DeleteAccount(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockRelayAccountService) DecryptAPIKey(ctx context.Context, accountID uuid.UUID) (string, error) {
	return m.decryptAPIKeyFunc(ctx, accountID)
}

type mockForwarder struct {
	forwardFunc func(ctx context.Context, req *http.Request) (*http.Response, error)
}

func (m *mockForwarder) Forward(ctx context.Context, req *http.Request) (*http.Response, error) {
	return m.forwardFunc(ctx, req)
}

type mockBillingService struct {
	preDeductFunc func(ctx context.Context, userID, tokenID uuid.UUID, model *ent.AIModel, requestBody []byte) (*repository.BillingHold, error)
	settleFunc    func(ctx context.Context, hold *repository.BillingHold, model *ent.AIModel, usage RelayUsage, callLogID uuid.UUID) error
	refundFunc    func(ctx context.Context, hold *repository.BillingHold, callLogID *uuid.UUID) error
}

func (m *mockBillingService) PreDeduct(ctx context.Context, userID, tokenID uuid.UUID, model *ent.AIModel, requestBody []byte) (*repository.BillingHold, error) {
	if m.preDeductFunc != nil {
		return m.preDeductFunc(ctx, userID, tokenID, model, requestBody)
	}
	return &repository.BillingHold{UserID: userID, TokenID: tokenID}, nil
}

func (m *mockBillingService) Settle(ctx context.Context, hold *repository.BillingHold, model *ent.AIModel, usage RelayUsage, callLogID uuid.UUID) error {
	if m.settleFunc != nil {
		return m.settleFunc(ctx, hold, model, usage, callLogID)
	}
	return nil
}

func (m *mockBillingService) Refund(ctx context.Context, hold *repository.BillingHold, callLogID *uuid.UUID) error {
	if m.refundFunc != nil {
		return m.refundFunc(ctx, hold, callLogID)
	}
	return nil
}

func TestRelayService_ChatCompletions(t *testing.T) {
	key := make([]byte, 32)
	platformID := uuid.New()
	accountID := uuid.New()
	groupID := uuid.New()

	modelRepo := &mockRelayAIModelRepo{
		getByNameFunc: func(ctx context.Context, name string) (*ent.AIModel, error) {
			return &ent.AIModel{Name: "gpt-4o", UpstreamName: "gpt-4o-2024-08-06", IsEnabled: true}, nil
		},
	}
	groupPlatformRepo := &mockRelayGroupPlatformRepo{
		listPlatformsByGroupIDFunc: func(ctx context.Context, gid uuid.UUID) ([]uuid.UUID, error) {
			return []uuid.UUID{platformID}, nil
		},
	}
	platformRepo := &mockRelayPlatformRepo{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*ent.Platform, error) {
			return &ent.Platform{ID: platformID, BaseURL: "https://api.openai.com/v1", Type: platform.TypeOpenai, Status: platform.StatusActive}, nil
		},
	}
	accountRepo := &mockRelayAccountRepo{
		listByPlatformIDFunc: func(ctx context.Context, pid uuid.UUID) ([]*ent.Account, error) {
			return []*ent.Account{
				{ID: accountID, PlatformID: platformID, Status: account.StatusActive, Weight: 1},
			}, nil
		},
	}
	callLogRepo := &mockRelayCallLogRepo{
		createFunc: func(ctx context.Context, input repository.CreateCallLogInput) (*ent.CallLog, error) {
			if input.UserID == uuid.Nil {
				t.Error("expected user_id")
			}
			if input.AccountID != accountID {
				t.Errorf("expected account_id %s, got %s", accountID, input.AccountID)
			}
			if input.PromptTokens != 5 {
				t.Errorf("expected prompt_tokens 5, got %d", input.PromptTokens)
			}
			if input.TotalTokens != 10 {
				t.Errorf("expected total_tokens 10, got %d", input.TotalTokens)
			}
			return &ent.CallLog{}, nil
		},
	}
	accountService := &mockRelayAccountService{
		decryptAPIKeyFunc: func(ctx context.Context, id uuid.UUID) (string, error) {
			return "sk-upstream", nil
		},
	}
	forwarder := &mockForwarder{
		forwardFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
			body, _ := io.ReadAll(req.Body)
			if !bytes.Contains(body, []byte(`"gpt-4o-2024-08-06"`)) {
				t.Errorf("expected upstream model name in body, got %s", string(body))
			}
			if req.Header.Get("Authorization") != "Bearer sk-upstream" {
				t.Errorf("unexpected authorization: %s", req.Header.Get("Authorization"))
			}
			return &http.Response{
				StatusCode: 200,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"usage":{"prompt_tokens":5,"completion_tokens":5,"total_tokens":10}}`))),
			}, nil
		},
	}

	svc := NewRelayService(modelRepo, groupPlatformRepo, platformRepo, accountRepo, callLogRepo, accountService, &mockBillingService{}, forwarder, key, 50*1024*1024, 2)
	result, err := svc.ChatCompletions(context.Background(), RelayInput{
		UserID:      uuid.New(),
		TokenID:     uuid.New(),
		GroupID:     groupID,
		RequestBody: []byte(`{"model":"gpt-4o","messages":[{"role":"user","content":"hi"}]}`),
	})
	if err != nil {
		t.Fatalf("chat completions: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("expected 200, got %d", result.StatusCode)
	}
	if result.Usage.TotalTokens != 10 {
		t.Errorf("expected total tokens 10, got %d", result.Usage.TotalTokens)
	}
}

func TestRelayService_ModelNotFound(t *testing.T) {
	modelRepo := &mockRelayAIModelRepo{
		getByNameFunc: func(ctx context.Context, name string) (*ent.AIModel, error) {
			return nil, &ent.NotFoundError{}
		},
	}
	svc := NewRelayService(modelRepo, nil, nil, nil, nil, nil, &mockBillingService{}, nil, make([]byte, 32), 50*1024*1024, 2)
	_, err := svc.ChatCompletions(context.Background(), RelayInput{
		UserID:      uuid.New(),
		TokenID:     uuid.New(),
		GroupID:     uuid.New(),
		RequestBody: []byte(`{"model":"unknown","messages":[]}`),
	})
	appErr := domain.AsAppError(err)
	if appErr.Code != 400 || appErr.BizCode != "model_not_found" {
		t.Errorf("expected model_not_found 400, got %v", appErr)
	}
}

func TestRelayService_NoAvailableAccount(t *testing.T) {
	platformID := uuid.New()
	modelRepo := &mockRelayAIModelRepo{
		getByNameFunc: func(ctx context.Context, name string) (*ent.AIModel, error) {
			return &ent.AIModel{Name: "gpt-4o", UpstreamName: "gpt-4o-2024-08-06", IsEnabled: true}, nil
		},
	}
	groupPlatformRepo := &mockRelayGroupPlatformRepo{
		listPlatformsByGroupIDFunc: func(ctx context.Context, gid uuid.UUID) ([]uuid.UUID, error) {
			return []uuid.UUID{platformID}, nil
		},
	}
	platformRepo := &mockRelayPlatformRepo{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*ent.Platform, error) {
			return &ent.Platform{ID: platformID, BaseURL: "https://api.openai.com/v1", Type: platform.TypeOpenai, Status: platform.StatusActive}, nil
		},
	}
	accountRepo := &mockRelayAccountRepo{
		listByPlatformIDFunc: func(ctx context.Context, pid uuid.UUID) ([]*ent.Account, error) {
			return []*ent.Account{}, nil
		},
	}

	svc := NewRelayService(modelRepo, groupPlatformRepo, platformRepo, accountRepo, nil, nil, &mockBillingService{}, nil, make([]byte, 32), 50*1024*1024, 2)
	_, err := svc.ChatCompletions(context.Background(), RelayInput{
		UserID:      uuid.New(),
		TokenID:     uuid.New(),
		GroupID:     uuid.New(),
		RequestBody: []byte(`{"model":"gpt-4o","messages":[]}`),
	})
	appErr := domain.AsAppError(err)
	if appErr.Code != 503 || appErr.BizCode != "service_unavailable" {
		t.Errorf("expected service_unavailable 503, got %v", appErr)
	}
}

func TestRelayService_ListModels(t *testing.T) {
	platformID := uuid.New()
	groupID := uuid.New()

	modelRepo := &mockRelayAIModelRepo{
		listFunc: func(ctx context.Context) ([]*ent.AIModel, error) {
			return []*ent.AIModel{
				{Name: "gpt-4o", IsEnabled: true},
				{Name: "gpt-3.5-turbo", IsEnabled: false},
			}, nil
		},
	}
	groupPlatformRepo := &mockRelayGroupPlatformRepo{
		listPlatformsByGroupIDFunc: func(ctx context.Context, gid uuid.UUID) ([]uuid.UUID, error) {
			return []uuid.UUID{platformID}, nil
		},
	}
	platformRepo := &mockRelayPlatformRepo{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*ent.Platform, error) {
			return &ent.Platform{ID: platformID, Type: platform.TypeOpenai, Status: platform.StatusActive}, nil
		},
	}

	svc := NewRelayService(modelRepo, groupPlatformRepo, platformRepo, nil, nil, nil, &mockBillingService{}, nil, nil, 0, 0)
	models, err := svc.ListModels(context.Background(), groupID)
	if err != nil {
		t.Fatalf("list models: %v", err)
	}
	if len(models) != 1 || models[0].ID != "gpt-4o" {
		t.Errorf("expected 1 enabled model gpt-4o, got %v", models)
	}
}

func TestRelayService_ListModelsNoActivePlatform(t *testing.T) {
	platformID := uuid.New()
	groupID := uuid.New()

	groupPlatformRepo := &mockRelayGroupPlatformRepo{
		listPlatformsByGroupIDFunc: func(ctx context.Context, gid uuid.UUID) ([]uuid.UUID, error) {
			return []uuid.UUID{platformID}, nil
		},
	}
	platformRepo := &mockRelayPlatformRepo{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*ent.Platform, error) {
			return &ent.Platform{ID: platformID, Type: platform.TypeAnthropic, Status: platform.StatusInactive}, nil
		},
	}

	svc := NewRelayService(nil, groupPlatformRepo, platformRepo, nil, nil, nil, &mockBillingService{}, nil, nil, 0, 0)
	models, err := svc.ListModels(context.Background(), groupID)
	if err != nil {
		t.Fatalf("list models: %v", err)
	}
	if len(models) != 0 {
		t.Errorf("expected empty models, got %d", len(models))
	}
}

func TestRelayService_ForwardFailure(t *testing.T) {
	key := make([]byte, 32)
	platformID := uuid.New()
	accountID := uuid.New()

	modelRepo := &mockRelayAIModelRepo{
		getByNameFunc: func(ctx context.Context, name string) (*ent.AIModel, error) {
			return &ent.AIModel{Name: "gpt-4o", UpstreamName: "gpt-4o-2024-08-06", IsEnabled: true}, nil
		},
	}
	groupPlatformRepo := &mockRelayGroupPlatformRepo{
		listPlatformsByGroupIDFunc: func(ctx context.Context, gid uuid.UUID) ([]uuid.UUID, error) {
			return []uuid.UUID{platformID}, nil
		},
	}
	platformRepo := &mockRelayPlatformRepo{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*ent.Platform, error) {
			return &ent.Platform{ID: platformID, BaseURL: "https://api.openai.com/v1", Type: platform.TypeOpenai, Status: platform.StatusActive}, nil
		},
	}
	accountRepo := &mockRelayAccountRepo{
		listByPlatformIDFunc: func(ctx context.Context, pid uuid.UUID) ([]*ent.Account, error) {
			return []*ent.Account{{ID: accountID, PlatformID: platformID, Status: account.StatusActive, Weight: 1}}, nil
		},
	}
	accountService := &mockRelayAccountService{
		decryptAPIKeyFunc: func(ctx context.Context, id uuid.UUID) (string, error) {
			return "sk-upstream", nil
		},
	}
	forwarder := &mockForwarder{
		forwardFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
			return nil, errors.New("network error")
		},
	}

	svc := NewRelayService(modelRepo, groupPlatformRepo, platformRepo, accountRepo, nil, accountService, &mockBillingService{}, forwarder, key, 50*1024*1024, 2)
	_, err := svc.ChatCompletions(context.Background(), RelayInput{
		UserID:      uuid.New(),
		TokenID:     uuid.New(),
		GroupID:     uuid.New(),
		RequestBody: []byte(`{"model":"gpt-4o","messages":[]}`),
	})
	appErr := domain.AsAppError(err)
	if appErr.Code != 504 {
		t.Errorf("expected gateway_timeout 504, got %v", appErr)
	}
}

func TestRelayService_DecryptFailure(t *testing.T) {
	key := make([]byte, 32)
	platformID := uuid.New()
	accountID := uuid.New()

	modelRepo := &mockRelayAIModelRepo{
		getByNameFunc: func(ctx context.Context, name string) (*ent.AIModel, error) {
			return &ent.AIModel{Name: "gpt-4o", UpstreamName: "gpt-4o-2024-08-06", IsEnabled: true}, nil
		},
	}
	groupPlatformRepo := &mockRelayGroupPlatformRepo{
		listPlatformsByGroupIDFunc: func(ctx context.Context, gid uuid.UUID) ([]uuid.UUID, error) {
			return []uuid.UUID{platformID}, nil
		},
	}
	platformRepo := &mockRelayPlatformRepo{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*ent.Platform, error) {
			return &ent.Platform{ID: platformID, BaseURL: "https://api.openai.com/v1", Type: platform.TypeOpenai, Status: platform.StatusActive}, nil
		},
	}
	accountRepo := &mockRelayAccountRepo{
		listByPlatformIDFunc: func(ctx context.Context, pid uuid.UUID) ([]*ent.Account, error) {
			return []*ent.Account{{ID: accountID, PlatformID: platformID, Status: account.StatusActive, Weight: 1}}, nil
		},
	}
	accountService := &mockRelayAccountService{
		decryptAPIKeyFunc: func(ctx context.Context, id uuid.UUID) (string, error) {
			return "", errors.New("decrypt failed")
		},
	}

	svc := NewRelayService(modelRepo, groupPlatformRepo, platformRepo, accountRepo, nil, accountService, &mockBillingService{}, nil, key, 50*1024*1024, 2)
	_, err := svc.ChatCompletions(context.Background(), RelayInput{
		UserID:      uuid.New(),
		TokenID:     uuid.New(),
		GroupID:     uuid.New(),
		RequestBody: []byte(`{"model":"gpt-4o","messages":[]}`),
	})
	appErr := domain.AsAppError(err)
	if appErr.Code != 503 {
		t.Errorf("expected service_unavailable 503, got %v", appErr)
	}
}

func TestRelayService_ChatCompletionsStream(t *testing.T) {
	key := make([]byte, 32)
	platformID := uuid.New()
	accountID := uuid.New()
	groupID := uuid.New()
	userID := uuid.New()
	tokenID := uuid.New()

	modelRepo := &mockRelayAIModelRepo{
		getByNameFunc: func(ctx context.Context, name string) (*ent.AIModel, error) {
			return &ent.AIModel{Name: "gpt-4o", UpstreamName: "gpt-4o-2024-08-06", IsEnabled: true}, nil
		},
	}
	groupPlatformRepo := &mockRelayGroupPlatformRepo{
		listPlatformsByGroupIDFunc: func(ctx context.Context, gid uuid.UUID) ([]uuid.UUID, error) {
			return []uuid.UUID{platformID}, nil
		},
	}
	platformRepo := &mockRelayPlatformRepo{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*ent.Platform, error) {
			return &ent.Platform{ID: platformID, BaseURL: "https://api.openai.com/v1", Type: platform.TypeOpenai, Status: platform.StatusActive}, nil
		},
	}
	accountRepo := &mockRelayAccountRepo{
		listByPlatformIDFunc: func(ctx context.Context, pid uuid.UUID) ([]*ent.Account, error) {
			return []*ent.Account{{ID: accountID, PlatformID: platformID, Status: account.StatusActive, Weight: 1}}, nil
		},
	}
	accountService := &mockRelayAccountService{
		decryptAPIKeyFunc: func(ctx context.Context, id uuid.UUID) (string, error) {
			return "sk-upstream", nil
		},
	}
	forwarder := &mockForwarder{
		forwardFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
			body, _ := io.ReadAll(req.Body)
			if !bytes.Contains(body, []byte(`"gpt-4o-2024-08-06"`)) {
				t.Errorf("expected upstream model name in body, got %s", string(body))
			}
			return &http.Response{
				StatusCode: 200,
				Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
				Body:       io.NopCloser(bytes.NewReader([]byte("data: {\"usage\":{\"prompt_tokens\":1,\"completion_tokens\":2,\"total_tokens\":3}}\n\ndata: [DONE]\n\n"))),
			}, nil
		},
	}

	var logged *repository.CreateCallLogInput
	callLogRepo := &mockRelayCallLogRepo{
		createFunc: func(ctx context.Context, input repository.CreateCallLogInput) (*ent.CallLog, error) {
			logged = &input
			return &ent.CallLog{}, nil
		},
	}

	svc := NewRelayService(modelRepo, groupPlatformRepo, platformRepo, accountRepo, callLogRepo, accountService, &mockBillingService{}, forwarder, key, 50*1024*1024, 2)
	rec := httptest.NewRecorder()
	err := svc.ChatCompletionsStream(context.Background(), RelayInput{
		UserID:      userID,
		TokenID:     tokenID,
		GroupID:     groupID,
		RequestBody: []byte(`{"model":"gpt-4o","stream":true,"messages":[]}`),
	}, rec)
	if err != nil {
		t.Fatalf("stream: %v", err)
	}

	if rec.Code != 200 {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("[DONE]")) {
		t.Errorf("expected [DONE] in body, got %s", rec.Body.String())
	}
	if logged == nil {
		t.Fatal("expected call log to be created")
	}
	if logged.UserID != userID {
		t.Errorf("expected user_id %s, got %s", userID, logged.UserID)
	}
	if logged.TotalTokens != 3 {
		t.Errorf("expected total tokens 3, got %d", logged.TotalTokens)
	}
}

func TestRelayService_Messages(t *testing.T) {
	key := make([]byte, 32)
	platformID := uuid.New()
	accountID := uuid.New()

	modelRepo := &mockRelayAIModelRepo{
		getByNameFunc: func(ctx context.Context, name string) (*ent.AIModel, error) {
			return &ent.AIModel{Name: "claude-3-5-sonnet", UpstreamName: "claude-3-5-sonnet-20240620", IsEnabled: true}, nil
		},
	}
	groupPlatformRepo := &mockRelayGroupPlatformRepo{
		listPlatformsByGroupIDFunc: func(ctx context.Context, gid uuid.UUID) ([]uuid.UUID, error) {
			return []uuid.UUID{platformID}, nil
		},
	}
	platformRepo := &mockRelayPlatformRepo{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*ent.Platform, error) {
			return &ent.Platform{ID: platformID, BaseURL: "https://api.anthropic.com", Type: platform.TypeAnthropic, Status: platform.StatusActive}, nil
		},
	}
	accountRepo := &mockRelayAccountRepo{
		listByPlatformIDFunc: func(ctx context.Context, pid uuid.UUID) ([]*ent.Account, error) {
			return []*ent.Account{{ID: accountID, PlatformID: platformID, Status: account.StatusActive, Weight: 1}}, nil
		},
	}
	accountService := &mockRelayAccountService{
		decryptAPIKeyFunc: func(ctx context.Context, id uuid.UUID) (string, error) {
			return "sk-anthropic", nil
		},
	}
	forwarder := &mockForwarder{
		forwardFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
			if !strings.HasSuffix(req.URL.Path, "/v1/messages") {
				t.Errorf("expected /v1/messages path, got %s", req.URL.Path)
			}
			if req.Header.Get("x-api-key") != "sk-anthropic" {
				t.Errorf("expected x-api-key header, got %s", req.Header.Get("x-api-key"))
			}
			return &http.Response{
				StatusCode: 200,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"id":"msg_1","usage":{"input_tokens":2,"output_tokens":3}}`))),
			}, nil
		},
	}
	callLogRepo := &mockRelayCallLogRepo{
		createFunc: func(ctx context.Context, input repository.CreateCallLogInput) (*ent.CallLog, error) {
			if input.PromptTokens != 2 {
				t.Errorf("expected prompt_tokens 2, got %d", input.PromptTokens)
			}
			if input.CompletionTokens != 3 {
				t.Errorf("expected completion_tokens 3, got %d", input.CompletionTokens)
			}
			if input.TotalTokens != 5 {
				t.Errorf("expected total_tokens 5, got %d", input.TotalTokens)
			}
			return &ent.CallLog{}, nil
		},
	}

	svc := NewRelayService(modelRepo, groupPlatformRepo, platformRepo, accountRepo, callLogRepo, accountService, &mockBillingService{}, forwarder, key, 50*1024*1024, 2)
	result, err := svc.Messages(context.Background(), RelayInput{
		UserID:      uuid.New(),
		TokenID:     uuid.New(),
		GroupID:     uuid.New(),
		RequestBody: []byte(`{"model":"claude-3-5-sonnet","max_tokens":1024,"messages":[]}`),
	})
	if err != nil {
		t.Fatalf("messages: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("expected 200, got %d", result.StatusCode)
	}
	if result.Usage.TotalTokens != 5 {
		t.Errorf("expected total tokens 5, got %d", result.Usage.TotalTokens)
	}
}

func TestRelayService_MessagesStream(t *testing.T) {
	key := make([]byte, 32)
	platformID := uuid.New()
	accountID := uuid.New()

	modelRepo := &mockRelayAIModelRepo{
		getByNameFunc: func(ctx context.Context, name string) (*ent.AIModel, error) {
			return &ent.AIModel{Name: "claude-3-5-sonnet", UpstreamName: "claude-3-5-sonnet-20240620", IsEnabled: true}, nil
		},
	}
	groupPlatformRepo := &mockRelayGroupPlatformRepo{
		listPlatformsByGroupIDFunc: func(ctx context.Context, gid uuid.UUID) ([]uuid.UUID, error) {
			return []uuid.UUID{platformID}, nil
		},
	}
	platformRepo := &mockRelayPlatformRepo{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*ent.Platform, error) {
			return &ent.Platform{ID: platformID, BaseURL: "https://api.anthropic.com", Type: platform.TypeAnthropic, Status: platform.StatusActive}, nil
		},
	}
	accountRepo := &mockRelayAccountRepo{
		listByPlatformIDFunc: func(ctx context.Context, pid uuid.UUID) ([]*ent.Account, error) {
			return []*ent.Account{{ID: accountID, PlatformID: platformID, Status: account.StatusActive, Weight: 1}}, nil
		},
	}
	accountService := &mockRelayAccountService{
		decryptAPIKeyFunc: func(ctx context.Context, id uuid.UUID) (string, error) {
			return "sk-anthropic", nil
		},
	}
	forwarder := &mockForwarder{
		forwardFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
				Body:       io.NopCloser(bytes.NewReader([]byte("data: {\"usage\":{\"input_tokens\":1,\"output_tokens\":2}}\n\n"))),
			}, nil
		},
	}
	var logged *repository.CreateCallLogInput
	callLogRepo := &mockRelayCallLogRepo{
		createFunc: func(ctx context.Context, input repository.CreateCallLogInput) (*ent.CallLog, error) {
			logged = &input
			return &ent.CallLog{}, nil
		},
	}

	svc := NewRelayService(modelRepo, groupPlatformRepo, platformRepo, accountRepo, callLogRepo, accountService, &mockBillingService{}, forwarder, key, 50*1024*1024, 2)
	rec := httptest.NewRecorder()
	err := svc.MessagesStream(context.Background(), RelayInput{
		UserID:      uuid.New(),
		TokenID:     uuid.New(),
		GroupID:     uuid.New(),
		RequestBody: []byte(`{"model":"claude-3-5-sonnet","stream":true,"messages":[]}`),
	}, rec)
	if err != nil {
		t.Fatalf("messages stream: %v", err)
	}
	if rec.Code != 200 {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if logged == nil {
		t.Fatal("expected call log")
	}
	if logged.TotalTokens != 3 {
		t.Errorf("expected total tokens 3, got %d", logged.TotalTokens)
	}
}

// Ensure mockForwarder satisfies the interface at compile time.
var _ relay.Forwarder = (*mockForwarder)(nil)
var _ AccountService = (*mockRelayAccountService)(nil)

func TestRelayService_PickWeightedCandidate(t *testing.T) {
	svc := NewRelayService(nil, nil, nil, nil, nil, nil, &mockBillingService{}, nil, nil, 0, 0).(*relayService)
	ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	candidates := []*candidate{
		{account: &ent.Account{ID: ids[0], Weight: 1}},
		{account: &ent.Account{ID: ids[1], Weight: 5}},
		{account: &ent.Account{ID: ids[2], Weight: 10}},
	}

	picked, remaining, err := svc.pickWeightedCandidate(candidates)
	if err != nil {
		t.Fatalf("pick: %v", err)
	}
	if len(remaining) != 2 {
		t.Errorf("expected 2 remaining candidates, got %d", len(remaining))
	}
	found := false
	for _, id := range ids {
		if picked.account.ID == id {
			found = true
		}
	}
	if !found {
		t.Errorf("picked candidate not in original list")
	}
	for _, c := range remaining {
		if c.account.ID == picked.account.ID {
			t.Errorf("remaining contains picked candidate")
		}
	}
}

func TestRelayService_FailoverNonStreamSuccess(t *testing.T) {
	key := make([]byte, 32)
	platformID := uuid.New()
	accountA := uuid.New()
	accountB := uuid.New()
	groupID := uuid.New()

	// Force deterministic selection: always pick the first candidate first.
	origRandInt := randInt
	randInt = func(max *big.Int) (*big.Int, error) { return big.NewInt(0), nil }
	defer func() { randInt = origRandInt }()

	modelRepo := &mockRelayAIModelRepo{
		getByNameFunc: func(ctx context.Context, name string) (*ent.AIModel, error) {
			return &ent.AIModel{Name: "gpt-4o", UpstreamName: "gpt-4o-2024-08-06", IsEnabled: true}, nil
		},
	}
	groupPlatformRepo := &mockRelayGroupPlatformRepo{
		listPlatformsByGroupIDFunc: func(ctx context.Context, gid uuid.UUID) ([]uuid.UUID, error) {
			return []uuid.UUID{platformID}, nil
		},
	}
	platformRepo := &mockRelayPlatformRepo{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*ent.Platform, error) {
			return &ent.Platform{ID: platformID, BaseURL: "https://api.openai.com/v1", Type: platform.TypeOpenai, Status: platform.StatusActive}, nil
		},
	}
	accountRepo := &mockRelayAccountRepo{
		listByPlatformIDFunc: func(ctx context.Context, pid uuid.UUID) ([]*ent.Account, error) {
			return []*ent.Account{
				{ID: accountA, PlatformID: platformID, Status: account.StatusActive, Weight: 1},
				{ID: accountB, PlatformID: platformID, Status: account.StatusActive, Weight: 1},
			}, nil
		},
		incrementErrorCountFunc: func(ctx context.Context, id uuid.UUID) error {
			if id != accountA {
				t.Errorf("expected error count increment for account A, got %s", id)
			}
			return nil
		},
	}
	accountService := &mockRelayAccountService{
		decryptAPIKeyFunc: func(ctx context.Context, id uuid.UUID) (string, error) {
			if id == accountA {
				return "key-a", nil
			}
			return "key-b", nil
		},
	}

	var calls int
	forwarder := &mockForwarder{
		forwardFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
			calls++
			if req.Header.Get("Authorization") == "Bearer key-a" {
				return &http.Response{
					StatusCode: 503,
					Header:     http.Header{"Content-Type": []string{"application/json"}},
					Body:       io.NopCloser(bytes.NewReader([]byte(`{"error":"overload"}`))),
				}, nil
			}
			return &http.Response{
				StatusCode: 200,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`))),
			}, nil
		},
	}
	callLogRepo := &mockRelayCallLogRepo{
		createFunc: func(ctx context.Context, input repository.CreateCallLogInput) (*ent.CallLog, error) {
			return &ent.CallLog{}, nil
		},
	}

	svc := NewRelayService(modelRepo, groupPlatformRepo, platformRepo, accountRepo, callLogRepo, accountService, &mockBillingService{}, forwarder, key, 50*1024*1024, 2)
	result, err := svc.ChatCompletions(context.Background(), RelayInput{
		UserID:      uuid.New(),
		TokenID:     uuid.New(),
		GroupID:     groupID,
		RequestBody: []byte(`{"model":"gpt-4o","messages":[]}`),
	})
	if err != nil {
		t.Fatalf("chat completions: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("expected 200 after failover, got %d", result.StatusCode)
	}
	if calls != 2 {
		t.Errorf("expected 2 upstream attempts, got %d", calls)
	}
}

func TestRelayService_FailoverExhausted(t *testing.T) {
	key := make([]byte, 32)
	platformID := uuid.New()
	accountA := uuid.New()
	accountB := uuid.New()

	origRandInt := randInt
	randInt = func(max *big.Int) (*big.Int, error) { return big.NewInt(0), nil }
	defer func() { randInt = origRandInt }()

	modelRepo := &mockRelayAIModelRepo{
		getByNameFunc: func(ctx context.Context, name string) (*ent.AIModel, error) {
			return &ent.AIModel{Name: "gpt-4o", UpstreamName: "gpt-4o-2024-08-06", IsEnabled: true}, nil
		},
	}
	groupPlatformRepo := &mockRelayGroupPlatformRepo{
		listPlatformsByGroupIDFunc: func(ctx context.Context, gid uuid.UUID) ([]uuid.UUID, error) {
			return []uuid.UUID{platformID}, nil
		},
	}
	platformRepo := &mockRelayPlatformRepo{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*ent.Platform, error) {
			return &ent.Platform{ID: platformID, BaseURL: "https://api.openai.com/v1", Type: platform.TypeOpenai, Status: platform.StatusActive}, nil
		},
	}

	var incremented []uuid.UUID
	accountRepo := &mockRelayAccountRepo{
		listByPlatformIDFunc: func(ctx context.Context, pid uuid.UUID) ([]*ent.Account, error) {
			return []*ent.Account{
				{ID: accountA, PlatformID: platformID, Status: account.StatusActive, Weight: 1},
				{ID: accountB, PlatformID: platformID, Status: account.StatusActive, Weight: 1},
			}, nil
		},
		incrementErrorCountFunc: func(ctx context.Context, id uuid.UUID) error {
			incremented = append(incremented, id)
			return nil
		},
	}
	accountService := &mockRelayAccountService{
		decryptAPIKeyFunc: func(ctx context.Context, id uuid.UUID) (string, error) { return "sk-upstream", nil },
	}
	forwarder := &mockForwarder{
		forwardFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 503,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"error":"overload"}`))),
			}, nil
		},
	}
	callLogRepo := &mockRelayCallLogRepo{
		createFunc: func(ctx context.Context, input repository.CreateCallLogInput) (*ent.CallLog, error) {
			return &ent.CallLog{}, nil
		},
	}

	svc := NewRelayService(modelRepo, groupPlatformRepo, platformRepo, accountRepo, callLogRepo, accountService, &mockBillingService{}, forwarder, key, 50*1024*1024, 2)
	result, err := svc.ChatCompletions(context.Background(), RelayInput{
		UserID:      uuid.New(),
		TokenID:     uuid.New(),
		GroupID:     uuid.New(),
		RequestBody: []byte(`{"model":"gpt-4o","messages":[]}`),
	})
	if err != nil {
		t.Fatalf("expected last upstream response to be returned, got error: %v", err)
	}
	if result.StatusCode != 503 {
		t.Errorf("expected status 503, got %d", result.StatusCode)
	}
	if len(incremented) != 2 {
		t.Errorf("expected 2 error count increments, got %d", len(incremented))
	}
}

// Helper for encrypted key in integration tests.
func encryptKey(t *testing.T, key []byte, plaintext string) string {
	t.Helper()
	out, err := encryption.Encrypt(key, plaintext)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	return out
}
