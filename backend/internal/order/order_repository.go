// Package order provides database persistence layer operations for the order domain.
package order

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

// Repository implements persistence queries and mutations for orders, order items, and status history.
type Repository struct{}

// NewRepository constructs a new Order Repository instance.
func NewRepository() *Repository {
	return &Repository{}
}

// --- Order CRUD ---

// CreateOrder inserts a new order record into the orders table within the provided QueryExecutor transaction.
// It serializes the structured ShippingAddress value to JSONB and scans the server-assigned timestamps.
func (r *Repository) CreateOrder(
	ctx context.Context,
	qe database.QueryExecutor,
	o *model.Order,
) error {
	shippingJSON, err := json.Marshal(o.ShippingAddress)
	if err != nil {
		return fmt.Errorf("marshal shipping address: %w", err)
	}

	query := `
		INSERT INTO orders (
			id, customer_id, customer_email, shipping_address,
			status, coupon_id, discount_amount, total_amount, currency, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, now(), now())
		RETURNING created_at, updated_at
	`
	err = qe.QueryRow(
		ctx,
		query,
		o.ID,
		o.CustomerID,
		o.CustomerEmail,
		shippingJSON,
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

// AddOrderItem inserts a single order line item into the order_items table.
// It records the variant ID, quantity, purchase price, and the immutable product snapshot.
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

// AddStatusHistory inserts an audit log entry into the order_status_history table,
// recording state transitions, optional operational notes, and timestamps.
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

// GetOrderByID retrieves an order record by its primary key UUID.
// It also queries and hydrates the associated order line items via GetOrderItems.
func (r *Repository) GetOrderByID(
	ctx context.Context,
	qe database.QueryExecutor,
	id uuid.UUID,
) (*model.Order, error) {
	query := `
		SELECT id, customer_id, customer_email, shipping_address,
		       status, coupon_id, discount_amount, total_amount, currency, created_at, updated_at
		FROM orders
		WHERE id = $1
	`
	row := qe.QueryRow(ctx, query, id)

	var o model.Order
	var shippingBytes []byte

	err := row.Scan(
		&o.ID,
		&o.CustomerID,
		&o.CustomerEmail,
		&shippingBytes,
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

	items, err := r.GetOrderItems(ctx, qe, o.ID)
	if err != nil {
		return nil, fmt.Errorf("get order items: %w", err)
	}
	o.Items = items

	return &o, nil
}

// GetOrderByIDForUpdate retrieves an order record by its primary key UUID and acquires
// an exclusive row-level lock (FOR UPDATE) within an active transaction.
func (r *Repository) GetOrderByIDForUpdate(
	ctx context.Context,
	qe database.QueryExecutor,
	id uuid.UUID,
) (*model.Order, error) {
	query := `
		SELECT id, customer_id, customer_email, shipping_address,
		       status, coupon_id, discount_amount, total_amount, currency, created_at, updated_at
		FROM orders
		WHERE id = $1
		FOR UPDATE
	`
	row := qe.QueryRow(ctx, query, id)

	var o model.Order
	var shippingBytes []byte

	err := row.Scan(
		&o.ID,
		&o.CustomerID,
		&o.CustomerEmail,
		&shippingBytes,
		&o.Status,
		&o.CouponID,
		&o.DiscountAmount,
		&o.TotalAmount,
		&o.Currency,
		&o.CreatedAt,
		&o.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get order by id for update: %w", err)
	}

	if err := json.Unmarshal(shippingBytes, &o.ShippingAddress); err != nil {
		return nil, fmt.Errorf("unmarshal shipping address: %w", err)
	}

	items, err := r.GetOrderItems(ctx, qe, o.ID)
	if err != nil {
		return nil, fmt.Errorf("get order items: %w", err)
	}
	o.Items = items

	return &o, nil
}

// GetOrdersByCustomerID fetches all orders belonging to a given customer ID, sorted descending by created_at.
// It performs N+1 queries by calling GetOrderItems for each returned order.
func (r *Repository) GetOrdersByCustomerID(
	ctx context.Context,
	qe database.QueryExecutor,
	customerID uuid.UUID,
) ([]model.Order, error) {
	query := `
		SELECT id, customer_id, customer_email, shipping_address,
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
		var shippingBytes []byte

		if err := rows.Scan(
			&o.ID,
			&o.CustomerID,
			&o.CustomerEmail,
			&shippingBytes,
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

		orders = append(orders, o)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate orders: %w", err)
	}

	if len(orders) > 0 {
		orderIDs := make([]uuid.UUID, len(orders))
		for i := range orders {
			orderIDs[i] = orders[i].ID
		}

		itemsMap, err := r.GetOrderItemsByOrderIDs(ctx, qe, orderIDs)
		if err != nil {
			return nil, fmt.Errorf("batch get order items: %w", err)
		}

		for i := range orders {
			orders[i].Items = itemsMap[orders[i].ID]
		}
	}

	return orders, nil
}

// GetOrderItems fetches all line item rows associated with a given order ID.
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

// GetOrderItemsByOrderIDs fetches line items for multiple order IDs in a single query.
func (r *Repository) GetOrderItemsByOrderIDs(
	ctx context.Context,
	qe database.QueryExecutor,
	orderIDs []uuid.UUID,
) (map[uuid.UUID][]model.OrderItem, error) {
	result := make(map[uuid.UUID][]model.OrderItem, len(orderIDs))
	if len(orderIDs) == 0 {
		return result, nil
	}

	query := `
		SELECT id, order_id, variant_id, quantity, price_at_purchase, product_snapshot
		FROM order_items
		WHERE order_id = ANY($1)
	`
	rows, err := qe.Query(ctx, query, orderIDs)
	if err != nil {
		return nil, fmt.Errorf("batch get order items: %w", err)
	}
	defer rows.Close()

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
		result[item.OrderID] = append(result[item.OrderID], item)
	}

	return result, rows.Err()
}

// UpdateOrderStatus updates the status enum and updated_at timestamp of an order.
// Returns pgx.ErrNoRows if the target order ID was not found.
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

// OrderListFilter specifies query parameters for filtering and paginating administrative order lists.
type OrderListFilter struct {
	// CustomerID optionally restricts listing to a specific customer UUID.
	CustomerID *uuid.UUID
	// Status optionally restricts listing to a specific order status string.
	Status *string
	// Limit controls page size.
	Limit int
	// Offset controls page offset.
	Offset int
}

// ListOrders fetches a filtered, paginated slice of orders from the database.
func (r *Repository) ListOrders(
	ctx context.Context,
	qe database.QueryExecutor,
	filter OrderListFilter,
) ([]model.Order, error) {
	query := `
		SELECT id, customer_id, customer_email, shipping_address,
		       status, coupon_id, discount_amount, total_amount, currency, created_at, updated_at
		FROM orders
		WHERE 1=1
	`
	args := []any{}
	i := 1

	if filter.CustomerID != nil {
		query += fmt.Sprintf(" AND customer_id = $%d", i)
		args = append(args, *filter.CustomerID)
		i++
	}

	if filter.Status != nil && *filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d::order_status", i)
		args = append(args, *filter.Status)
		i++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", i, i+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := qe.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var o model.Order
		var shippingBytes []byte

		if err := rows.Scan(
			&o.ID,
			&o.CustomerID,
			&o.CustomerEmail,
			&shippingBytes,
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
			return nil, fmt.Errorf("unmarshal shipping address for order %s: %w", o.ID, err)
		}

		orders = append(orders, o)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate orders: %w", err)
	}

	if len(orders) > 0 {
		orderIDs := make([]uuid.UUID, len(orders))
		for i := range orders {
			orderIDs[i] = orders[i].ID
		}

		itemsMap, err := r.GetOrderItemsByOrderIDs(ctx, qe, orderIDs)
		if err != nil {
			return nil, fmt.Errorf("batch get order items: %w", err)
		}

		for i := range orders {
			orders[i].Items = itemsMap[orders[i].ID]
		}
	}

	return orders, nil
}

// CountOrders returns the total number of orders matching the specified filter criteria.
func (r *Repository) CountOrders(
	ctx context.Context,
	qe database.QueryExecutor,
	filter OrderListFilter,
) (int, error) {
	query := `SELECT COUNT(*) FROM orders WHERE 1=1`
	args := []any{}
	i := 1

	if filter.CustomerID != nil {
		query += fmt.Sprintf(" AND customer_id = $%d", i)
		args = append(args, *filter.CustomerID)
		i++
	}

	if filter.Status != nil && *filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d::order_status", i)
		args = append(args, *filter.Status)
		i++
	}

	var count int
	if err := qe.QueryRow(ctx, query, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("count orders: %w", err)
	}

	return count, nil
}
