package cart

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
)

type CartRepository struct{}

func NewRepository() *CartRepository {
	return &CartRepository{}
}

func (r *CartRepository) CreateCart(
	ctx context.Context,
	qe database.QueryExecutor,
	cart *model.Cart,
) error {
	query := `
		INSERT INTO carts (id, user_id, session_id, created_at, updated_at)
		VALUES ($1, $2, $3, now(), now())
		RETURNING created_at, updated_at
	`
	err := qe.QueryRow(ctx, query, cart.ID, cart.UserID, cart.SessionID).Scan(&cart.CreatedAt, &cart.UpdatedAt)
	return err
}

func (r *CartRepository) GetCartByID(
	ctx context.Context,
	qe database.QueryExecutor,
	id uuid.UUID,
) (*model.Cart, error) {
	query := `
		SELECT id, user_id, session_id, created_at, updated_at, deleted_at
		FROM carts
		WHERE id = $1 AND deleted_at IS NULL
	`
	row := qe.QueryRow(ctx, query, id)
	cart := &model.Cart{}
	err := row.Scan(&cart.ID, &cart.UserID, &cart.SessionID, &cart.CreatedAt, &cart.UpdatedAt, &cart.DeletedAt)
	if err != nil {
		return nil, err
	}
	return cart, nil
}

func (r *CartRepository) GetCartByUserID(
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
	`
	row := qe.QueryRow(ctx, query, userID)
	cart := &model.Cart{}
	err := row.Scan(&cart.ID, &cart.UserID, &cart.SessionID, &cart.CreatedAt, &cart.UpdatedAt, &cart.DeletedAt)
	if err != nil {
		return nil, err
	}
	return cart, nil
}

func (r *CartRepository) GetCartBySessionID(
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
	`
	row := qe.QueryRow(ctx, query, sessionID)
	cart := &model.Cart{}
	err := row.Scan(&cart.ID, &cart.UserID, &cart.SessionID, &cart.CreatedAt, &cart.UpdatedAt, &cart.DeletedAt)
	if err != nil {
		return nil, err
	}
	return cart, nil
}

func (r *CartRepository) GetCartItems(
	ctx context.Context,
	qe database.QueryExecutor,
	cartID uuid.UUID,
) ([]model.CartItem, error) {
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

	items := make([]model.CartItem, 0)
	for rows.Next() {
		var item model.CartItem
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
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *CartRepository) GetItemByVariantID(
	ctx context.Context,
	qe database.QueryExecutor,
	cartID uuid.UUID,
	variantID uuid.UUID,
) (*model.CartItem, error) {
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

func (r *CartRepository) AddItem(
	ctx context.Context,
	qe database.QueryExecutor,
	item *model.CartItem,
) error {
	query := `
		INSERT INTO cart_items (id, cart_id, variant_id, quantity, price_at_purchase, currency, carted_at)
		VALUES ($1, $2, $3, $4, $5, $6, now())
		RETURNING carted_at
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
	).Scan(&item.CartedAt)
	return err
}

func (r *CartRepository) UpdateItemQuantity(
	ctx context.Context,
	qe database.QueryExecutor,
	cartID uuid.UUID,
	itemID uuid.UUID,
	quantity int,
) error {
	query := `
		UPDATE cart_items
		SET quantity = $1
		WHERE id = $2 AND cart_id = $3 AND deleted_at IS NULL
	`
	tag, err := qe.Exec(ctx, query, quantity, itemID, cartID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *CartRepository) RemoveItem(
	ctx context.Context,
	qe database.QueryExecutor,
	cartID uuid.UUID,
	itemID uuid.UUID,
) error {
	query := `
		DELETE FROM cart_items
		WHERE id = $1 AND cart_id = $2
	`
	tag, err := qe.Exec(ctx, query, itemID, cartID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *CartRepository) ClearCart(
	ctx context.Context,
	qe database.QueryExecutor,
	cartID uuid.UUID,
) error {
	query := `
		DELETE FROM cart_items
		WHERE cart_id = $1
	`
	_, err := qe.Exec(ctx, query, cartID)
	return err
}

func (r *CartRepository) DeleteCart(
	ctx context.Context,
	qe database.QueryExecutor,
	cartID uuid.UUID,
) error {
	query := `DELETE FROM carts WHERE id = $1`
	_, err := qe.Exec(ctx, query, cartID)
	return err
}
