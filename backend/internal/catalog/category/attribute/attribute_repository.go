package attribute

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

func (r *AttributeRepository) ListAttributes(ctx context.Context) ([]AttributeItem, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}
	query := `
		SELECT ca.id, ca.category_id, COALESCE(pc.name, 'Global'), ca.key, ca.label, ca.value_type,
		       ca.allowed_values, ca.is_filterable, ca.sort_order, ca.created_at, ca.updated_at
		FROM category_attributes ca
		LEFT JOIN product_categories pc ON ca.category_id = pc.id
		ORDER BY ca.created_at DESC
	`
	rows, err := executor.QueryContext(ctx, query)
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

func (r *AttributeRepository) GetByID(ctx context.Context, id uuid.UUID) (*AttributeItem, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}
	query := `
		SELECT ca.id, ca.category_id, COALESCE(pc.name, 'Global'), ca.key, ca.label, ca.value_type,
		       ca.allowed_values, ca.is_filterable, ca.sort_order, ca.created_at, ca.updated_at
		FROM category_attributes ca
		LEFT JOIN product_categories pc ON ca.category_id = pc.id
		WHERE ca.id = $1
	`
	var ai AttributeItem
	err = executor.QueryRowContext(ctx, query, id).Scan(
		&ai.ID, &ai.CategoryID, &ai.CategoryName, &ai.Key, &ai.Label, &ai.ValueType,
		&ai.AllowedValues, &ai.IsFilterable, &ai.SortOrder, &ai.CreatedAt, &ai.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get attribute by id: %w", err)
	}
	return &ai, nil
}

// ListByCategory returns the attribute definitions that belong to a category.
// It is used to validate product variation axes and their allowed values.
func (r *AttributeRepository) ListByCategory(ctx context.Context, categoryID uuid.UUID) ([]AttributeItem, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}
	query := `
		SELECT ca.id, ca.category_id, COALESCE(pc.name, 'Global'), ca.key, ca.label, ca.value_type,
		       ca.allowed_values, ca.is_filterable, ca.sort_order, ca.created_at, ca.updated_at
		FROM category_attributes ca
		LEFT JOIN product_categories pc ON ca.category_id = pc.id
		WHERE ca.category_id = $1
		ORDER BY ca.sort_order ASC, ca.key ASC
	`
	rows, err := executor.QueryContext(ctx, query, categoryID)
	if err != nil {
		return nil, fmt.Errorf("list category attributes: %w", err)
	}
	defer rows.Close()

	var list []AttributeItem
	for rows.Next() {
		var ai AttributeItem
		if err := rows.Scan(
			&ai.ID, &ai.CategoryID, &ai.CategoryName, &ai.Key, &ai.Label, &ai.ValueType,
			&ai.AllowedValues, &ai.IsFilterable, &ai.SortOrder, &ai.CreatedAt, &ai.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan category attribute: %w", err)
		}
		list = append(list, ai)
	}
	return list, rows.Err()
}

func (r *AttributeRepository) CreateAttribute(ctx context.Context, ai *AttributeItem) error {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}
	query := `
		INSERT INTO category_attributes (id, category_id, key, label, value_type, allowed_values, is_filterable, sort_order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5::attribute_value_type, $6, $7, $8, NOW(), NOW())
		RETURNING created_at, updated_at
	`
	return tx.QueryRowContext(ctx, query,
		ai.ID, ai.CategoryID, ai.Key, ai.Label, ai.ValueType, ai.AllowedValues, ai.IsFilterable, ai.SortOrder,
	).Scan(&ai.CreatedAt, &ai.UpdatedAt)
}

// UpdateAttributeFields holds the optional columns to patch on an attribute.
// Nil pointers and nil raw messages are treated as "leave unchanged".
type UpdateAttributeFields struct {
	Label         *string
	ValueType     *string
	AllowedValues json.RawMessage
	IsFilterable  *bool
	SortOrder     *int
}

// Update applies present fields to an attribute, returning pgx.ErrNoRows when
// the attribute does not exist.
func (r *AttributeRepository) UpdateAttribute(
	ctx context.Context,
	id uuid.UUID,
	fields UpdateAttributeFields,
) error {
	if id == uuid.Nil {
		panic("update attribute: id is required")
	}

	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}

	setClauses := make([]string, 0, 5)
	args := make([]any, 0, 5)
	argIdx := 1

	if fields.Label != nil {
		setClauses = append(setClauses, fmt.Sprintf("label = $%d", argIdx))
		args = append(args, *fields.Label)
		argIdx++
	}
	if fields.ValueType != nil {
		setClauses = append(setClauses, fmt.Sprintf("value_type = $%d::attribute_value_type", argIdx))
		args = append(args, *fields.ValueType)
		argIdx++
	}
	if fields.AllowedValues != nil {
		setClauses = append(setClauses, fmt.Sprintf("allowed_values = $%d", argIdx))
		args = append(args, fields.AllowedValues)
		argIdx++
	}
	if fields.IsFilterable != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_filterable = $%d", argIdx))
		args = append(args, *fields.IsFilterable)
		argIdx++
	}
	if fields.SortOrder != nil {
		setClauses = append(setClauses, fmt.Sprintf("sort_order = $%d", argIdx))
		args = append(args, *fields.SortOrder)
		argIdx++
	}

	if len(setClauses) == 0 {
		return nil // No updates supplied
	}

	setClauses = append(setClauses, "updated_at = now()")

	query := fmt.Sprintf(`
		UPDATE category_attributes
		SET %s
		WHERE id = $%d
	`, strings.Join(setClauses, ", "), argIdx)

	args = append(args, id)

	cmd, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update attribute: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *AttributeRepository) DeleteAttribute(ctx context.Context, id uuid.UUID) error {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}
	query := `DELETE FROM category_attributes WHERE id = $1`
	cmd, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete attribute: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// ListBySellerCategory returns one merchant's custom attribute definitions
