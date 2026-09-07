package service

import (
	"context"
	"fmt"
	"net/mail"
	"strings"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/user"
	"github.com/school-api/school-api-v1/internal/auth"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/repository"
)

// MaxBulkImportRows caps a single bulk-import request. A class roster is at
// most a few hundred rows; the cap exists to bound request parsing and the
// per-row insert loop.
const MaxBulkImportRows = 1000

// BulkImportRow is one student line of an import. Row is the caller-visible
// line number: array index + 1 for the JSON payload, the real Excel line
// number for XLSX uploads (header on line 1, data starts at 2).
type BulkImportRow struct {
	Row      int    `json:"row"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Gender   string `json:"gender"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
}

// BulkImportInput is a class-scoped bulk import: every row becomes a user in
// the same department (the class). DefaultPassword is the shared initial
// password the administrator sets for the batch.
type BulkImportInput struct {
	DepartmentID    *uuid.UUID
	DefaultPassword string
	Rows            []BulkImportRow
}

// BulkImportRowResult reports the per-row outcome. Failed rows carry a
// human-readable Chinese error; created rows carry the new user id.
type BulkImportRowResult struct {
	Row      int    `json:"row"`
	Username string `json:"username"`
	Status   string `json:"status"` // created | failed
	UserID   string `json:"user_id,omitempty"`
	Error    string `json:"error,omitempty"`
}

// BulkImportResult is the summary plus every row outcome. Rows are inserted
// independently: one bad line never rolls back the others.
type BulkImportResult struct {
	Total   int                   `json:"total"`
	Created int                   `json:"created"`
	Failed  int                   `json:"failed"`
	Results []BulkImportRowResult `json:"results"`
}

// BulkImport validates each row, creates users one by one, and returns the
// per-row outcomes. Imported users get role=student, status=active and no
// token; quota is granted separately via the existing token endpoints.
func (s *UserService) BulkImport(ctx context.Context, input BulkImportInput) (*BulkImportResult, error) {
	if input.DepartmentID == nil {
		return nil, domain.NewValidationError(map[string]string{"department_id": "必须指定班级（部门）"})
	}
	if l := len(input.DefaultPassword); l < 8 || l > 72 {
		return nil, domain.NewValidationError(map[string]string{"default_password": "初始密码长度须为 8-72 个字符"})
	}
	if n := len(input.Rows); n == 0 {
		return nil, domain.NewValidationError(map[string]string{"users": "导入列表不能为空"})
	} else if n > MaxBulkImportRows {
		return nil, domain.NewValidationError(map[string]string{"users": fmt.Sprintf("单次最多导入 %d 行", MaxBulkImportRows)})
	}

	result := &BulkImportResult{Total: len(input.Rows), Results: make([]BulkImportRowResult, 0, len(input.Rows))}
	seenUser := map[string]int{}  // username -> row of first occurrence
	seenEmail := map[string]int{} // email -> row of first occurrence
	seenPhone := map[string]int{} // phone -> row of first occurrence

	for _, row := range input.Rows {
		res := BulkImportRowResult{Row: row.Row, Username: row.Username, Status: "created"}
		if err := s.importRow(ctx, input, row, seenUser, seenEmail, seenPhone, &res); err != nil {
			res.Status = "failed"
			res.Error = err.Error()
			res.UserID = ""
			result.Failed++
		} else {
			result.Created++
		}
		result.Results = append(result.Results, res)
	}
	return result, nil
}

// importRow validates and inserts a single row. Duplicate maps are mutated so
// later rows see earlier ones; the error return marks the row failed.
func (s *UserService) importRow(
	ctx context.Context,
	input BulkImportInput,
	row BulkImportRow,
	seenUser, seenEmail, seenPhone map[string]int,
	res *BulkImportRowResult,
) error {
	username := strings.TrimSpace(row.Username)
	name := strings.TrimSpace(row.Name)
	email := strings.TrimSpace(row.Email)
	phone := strings.TrimSpace(row.Phone)
	gender := strings.TrimSpace(row.Gender)

	if l := len(username); l < 3 || l > 50 {
		return fmt.Errorf("学号长度须为 3-50 个字符")
	}
	if l := len(name); l == 0 || l > 100 {
		return fmt.Errorf("姓名不能为空且不超过 100 个字符")
	}
	if l := len(gender); l > 10 {
		return fmt.Errorf("性别长度不超过 10 个字符")
	}
	if email != "" {
		if len(email) > 255 {
			return fmt.Errorf("邮箱长度不超过 255 个字符")
		}
		if _, err := mail.ParseAddress(email); err != nil {
			return fmt.Errorf("邮箱格式无效")
		}
	}
	if len(phone) > 20 {
		return fmt.Errorf("手机号长度不超过 20 个字符")
	}

	if prev, dup := seenUser[username]; dup {
		return fmt.Errorf("批次内学号与第 %d 行重复", prev)
	}
	if email != "" {
		if prev, dup := seenEmail[email]; dup {
			return fmt.Errorf("批次内邮箱与第 %d 行重复", prev)
		}
	}
	if phone != "" {
		if prev, dup := seenPhone[phone]; dup {
			return fmt.Errorf("批次内手机号与第 %d 行重复", prev)
		}
	}
	seenUser[username] = row.Row
	if email != "" {
		seenEmail[email] = row.Row
	}
	if phone != "" {
		seenPhone[phone] = row.Row
	}

	hash, err := auth.HashPassword(s.bcryptCost, input.DefaultPassword)
	if err != nil {
		return fmt.Errorf("密码哈希失败")
	}

	u, err := s.repo.Create(ctx, repository.CreateUserInput{
		Username:     username,
		PasswordHash: hash,
		Name:         name,
		Email:        email,
		Phone:        phone,
		Role:         user.RoleStudent,
		Gender:       gender,
		Status:       user.StatusActive,
		DepartmentID: input.DepartmentID,
	})
	if err != nil {
		if ent.IsConstraintError(err) {
			return fmt.Errorf("学号、邮箱或手机号已存在")
		}
		return fmt.Errorf("写入失败")
	}
	res.UserID = u.ID.String()
	return nil
}
