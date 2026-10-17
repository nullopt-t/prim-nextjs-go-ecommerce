// Package cart manages persistent storage and retrieval for shopping carts and cart line items.
package cart

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

// CartRepository handles database operations against the 'carts' and 'cart_items' tables.
// It accepts a database.QueryExecutor to support execution within or outside database transactions.
type CartRepository struct{}

// NewRepository constructs and returns a new CartRepository instance.
func NewRepository() *CartRepository {
	return &CartRepository{}
}

// CreateCart inserts a new cart record into the 'carts' table.
// It stores the cart ID, optional customer userID, and optional guest sessionID.
// The created_at and updated_at timestamps are populated by PostgreSQL now() and scanned back into the model.
func (r *CartRepository) CreateCart(
	ctx context.Context,
	qe database.QueryExecutor,
	cart *model.Cart,
) error {
	// SQL query to persist the base cart entity.
	query := `
		INSERT INTO carts (id, user_id, session_id, created_at, updated_at)
		VALUES ($1, $2, $3, now(), now())
		RETURNING created_at, updated_at
	`
	// Execute insertion and assign returned database timestamps to the passed cart pointer.
	err := qe.QueryRow(ctx, query, cart.ID, cart.UserID, cart.SessionID).Scan(&cart.CreatedAt, &cart.UpdatedAt)
	return err
}

// GetCartByID retrieves a non-deleted cart record by its primary key UUID.
// Returns pgx.ErrNoRows if no active cart matches the given identifier.
func (r *CartRepository) GetCartByID(
	ctx context.Context,
	qe database.QueryExecutor,
	id uuid.UUID,
) (*model.Cart, error) {
	// Select columns for an active, non-soft-deleted cart record.
	query := `
		SELECT id, user_id, session_id, created_at, updated_at, deleted_at
		FROM carts
		WHERE id = $1 AND deleted_at IS NULL
	`
	// Execute query using provided QueryExecutor.
	row := qe.QueryRow(ctx, query, id)
	cart := &model.Cart{}
	// Scan database fields into the model.Cart struct fields.
	err := row.Scan(&cart.ID, &cart.UserID, &cart.SessionID, &cart.CreatedAt, &cart.UpdatedAt, &cart.DeletedAt)
	if err != nil {
		return nil, err
	}
	return cart, nil
}

// GetCartByUserID retrieves the most recently created active cart for a registered customer.
func (r *CartRepository) GetCartByUserID(
	ctx context.Context,
	qe database.QueryExecutor,
	userID uuid.UUID,
) (*model.Cart, error) {
	// Query the newest active cart belonging to the customer.
	query := `
		SELECT id, user_id, session_id, created_at, updated_at, deleted_at
		FROM carts
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`
	row := qe.QueryRow(ctx, query, userID)
	cart := &model.Cart{}
	// Unmarshal row data into cart model.
	err := row.Scan(&cart.ID, &cart.UserID, &cart.SessionID, &cart.CreatedAt, &cart.UpdatedAt, &cart.DeletedAt)
	if err != nil {
		return nil, err
	}
	return cart, nil
}

// GetCartByUserIDForUpdate retrieves the most recently created active cart for a registered customer with row-level locking.
func (r *CartRepository) GetCartByUserIDForUpdate(
	ctx context.Context,
	qe database.QueryExecutor,
	userID uuid.UUID,
) (*model.Cart, error) {
	query := `
		SELECT id, user_id, session_id, created_at, updated_at, deleted_at
		FROM carts
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
		FOR UPDATE
	`
	row := qe.QueryRow(ctx, query, userID)
	cart := &model.Cart{}
	err := row.Scan(&cart.ID, &cart.UserID, &cart.SessionID, &cart.CreatedAt, &cart.UpdatedAt, &cart.DeletedAt)
	if err != nil {
		return nil, err
	}
	return cart, nil
}

// GetCartBySessionID retrieves the most recently created active anonymous guest cart for a session ID.
// Filters exclusively for user_id IS NULL to prevent leaking authenticated carts associated with obsolete cookies.
func (r *CartRepository) GetCartBySessionID(
	ctx context.Context,
	qe database.QueryExecutor,
	sessionID string,
) (*model.Cart, error) {
	// Query active guest cart where user_id is null.
	query := `
		SELECT id, user_id, session_id, created_at, updated_at, deleted_at
		FROM carts
		WHERE session_id = $1 AND user_id IS NULL AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`
	row := qe.QueryRow(ctx, query, sessionID)
	cart := &model.Cart{}
	// Scan the database row into the cart model.
	err := row.Scan(&cart.ID, &cart.UserID, &cart.SessionID, &cart.CreatedAt, &cart.UpdatedAt, &cart.DeletedAt)
	if err != nil {
		return nil, err
	}
	return cart, nil
}

