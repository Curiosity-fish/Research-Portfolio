//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/internal/repository"
	"github.com/school-api/school-api-v1/internal/service"
)

func TestDepartment_CRUD(t *testing.T) {
	dbURL := os.Getenv("APP_DATABASE__URL")
	if dbURL == "" {
		t.Skip("APP_DATABASE__URL not set")
	}

	ctx := context.Background()

	client, err := repository.OpenEntClient(dbURL)
	if err != nil {
		t.Fatalf("open ent client: %v", err)
	}
	defer client.Close()

	if err := client.Schema.Create(ctx); err != nil {
		t.Fatalf("create schema: %v", err)
	}

	engine := buildUserEngine(t, client)
	token := createAdminAndLogin(t, ctx, client, engine)

	// Create root department.
	body := `{"name":"教务处","code":"jwc","sort_order":1}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/departments", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create department expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var rootResp struct {
		Data service.DepartmentResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &rootResp); err != nil {
		t.Fatalf("decode department response: %v", err)
	}
	rootID := rootResp.Data.ID
	if rootResp.Data.Level != 0 {
		t.Errorf("expected level 0, got %d", rootResp.Data.Level)
	}
	t.Cleanup(func() {
		_ = client.Department.DeleteOneID(uuid.MustParse(rootID)).Exec(context.Background())
	})

	// Create child department.
	childBody := `{"name":"教学科","code":"jxk","parent_id":"` + rootID + `"}`
	req = httptest.NewRequest(http.MethodPost, "/api/v1/admin/departments", bytes.NewReader([]byte(childBody)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create child department expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var childResp struct {
		Data service.DepartmentResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &childResp); err != nil {
		t.Fatalf("decode child response: %v", err)
	}
	childID := childResp.Data.ID
	if childResp.Data.Level != 1 {
		t.Errorf("expected child level 1, got %d", childResp.Data.Level)
	}
	if childResp.Data.ParentID != rootID {
		t.Errorf("expected parent_id %s, got %s", rootID, childResp.Data.ParentID)
	}

	// List departments.
	req = httptest.NewRequest(http.MethodGet, "/api/v1/admin/departments", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list departments expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var listResp struct {
		Data struct {
			List []service.DepartmentResponse `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(listResp.Data.List) != 2 {
		t.Errorf("expected 2 departments, got %d", len(listResp.Data.List))
	}

	// Update department.
	updateBody := `{"name":"教务处（更新）"}`
	req = httptest.NewRequest(http.MethodPut, "/api/v1/admin/departments/"+rootID, bytes.NewReader([]byte(updateBody)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("update department expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Delete child department.
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/admin/departments/"+childID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete department expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

// entClient is an alias to avoid repeating the import in test signatures.
type entClient = ent.Client
