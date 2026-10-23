package report

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/pkg/api"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
)

// Handler exposes reporting and export endpoints.
type Handler struct {
	service *Service
}

// NewHandler constructs a report HTTP handler.
func NewHandler(s *Service) *Handler {
	return &Handler{service: s}
}

// GetSalesReport godoc
//
//	@Summary		Sales report
//	@Description	Returns revenue/order totals and a day or month time series for the given date range.
//	@Tags			Admin Reports
//	@Produce		json
//	@Param			dateFrom	query	string	false	"Start of range (RFC3339 or YYYY-MM-DD)"
//	@Param			dateTo		query	string	false	"End of range (RFC3339 or YYYY-MM-DD)"
//	@Param			groupBy		query	string	false	"Bucket size: day (default) or month"
//	@Success		200	{object}	api.DataResponse{data=SalesReport}	"Sales report"
//	@Failure		400	{object}	api.BadRequestErrorResponse	"Invalid filter"
//	@Failure		403	{object}	api.ForbiddenErrorResponse	"Admin privileges required"
//	@Router			/admin/reports/sales [get]
func (h *Handler) GetSalesReport(c *gin.Context) {
	filter, err := parseDateRange(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	rep, err := h.service.GetSalesReport(c.Request.Context(), filter, c.Query("groupBy"))
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: rep})
}

// ExportReport godoc
//
//	@Summary		Export a dataset as CSV
//	@Description	Streams orders, payments, customers, or inventory as a CSV attachment.
//	@Tags			Admin Reports
//	@Produce		text/csv
//	@Param			type		query	string	true	"Dataset: orders | payments | customers | inventory"
//	@Param			dateFrom	query	string	false	"Start of range (RFC3339 or YYYY-MM-DD)"
//	@Param			dateTo		query	string	false	"End of range (RFC3339 or YYYY-MM-DD)"
//	@Param			status		query	string	false	"Status filter (orders/payments only)"
//	@Success		200	{file}		file	"CSV attachment"
//	@Failure		400	{object}	api.BadRequestErrorResponse	"Invalid export type or filter"
//	@Failure		403	{object}	api.ForbiddenErrorResponse	"Admin privileges required"
//	@Router			/admin/reports/export [get]
func (h *Handler) ExportReport(c *gin.Context) {
	exportType, ok := ParseExportType(c.Query("type"))
	if !ok {
		_ = c.Error(apierr.BadRequestError("Invalid type (expected orders, payments, customers or inventory)").
			WithCode(apierr.CodeInvalidInput))
		return
	}

	filter, err := parseDateRange(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	data, err := h.service.BuildCSV(c.Request.Context(), exportType, filter)
	if err != nil {
		_ = c.Error(err)
		return
	}

	filename := fmt.Sprintf("%s-%s.csv", exportType, time.Now().UTC().Format("20060102-150405"))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.Data(http.StatusOK, "text/csv; charset=utf-8", data)
}

// parseDateRange reads dateFrom/dateTo/status. Each date accepts RFC3339 or a
// plain YYYY-MM-DD date (normalized to the start and end of that day).
func parseDateRange(c *gin.Context) (ReportFilter, error) {
	var f ReportFilter

	if v := c.Query("dateFrom"); v != "" {
		t, err := parseReportTime(v, false)
		if err != nil {
			return f, apierr.BadRequestError("Invalid dateFrom (RFC3339 or YYYY-MM-DD expected)").
				WithCode(apierr.CodeInvalidInput)
		}
		f.From = &t
	}
	if v := c.Query("dateTo"); v != "" {
		t, err := parseReportTime(v, true)
		if err != nil {
			return f, apierr.BadRequestError("Invalid dateTo (RFC3339 or YYYY-MM-DD expected)").
				WithCode(apierr.CodeInvalidInput)
		}
		f.To = &t
	}

	f.Status = c.Query("status")
	return f, nil
}

func parseReportTime(v string, endOfDay bool) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, v); err == nil {
		return t, nil
	}
	t, err := time.Parse("2006-01-02", v)
	if err != nil {
		return time.Time{}, err
	}
	if endOfDay {
		t = t.Add(24*time.Hour - time.Second)
	}
	return t, nil
}