// GetCartBySessionIDForUpdate retrieves the active anonymous guest cart with row-level locking.
func (r *CartRepository) GetCartBySessionIDForUpdate(
	ctx context.Context,
	qe database.QueryExecutor,
	sessionID string,
) (*model.Cart, error) {
	query := `
		SELECT id, user_id, session_id, created_at, updated_at, deleted_at
		FROM carts
		WHERE session_id = $1 AND user_id IS NULL AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
		FOR UPDATE
	`
	row := qe.QueryRow(ctx, query, sessionID)
	cart := &model.Cart{}
	err := row.Scan(&cart.ID, &cart.UserID, &cart.SessionID, &cart.CreatedAt, &cart.UpdatedAt, &cart.DeletedAt)
	if err != nil {
		return nil, err
	}
	return cart, nil
}

// GetCartItems retrieves all active (non-soft-deleted) items in chronological order of addition for a cart.
func (r *CartRepository) GetCartItems(
	ctx context.Context,
	qe database.QueryExecutor,
	cartID uuid.UUID,
) ([]model.CartItem, error) {
	// Query all active cart items linked to the cart.
	query := `
		SELECT id, cart_id, variant_id, quantity, price_at_purchase, currency, carted_at, deleted_at
		FROM cart_items
		WHERE cart_id = $1 AND deleted_at IS NULL
		ORDER BY carted_at ASC
	`
	rows, err := qe.Query(ctx, query, cartID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Initialize empty slice to prevent returning nil in JSON responses when cart is empty.
	items := make([]model.CartItem, 0)
	for rows.Next() {
		var item model.CartItem
		// Scan each column into the model.CartItem struct.
		if err := rows.Scan(
			&item.ID,
			&item.CartID,
			&item.VariantID,
			&item.Quantity,
			&item.PriceAtPurchase,
			&item.Currency,
			&item.CartedAt,
			&item.DeletedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	// Verify that iteration encountered no I/O or protocol errors.
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// GetItemByID fetches an active cart item matching either the item record UUID or the variant UUID within the specified cart.
// Supporting both ID types provides backward compatibility for client operations targeting items by variant SKU.
func (r *CartRepository) GetItemByID(
	ctx context.Context,
	qe database.QueryExecutor,
	cartID uuid.UUID,
	itemID uuid.UUID,
) (*model.CartItem, error) {
	// Query item by item ID or variant ID within the scoped cart.
	query := `
		SELECT id, cart_id, variant_id, quantity, price_at_purchase, currency, carted_at, deleted_at
		FROM cart_items
		WHERE cart_id = $1 AND (id = $2 OR variant_id = $2) AND deleted_at IS NULL
		LIMIT 1
	`
	row := qe.QueryRow(ctx, query, cartID, itemID)
	item := &model.CartItem{}
	err := row.Scan(
		&item.ID,
		&item.CartID,
		&item.VariantID,
		&item.Quantity,
		&item.PriceAtPurchase,
		&item.Currency,
		&item.CartedAt,
		&item.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return item, nil
}

// GetItemByVariantID fetches an active cart item for a specific product variant in the specified cart.
func (r *CartRepository) GetItemByVariantID(
	ctx context.Context,
	qe database.QueryExecutor,
	cartID uuid.UUID,
	variantID uuid.UUID,
) (*model.CartItem, error) {
	// Locate line item strictly by cart_id and variant_id.
	query := `
		SELECT id, cart_id, variant_id, quantity, price_at_purchase, currency, carted_at, deleted_at
		FROM cart_items
		WHERE cart_id = $1 AND variant_id = $2 AND deleted_at IS NULL
	`
	row := qe.QueryRow(ctx, query, cartID, variantID)
	item := &model.CartItem{}
	err := row.Scan(
		&item.ID,
		&item.CartID,
		&item.VariantID,
		&item.Quantity,
		&item.PriceAtPurchase,
		&item.Currency,
		&item.CartedAt,
		&item.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return item, nil
}

// AddItem inserts or updates a line item in the 'cart_items' table.
// If the variant already exists in the cart, its quantity is incremented atomically.
// Also updates the parent cart's updated_at timestamp.
func (r *CartRepository) AddItem(
	ctx context.Context,
	qe database.QueryExecutor,
	item *model.CartItem,
) error {
	query := `
		INSERT INTO cart_items (id, cart_id, variant_id, quantity, price_at_purchase, currency, carted_at, deleted_at)
		VALUES ($1, $2, $3, $4, $5, $6, now(), NULL)
		ON CONFLICT (cart_id, variant_id) DO UPDATE
		SET quantity = CASE
				WHEN cart_items.deleted_at IS NOT NULL THEN EXCLUDED.quantity
				ELSE cart_items.quantity + EXCLUDED.quantity
			END,
			price_at_purchase = EXCLUDED.price_at_purchase,
			currency = EXCLUDED.currency,
			deleted_at = NULL
		RETURNING id, quantity, carted_at
	`
	err := qe.QueryRow(
		ctx,
		query,
		item.ID,
		item.CartID,
		item.VariantID,
		item.Quantity,
		item.PriceAtPurchase,
		item.Currency,
	).Scan(&item.ID, &item.Quantity, &item.CartedAt)
	if err != nil {
		return err
	}

	updateCartQuery := `UPDATE carts SET updated_at = now() WHERE id = $1`
	_, _ = qe.Exec(ctx, updateCartQuery, item.CartID)
	return nil
}

// UpdateItemQuantity updates the quantity of an existing line item in the cart.
// If no rows are affected (e.g. item does not exist or has been deleted), returns pgx.ErrNoRows.
// Also refreshes the parent cart's updated_at timestamp.
func (r *CartRepository) UpdateItemQuantity(
	ctx context.Context,
	qe database.QueryExecutor,
	cartID uuid.UUID,
	itemID uuid.UUID,
	quantity int,
) error {
	// Execute quantity update matching either item record ID or variant ID within the designated cart.
	query := `
		UPDATE cart_items
		SET quantity = $1
		WHERE (id = $2 OR variant_id = $2) AND cart_id = $3 AND deleted_at IS NULL
	`
	tag, err := qe.Exec(ctx, query, quantity, itemID, cartID)
	if err != nil {
		return err
	}
	// If 0 rows affected, report ErrNoRows so service can map to a 404 CartItemNotFound error.
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	updateCartQuery := `UPDATE carts SET updated_at = now() WHERE id = $1`
	_, _ = qe.Exec(ctx, updateCartQuery, cartID)
	return nil
}

// RemoveItem removes a line item from the cart matching item UUID or variant UUID via soft-delete.
// Also updates parent cart's updated_at timestamp.
func (r *CartRepository) RemoveItem(
	ctx context.Context,
	qe database.QueryExecutor,
	cartID uuid.UUID,
	itemID uuid.UUID,
) error {
	query := `
		UPDATE cart_items
		SET deleted_at = now()
		WHERE (id = $1 OR variant_id = $1) AND cart_id = $2 AND deleted_at IS NULL
	`
	tag, err := qe.Exec(ctx, query, itemID, cartID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	updateCartQuery := `UPDATE carts SET updated_at = now() WHERE id = $1`
	_, _ = qe.Exec(ctx, updateCartQuery, cartID)
	return nil
}

// ClearCart soft-deletes all items belonging to the given cart ID and refreshes updated_at.
func (r *CartRepository) ClearCart(
	ctx context.Context,
	qe database.QueryExecutor,
	cartID uuid.UUID,
) error {
	query := `
		UPDATE cart_items
		SET deleted_at = now()
		WHERE cart_id = $1 AND deleted_at IS NULL
	`
	_, err := qe.Exec(ctx, query, cartID)
	if err != nil {
		return err
	}

	updateCartQuery := `UPDATE carts SET updated_at = now() WHERE id = $1`
	_, _ = qe.Exec(ctx, updateCartQuery, cartID)
	return nil
}

// DeleteCart soft-deletes the cart record from the database.
func (r *CartRepository) DeleteCart(
	ctx context.Context,
	qe database.QueryExecutor,
	cartID uuid.UUID,
) error {
	query := `UPDATE carts SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL`
	_, err := qe.Exec(ctx, query, cartID)
	return err
}
