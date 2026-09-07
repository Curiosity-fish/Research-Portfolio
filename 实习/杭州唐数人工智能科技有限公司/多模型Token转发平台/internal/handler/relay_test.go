package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/internal/auth"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/service"
)

type mockRelayService struct {
	chatFunc         func(ctx context.Context, input service.RelayInput) (*service.RelayResult, error)
	chatStreamFunc   func(ctx context.Context, input service.RelayInput, w http.ResponseWriter) error
	messagesFunc     func(ctx context.Context, input service.RelayInput) (*service.RelayResult, error)
	messagesStreamFunc func(ctx context.Context, input service.RelayInput, w http.ResponseWriter) error
	listModelsFunc   func(ctx context.Context, groupID uuid.UUID) ([]service.ModelInfo, error)
}

func (m *mockRelayService) ChatCompletions(ctx context.Context, input service.RelayInput) (*service.RelayResult, error) {
	return m.chatFunc(ctx, input)
}

func (m *mockRelayService) ChatCompletionsStream(ctx context.Context, input service.RelayInput, w http.ResponseWriter) error {
	return m.chatStreamFunc(ctx, input, w)
}

func (m *mockRelayService) Messages(ctx context.Context, input service.RelayInput) (*service.RelayResult, error) {
	return m.messagesFunc(ctx, input)
}

func (m *mockRelayService) MessagesStream(ctx context.Context, input service.RelayInput, w http.ResponseWriter) error {
	return m.messagesStreamFunc(ctx, input, w)
}

func (m *mockRelayService) ListModels(ctx context.Context, groupID uuid.UUID) ([]service.ModelInfo, error) {
	return m.listModelsFunc(ctx, groupID)
}

func newRelayEngine(t *testing.T, svc service.RelayService) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	h := NewRelayHandler(svc)

	// Simulate API key auth by setting context values directly.
	engine.POST("/v1/chat/completions", setAuthContext, h.ChatCompletions)
	engine.POST("/v1/messages", setAuthContext, h.Messages)
	engine.GET("/v1/models", setAuthContext, h.ListModels)

	return engine
}

func setAuthContext(c *gin.Context) {
	auth.SetAPIKeyContext(c, &auth.APIKeyInfo{
		UserID:  uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		TokenID: uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		GroupID: uuid.MustParse("33333333-3333-3333-3333-333333333333"),
	})
}

func TestRelayHandler_ChatCompletions(t *testing.T) {
	svc := &mockRelayService{
		chatFunc: func(ctx context.Context, input service.RelayInput) (*service.RelayResult, error) {
			if string(input.RequestBody) != `{"model":"gpt-4o"}` {
				t.Errorf("unexpected request body: %s", string(input.RequestBody))
			}
			return &service.RelayResult{
				StatusCode: 200,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       []byte(`{"id":"chatcmpl-1"}`),
			}, nil
		},
	}

	engine := newRelayEngine(t, svc)
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(`{"model":"gpt-4o"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if rec.Body.String() != `{"id":"chatcmpl-1"}` {
		t.Errorf("unexpected body: %s", rec.Body.String())
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("unexpected content-type: %s", rec.Header().Get("Content-Type"))
	}
}

func TestRelayHandler_ChatCompletionsStream(t *testing.T) {
	svc := &mockRelayService{
		chatStreamFunc: func(ctx context.Context, input service.RelayInput, w http.ResponseWriter) error {
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(200)
			_, _ = w.Write([]byte("data: hello\n\n"))
			return nil
		},
	}

	engine := newRelayEngine(t, svc)
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(`{"model":"gpt-4o","stream":true}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if rec.Body.String() != "data: hello\n\n" {
		t.Errorf("unexpected body: %s", rec.Body.String())
	}
	if rec.Header().Get("Content-Type") != "text/event-stream" {
		t.Errorf("unexpected content-type: %s", rec.Header().Get("Content-Type"))
	}
}

func TestRelayHandler_ListModels(t *testing.T) {
	svc := &mockRelayService{
		listModelsFunc: func(ctx context.Context, groupID uuid.UUID) ([]service.ModelInfo, error) {
			return []service.ModelInfo{{ID: "gpt-4o", Object: "model"}}, nil
		},
	}

	engine := newRelayEngine(t, svc)
	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"id":"gpt-4o"`)) {
		t.Errorf("unexpected body: %s", rec.Body.String())
	}
}

func TestRelayHandler_ServiceError(t *testing.T) {
	svc := &mockRelayService{
		chatFunc: func(ctx context.Context, input service.RelayInput) (*service.RelayResult, error) {
			return nil, domain.NewAppError(503, "service_unavailable", "No available account")
		},
	}

	engine := newRelayEngine(t, svc)
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(`{"model":"gpt-4o"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != 503 {
		t.Fatalf("expected 503, got %d: %s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"type":"api_error"`)) {
		t.Errorf("expected OpenAI error type, got: %s", rec.Body.String())
	}
}

func TestRelayHandler_Messages(t *testing.T) {
	svc := &mockRelayService{
		messagesFunc: func(ctx context.Context, input service.RelayInput) (*service.RelayResult, error) {
			if string(input.RequestBody) != `{"model":"claude-3-5-sonnet"}` {
				t.Errorf("unexpected request body: %s", string(input.RequestBody))
			}
			return &service.RelayResult{
				StatusCode: 200,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       []byte(`{"id":"msg_1"}`),
			}, nil
		},
	}

	engine := newRelayEngine(t, svc)
	req := httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewBufferString(`{"model":"claude-3-5-sonnet"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if rec.Body.String() != `{"id":"msg_1"}` {
		t.Errorf("unexpected body: %s", rec.Body.String())
	}
}

func TestRelayHandler_MessagesStream(t *testing.T) {
	svc := &mockRelayService{
		messagesStreamFunc: func(ctx context.Context, input service.RelayInput, w http.ResponseWriter) error {
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(200)
			_, _ = w.Write([]byte("data: hello\n\n"))
			return nil
		},
	}

	engine := newRelayEngine(t, svc)
	req := httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewBufferString(`{"model":"claude-3-5-sonnet","stream":true}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if rec.Body.String() != "data: hello\n\n" {
		t.Errorf("unexpected body: %s", rec.Body.String())
	}
	if rec.Header().Get("Content-Type") != "text/event-stream" {
		t.Errorf("unexpected content-type: %s", rec.Header().Get("Content-Type"))
	}
}

// Ensure the mock satisfies the interface.
var _ service.RelayService = (*mockRelayService)(nil)
