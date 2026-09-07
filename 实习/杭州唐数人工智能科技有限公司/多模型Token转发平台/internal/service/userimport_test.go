package service

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/repository"
)

// importDeptID creates a real department: the SQLite test client enforces
// foreign keys, so user.department_id must reference an existing row.
func importDeptID(t *testing.T, ctx context.Context, client *ent.Client) *uuid.UUID {
	t.Helper()
	dep, err := client.Department.Create().
		SetName("导入测试班-" + uuid.NewString()[:8]).
		SetCode("imp-" + uuid.NewString()[:8]).
		Save(ctx)
	if err != nil {
		t.Fatalf("create department: %v", err)
	}
	return &dep.ID
}

func TestUserService_BulkImportCreatesUsers(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	svc := NewUserService(repository.NewEntUserRepository(client), 4)

	dept := importDeptID(t, ctx, client)
	result, err := svc.BulkImport(ctx, BulkImportInput{
		DepartmentID:    dept,
		DefaultPassword: "password123",
		Rows: []BulkImportRow{
			{Row: 1, Username: "imp001", Name: "张三", Gender: "男", Email: "imp001@example.com", Phone: "13800000001"},
			{Row: 2, Username: "imp002", Name: "李四"},
		},
	})
	if err != nil {
		t.Fatalf("bulk import: %v", err)
	}
	if result.Total != 2 || result.Created != 2 || result.Failed != 0 {
		t.Fatalf("unexpected result: %+v", result)
	}
	for _, r := range result.Results {
		if r.Status != "created" || r.UserID == "" {
			t.Fatalf("unexpected row result: %+v", r)
		}
	}

	created, err := client.User.Query().Where().All(ctx)
	if err != nil {
		t.Fatalf("query users: %v", err)
	}
	if len(created) != 2 {
		t.Fatalf("expected 2 users stored, got %d", len(created))
	}
	for _, u := range created {
		if u.DepartmentID == nil || *u.DepartmentID != *dept {
			t.Fatalf("expected department on %s, got %v", u.Username, u.DepartmentID)
		}
		if u.Role != "student" || u.Status != "active" {
			t.Fatalf("expected student/active on %s, got %s/%s", u.Username, u.Role, u.Status)
		}
	}
}

func TestUserService_BulkImportRowFailures(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	svc := NewUserService(repository.NewEntUserRepository(client), 4)

	result, err := svc.BulkImport(ctx, BulkImportInput{
		DepartmentID:    importDeptID(t, ctx, client),
		DefaultPassword: "password123",
		Rows: []BulkImportRow{
			{Row: 1, Username: "ab", Name: "太短学号"},                                // username too short
			{Row: 2, Username: "dup001", Name: "重复"},                                 // ok
			{Row: 3, Username: "dup001", Name: "重复"},                                 // duplicate within batch
			{Row: 4, Username: "bad001", Name: "坏邮箱", Email: "not-an-email"},         // invalid email
			{Row: 5, Username: "noemail", Name: "无邮箱", Email: ""},                     // empty optional email is fine
			{Row: 6, Username: "bad002", Name: "", Email: ""},                          // missing name
			{Row: 7, Username: "dupmail", Name: "重复邮箱", Email: "dup@example.com"},   // duplicate email in batch
			{Row: 8, Username: "dupmail2", Name: "重复邮箱2", Email: "dup@example.com"}, // duplicate email in batch
		},
	})
	if err != nil {
		t.Fatalf("bulk import: %v", err)
	}
	if result.Created != 3 || result.Failed != 5 || result.Total != 8 {
		t.Fatalf("unexpected counts: total=%d created=%d failed=%d", result.Total, result.Created, result.Failed)
	}

	byRow := map[int]BulkImportRowResult{}
	for _, r := range result.Results {
		byRow[r.Row] = r
	}
	expectFailed := map[int]string{
		1: "学号",
		3: "批次内学号",
		4: "邮箱格式",
		6: "姓名",
		8: "批次内邮箱",
	}
	for row, frag := range expectFailed {
		r := byRow[row]
		if r.Status != "failed" || !strings.Contains(r.Error, frag) {
			t.Fatalf("row %d expected failed containing %q, got %+v", row, frag, r)
		}
	}
	// Rows 2, 5 and 7 are the first occurrence of their username/email and
	// must succeed despite later duplicates.
	for _, row := range []int{2, 5, 7} {
		if byRow[row].Status != "created" {
			t.Fatalf("expected row %d created, got %+v", row, byRow[row])
		}
	}
}

func TestUserService_BulkImportExistingUsername(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	svc := NewUserService(repository.NewEntUserRepository(client), 4)

	if _, err := svc.CreateUser(ctx, CreateUserInput{
		Username: "existing01",
		Password: "password123",
		Name:     "既有用户",
	}); err != nil {
		t.Fatalf("seed user: %v", err)
	}

	result, err := svc.BulkImport(ctx, BulkImportInput{
		DepartmentID:    importDeptID(t, ctx, client),
		DefaultPassword: "password123",
		Rows: []BulkImportRow{
			{Row: 1, Username: "existing01", Name: "同名"},
			{Row: 2, Username: "fresh001", Name: "新用户"},
		},
	})
	if err != nil {
		t.Fatalf("bulk import: %v", err)
	}
	if result.Created != 1 || result.Failed != 1 {
		t.Fatalf("unexpected counts: %+v", result)
	}
	if result.Results[0].Status != "failed" || !strings.Contains(result.Results[0].Error, "已存在") {
		t.Fatalf("expected duplicate-username failure, got %+v", result.Results[0])
	}
	if result.Results[1].Status != "created" {
		t.Fatalf("expected second row created, got %+v", result.Results[1])
	}
}