// for a category, newest first.
func (r *AttributeRepository) ListBySellerCategory(ctx context.Context, sellerID, categoryID uuid.UUID) ([]AttributeItem, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}
	query := `
		SELECT sa.id, sa.category_id, COALESCE(pc.name, ''), sa.key, sa.label, sa.value_type,
		       sa.allowed_values, sa.is_filterable, sa.sort_order, sa.created_at, sa.updated_at
		FROM seller_attributes sa
		LEFT JOIN product_categories pc ON sa.category_id = pc.id
		WHERE sa.seller_id = $1 AND sa.category_id = $2
		ORDER BY sa.sort_order ASC, sa.key ASC
	`
	rows, err := executor.QueryContext(ctx, query, sellerID, categoryID)
	if err != nil {
		return nil, fmt.Errorf("list seller attributes: %w", err)
	}
	defer rows.Close()

	var list []AttributeItem
	for rows.Next() {
		var ai AttributeItem
		if err := rows.Scan(
			&ai.ID, &ai.CategoryID, &ai.CategoryName, &ai.Key, &ai.Label, &ai.ValueType,
			&ai.AllowedValues, &ai.IsFilterable, &ai.SortOrder, &ai.CreatedAt, &ai.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan seller attribute: %w", err)
		}
		list = append(list, ai)
	}
	return list, rows.Err()
}

// CreateSellerAttribute inserts one merchant-owned attribute definition.
func (r *AttributeRepository) CreateSellerAttribute(ctx context.Context, sellerID uuid.UUID, ai *AttributeItem) error {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}
	query := `
		INSERT INTO seller_attributes (id, seller_id, category_id, key, label, value_type, allowed_values, is_filterable, sort_order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6::attribute_value_type, $7, $8, $9, NOW(), NOW())
		RETURNING created_at, updated_at
	`
	return tx.QueryRowContext(ctx, query,
		ai.ID, sellerID, ai.CategoryID, ai.Key, ai.Label, ai.ValueType, ai.AllowedValues, ai.IsFilterable, ai.SortOrder,
	).Scan(&ai.CreatedAt, &ai.UpdatedAt)
}

// UpdateSellerAttribute applies present fields to a merchant-owned attribute.
func (r *AttributeRepository) UpdateSellerAttribute(
	ctx context.Context,
	sellerID, id uuid.UUID,
	fields UpdateAttributeFields,
) error {
	if id == uuid.Nil {
		panic("update seller attribute: id is required")
	}

	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}

	setClauses := make([]string, 0, 5)
	args := make([]any, 0, 5)
	argIdx := 1

	if fields.Label != nil {
		setClauses = append(setClauses, fmt.Sprintf("label = $%d", argIdx))
		args = append(args, *fields.Label)
		argIdx++
	}
	if fields.ValueType != nil {
		setClauses = append(setClauses, fmt.Sprintf("value_type = $%d::attribute_value_type", argIdx))
		args = append(args, *fields.ValueType)
		argIdx++
	}
	if fields.AllowedValues != nil {
		setClauses = append(setClauses, fmt.Sprintf("allowed_values = $%d", argIdx))
		args = append(args, fields.AllowedValues)
		argIdx++
	}
	if fields.IsFilterable != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_filterable = $%d", argIdx))
		args = append(args, *fields.IsFilterable)
		argIdx++
	}
	if fields.SortOrder != nil {
		setClauses = append(setClauses, fmt.Sprintf("sort_order = $%d", argIdx))
		args = append(args, *fields.SortOrder)
		argIdx++
	}

	if len(setClauses) == 0 {
		return nil // No updates supplied
	}

	setClauses = append(setClauses, "updated_at = now()")

	query := fmt.Sprintf(`
		UPDATE seller_attributes
		SET %s
		WHERE seller_id = $%d AND id = $%d
	`, strings.Join(setClauses, ", "), argIdx, argIdx+1)
	args = append(args, sellerID, id)

	cmd, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update seller attribute: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// DeleteSellerAttribute removes a merchant-owned attribute definition.
func (r *AttributeRepository) DeleteSellerAttribute(ctx context.Context, sellerID, id uuid.UUID) error {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}
	query := `DELETE FROM seller_attributes WHERE seller_id = $1 AND id = $2`
	cmd, err := tx.ExecContext(ctx, query, sellerID, id)
	if err != nil {
		return fmt.Errorf("delete seller attribute: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
