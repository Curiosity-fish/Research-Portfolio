package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"

	"github.com/school-api/school-api-v1/internal/auth"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/server/respond"
	"github.com/school-api/school-api-v1/internal/service"
)

// ExportService is the subset of the export service used by ExportHandler.
type ExportService interface {
	UsageExport(ctx context.Context, start, end *time.Time, groupBy string) ([][]string, error)
	SummaryExport(ctx context.Context, days int) ([]service.ExportSheet, error)
	BalanceRecordsExport(ctx context.Context, userID *uuid.UUID) ([][]string, error)
	RechargeOrdersExport(ctx context.Context) ([][]string, error)
}

// ExportHandler handles XLSX export endpoints.
type ExportHandler struct {
	svc ExportService
}

// NewExportHandler creates a new ExportHandler.
func NewExportHandler(svc ExportService) *ExportHandler {
	return &ExportHandler{svc: svc}
}

// AdminUsageExport handles GET /api/v1/admin/export/usage.
func (h *ExportHandler) AdminUsageExport(c *gin.Context) {
	start, end, err := parseDateRange(c)
	if err != nil {
		respond.Error(c, err)
		return
	}
	groupBy := c.Query("group_by")
	if groupBy == "" {
		respond.ErrorWithStatus(c, http.StatusBadRequest, "INVALID_GROUP_BY", "group_by 必须是 model、user 或 day")
		return
	}

	rows, err := h.svc.UsageExport(c.Request.Context(), start, end, groupBy)
	if err != nil {
		respond.Error(c, err)
		return
	}
	writeXLSX(c, exportFilename("usage-"+groupBy), []service.ExportSheet{{Name: "使用统计", Rows: rows}})
}

// AdminSummaryExport handles GET /api/v1/admin/export/summary.
func (h *ExportHandler) AdminSummaryExport(c *gin.Context) {
	days, _, err := parseStatsWindow(c)
	if err != nil {
		respond.Error(c, err)
		return
	}

	sheets, err := h.svc.SummaryExport(c.Request.Context(), days)
	if err != nil {
		respond.Error(c, err)
		return
	}
	writeXLSX(c, exportFilename("summary"), sheets)
}

// AdminBalanceRecordsExport handles GET /api/v1/admin/export/balance-records.
func (h *ExportHandler) AdminBalanceRecordsExport(c *gin.Context) {
	rows, err := h.svc.BalanceRecordsExport(c.Request.Context(), nil)
	if err != nil {
		respond.Error(c, err)
		return
	}
	writeXLSX(c, exportFilename("balance-records"), []service.ExportSheet{{Name: "计费流水", Rows: rows}})
}

// AdminRechargeOrdersExport handles GET /api/v1/admin/export/recharge-orders.
func (h *ExportHandler) AdminRechargeOrdersExport(c *gin.Context) {
	rows, err := h.svc.RechargeOrdersExport(c.Request.Context())
	if err != nil {
		respond.Error(c, err)
		return
	}
	writeXLSX(c, exportFilename("recharge-orders"), []service.ExportSheet{{Name: "充值订单", Rows: rows}})
}

// UserBalanceRecordsExport handles GET /api/v1/export/balance-records
// (API-key auth, scoped to the authenticated user).
func (h *ExportHandler) UserBalanceRecordsExport(c *gin.Context) {
	userID, ok := auth.UserID(c)
	if !ok {
		respond.Error(c, domain.ErrUnauthorized)
		return
	}
	rows, err := h.svc.BalanceRecordsExport(c.Request.Context(), &userID)
	if err != nil {
		respond.Error(c, err)
		return
	}
	writeXLSX(c, exportFilename("balance-records"), []service.ExportSheet{{Name: "余额流水", Rows: rows}})
}

// exportFilename builds a download name like "export-usage-model-20260907.xlsx".
func exportFilename(kind string) string {
	return fmt.Sprintf("export-%s-%s.xlsx", kind, time.Now().Format("20060102"))
}

// writeXLSX renders sheets into an XLSX workbook and streams it as the
// response body.
func writeXLSX(c *gin.Context, filename string, sheets []service.ExportSheet) {
	if len(sheets) == 0 {
		respond.ErrorWithStatus(c, http.StatusInternalServerError, "EXPORT_EMPTY", "导出内容为空")
		return
	}

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()

	for i, sheet := range sheets {
		name := sheetName(sheet.Name, i)
		var index int
		var err error
		if i == 0 {
			index, err = file.GetSheetIndex("Sheet1")
			if err == nil {
				err = file.SetSheetName("Sheet1", name)
			}
		} else {
			index, err = file.NewSheet(name)
		}
		if err != nil {
			respond.ErrorWithStatus(c, http.StatusInternalServerError, "EXPORT_FAILED", "生成导出文件失败")
			return
		}

		for rowIdx, row := range sheet.Rows {
			cell, err := excelize.CoordinatesToCellName(1, rowIdx+1)
			if err != nil {
				respond.ErrorWithStatus(c, http.StatusInternalServerError, "EXPORT_FAILED", "生成导出文件失败")
				return
			}
			if err := file.SetSheetRow(name, cell, &row); err != nil {
				respond.ErrorWithStatus(c, http.StatusInternalServerError, "EXPORT_FAILED", "生成导出文件失败")
				return
			}
		}
		file.SetActiveSheet(index)
	}

	buf, err := file.WriteToBuffer()
	if err != nil {
		respond.ErrorWithStatus(c, http.StatusInternalServerError, "EXPORT_FAILED", "生成导出文件失败")
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

// sheetName sanitises a sheet title: Excel forbids []:*?/\ and caps names at
// 31 runes. The index keeps names unique after sanitisation.
func sheetName(name string, index int) string {
	sanitised := make([]rune, 0, len(name))
	for _, r := range name {
		switch r {
		case '[', ']', ':', '*', '?', '/', '\\':
			continue
		default:
			sanitised = append(sanitised, r)
		}
	}
	if len(sanitised) == 0 {
		return fmt.Sprintf("Sheet%d", index+1)
	}
	if len(sanitised) > 31 {
		sanitised = sanitised[:31]
	}
	return string(sanitised)
}
