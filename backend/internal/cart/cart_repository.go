// Package cart manages persistent storage and retrieval for shopping carts and cart line items.
package cart

import (
	"context"
	"errors"

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

// getCartItems retrieves all items belonging to a cart ordered by carted_at ascending.
func (r *CartRepository) getCartItems(
	ctx context.Context,
	qe database.QueryExecutor,
	cartID uuid.UUID,
) ([]model.CartItem, error) {
	query := `
		SELECT id, variant_id, quantity, price_at_purchase, currency, carted_at
		FROM cart_items
		WHERE cart_id = $1
		ORDER BY carted_at ASC
	`
	rows, err := qe.Query(ctx, query, cartID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.CartItem, 0)
	for rows.Next() {
		var item model.CartItem
		item.CartID = cartID
		err := rows.Scan(
			&item.ID,
			&item.VariantID,
			&item.Quantity,
			&item.PriceAtPurchase,
			&item.Currency,
			&item.CartedAt,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// GetCartByID retrieves a cart record by its primary key UUID along with its items.
func (r *CartRepository) GetCartByID(
	ctx context.Context,
	qe database.QueryExecutor,
	id uuid.UUID,
) (*model.Cart, error) {
	query := `
		SELECT id, user_id, session_id, created_at, updated_at
		FROM carts
		WHERE id = $1
	`
	cart := &model.Cart{Items: make([]model.CartItem, 0)}
	err := qe.QueryRow(ctx, query, id).Scan(
		&cart.ID,
		&cart.UserID,
		&cart.SessionID,
		&cart.CreatedAt,
		&cart.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, database.ErrNotFound
		}
		return nil, err
	}

	items, err := r.getCartItems(ctx, qe, cart.ID)
	if err != nil {
		return nil, err
	}
	cart.Items = items

	return cart, nil
}

// GetCartByUserID retrieves the most recently created cart for a registered customer along with all items.
func (r *CartRepository) GetCartByUserID(
	ctx context.Context,
	qe database.QueryExecutor,
	userID uuid.UUID,
) (*model.Cart, error) {
	query := `
		SELECT id, user_id, session_id, created_at, updated_at
		FROM carts
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	cart := &model.Cart{Items: make([]model.CartItem, 0)}
	err := qe.QueryRow(ctx, query, userID).Scan(
		&cart.ID,
		&cart.UserID,
		&cart.SessionID,
		&cart.CreatedAt,
		&cart.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, database.ErrNotFound
		}
		return nil, err
	}

	items, err := r.getCartItems(ctx, qe, cart.ID)
	if err != nil {
		return nil, err
	}
	cart.Items = items

	return cart, nil
}

// GetCartBySessionID retrieves the most recently created anonymous guest cart for a session ID.
func (r *CartRepository) GetCartBySessionID(
	ctx context.Context,
	qe database.QueryExecutor,
	sessionID string,
) (*model.Cart, error) {
	query := `
		SELECT id, user_id, session_id, created_at, updated_at
		FROM carts
		WHERE session_id = $1 AND user_id IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`
	cart := &model.Cart{Items: make([]model.CartItem, 0)}
	err := qe.QueryRow(ctx, query, sessionID).Scan(
		&cart.ID,
		&cart.UserID,
		&cart.SessionID,
		&cart.CreatedAt,
		&cart.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, database.ErrNotFound
		}
		return nil, err
	}

	items, err := r.getCartItems(ctx, qe, cart.ID)
	if err != nil {
		return nil, err
	}
	cart.Items = items

	return cart, nil
}

func (r *CartRepository) AddItem(
	ctx context.Context,
	qe *database.Tx,
	item *model.CartItem,
) error {
	query := `
		INSERT INTO cart_items (cart_id, variant_id, quantity, price_at_purchase, currency, carted_at)
		VALUES ($1, $2, $3, $4, $5, now())
		ON CONFLICT (cart_id, variant_id) DO UPDATE
		SET 
			quantity = cart_items.quantity + EXCLUDED.quantity,
			price_at_purchase = EXCLUDED.price_at_purchase,
			currency = EXCLUDED.currency,
			carted_at = now()
		RETURNING id, quantity, carted_at
	`
	err := qe.QueryRow(
		ctx,
		query,
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

func (r *CartRepository) UpdateItemQuantity(
	ctx context.Context,
	tx *database.Tx,
	cartID uuid.UUID,
	itemID int64,
	quantity int,
) error {
	query := `
		UPDATE cart_items
		SET quantity = $1
		WHERE id = $2 AND cart_id = $3
	`
	tag, err := tx.Exec(ctx, query, quantity, itemID, cartID)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	updateCartQuery := `UPDATE carts SET updated_at = now() WHERE id = $1`
	_, _ = tx.Exec(ctx, updateCartQuery, cartID)
	return nil
}

func (r *CartRepository) RemoveItem(
	ctx context.Context,
	tx *database.Tx,
	cartID uuid.UUID,
	itemID int64,
) error {
	query := `
		DELETE FROM cart_items
		WHERE id = $1 AND cart_id = $2
	`
	tag, err := tx.Exec(ctx, query, itemID, cartID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	updateCartQuery := `UPDATE carts SET updated_at = now() WHERE id = $1`
	_, _ = tx.Exec(ctx, updateCartQuery, cartID)
	return nil
}

func (r *CartRepository) ClearCart(
	ctx context.Context,
	qe database.QueryExecutor,
	cartID uuid.UUID,
) error {
	query := `DELETE FROM cart_items WHERE cart_id = $1`
	_, err := qe.Exec(ctx, query, cartID)
	if err != nil {
		return err
	}

	updateCartQuery := `UPDATE carts SET updated_at = now() WHERE id = $1`
	_, _ = qe.Exec(ctx, updateCartQuery, cartID)
	return nil
}

func (r *CartRepository) DeleteCart(
	ctx context.Context,
	tx database.QueryExecutor,
	cartID uuid.UUID,
) error {
	query := `DELETE FROM carts WHERE id = $1`
	_, err := tx.Exec(ctx, query, cartID)
	return err
}
