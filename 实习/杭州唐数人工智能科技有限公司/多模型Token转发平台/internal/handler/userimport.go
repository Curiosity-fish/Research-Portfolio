package handler

import (
	"mime/multipart"
	"strings"

	"github.com/xuri/excelize/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/server/respond"
	"github.com/school-api/school-api-v1/internal/service"
)

// maxImportFileSize caps the XLSX upload at 5 MiB; rosters far below that.
const maxImportFileSize = 5 << 20

// BulkImportRequest is the JSON body for POST /api/v1/admin/users/bulk-import.
// Per-row format rules are deliberately not binding tags: bad rows must come
// back as failed row results, not reject the whole request.
type BulkImportRequest struct {
	DepartmentID    string                  `json:"department_id" binding:"required,uuid"`
	DefaultPassword string                  `json:"default_password" binding:"required,min=8,max=72"`
	Users           []service.BulkImportRow `json:"users" binding:"required,min=1,max=1000"`
}

// BulkImport handles POST /api/v1/admin/users/bulk-import: a class roster as
// a JSON array; every row becomes a user in the given department.
func (h *UserHandler) BulkImport(c *gin.Context) {
	var req BulkImportRequest
	if !bindAndValidate(c, &req) {
		return
	}
	if len([]byte(req.DefaultPassword)) > 72 {
		respond.Error(c, domain.NewValidationError(map[string]string{"default_password": "密码长度超过最大字节限制"}))
		return
	}
	deptID, err := uuid.Parse(req.DepartmentID)
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"department_id": "部门 ID 格式错误"}))
		return
	}

	rows := make([]service.BulkImportRow, len(req.Users))
	for i, r := range req.Users {
		r.Row = i + 1
		rows[i] = r
	}

	resp, err := h.userService.BulkImport(c.Request.Context(), service.BulkImportInput{
		DepartmentID:    &deptID,
		DefaultPassword: req.DefaultPassword,
		Rows:            rows,
	})
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.OK(c, resp)
}

// xlsxHeaderAliases maps a header cell (case-insensitive) to a row field.
// Both the Chinese template headers and plain English names are accepted.
var xlsxHeaderAliases = map[string]string{
	"学号": "username", "username": "username", "user_name": "username",
	"姓名": "name", "name": "name",
	"性别": "gender", "gender": "gender",
	"邮箱": "email", "email": "email", "e-mail": "email",
	"手机号": "phone", "手机": "phone", "电话": "phone", "phone": "phone", "mobile": "phone",
}

// BulkImportFile handles POST /api/v1/admin/users/bulk-import-file: a
// multipart XLSX upload. The first worksheet must carry a header row; data
// starts on line 2 and row results report real Excel line numbers.
func (h *UserHandler) BulkImportFile(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		respond.Error(c, domain.NewValidationError(map[string]string{"file": "必须上传 XLSX 文件"}))
		return
	}
	if fileHeader.Size > maxImportFileSize {
		respond.Error(c, domain.NewValidationError(map[string]string{"file": "文件大小超过 5MB 限制"}))
		return
	}

	deptID, defaultPassword, err := parseImportForm(c)
	if err != nil {
		respond.Error(c, err)
		return
	}

	rows, err := parseImportXLSX(fileHeader)
	if err != nil {
		respond.Error(c, err)
		return
	}

	resp, err := h.userService.BulkImport(c.Request.Context(), service.BulkImportInput{
		DepartmentID:    deptID,
		DefaultPassword: defaultPassword,
		Rows:            rows,
	})
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.OK(c, resp)
}

// parseImportForm extracts and validates the shared multipart form fields.
func parseImportForm(c *gin.Context) (*uuid.UUID, string, error) {
	departmentID := strings.TrimSpace(c.PostForm("department_id"))
	if departmentID == "" {
		return nil, "", domain.NewValidationError(map[string]string{"department_id": "必须指定班级（部门）"})
	}
	deptID, err := uuid.Parse(departmentID)
	if err != nil {
		return nil, "", domain.NewValidationError(map[string]string{"department_id": "部门 ID 格式错误"})
	}
	defaultPassword := c.PostForm("default_password")
	if l := len(defaultPassword); l < 8 || l > 72 {
		return nil, "", domain.NewValidationError(map[string]string{"default_password": "初始密码长度须为 8-72 个字符"})
	}
	return &deptID, defaultPassword, nil
}

// parseImportXLSX reads the first worksheet, maps header columns to row
// fields and returns the data rows. Empty lines are skipped so trailing blank
// Excel lines do not surface as validation failures.
func parseImportXLSX(fileHeader *multipart.FileHeader) ([]service.BulkImportRow, error) {
	src, err := fileHeader.Open()
	if err != nil {
		return nil, domain.NewValidationError(map[string]string{"file": "无法读取上传文件"})
	}
	defer src.Close()

	xl, err := excelize.OpenReader(src)
	if err != nil {
		return nil, domain.NewValidationError(map[string]string{"file": "无法解析 XLSX 文件"})
	}
	defer xl.Close()

	sheets := xl.GetSheetList()
	if len(sheets) == 0 {
		return nil, domain.NewValidationError(map[string]string{"file": "文件中没有工作表"})
	}
	rawRows, err := xl.GetRows(sheets[0])
	if err != nil {
		return nil, domain.NewValidationError(map[string]string{"file": "读取工作表失败"})
	}
	if len(rawRows) == 0 {
		return nil, domain.NewValidationError(map[string]string{"file": "文件中没有表头"})
	}

	// Header row: map column index to a row field, ignoring unknown columns.
	colField := map[int]string{}
	hasUsername, hasName := false, false
	for col, cell := range rawRows[0] {
		field, ok := xlsxHeaderAliases[strings.ToLower(strings.TrimSpace(cell))]
		if !ok {
			continue
		}
		colField[col] = field
		if field == "username" {
			hasUsername = true
		}
		if field == "name" {
			hasName = true
		}
	}
	if !hasUsername || !hasName {
		return nil, domain.NewValidationError(map[string]string{"file": "表头缺少必需列：学号、姓名"})
	}

	rows := make([]service.BulkImportRow, 0, len(rawRows))
	for i := 1; i < len(rawRows); i++ {
		cells := rawRows[i]
		row := service.BulkImportRow{Row: i + 1}
		empty := true
		for col, field := range colField {
			v := ""
			if col < len(cells) {
				v = strings.TrimSpace(cells[col])
			}
			if v != "" {
				empty = false
			}
			switch field {
			case "username":
				row.Username = v
			case "name":
				row.Name = v
			case "gender":
				row.Gender = v
			case "email":
				row.Email = v
			case "phone":
				row.Phone = v
			}
		}
		if empty {
			continue
		}
		rows = append(rows, row)
	}
	return rows, nil
}
