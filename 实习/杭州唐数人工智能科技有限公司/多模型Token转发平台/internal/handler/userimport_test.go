package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"

	"github.com/school-api/school-api-v1/internal/service"
)

func TestUserHandler_BulkImportParsesRows(t *testing.T) {
	dept := uuid.New()
	var captured service.BulkImportInput
	svc := &mockUserService{
		bulkImportFunc: func(_ context.Context, input service.BulkImportInput) (*service.BulkImportResult, error) {
			captured = input
			return &service.BulkImportResult{Total: 2, Created: 2}, nil
		},
	}
	engine := newUserEngine(t, svc)

	body := `{"department_id":"` + dept.String() + `","default_password":"password123","users":[` +
		`{"username":"2023001","name":"张三","gender":"男","email":"zs@example.com","phone":"13800000001"},` +
		`{"username":"2023002","name":"李四"}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/bulk-import", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if captured.DepartmentID == nil || *captured.DepartmentID != dept {
		t.Fatalf("department id not parsed: %+v", captured)
	}
	if len(captured.Rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(captured.Rows))
	}
	first := captured.Rows[0]
	if first.Row != 1 || first.Username != "2023001" || first.Name != "张三" || first.Gender != "男" || first.Email != "zs@example.com" {
		t.Fatalf("first row not parsed: %+v", first)
	}
	if captured.Rows[1].Row != 2 || captured.Rows[1].Username != "2023002" {
		t.Fatalf("second row numbering wrong: %+v", captured.Rows[1])
	}
	var resp struct {
		Data service.BulkImportResult `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Data.Created != 2 {
		t.Fatalf("expected created 2, got %d", resp.Data.Created)
	}
}

func TestUserHandler_BulkImportBindingValidation(t *testing.T) {
	called := false
	svc := &mockUserService{
		bulkImportFunc: func(context.Context, service.BulkImportInput) (*service.BulkImportResult, error) {
			called = true
			return &service.BulkImportResult{}, nil
		},
	}
	engine := newUserEngine(t, svc)

	dept := uuid.NewString()
	cases := []string{
		`{"default_password":"password123","users":[{"username":"2023001","name":"张三"}]}`,                         // missing department_id
		`{"department_id":"` + dept + `","default_password":"short","users":[{"username":"2023001","name":"张三"}]}`,  // short password
		`{"department_id":"` + dept + `","default_password":"password123"}`,                                        // missing users
		`{"department_id":"not-a-uuid","default_password":"password123","users":[{"username":"2023001","name":"张三"}]}`, // bad uuid
	}
	for _, body := range cases {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/bulk-import", bytes.NewBufferString(body))
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

// buildImportXLSX renders rows (first element = header) into an in-memory
// workbook so the multipart handler can be exercised end to end.
func buildImportXLSX(t *testing.T, rows [][]interface{}) []byte {
	t.Helper()
	f := excelize.NewFile()
	sheet := f.GetSheetList()[0]
	for i := range rows {
		cell, err := excelize.CoordinatesToCellName(1, i+1)
		if err != nil {
			t.Fatalf("cell name: %v", err)
		}
		if err := f.SetSheetRow(sheet, cell, &rows[i]); err != nil {
			t.Fatalf("set row: %v", err)
		}
	}
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatalf("write xlsx: %v", err)
	}
	return buf.Bytes()
}

func TestUserHandler_BulkImportFileParsesXLSX(t *testing.T) {
	dept := uuid.New()
	var captured service.BulkImportInput
	svc := &mockUserService{
		bulkImportFunc: func(_ context.Context, input service.BulkImportInput) (*service.BulkImportResult, error) {
			captured = input
			return &service.BulkImportResult{Total: 2, Created: 2}, nil
		},
	}
	engine := newUserEngine(t, svc)

	xlsx := buildImportXLSX(t, [][]interface{}{
		{"学号", "姓名", "性别", "邮箱", "手机号"},
		{"2023011", "王五", "男", "ww@example.com", "13900000001"},
		{"2023012", "赵六", "女", "", ""},
	})

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("file", "roster.xlsx")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := fw.Write(xlsx); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	_ = w.WriteField("department_id", dept.String())
	_ = w.WriteField("default_password", "password123")
	w.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/bulk-import-file", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if captured.DepartmentID == nil || *captured.DepartmentID != dept {
		t.Fatalf("department id not parsed: %+v", captured)
	}
	if captured.DefaultPassword != "password123" {
		t.Fatalf("password not parsed: %q", captured.DefaultPassword)
	}
	if len(captured.Rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(captured.Rows))
	}
	first := captured.Rows[0]
	if first.Row != 2 || first.Username != "2023011" || first.Name != "王五" || first.Gender != "男" || first.Email != "ww@example.com" || first.Phone != "13900000001" {
		t.Fatalf("first row not mapped: %+v", first)
	}
	second := captured.Rows[1]
	if second.Row != 3 || second.Username != "2023012" || second.Email != "" {
		t.Fatalf("second row not mapped: %+v", second)
	}
}

func TestUserHandler_BulkImportFileMissingRequiredHeader(t *testing.T) {
	svc := &mockUserService{}
	engine := newUserEngine(t, svc)

	xlsx := buildImportXLSX(t, [][]interface{}{
		{"姓名", "性别"},
		{"王五", "男"},
	})

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, _ := w.CreateFormFile("file", "roster.xlsx")
	_, _ = fw.Write(xlsx)
	_ = w.WriteField("department_id", uuid.NewString())
	_ = w.WriteField("default_password", "password123")
	w.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/bulk-import-file", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing header, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUserHandler_ListParsesDepartmentFilter(t *testing.T) {
	dept := uuid.New()
	var captured service.ListUsersFilter
	svc := &mockUserService{
		listFunc: func(_ context.Context, filter service.ListUsersFilter) (*service.ListUsersResponse, error) {
			captured = filter
			return &service.ListUsersResponse{}, nil
		},
	}
	engine := newUserEngine(t, svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users?department_id="+dept.String(), nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if captured.DepartmentID == nil || *captured.DepartmentID != dept {
		t.Fatalf("department filter not parsed: %+v", captured)
	}
}
