package order

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
)

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

// --- Order CRUD ---

func (r *Repository) CreateOrder(
	ctx context.Context,
	qe database.QueryExecutor,
	o *model.Order,
) error {
	shippingJSON, err := json.Marshal(o.ShippingAddress)
	if err != nil {
		return fmt.Errorf("marshal shipping address: %w", err)
	}
	billingJSON, err := json.Marshal(o.BillingAddress)
	if err != nil {
		return fmt.Errorf("marshal billing address: %w", err)
	}

	query := `
		INSERT INTO orders (
			id, customer_id, customer_email, shipping_address, billing_address,
			status, coupon_id, discount_amount, total_amount, currency, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, now(), now())
		RETURNING created_at, updated_at
	`
	err = qe.QueryRow(
		ctx,
		query,
		o.ID,
		o.CustomerID,
		o.CustomerEmail,
		shippingJSON,
		billingJSON,
		o.Status,
		o.CouponID,
		o.DiscountAmount,
		o.TotalAmount,
		o.Currency,
	).Scan(&o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create order: %w", err)
	}
	return nil
}

func (r *Repository) AddOrderItem(
	ctx context.Context,
	qe database.QueryExecutor,
	item *model.OrderItem,
) error {
	query := `
		INSERT INTO order_items (id, order_id, variant_id, quantity, price_at_purchase, product_snapshot)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := qe.Exec(
		ctx,
		query,
		item.ID,
		item.OrderID,
		item.VariantID,
		item.Quantity,
		item.PriceAtPurchase,
		item.ProductSnapshot,
	)
	if err != nil {
		return fmt.Errorf("add order item: %w", err)
	}
	return nil
}

func (r *Repository) AddStatusHistory(
	ctx context.Context,
	qe database.QueryExecutor,
	orderID uuid.UUID,
	status model.OrderStatus,
	notes *string,
) error {
	query := `
		INSERT INTO order_status_history (id, order_id, status, notes, created_at)
		VALUES ($1, $2, $3, $4, now())
	`
	_, err := qe.Exec(ctx, query, uuid.New(), orderID, status, notes)
	if err != nil {
		return fmt.Errorf("add status history: %w", err)
	}
	return nil
}

func (r *Repository) GetOrderByID(
	ctx context.Context,
	qe database.QueryExecutor,
	id uuid.UUID,
) (*model.Order, error) {
	query := `
		SELECT id, customer_id, customer_email, shipping_address, billing_address,
		       status, coupon_id, discount_amount, total_amount, currency, created_at, updated_at
		FROM orders
		WHERE id = $1
	`
	row := qe.QueryRow(ctx, query, id)

	var o model.Order
	var shippingBytes, billingBytes []byte

	err := row.Scan(
		&o.ID,
		&o.CustomerID,
		&o.CustomerEmail,
		&shippingBytes,
		&billingBytes,
		&o.Status,
		&o.CouponID,
		&o.DiscountAmount,
		&o.TotalAmount,
		&o.Currency,
		&o.CreatedAt,
		&o.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get order by id: %w", err)
	}

	if err := json.Unmarshal(shippingBytes, &o.ShippingAddress); err != nil {
		return nil, fmt.Errorf("unmarshal shipping address: %w", err)
	}
	if err := json.Unmarshal(billingBytes, &o.BillingAddress); err != nil {
		return nil, fmt.Errorf("unmarshal billing address: %w", err)
	}

	items, err := r.GetOrderItems(ctx, qe, o.ID)
	if err == nil {
		o.Items = items
	}

	return &o, nil
}

func (r *Repository) GetOrdersByCustomerID(
	ctx context.Context,
	qe database.QueryExecutor,
	customerID uuid.UUID,
) ([]model.Order, error) {
	query := `
		SELECT id, customer_id, customer_email, shipping_address, billing_address,
		       status, coupon_id, discount_amount, total_amount, currency, created_at, updated_at
		FROM orders
		WHERE customer_id = $1
		ORDER BY created_at DESC
	`
	rows, err := qe.Query(ctx, query, customerID)
	if err != nil {
		return nil, fmt.Errorf("get orders by customer id: %w", err)
	}
	defer rows.Close()

	orders := make([]model.Order, 0)
	for rows.Next() {
		var o model.Order
		var shippingBytes, billingBytes []byte

		if err := rows.Scan(
			&o.ID,
			&o.CustomerID,
			&o.CustomerEmail,
			&shippingBytes,
			&billingBytes,
			&o.Status,
			&o.CouponID,
			&o.DiscountAmount,
			&o.TotalAmount,
			&o.Currency,
			&o.CreatedAt,
			&o.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}

		if err := json.Unmarshal(shippingBytes, &o.ShippingAddress); err != nil {
			return nil, fmt.Errorf("unmarshal shipping address: %w", err)
		}
		if err := json.Unmarshal(billingBytes, &o.BillingAddress); err != nil {
			return nil, fmt.Errorf("unmarshal billing address: %w", err)
		}

		orders = append(orders, o)
	}

	return orders, rows.Err()
}

func (r *Repository) GetOrderItems(
	ctx context.Context,
	qe database.QueryExecutor,
	orderID uuid.UUID,
) ([]model.OrderItem, error) {
	query := `
		SELECT id, order_id, variant_id, quantity, price_at_purchase, product_snapshot
		FROM order_items
		WHERE order_id = $1
	`
	rows, err := qe.Query(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("get order items: %w", err)
	}
	defer rows.Close()

	items := make([]model.OrderItem, 0)
	for rows.Next() {
		var item model.OrderItem
		if err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.VariantID,
			&item.Quantity,
			&item.PriceAtPurchase,
			&item.ProductSnapshot,
		); err != nil {
			return nil, fmt.Errorf("scan order item: %w", err)
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *Repository) UpdateOrderStatus(
	ctx context.Context,
	qe database.QueryExecutor,
	orderID uuid.UUID,
	status model.OrderStatus,
) error {
	query := `
		UPDATE orders
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`
	tag, err := qe.Exec(ctx, query, status, orderID)
	if err != nil {
		return fmt.Errorf("update order status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
