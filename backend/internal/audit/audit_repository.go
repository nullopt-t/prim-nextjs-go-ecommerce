package audit

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

// Repository persists and reads audit_log rows.
type Repository struct{}

// NewRepository constructs an audit repository.
func NewRepository() *Repository {
	return &Repository{}
}

// Insert appends a single audit entry.
func (r *Repository) Insert(
	ctx context.Context,
	entry *model.AuditLog,
) error {
	executor, err := database.Executor(ctx)
	if err != nil {
		return err
	}
	err = executor.QueryRowContext(ctx, `
		INSERT INTO audit_log (
			actor_id, actor_role, action, entity_type, entity_id,
			method, path, status_code, ip_address, user_agent,
			before_state, after_state, metadata, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id
	`,
		entry.ActorID,
		entry.ActorRole,
		entry.Action,
		entry.EntityType,
		entry.EntityID,
		entry.Method,
		entry.Path,
		entry.StatusCode,
		entry.IPAddress,
		entry.UserAgent,
		entry.Before,
		entry.After,
		entry.Metadata,
		entry.CreatedAt,
	).Scan(&entry.ID)
	if err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}
	return nil
}

// Filter narrows an audit query. Zero values mean "no constraint".
type Filter struct {
	ActorID    *uuid.UUID
	Action     string
	EntityType string
	EntityID   string
	From       *time.Time
	To         *time.Time
}

// List returns a keyset-paginated slice of audit entries, newest first.
func (r *Repository) List(
	ctx context.Context,
	f Filter,
	limit int,
	cursor *pagination.Keyset,
) ([]*model.AuditLog, *pagination.CursorPage, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, nil, err
	}
	query := `
		SELECT id, actor_id, actor_role, action, entity_type, entity_id,
		       method, path, status_code, ip_address, user_agent,
		       before_state, after_state, metadata, created_at
		FROM audit_log
		WHERE 1 = 1
	`
	args := []any{}
	i := 1

	if f.ActorID != nil {
		query += fmt.Sprintf(" AND actor_id = $%d", i)
		args = append(args, *f.ActorID)
		i++
	}
	if f.Action != "" {
		query += fmt.Sprintf(" AND action = $%d", i)
		args = append(args, f.Action)
		i++
	}
	if f.EntityType != "" {
		query += fmt.Sprintf(" AND entity_type = $%d", i)
		args = append(args, f.EntityType)
		i++
	}
	if f.EntityID != "" {
		query += fmt.Sprintf(" AND entity_id = $%d", i)
		args = append(args, f.EntityID)
		i++
	}
	if f.From != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", i)
		args = append(args, *f.From)
		i++
	}
	if f.To != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", i)
		args = append(args, *f.To)
		i++
	}

	keysetClause, keysetArgs := pagination.KeysetClause("", cursor, i)
	query += keysetClause
	args = append(args, keysetArgs...)
	i += len(keysetArgs)

	query += pagination.KeysetOrderBy("")
	query += fmt.Sprintf(" LIMIT $%d", i)
	args = append(args, limit+1)

	rows, err := executor.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("list audit log: %w", err)
	}
	defer rows.Close()

	var entries []*model.AuditLog
	for rows.Next() {
		e := &model.AuditLog{}
		if err := rows.Scan(
			&e.ID, &e.ActorID, &e.ActorRole, &e.Action, &e.EntityType, &e.EntityID,
			&e.Method, &e.Path, &e.StatusCode, &e.IPAddress, &e.UserAgent,
			&e.Before, &e.After, &e.Metadata, &e.CreatedAt,
		); err != nil {
			return nil, nil, fmt.Errorf("scan audit log: %w", err)
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("iterate audit log: %w", err)
	}

	entries, page := pagination.NewCursorPageFrom(entries, limit, func(e *model.AuditLog) pagination.Keyset {
		return pagination.Keyset{CreatedAt: e.CreatedAt, IDInt: e.ID}
	})
	return entries, page, nil
}