func TestUserService_BulkImportRequestValidation(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	svc := NewUserService(repository.NewEntUserRepository(client), 4)

	rows := func(n int) []BulkImportRow {
		out := make([]BulkImportRow, n)
		for i := range out {
			out[i] = BulkImportRow{Row: i + 1, Username: "u" + strings.Repeat("0", 4) + string(rune('a'+i%26)), Name: "用户"}
		}
		return out
	}

	cases := []struct {
		name  string
		input BulkImportInput
	}{
		{"missing department", BulkImportInput{DefaultPassword: "password123", Rows: rows(1)}},
		{"short password", BulkImportInput{DepartmentID: importDeptID(t, ctx, client), DefaultPassword: "short", Rows: rows(1)}},
		{"empty rows", BulkImportInput{DepartmentID: importDeptID(t, ctx, client), DefaultPassword: "password123"}},
		{"too many rows", BulkImportInput{DepartmentID: importDeptID(t, ctx, client), DefaultPassword: "password123", Rows: rows(MaxBulkImportRows + 1)}},
	}
	for _, tc := range cases {
		_, err := svc.BulkImport(ctx, tc.input)
		appErr, ok := err.(*domain.AppError)
		if !ok || appErr.BizCode != "VALIDATION_ERROR" {
			t.Fatalf("%s: expected validation error, got %v", tc.name, err)
		}
	}
}

func TestUserService_ListDepartmentFilterAndQuotaAggregate(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	svc := NewUserService(repository.NewEntUserRepository(client), 4)

	classA := importDeptID(t, ctx, client)
	classB := importDeptID(t, ctx, client)

	userA, err := client.User.Create().
		SetUsername("agg-a").
		SetPasswordHash("hash").
		SetName("聚合A").
		SetDepartmentID(*classA).
		Save(ctx)
	if err != nil {
		t.Fatalf("create user a: %v", err)
	}
	userB, err := client.User.Create().
		SetUsername("agg-b").
		SetPasswordHash("hash").
		SetName("聚合B").
		SetDepartmentID(*classB).
		Save(ctx)
	if err != nil {
		t.Fatalf("create user b: %v", err)
	}

	// User A: enabled token (limit 1000, used 300) + disabled token (ignored).
	enabled := int64(1000)
	if _, err := client.UserToken.Create().
		SetUserID(userA.ID).
		SetName("t1").
		SetTokenHash("hash-a1").
		SetTokenLast4("a001").
		SetQuotaLimit(enabled).
		SetQuotaUsed(300).
		Save(ctx); err != nil {
		t.Fatalf("create token a1: %v", err)
	}
	if _, err := client.UserToken.Create().
		SetUserID(userA.ID).
		SetName("t2").
		SetTokenHash("hash-a2").
		SetTokenLast4("a002").
		SetQuotaLimit(5000).
		SetQuotaUsed(999).
		SetIsEnabled(false).
		Save(ctx); err != nil {
		t.Fatalf("create token a2: %v", err)
	}

	// User B: one unlimited enabled token + one finite enabled token.
	finite := int64(500)
	if _, err := client.UserToken.Create().
		SetUserID(userB.ID).
		SetName("t1").
		SetTokenHash("hash-b1").
		SetTokenLast4("b001").
		Save(ctx); err != nil { // no quota limit -> unlimited
		t.Fatalf("create token b1: %v", err)
	}
	if _, err := client.UserToken.Create().
		SetUserID(userB.ID).
		SetName("t2").
		SetTokenHash("hash-b2").
		SetTokenLast4("b002").
		SetQuotaLimit(finite).
		SetQuotaUsed(120).
		Save(ctx); err != nil {
		t.Fatalf("create token b2: %v", err)
	}

	// Class filter returns only user A with the aggregate fields.
	resp, err := svc.ListUsers(ctx, ListUsersFilter{DepartmentID: classA, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list by department: %v", err)
	}
	if resp.Total != 1 || len(resp.List) != 1 {
		t.Fatalf("expected 1 user in class A, got total=%d len=%d", resp.Total, len(resp.List))
	}
	a := resp.List[0]
	if a.Username != "agg-a" {
		t.Fatalf("expected agg-a, got %s", a.Username)
	}
	if a.TokenCount != 1 {
		t.Errorf("expected token_count 1 (disabled ignored), got %d", a.TokenCount)
	}
	if a.QuotaLimitTotal == nil || *a.QuotaLimitTotal != 1000 {
		t.Errorf("expected quota_limit_total 1000, got %v", a.QuotaLimitTotal)
	}
	if a.QuotaUsedTotal != 300 {
		t.Errorf("expected quota_used_total 300, got %d", a.QuotaUsedTotal)
	}

	// User B has an unlimited token: total reported as nil despite the finite one.
	resp, err = svc.ListUsers(ctx, ListUsersFilter{DepartmentID: classB, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list by department b: %v", err)
	}
	if resp.Total != 1 {
		t.Fatalf("expected 1 user in class B, got %d", resp.Total)
	}
	b := resp.List[0]
	if b.TokenCount != 2 {
		t.Errorf("expected token_count 2, got %d", b.TokenCount)
	}
	if b.QuotaLimitTotal != nil {
		t.Errorf("expected nil quota_limit_total for unlimited token, got %v", *b.QuotaLimitTotal)
	}
	if b.QuotaUsedTotal != 120 {
		t.Errorf("expected quota_used_total 120, got %d", b.QuotaUsedTotal)
	}
}
