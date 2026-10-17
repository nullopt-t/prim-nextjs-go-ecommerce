// Package audit provides an append-only accountability trail for privileged
// back-office mutations. A middleware records admin API writes automatically
// (actor, action, entity, request context), and superadmins read the trail
// through GET /admin/audit.
package audit

import (
	"context"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
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

// Record persists a single audit entry, assigning an ID when absent. A failure
// here must never break the originating request: callers are expected to log it
// and continue.
func (s *Service) Record(ctx context.Context, entry *model.AuditLog) error {
	if entry.ID == uuid.Nil {
		entry.ID = uuid.New()
	}
	return s.runner.WithDB(ctx, func(db database.QueryExecutor) error {
		return s.repo.Insert(ctx, db, entry)
	})
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
	err := s.runner.WithDB(ctx, func(db database.QueryExecutor) error {
		var err error
		entries, page, err = s.repo.List(ctx, db, f, limit, cursor)
		return err
	})
	return entries, page, err
}
