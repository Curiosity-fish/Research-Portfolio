//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/user"
	"github.com/school-api/school-api-v1/internal/auth"
	"github.com/school-api/school-api-v1/internal/handler"
	"github.com/school-api/school-api-v1/internal/repository"
	"github.com/school-api/school-api-v1/internal/server/middleware"
	"github.com/school-api/school-api-v1/internal/service"
)

func buildUserImportEngine(t *testing.T, client *ent.Client) *gin.Engine {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	gin.SetMode(gin.TestMode)

	adminRepo := repository.NewEntAdminRepository(client)
	tokenManager := auth.NewTokenManager("userimport-integration-secret", time.Hour)
	authService := service.NewAuthService(adminRepo, tokenManager, auth.BcryptCostForEnv("test"), time.Hour)
	authHandler := handler.NewAuthHandler(authService)

	userRepo := repository.NewEntUserRepository(client)
	userService := service.NewUserService(userRepo, auth.BcryptCostForEnv("test"))
	userHandler := handler.NewUserHandler(userService)

	engine := gin.New()
	engine.Use(
		middleware.Recovery(logger),
		middleware.RequestID(),
		middleware.RequestLogger(logger),
	)

	engine.POST("/api/v1/admin/login", authHandler.Login)

	adminGroup := engine.Group("/api/v1/admin", middleware.JWTAuth(tokenManager))
	adminGroup.POST("/users", userHandler.Create)
	adminGroup.GET("/users", userHandler.List)
	adminGroup.POST("/users/bulk-import", userHandler.BulkImport)
	adminGroup.POST("/users/bulk-import-file", userHandler.BulkImportFile)

	return engine
}

func TestUserImport_BulkImportAndClassView(t *testing.T) {
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

	engine := buildUserImportEngine(t, client)
	adminToken := createAdminAndLogin(t, ctx, client, engine)

	// Unique per-run prefix keeps usernames conflict-free across reruns.
	run := "stu" + uuid.NewString()[:6]
	t.Cleanup(func() {
		_, _ = client.User.Delete().Where(user.UsernameHasPrefix(run)).Exec(context.Background())
	})

	// The class every imported student lands in.
	dep, err := client.Department.Create().
		SetName("集成导入班-" + uuid.NewString()[:8]).
		SetCode("imp-" + uuid.NewString()[:8]).
		Save(ctx)
	if err != nil {
		t.Fatalf("create department: %v", err)
	}
	t.Cleanup(func() { _ = client.Department.DeleteOneID(dep.ID).Exec(context.Background()) })

	// JSON import: one good row, one duplicate username within the batch.
	body := fmt.Sprintf(`{"department_id":%q,"default_password":"password123","users":[`+
		`{"username":"%s01","name":"学生一","gender":"女","email":"%s01@example.com","phone":"13800000001"},`+
		`{"username":"%s01","name":"学生一替身"}]}`, dep.ID.String(), run, run, run)
	rec := settingsIntegrationRequest(t, engine, http.MethodPost, "/api/v1/admin/users/bulk-import", adminToken, body)
	if rec.Code != http.StatusOK {
		t.Fatalf("json import expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var jsonResult struct {
		Data service.BulkImportResult `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &jsonResult); err != nil {
		t.Fatalf("decode json import result: %v", err)
	}
	if jsonResult.Data.Created != 1 || jsonResult.Data.Failed != 1 {
		t.Fatalf("unexpected json import result: %+v", jsonResult.Data)
	}
	if jsonResult.Data.Results[1].Error == "" {
		t.Fatal("expected duplicate row to carry an error message")
	}

	// XLSX import: Chinese headers, two students, optional cells left empty.
	xlsx := excelize.NewFile()
	sheet := xlsx.GetSheetList()[0]
	rows := [][]interface{}{
		{"学号", "姓名", "性别", "邮箱", "手机号"},
		{run + "0101", "王五", "男", run + "0101@example.com", "13800000002"},
		{run + "0102", "赵六", "女", "", ""},
	}
	for i, values := range rows {
		cell, _ := excelize.CoordinatesToCellName(1, i+1)
		if err := xlsx.SetSheetRow(sheet, cell, &values); err != nil {
			t.Fatalf("set row: %v", err)
		}
	}
	var xlsxBuf bytes.Buffer
	if err := xlsx.Write(&xlsxBuf); err != nil {
		t.Fatalf("write xlsx: %v", err)
	}

	var form bytes.Buffer
	w := multipart.NewWriter(&form)
	fw, err := w.CreateFormFile("file", "roster.xlsx")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := fw.Write(xlsxBuf.Bytes()); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	_ = w.WriteField("department_id", dep.ID.String())
	_ = w.WriteField("default_password", "password123")
	w.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/bulk-import-file", &form)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+adminToken)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("xlsx import expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var xlsxResult struct {
		Data service.BulkImportResult `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &xlsxResult); err != nil {
		t.Fatalf("decode xlsx import result: %v", err)
	}
	if xlsxResult.Data.Created != 2 || xlsxResult.Data.Failed != 0 {
		t.Fatalf("unexpected xlsx import result: %+v", xlsxResult.Data)
	}
	// Excel line numbers: data starts under the header on line 2.
	if xlsxResult.Data.Results[0].Row != 2 {
		t.Fatalf("expected first xlsx row number 2, got %d", xlsxResult.Data.Results[0].Row)
	}

	// Class view: list filtered by department returns the three imported
	// students with quota aggregates (no tokens yet -> zero counts).
	rec = settingsIntegrationRequest(t, engine, http.MethodGet,
		"/api/v1/admin/users?department_id="+dep.ID.String()+"&page_size=50", adminToken, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("list by department expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var listResp struct {
		Data struct {
			Total int                    `json:"total"`
			List  []service.UserResponse `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if listResp.Data.Total != 3 {
		t.Fatalf("expected 3 students in class, got %d", listResp.Data.Total)
	}

	// Grant one student a token and verify the aggregate reflects it.
	stu, err := client.User.Query().
		Where(user.DepartmentID(dep.ID)).
		First(ctx)
	if err != nil {
		t.Fatalf("query student: %v", err)
	}
	limit := int64(2000)
	tok, err := client.UserToken.Create().
		SetUserID(stu.ID).
		SetName("primary").
		SetTokenHash(uuid.NewString()).
		SetTokenLast4("abcd").
		SetQuotaLimit(limit).
		SetQuotaUsed(500).
		Save(ctx)
	if err != nil {
		t.Fatalf("create token: %v", err)
	}
	t.Cleanup(func() { _ = client.UserToken.DeleteOneID(tok.ID).Exec(context.Background()) })

	rec = settingsIntegrationRequest(t, engine, http.MethodGet,
		"/api/v1/admin/users?department_id="+dep.ID.String()+"&page_size=50", adminToken, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("re-list by department expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	listResp = struct {
		Data struct {
			Total int                    `json:"total"`
			List  []service.UserResponse `json:"list"`
		} `json:"data"`
	}{}
	if err := json.Unmarshal(rec.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("decode re-list: %v", err)
	}
	var withToken *service.UserResponse
	for i := range listResp.Data.List {
		if listResp.Data.List[i].ID == stu.ID.String() {
			withToken = &listResp.Data.List[i]
		}
	}
	if withToken == nil {
		t.Fatal("token owner missing from class list")
	}
	if withToken.TokenCount != 1 || withToken.QuotaLimitTotal == nil || *withToken.QuotaLimitTotal != 2000 || withToken.QuotaUsedTotal != 500 {
		t.Fatalf("unexpected aggregate: %+v", withToken)
	}
}
