// Package order provides database persistence layer operations for the order domain.
package order

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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
			status, coupon_id, discount_amount, shipping_amount, tax_amount, total_amount, currency,
			payment_method, tracking_number, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, now(), now())
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
		o.ShippingAmount,
		o.TaxAmount,
		o.TotalAmount,
		o.Currency,
		o.PaymentMethod,
		o.TrackingNumber,
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
		SELECT o.id, o.customer_id, o.customer_email, o.shipping_address,
		       o.status, o.coupon_id, o.discount_amount, o.shipping_amount, o.tax_amount,
		       o.total_amount, o.currency, o.payment_method, o.tracking_number,
		       o.created_at, o.updated_at,
		       c.code
		FROM orders o
		LEFT JOIN coupons c ON o.coupon_id = c.id
		WHERE o.id = $1
	`
	row := qe.QueryRow(ctx, query, id)

	var o model.Order
	var shippingBytes []byte
	var couponCode *string
	var paymentMethod, trackingNumber *string

	err := row.Scan(
		&o.ID,
		&o.CustomerID,
		&o.CustomerEmail,
		&shippingBytes,
		&o.Status,
		&o.CouponID,
		&o.DiscountAmount,
		&o.ShippingAmount,
		&o.TaxAmount,
		&o.TotalAmount,
		&o.Currency,
		&paymentMethod,
		&trackingNumber,
		&o.CreatedAt,
		&o.UpdatedAt,
		&couponCode,
	)
	o.CouponCode = couponCode
	o.PaymentMethod = derefString(paymentMethod)
	o.TrackingNumber = derefString(trackingNumber)
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
		SELECT o.id, o.customer_id, o.customer_email, o.shipping_address,
		       o.status, o.coupon_id, o.discount_amount, o.shipping_amount, o.tax_amount,
		       o.total_amount, o.currency, o.payment_method, o.tracking_number,
		       o.created_at, o.updated_at,
		       c.code
		FROM orders o
		LEFT JOIN coupons c ON o.coupon_id = c.id
		WHERE o.id = $1
		FOR UPDATE OF o
	`
	row := qe.QueryRow(ctx, query, id)

	var o model.Order
	var shippingBytes []byte
	var couponCode *string
	var paymentMethod, trackingNumber *string

	err := row.Scan(
		&o.ID,
		&o.CustomerID,
		&o.CustomerEmail,
		&shippingBytes,
		&o.Status,
		&o.CouponID,
		&o.DiscountAmount,
		&o.ShippingAmount,
		&o.TaxAmount,
		&o.TotalAmount,
		&o.Currency,
		&paymentMethod,
		&trackingNumber,
		&o.CreatedAt,
		&o.UpdatedAt,
		&couponCode,
	)
	o.CouponCode = couponCode
	o.PaymentMethod = derefString(paymentMethod)
	o.TrackingNumber = derefString(trackingNumber)
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
		SELECT o.id, o.customer_id, o.customer_email, o.shipping_address,
		       o.status, o.coupon_id, o.discount_amount, o.shipping_amount, o.tax_amount,
		       o.total_amount, o.currency, o.payment_method, o.tracking_number,
		       o.created_at, o.updated_at,
		       c.code
		FROM orders o
		LEFT JOIN coupons c ON o.coupon_id = c.id
		WHERE o.customer_id = $1
		ORDER BY o.created_at DESC
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
		var couponCode *string
		var paymentMethod, trackingNumber *string

		if err := rows.Scan(
			&o.ID,
			&o.CustomerID,
			&o.CustomerEmail,
			&shippingBytes,
			&o.Status,
			&o.CouponID,
			&o.DiscountAmount,
			&o.ShippingAmount,
			&o.TaxAmount,
			&o.TotalAmount,
			&o.Currency,
			&paymentMethod,
			&trackingNumber,
			&o.CreatedAt,
			&o.UpdatedAt,
			&couponCode,
		); err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		o.CouponCode = couponCode
		o.PaymentMethod = derefString(paymentMethod)
		o.TrackingNumber = derefString(trackingNumber)

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
		SELECT o.id, o.customer_id, o.customer_email, o.shipping_address,
		       o.status, o.coupon_id, o.discount_amount, o.shipping_amount, o.tax_amount,
		       o.total_amount, o.currency, o.payment_method, o.tracking_number,
		       o.created_at, o.updated_at,
		       c.code
		FROM orders o
		LEFT JOIN coupons c ON o.coupon_id = c.id
		WHERE 1=1
	`
	args := []any{}
	i := 1

	if filter.CustomerID != nil {
		query += fmt.Sprintf(" AND o.customer_id = $%d", i)
		args = append(args, *filter.CustomerID)
		i++
	}

	if filter.Status != nil && *filter.Status != "" {
		query += fmt.Sprintf(" AND o.status = $%d::order_status", i)
		args = append(args, *filter.Status)
		i++
	}

	query += fmt.Sprintf(" ORDER BY o.created_at DESC LIMIT $%d OFFSET $%d", i, i+1)
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
		var couponCode *string
		var paymentMethod, trackingNumber *string

		if err := rows.Scan(
			&o.ID,
			&o.CustomerID,
			&o.CustomerEmail,
			&shippingBytes,
			&o.Status,
			&o.CouponID,
			&o.DiscountAmount,
			&o.ShippingAmount,
			&o.TaxAmount,
			&o.TotalAmount,
			&o.Currency,
			&paymentMethod,
			&trackingNumber,
			&o.CreatedAt,
			&o.UpdatedAt,
			&couponCode,
		); err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		o.CouponCode = couponCode
		o.PaymentMethod = derefString(paymentMethod)
		o.TrackingNumber = derefString(trackingNumber)

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

// --- Dashboard & Payments Aggregates ---

type DashboardStats struct {
	TotalRevenue      int64        `json:"totalRevenue"`
	TotalOrders       int          `json:"totalOrders"`
	TotalCustomers    int          `json:"totalCustomers"`
	TotalProducts     int          `json:"totalProducts"`
	AverageOrderValue float64      `json:"averageOrderValue"`
	PendingOrders     int          `json:"pendingOrders"`
	ProcessingOrders  int          `json:"processingOrders"`
	ShippedOrders     int          `json:"shippedOrders"`
	DeliveredOrders   int          `json:"deliveredOrders"`
	CanceledOrders    int          `json:"canceledOrders"`
	RecentSales       []DailySales `json:"recentSales"`
}

type DailySales struct {
	Date   string `json:"date"`
	Sales  int64  `json:"sales"`
	Orders int    `json:"orders"`
}

type PaymentItem struct {
	ID                    uuid.UUID `json:"id"`
	OrderID               uuid.UUID `json:"orderId"`
	Amount                int64     `json:"amount"`
	Currency              string    `json:"currency"`
	Status                string    `json:"status"`
	Provider              string    `json:"provider"`
	ProviderTransactionID string    `json:"providerTransactionId"`
	CustomerEmail         string    `json:"customerEmail"`
	CreatedAt             time.Time `json:"createdAt"`
	UpdatedAt             time.Time `json:"updatedAt"`
}

func (r *Repository) GetDashboardStats(ctx context.Context, qe database.QueryExecutor) (*DashboardStats, error) {
	stats := &DashboardStats{}

	query := `
		SELECT 
			COALESCE(SUM(total_amount) FILTER (WHERE status != 'cancelled'), 0),
			COUNT(*),
			COUNT(DISTINCT customer_email),
			COALESCE(AVG(total_amount) FILTER (WHERE status != 'cancelled'), 0),
			COUNT(*) FILTER (WHERE status = 'pending'),
			COUNT(*) FILTER (WHERE status = 'processing'),
			COUNT(*) FILTER (WHERE status = 'shipped'),
			COUNT(*) FILTER (WHERE status = 'delivered'),
			COUNT(*) FILTER (WHERE status = 'cancelled')
		FROM orders
	`
	err := qe.QueryRow(ctx, query).Scan(
		&stats.TotalRevenue,
		&stats.TotalOrders,
		&stats.TotalCustomers,
		&stats.AverageOrderValue,
		&stats.PendingOrders,
		&stats.ProcessingOrders,
		&stats.ShippedOrders,
		&stats.DeliveredOrders,
		&stats.CanceledOrders,
	)
	if err != nil {
		return nil, fmt.Errorf("get order aggregates: %w", err)
	}

	_ = qe.QueryRow(ctx, `SELECT COUNT(*) FROM products WHERE deleted_at IS NULL`).Scan(&stats.TotalProducts)

	salesQuery := `
		SELECT 
			TO_CHAR(created_at, 'YYYY-MM-DD') as day,
			COALESCE(SUM(total_amount) FILTER (WHERE status != 'cancelled'), 0) as daily_sales,
			COUNT(*) FILTER (WHERE status != 'cancelled') as daily_orders
		FROM orders
		GROUP BY TO_CHAR(created_at, 'YYYY-MM-DD')
		ORDER BY day DESC
		LIMIT 14
	`
	rows, err := qe.Query(ctx, salesQuery)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var ds DailySales
			if err := rows.Scan(&ds.Date, &ds.Sales, &ds.Orders); err == nil {
				stats.RecentSales = append(stats.RecentSales, ds)
			}
		}
	}

	return stats, nil
}

func (r *Repository) ListPayments(ctx context.Context, qe database.QueryExecutor, limit, offset int) ([]PaymentItem, int, error) {
	var total int
	_ = qe.QueryRow(ctx, `SELECT COUNT(*) FROM payments`).Scan(&total)

	query := `
		SELECT p.id, p.order_id, p.amount, p.currency, p.status, p.provider, p.provider_transaction_id,
		       COALESCE(o.customer_email, 'customer@example.com'), p.created_at, p.updated_at
		FROM payments p
		LEFT JOIN orders o ON p.order_id = o.id
		ORDER BY p.created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := qe.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list payments: %w", err)
	}
	defer rows.Close()

	var payments []PaymentItem
	for rows.Next() {
		var pi PaymentItem
		if err := rows.Scan(
			&pi.ID, &pi.OrderID, &pi.Amount, &pi.Currency, &pi.Status, &pi.Provider, &pi.ProviderTransactionID,
			&pi.CustomerEmail, &pi.CreatedAt, &pi.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan payment: %w", err)
		}
		payments = append(payments, pi)
	}
	return payments, total, nil
}

func (r *Repository) UpdatePaymentStatus(ctx context.Context, qe database.QueryExecutor, id uuid.UUID, status string) error {
	query := `UPDATE payments SET status = $1::payment_status, updated_at = NOW() WHERE id = $2`
	_, err := qe.Exec(ctx, query, status, id)
	return err
}

func (r *Repository) UpdateOrderTracking(ctx context.Context, qe database.QueryExecutor, orderID uuid.UUID, trackingNumber string) error {
	query := `UPDATE orders SET tracking_number = $1, updated_at = NOW() WHERE id = $2`
	tag, err := qe.Exec(ctx, query, trackingNumber, orderID)
	if err != nil {
		return fmt.Errorf("update order tracking: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// derefString returns the string value of a *string, or "" if nil.
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

