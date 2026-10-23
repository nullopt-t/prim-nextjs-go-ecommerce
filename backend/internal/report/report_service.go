// Package report provides date-ranged sales reporting and CSV export of the
// operational datasets (orders, payments, customers, inventory) for the admin
// back office.
package report

import (
	"context"
	"encoding/csv"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

// Service builds reports and export payloads.
type Service struct {
	runner database.Runner
	repo   *Repository
}

// NewService constructs a report service.
func NewService(runner database.Runner, repo *Repository) *Service {
	return &Service{runner: runner, repo: repo}
}

// handleError maps database, driver, and context errors into typed domain APIErrors.
func (s *Service) handleError(err error, fallbackMsg string) error {
	if err == nil {
		return nil
	}

	var apiErr *apierr.APIError
	if errors.As(err, &apiErr) {
		return apiErr
	}

	mappedErr := database.MapError(err)
	switch {
	case errors.Is(mappedErr, database.ErrNotFound) || errors.Is(err, pgx.ErrNoRows):
		return apierr.NotFoundError("Report data not found").
			WithCode(apierr.CodeNotFound).
			Wrap(err)
	case errors.Is(mappedErr, database.ErrConflict):
		return apierr.ConflictError("Report conflict").
			WithCode(apierr.CodeResourceConflict).
			Wrap(err)
	case errors.Is(mappedErr, database.ErrForeignKeyViolation):
		return apierr.BadRequestError("Referenced entity does not exist").
			WithCode(apierr.CodeInvalidReference).
			Wrap(err)
	case errors.Is(mappedErr, database.ErrCheckViolation),
		errors.Is(mappedErr, database.ErrNotNullViolation),
		errors.Is(mappedErr, database.ErrInvalidInput):
		return apierr.BadRequestError("Invalid report parameters").
			WithCode(apierr.CodeInvalidInput).
			Wrap(err)
	case errors.Is(mappedErr, database.ErrTimeout):
		return apierr.GatewayTimeoutError("Report operation timed out").
			Wrap(err)
	case errors.Is(mappedErr, database.ErrCanceled):
		return apierr.GatewayTimeoutError("Report operation was cancelled").
			Wrap(err)
	default:
		return apierr.InternalServerError(fallbackMsg).
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}
}

// GetSalesReport returns the sales summary for the period. groupBy is "day" or
// "month" and defaults to "day".
func (s *Service) GetSalesReport(ctx context.Context, f ReportFilter, groupBy string) (*SalesReport, error) {
	if groupBy == "" {
		groupBy = "day"
	}
	switch groupBy {
	case "day", "month":
	default:
		return nil, apierr.BadRequestError("Invalid groupBy (expected 'day' or 'month')").
			WithCode(apierr.CodeInvalidInput)
	}

	var rep *SalesReport
	err := s.runner.WithDB(ctx, func(db database.QueryExecutor) error {
		var err error
		rep, err = s.repo.Sales(ctx, db, f, groupBy)
		return err
	})
	if err != nil {
		return nil, s.handleError(err, "Failed to generate sales report")
	}
	return rep, nil
}

// BuildCSV renders the requested dataset as CSV bytes (header + rows).
func (s *Service) BuildCSV(ctx context.Context, t ExportType, f ReportFilter) ([]byte, error) {
	var header []string
	var rows [][]string
	err := s.runner.WithDB(ctx, func(db database.QueryExecutor) error {
		var err error
		header, rows, err = s.repo.ExportDataset(ctx, db, t, f)
		return err
	})
	if err != nil {
		return nil, s.handleError(err, "Failed to export report dataset")
	}

	var sb strings.Builder
	w := csv.NewWriter(&sb)
	if err := w.Write(header); err != nil {
		return nil, s.handleError(err, "Failed to write CSV header")
	}
	for _, row := range rows {
		if err := w.Write(row); err != nil {
			return nil, s.handleError(err, "Failed to write CSV row")
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, s.handleError(err, "Failed to flush CSV output")
	}
	return []byte(sb.String()), nil
}
