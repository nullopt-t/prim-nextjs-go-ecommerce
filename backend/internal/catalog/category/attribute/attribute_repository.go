package attribute

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

type AttributeItem struct {
	ID            uuid.UUID       `json:"id"`
	CategoryID    *uuid.UUID      `json:"categoryId"`
	CategoryName  string          `json:"categoryName"`
	Key           string          `json:"key"`
	Label         string          `json:"label"`
	ValueType     string          `json:"valueType"`
	AllowedValues json.RawMessage `json:"allowedValues"`
	IsFilterable  bool            `json:"isFilterable"`
	SortOrder     int             `json:"sortOrder"`
	CreatedAt     time.Time       `json:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`
}

type AttributeRepository struct{}

func NewRepository() *AttributeRepository {
	return &AttributeRepository{}
}

func (r *AttributeRepository) ListAttributes(ctx context.Context, qe database.QueryExecutor) ([]AttributeItem, error) {
	query := `
		SELECT ca.id, ca.category_id, COALESCE(pc.name, 'Global'), ca.key, ca.label, ca.value_type,
		       ca.allowed_values, ca.is_filterable, ca.sort_order, ca.created_at, ca.updated_at
		FROM category_attributes ca
		LEFT JOIN product_categories pc ON ca.category_id = pc.id
		ORDER BY ca.created_at DESC
	`
	rows, err := qe.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list attributes: %w", err)
	}
	defer rows.Close()

	var list []AttributeItem
	for rows.Next() {
		var ai AttributeItem
		if err := rows.Scan(
			&ai.ID, &ai.CategoryID, &ai.CategoryName, &ai.Key, &ai.Label, &ai.ValueType,
			&ai.AllowedValues, &ai.IsFilterable, &ai.SortOrder, &ai.CreatedAt, &ai.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan attribute: %w", err)
		}
		list = append(list, ai)
	}
	return list, nil
}

func (r *AttributeRepository) CreateAttribute(ctx context.Context, qe database.QueryExecutor, ai *AttributeItem) error {
	query := `
		INSERT INTO category_attributes (id, category_id, key, label, value_type, allowed_values, is_filterable, sort_order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5::attribute_value_type, $6, $7, $8, NOW(), NOW())
		RETURNING created_at, updated_at
	`
	return qe.QueryRow(ctx, query,
		ai.ID, ai.CategoryID, ai.Key, ai.Label, ai.ValueType, ai.AllowedValues, ai.IsFilterable, ai.SortOrder,
	).Scan(&ai.CreatedAt, &ai.UpdatedAt)
}

func (r *AttributeRepository) DeleteAttribute(ctx context.Context, qe database.QueryExecutor, id uuid.UUID) error {
	query := `DELETE FROM category_attributes WHERE id = $1`
	_, err := qe.Exec(ctx, query, id)
	return err
}
