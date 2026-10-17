// Package report provides date-ranged sales reporting and CSV export of the
// operational datasets (orders, payments, customers, inventory) for the admin
// back office.
package report

import (
	"context"
	"encoding/csv"
	"strings"

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
		return nil, err
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
		return nil, err
	}

	var sb strings.Builder
	w := csv.NewWriter(&sb)
	if err := w.Write(header); err != nil {
		return nil, err
	}
	for _, row := range rows {
		if err := w.Write(row); err != nil {
			return nil, err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return []byte(sb.String()), nil
}
