// Package audit provides an append-only accountability trail for privileged
// back-office mutations. A middleware records admin API writes automatically
// (actor, action, entity, request context), and superadmins read the trail
// through GET /admin/audit.
package audit

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

// Service records and queries audit entries.
type Service struct {
	runner database.Runner
	repo   *Repository
}

// NewService constructs an audit service.
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
		return apierr.NotFoundError("Audit log not found").
			WithCode(apierr.CodeNotFound).
			Wrap(err)
	case errors.Is(mappedErr, database.ErrConflict):
		return apierr.ConflictError("Audit log conflict").
			WithCode(apierr.CodeResourceConflict).
			Wrap(err)
	case errors.Is(mappedErr, database.ErrForeignKeyViolation):
		return apierr.BadRequestError("Referenced entity does not exist").
			WithCode(apierr.CodeInvalidReference).
			Wrap(err)
	case errors.Is(mappedErr, database.ErrCheckViolation),
		errors.Is(mappedErr, database.ErrNotNullViolation),
		errors.Is(mappedErr, database.ErrInvalidInput):
		return apierr.BadRequestError("Invalid input parameter").
			WithCode(apierr.CodeInvalidInput).
			Wrap(err)
	case errors.Is(mappedErr, database.ErrTimeout):
		return apierr.GatewayTimeoutError("Audit operation timed out").
			Wrap(err)
	case errors.Is(mappedErr, database.ErrCanceled):
		return apierr.GatewayTimeoutError("Audit operation was cancelled").
			Wrap(err)
	default:
		return apierr.InternalServerError(fallbackMsg).
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}
}

// Record persists a single audit entry. The database assigns the entry id. A
// failure here must never break the originating request: callers are expected
// to log it and continue.
func (s *Service) Record(ctx context.Context, entry *model.AuditLog) error {
	err := s.runner.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {
		return s.repo.Insert(ctx, tx, entry)
	})
	if err != nil {
		return s.handleError(err, "Failed to record audit log")
	}
	return nil
}

// List returns audit entries matching the filter, newest first.
func (s *Service) List(
	ctx context.Context,
	f Filter,
	limit int,
	cursor *pagination.Keyset,
) ([]*model.AuditLog, *pagination.CursorPage, error) {
	var entries []*model.AuditLog
	var page *pagination.CursorPage
	err := s.runner.WithExecutor(ctx, func(exec database.QueryExecutor) error {
		var err error
		entries, page, err = s.repo.List(ctx, exec, f, limit, cursor)
		return err
	})
	if err != nil {
		return nil, nil, s.handleError(err, "Failed to list audit logs")
	}
	return entries, page, nil
}
