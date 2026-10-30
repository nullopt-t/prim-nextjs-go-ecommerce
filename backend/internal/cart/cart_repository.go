package cart

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

// CartRepository provides PostgreSQL data access methods for managing customer shopping carts and line items.
// All transactional operations accept a database.QueryExecutor to participate in an outer atomic transaction.
type CartRepository struct{}

// NewRepository initializes and returns a new CartRepository instance.
func NewRepository() *CartRepository {
	return &CartRepository{}
}

// CreateCart inserts a new cart record into the 'carts' table.
// It stores the cart ID, optional customer userID, and optional guest sessionID.
// The created_at and updated_at timestamps are populated by PostgreSQL now() and scanned back into the model.
func (r *CartRepository) CreateCart(
	ctx context.Context,
	cart *model.Cart,
) error {
	executor, err := database.Executor(ctx)
	if err != nil {
		return err
	}
	// SQL query to persist the base cart entity.
	query := `
		INSERT INTO carts (id, user_id, session_id, created_at, updated_at)
		VALUES ($1, $2, $3, now(), now())
		RETURNING created_at, updated_at
	`
	// Execute insertion and assign returned database timestamps to the passed cart pointer.
	return executor.QueryRow(ctx, query, cart.ID, cart.UserID, cart.SessionID).Scan(&cart.CreatedAt, &cart.UpdatedAt)
}

// getCartItems retrieves all items belonging to a cart ordered by carted_at ascending.
func (r *CartRepository) getCartItems(
	ctx context.Context,
	cartID uuid.UUID,
) ([]model.CartItem, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}
	query := `
		SELECT id, variant_id, quantity, price_at_purchase, currency, carted_at
		FROM cart_items
		WHERE cart_id = $1
		ORDER BY carted_at ASC
	`
	rows, err := executor.Query(ctx, query, cartID)
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
	id uuid.UUID,
) (*model.Cart, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}
	query := `
		SELECT id, user_id, session_id, created_at, updated_at
		FROM carts
		WHERE id = $1
	`
	cart := &model.Cart{Items: make([]model.CartItem, 0)}
	err = executor.QueryRow(ctx, query, id).Scan(
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

	items, err := r.getCartItems(ctx, cart.ID)
	if err != nil {
		return nil, err
	}
	cart.Items = items

	return cart, nil
}

// LockCartByUserID acquires a row-level write lock (SELECT ... FOR UPDATE) on
// the customer's active cart so concurrent checkouts serialize instead of both
// reading the same items and racing to create duplicate orders. It returns
// database.ErrNotFound when the customer has no cart.
func (r *CartRepository) LockCartByUserID(
	ctx context.Context,
	userID uuid.UUID,
) error {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}
	query := `
		SELECT id
		FROM carts
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1
		FOR UPDATE
	`
	var id uuid.UUID
	err = tx.QueryRow(ctx, query, userID).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return database.ErrNotFound
		}
		return err
	}
	return nil
}

// LockCartByID acquires a row-level write lock on a single cart so a mutation
// that inserts cart_items takes the cart -> items lock order that checkout also
// follows. This prevents a deadlock where checkout holds the cart row while
// waiting for a cart_items row that the mutation already locked.
func (r *CartRepository) LockCartByID(
	ctx context.Context,
	cartID uuid.UUID,
) error {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}
	var id uuid.UUID
	err = tx.QueryRow(ctx, `SELECT id FROM carts WHERE id = $1 FOR UPDATE`, cartID).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return database.ErrNotFound
		}
		return err
	}
	return nil
}

// GetCartByUserID retrieves the most recently created cart for a registered customer along with all items.
func (r *CartRepository) GetCartByUserID(
	ctx context.Context,
	userID uuid.UUID,
) (*model.Cart, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}
	query := `
		SELECT id, user_id, session_id, created_at, updated_at
		FROM carts
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	cart := &model.Cart{Items: make([]model.CartItem, 0)}
	err = executor.QueryRow(ctx, query, userID).Scan(
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

	items, err := r.getCartItems(ctx, cart.ID)
	if err != nil {
		return nil, err
	}
	cart.Items = items

	return cart, nil
}

// GetCartBySessionID retrieves the most recently created anonymous guest cart for a session ID.
func (r *CartRepository) GetCartBySessionID(
	ctx context.Context,
	sessionID string,
) (*model.Cart, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}
	query := `
		SELECT id, user_id, session_id, created_at, updated_at
		FROM carts
		WHERE session_id = $1 AND user_id IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`
	cart := &model.Cart{Items: make([]model.CartItem, 0)}
	err = executor.QueryRow(ctx, query, sessionID).Scan(
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

	items, err := r.getCartItems(ctx, cart.ID)
	if err != nil {
		return nil, err
	}
	cart.Items = items

	return cart, nil
}

func (r *CartRepository) AddItem(
	ctx context.Context,
	item *model.CartItem,
) error {
	// Take the cart -> items lock order shared with checkout before touching
	// cart_items; otherwise the child-first upsert can deadlock with checkout.
	if err := r.LockCartByID(ctx, item.CartID); err != nil {
		return err
	}

	executor, err := database.Executor(ctx)
	if err != nil {
		return err
	}

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
	err = executor.QueryRow(
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
	_, _ = executor.Exec(ctx, updateCartQuery, item.CartID)
	return nil
}

func (r *CartRepository) UpdateItemQuantity(
	ctx context.Context,
	cartID uuid.UUID,
	itemID int64,
	quantity int,
) error {
	if err := r.LockCartByID(ctx, cartID); err != nil {
		return err
	}

	executor, err := database.Executor(ctx)
	if err != nil {
		return err
	}

	query := `
		UPDATE cart_items
		SET quantity = $1
		WHERE id = $2 AND cart_id = $3
	`
	tag, err := executor.Exec(ctx, query, quantity, itemID, cartID)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	updateCartQuery := `UPDATE carts SET updated_at = now() WHERE id = $1`
	_, _ = executor.Exec(ctx, updateCartQuery, cartID)
	return nil
}

func (r *CartRepository) RemoveItem(
	ctx context.Context,
	cartID uuid.UUID,
	itemID int64,
) error {
	if err := r.LockCartByID(ctx, cartID); err != nil {
		return err
	}

	executor, err := database.Executor(ctx)
	if err != nil {
		return err
	}

	query := `
		DELETE FROM cart_items
		WHERE id = $1 AND cart_id = $2
	`
	tag, err := executor.Exec(ctx, query, itemID, cartID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	updateCartQuery := `UPDATE carts SET updated_at = now() WHERE id = $1`
	_, _ = executor.Exec(ctx, updateCartQuery, cartID)
	return nil
}

func (r *CartRepository) ClearCart(
	ctx context.Context,
	cartID uuid.UUID,
) error {
	if err := r.LockCartByID(ctx, cartID); err != nil {
		return err
	}

	executor, err := database.Executor(ctx)
	if err != nil {
		return err
	}

	query := `DELETE FROM cart_items WHERE cart_id = $1`
	_, err = executor.Exec(ctx, query, cartID)
	if err != nil {
		return err
	}

	updateCartQuery := `UPDATE carts SET updated_at = now() WHERE id = $1`
	_, _ = executor.Exec(ctx, updateCartQuery, cartID)
	return nil
}

func (r *CartRepository) DeleteCart(
	ctx context.Context,
	cartID uuid.UUID,
) error {
	executor, err := database.Executor(ctx)
	if err != nil {
		return err
	}
	query := `DELETE FROM carts WHERE id = $1`
	_, err = executor.Exec(ctx, query, cartID)
	return err
}
