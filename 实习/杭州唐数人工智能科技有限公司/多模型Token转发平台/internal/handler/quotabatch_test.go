package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/repository"
	"github.com/school-api/school-api-v1/internal/service"
)

type mockQuotaBatchService struct {
	apply func(ctx context.Context, input service.ApplyQuotaBatchInput) (*repository.QuotaBatchResult, error)
}

func (m *mockQuotaBatchService) Apply(ctx context.Context, input service.ApplyQuotaBatchInput) (*repository.QuotaBatchResult, error) {
	return m.apply(ctx, input)
}

func newQuotaBatchEngine(svc QuotaBatchService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	h := NewQuotaBatchHandler(svc)
	engine.POST("/api/v1/admin/quota-batch", h.Apply)
	return engine
}

func TestQuotaBatchHandler_ApplyParsesUserIDs(t *testing.T) {
	id1, id2 := uuid.New(), uuid.New()
	var captured service.ApplyQuotaBatchInput
	svc := &mockQuotaBatchService{
		apply: func(_ context.Context, input service.ApplyQuotaBatchInput) (*repository.QuotaBatchResult, error) {
			captured = input
			return &repository.QuotaBatchResult{MatchedUsers: 2, UpdatedTokens: 3}, nil
		},
	}
	engine := newQuotaBatchEngine(svc)

	body := `{"user_ids":["` + id1.String() + `","` + id2.String() + `"],"mode":"set","value":3000000}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/quota-batch", strReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if len(captured.UserIDs) != 2 || captured.UserIDs[0] != id1 || captured.UserIDs[1] != id2 {
		t.Fatalf("user ids not parsed: %+v", captured)
	}
	if captured.Mode != "set" || captured.Value != 3_000_000 {
		t.Fatalf("unexpected input: %+v", captured)
	}
	if !strContains(rec.Body.String(), `"updated_tokens":3`) {
		t.Fatalf("counts missing from response: %s", rec.Body.String())
	}
}

func TestQuotaBatchHandler_ApplyParsesDepartmentID(t *testing.T) {
	dep := uuid.New()
	var captured service.ApplyQuotaBatchInput
	svc := &mockQuotaBatchService{
		apply: func(_ context.Context, input service.ApplyQuotaBatchInput) (*repository.QuotaBatchResult, error) {
			captured = input
			return &repository.QuotaBatchResult{}, nil
		},
	}
	engine := newQuotaBatchEngine(svc)

	body := `{"department_id":"` + dep.String() + `","mode":"add","value":-500000}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/quota-batch", strReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if captured.DepartmentID == nil || *captured.DepartmentID != dep {
		t.Fatalf("department id not parsed: %+v", captured)
	}
}

func TestQuotaBatchHandler_BindingValidation(t *testing.T) {
	called := false
	svc := &mockQuotaBatchService{
		apply: func(context.Context, service.ApplyQuotaBatchInput) (*repository.QuotaBatchResult, error) {
			called = true
			return &repository.QuotaBatchResult{}, nil
		},
	}
	engine := newQuotaBatchEngine(svc)

	cases := []string{
		`{"user_ids":["not-a-uuid"],"mode":"set","value":1}`, // bad uuid
		`{"user_ids":["` + uuid.NewString() + `"],"mode":"swap","value":1}`, // bad mode
		`{"user_ids":["` + uuid.NewString() + `"],"value":1}`,               // missing mode
		`{"user_ids":["` + uuid.NewString() + `"],"mode":"set"}`,            // missing value
	}
	for _, body := range cases {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/quota-batch", strReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for %s, got %d: %s", body, rec.Code, rec.Body.String())
		}
	}
	if called {
		t.Fatal("service must not be called for invalid bodies")
	}
}

func TestQuotaBatchHandler_ServiceValidationSurfaces(t *testing.T) {
	svc := &mockQuotaBatchService{
		apply: func(context.Context, service.ApplyQuotaBatchInput) (*repository.QuotaBatchResult, error) {
			return nil, domain.NewValidationError(map[string]string{"value": "设置的配额必须在额度区间内"})
		},
	}
	engine := newQuotaBatchEngine(svc)

	body := `{"user_ids":["` + uuid.NewString() + `"],"mode":"set","value":4000000}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/quota-batch", strReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func strReader(body string) *bytes.Reader {
	return bytes.NewReader([]byte(body))
}

func strContains(haystack, needle string) bool {
	return bytes.Contains([]byte(haystack), []byte(needle))
}
