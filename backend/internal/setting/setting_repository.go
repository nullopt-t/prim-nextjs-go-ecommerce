package setting

import (
	"context"
	"fmt"
	"time"

	"github.com/nullopt-t/prim-backend/pkg/database"
)

// Item represents a single administrative setting key/value pair persisted in
// the admin_settings table. Values are stored as JSONB.
type Item struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	UpdatedAt time.Time   `json:"updatedAt"`
}

// Repository provides persistence access to the admin_settings table.
type Repository struct{}

// NewRepository constructs a stateless settings repository.
func NewRepository() *Repository {
	return &Repository{}
}

// GetAll returns every stored setting ordered by key.
func (r *Repository) GetAll(ctx context.Context, qe database.QueryExecutor) ([]Item, error) {
	const query = `SELECT key, value, updated_at FROM admin_settings ORDER BY key`

	rows, err := qe.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list settings: %w", err)
	}
	defer rows.Close()

	var list []Item
	for rows.Next() {
		var it Item
		if err := rows.Scan(&it.Key, &it.Value, &it.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan setting: %w", err)
		}
		list = append(list, it)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate settings: %w", err)
	}
	return list, nil
}

// Upsert inserts or updates a single setting by key.
func (r *Repository) Upsert(ctx context.Context, qe database.QueryExecutor, key string, value interface{}) error {
	const query = `
		INSERT INTO admin_settings (key, value, updated_at)
		VALUES ($1, $2, now())
		ON CONFLICT (key) DO UPDATE
		SET value = EXCLUDED.value, updated_at = now()
	`

	if _, err := qe.Exec(ctx, query, key, value); err != nil {
		return fmt.Errorf("upsert setting %s: %w", key, err)
	}
	return nil
}
