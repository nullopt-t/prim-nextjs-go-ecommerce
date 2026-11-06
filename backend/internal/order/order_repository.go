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
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

// Repository implements persistence queries and mutations for orders, order items, and status history.
type Repository struct{}

// NewRepository constructs a new Order Repository instance.
func NewRepository() *Repository {
	return &Repository{}
}

// --- Order CRUD ---

// CreateOrder inserts a new order record into the orders table within the active transaction.
// It serializes the structured ShippingAddress value to JSONB and scans the server-assigned timestamps.
func (r *Repository) CreateOrder(
	ctx context.Context,
	o *model.Order,
) error {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}
	shippingJSON, err := json.Marshal(o.ShippingAddress)
	if err != nil {
		return fmt.Errorf("marshal shipping address: %w", err)
	}

	locale := o.Locale
	if locale != "ar" && locale != "en" {
		locale = "en"
	}
	o.Locale = locale

	query := `
		INSERT INTO orders (
			id, customer_id, customer_email, shipping_address,
			status, coupon_id, discount_amount, shipping_amount, tax_amount, cod_fee, total_amount, currency,
			payment_method, tracking_number, locale, carrier, tracking_url, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, now(), now())
		RETURNING created_at, updated_at
	`
	err = tx.QueryRow(
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
		o.CODFee,
		o.TotalAmount,
		o.Currency,
		o.PaymentMethod,
		o.TrackingNumber,
		o.Locale,
		o.Carrier,
		o.TrackingURL,
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
	item *model.OrderItem,
) error {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}
	query := `
		INSERT INTO order_items (
			id, order_id, variant_id, quantity, price_at_purchase, product_snapshot,
			vendor_id, vendor_earning, commission_rate_bps, commission_amount
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err = tx.Exec(
		ctx,
		query,
		item.ID,
		item.OrderID,
		item.VariantID,
		item.Quantity,
		item.PriceAtPurchase,
		item.ProductSnapshot,
		item.VendorID,
		item.VendorEarning,
		item.CommissionRateBps,
		item.CommissionAmount,
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
	orderID uuid.UUID,
	status model.OrderStatus,
	notes *string,
) error {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}
	query := `
		INSERT INTO order_status_history (order_id, status, notes, created_at)
		VALUES ($1, $2, $3, now())
	`
	_, err = tx.Exec(ctx, query, orderID, status, notes)
	if err != nil {
		return fmt.Errorf("add status history: %w", err)
	}
	return nil
}

// GetOrderByID retrieves an order record by its primary key UUID.
// It also queries and hydrates the associated order line items via GetOrderItems.
func (r *Repository) GetOrderByID(
	ctx context.Context,
	id uuid.UUID,
) (*model.Order, error) {
	qe, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}
	query := `
		SELECT o.id, o.customer_id, o.customer_email, o.shipping_address,
		       o.status, o.coupon_id, o.discount_amount, o.shipping_amount, o.tax_amount, o.cod_fee,
		       o.total_amount, o.currency, o.payment_method, o.tracking_number,
		       o.locale, o.carrier, o.tracking_url,
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
	var paymentMethod, trackingNumber, locale, carrier, trackingURL *string

	err = row.Scan(
		&o.ID,
		&o.CustomerID,
		&o.CustomerEmail,
		&shippingBytes,
		&o.Status,
		&o.CouponID,
		&o.DiscountAmount,
		&o.ShippingAmount,
		&o.TaxAmount,
		&o.CODFee,
		&o.TotalAmount,
		&o.Currency,
		&paymentMethod,
		&trackingNumber,
		&locale,
		&carrier,
		&trackingURL,
		&o.CreatedAt,
		&o.UpdatedAt,
		&couponCode,
	)
	o.CouponCode = couponCode
	o.PaymentMethod = derefString(paymentMethod)
	o.TrackingNumber = derefString(trackingNumber)
	o.Locale = derefString(locale)
	if o.Locale == "" {
		o.Locale = "en"
	}
	o.Carrier = carrier
	o.TrackingURL = trackingURL
	if err != nil {
		return nil, fmt.Errorf("get order by id: %w", err)
	}

	if err := json.Unmarshal(shippingBytes, &o.ShippingAddress); err != nil {
		return nil, fmt.Errorf("unmarshal shipping address: %w", err)
	}

	items, err := r.GetOrderItems(ctx, o.ID)
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
	id uuid.UUID,
) (*model.Order, error) {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return nil, err
	}
	query := `
		SELECT o.id, o.customer_id, o.customer_email, o.shipping_address,
		       o.status, o.coupon_id, o.discount_amount, o.shipping_amount, o.tax_amount, o.cod_fee,
		       o.total_amount, o.currency, o.payment_method, o.tracking_number,
		       o.locale, o.carrier, o.tracking_url,
		       o.created_at, o.updated_at,
		       c.code
		FROM orders o
		LEFT JOIN coupons c ON o.coupon_id = c.id
		WHERE o.id = $1
		FOR UPDATE OF o
	`
	row := tx.QueryRow(ctx, query, id)

	var o model.Order
	var shippingBytes []byte
	var couponCode *string
	var paymentMethod, trackingNumber, locale, carrier, trackingURL *string

	err = row.Scan(
		&o.ID,
		&o.CustomerID,
		&o.CustomerEmail,
		&shippingBytes,
		&o.Status,
		&o.CouponID,
		&o.DiscountAmount,
		&o.ShippingAmount,
		&o.TaxAmount,
		&o.CODFee,
		&o.TotalAmount,
		&o.Currency,
		&paymentMethod,
		&trackingNumber,
		&locale,
		&carrier,
		&trackingURL,
		&o.CreatedAt,
		&o.UpdatedAt,
		&couponCode,
	)
	o.CouponCode = couponCode
	o.PaymentMethod = derefString(paymentMethod)
	o.TrackingNumber = derefString(trackingNumber)
	o.Locale = derefString(locale)
	if o.Locale == "" {
		o.Locale = "en"
	}
	o.Carrier = carrier
	o.TrackingURL = trackingURL
	if err != nil {
		return nil, fmt.Errorf("get order by id for update: %w", err)
	}

	if err := json.Unmarshal(shippingBytes, &o.ShippingAddress); err != nil {
		return nil, fmt.Errorf("unmarshal shipping address: %w", err)
	}

	items, err := r.GetOrderItems(ctx, o.ID)
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
	customerID uuid.UUID,
) ([]model.Order, error) {
	qe, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}
	query := `
		SELECT o.id, o.customer_id, o.customer_email, o.shipping_address,
		       o.status, o.coupon_id, o.discount_amount, o.shipping_amount, o.tax_amount, o.cod_fee,
		       o.total_amount, o.currency, o.payment_method, o.tracking_number,
		       o.locale, o.carrier, o.tracking_url,
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
		var paymentMethod, trackingNumber, locale, carrier, trackingURL *string

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
			&o.CODFee,
			&o.TotalAmount,
			&o.Currency,
			&paymentMethod,
			&trackingNumber,
			&locale,
			&carrier,
			&trackingURL,
			&o.CreatedAt,
			&o.UpdatedAt,
			&couponCode,
		); err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		o.CouponCode = couponCode
		o.PaymentMethod = derefString(paymentMethod)
		o.TrackingNumber = derefString(trackingNumber)
		o.Locale = derefString(locale)
		if o.Locale == "" {
			o.Locale = "en"
		}
		o.Carrier = carrier
		o.TrackingURL = trackingURL

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

		itemsMap, err := r.GetOrderItemsByOrderIDs(ctx, orderIDs)
		if err != nil {
			return nil, fmt.Errorf("batch get order items: %w", err)
		}

		for i := range orders {
			orders[i].Items = itemsMap[orders[i].ID]
		}
	}

	return orders, nil
}

// ListVendorOrders returns orders containing one vendor's lines, newest
// first. Only rows whose order_items carry the vendor's id are ever
// selected; the projection is what keeps vendors isolated from each
// other without leaking the admin /admin/orders surface.
func (r *Repository) ListVendorOrders(
	ctx context.Context,
	vendorID uuid.UUID,
	status *model.OrderStatus,
	q *pagination.ListQuery,
) ([]model.Order, int64, error) {
	qe, err := database.Executor(ctx)
	if err != nil {
		return nil, 0, err
	}

	where := "WHERE oi.vendor_id = $1"
	args := []any{vendorID}
	if status != nil && *status != "" {
		where += " AND o.status = $2::order_status"
		args = append(args, *status)
	}

	var total int64
	countQuery := `
		SELECT COUNT(DISTINCT o.id)
		FROM orders o
		JOIN order_items oi ON oi.order_id = o.id
		` + where
	if err := qe.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count vendor orders: %w", err)
	}

	limitIdx := len(args) + 1
	offsetIdx := len(args) + 2
	rows, err := qe.Query(ctx, `
		SELECT DISTINCT o.id, o.customer_id, o.customer_email, o.shipping_address,
		       o.status, o.coupon_id, o.discount_amount, o.shipping_amount, o.tax_amount, o.cod_fee,
		       o.total_amount, o.currency, o.payment_method, o.tracking_number,
		       o.locale, o.carrier, o.tracking_url,
		       o.created_at, o.updated_at,
		       c.code
		FROM orders o
		JOIN order_items oi ON oi.order_id = o.id
		LEFT JOIN coupons c ON o.coupon_id = c.id
		`+where+`
		ORDER BY o.created_at DESC, o.id DESC
		LIMIT $`+fmt.Sprint(limitIdx)+` OFFSET $`+fmt.Sprint(offsetIdx),
		append(args, q.PageSize, q.Offset)...,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list vendor orders: %w", err)
	}

	orders, err := scanOrders(rows)
	if err != nil {
		return nil, 0, err
	}

	// Vendor sees only their own lines: filter each order's items to those
	// carrying the vendor id, and never expose another vendor's lines that
	// share the same order.
	orderIDs := make([]uuid.UUID, len(orders))
	for i := range orders {
		orderIDs[i] = orders[i].ID
	}
	itemsMap, err := r.GetOrderItemsByOrderIDs(ctx, orderIDs)
	if err != nil {
		return nil, 0, fmt.Errorf("batch get vendor order items: %w", err)
	}
	for i := range orders {
		for _, item := range itemsMap[orders[i].ID] {
			if item.VendorID != nil && *item.VendorID == vendorID {
				orders[i].Items = append(orders[i].Items, item)
			}
		}
	}

	return orders, total, nil
}

// GetVendorOrder loads a single order when it contains a line owned by the
// vendor; otherwise it behaves as "not found" so existence of foreign or
// platform orders is never leaked to a vendor.
func (r *Repository) GetVendorOrder(ctx context.Context, vendorID, orderID uuid.UUID) (*model.Order, error) {
	qe, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}

	var owned bool
	err = qe.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM order_items
			WHERE order_id = $1 AND vendor_id = $2
		)
	`, orderID, vendorID).Scan(&owned)
	if err != nil {
		return nil, fmt.Errorf("check vendor order ownership: %w", err)
	}
	if !owned {
		return nil, pgx.ErrNoRows
	}

	order, err := r.GetOrderByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	items := make([]model.OrderItem, 0, len(order.Items))
	for _, item := range order.Items {
		if item.VendorID != nil && *item.VendorID == vendorID {
			items = append(items, item)
		}
	}
	order.Items = items
	return order, nil
}

// GetOrderItems fetches all line item rows associated with a given order ID.
func (r *Repository) GetOrderItems(
	ctx context.Context,
	orderID uuid.UUID,
) ([]model.OrderItem, error) {
	qe, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}
	query := `
		SELECT id, order_id, variant_id, quantity, price_at_purchase, product_snapshot,
			vendor_id, vendor_earning, commission_rate_bps, commission_amount
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
			&item.VendorID,
			&item.VendorEarning,
			&item.CommissionRateBps,
			&item.CommissionAmount,
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
	orderIDs []uuid.UUID,
) (map[uuid.UUID][]model.OrderItem, error) {
	result := make(map[uuid.UUID][]model.OrderItem, len(orderIDs))
	if len(orderIDs) == 0 {
		return result, nil
	}

	qe, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}
	query := `
		SELECT id, order_id, variant_id, quantity, price_at_purchase, product_snapshot,
			vendor_id, vendor_earning, commission_rate_bps, commission_amount
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
			&item.VendorID,
			&item.VendorEarning,
			&item.CommissionRateBps,
			&item.CommissionAmount,
		); err != nil {
			return nil, fmt.Errorf("scan order item: %w", err)
		}
		result[item.OrderID] = append(result[item.OrderID], item)
	}

	return result, rows.Err()
}

// GetVendorSalesStats projects one vendor's sales from order data. Earnings are
// the frozen per-line values stamped at checkout, so later commission changes
// never rewrite history. Cancelled and refunded orders are excluded from
// revenue and order counts.
func (r *Repository) GetVendorSalesStats(
	ctx context.Context,
	vendorID uuid.UUID,
) (*model.VendorSalesStats, error) {
	qe, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}

	stats := new(model.VendorSalesStats)
	err = qe.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(oi.price_at_purchase * oi.quantity)
				FILTER (WHERE o.status IN ('paid', 'processing', 'shipped', 'delivered')), 0),
			COALESCE(SUM(oi.vendor_earning)
				FILTER (WHERE o.status = 'delivered'), 0),
			COALESCE(SUM(oi.vendor_earning)
				FILTER (WHERE o.status IN ('paid', 'processing', 'shipped')), 0),
			COUNT(DISTINCT o.id)
				FILTER (WHERE o.status NOT IN ('pending', 'cancelled', 'refunded'))
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		WHERE oi.vendor_id = $1
	`, vendorID).Scan(
		&stats.GrossRevenue,
		&stats.SettledEarnings,
		&stats.PendingEarnings,
		&stats.OrderCount,
	)
	if err != nil {
		return nil, fmt.Errorf("get vendor sales stats: %w", err)
	}
	return stats, nil
}

// GetVendorSettledEarnings sums vendor earnings and derived sales counters for
// orders that reached a settled lifecycle state (delivered). Cancelled and
// refunded orders drop out automatically. The method is a read-only
// projection owned by the order domain and consumed by vendor via an
// interface, so the vendor package never issues SQL against order tables.
func (r *Repository) GetVendorSettledEarnings(
	ctx context.Context,
	vendorID uuid.UUID,
) (int64, error) {
	qe, err := database.Executor(ctx)
	if err != nil {
		return 0, err
	}

	var total int64
	err = qe.QueryRow(ctx, `
		SELECT COALESCE(SUM(oi.vendor_earning), 0)
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		WHERE oi.vendor_id = $1 AND o.status = 'delivered'
	`, vendorID).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("sum settled vendor earnings: %w", err)
	}
	return total, nil
}

// UpdateOrderStatus updates the status enum and updated_at timestamp of an order.
// Returns pgx.ErrNoRows if the target order ID was not found.
func (r *Repository) UpdateOrderStatus(
	ctx context.Context,
	orderID uuid.UUID,
	status model.OrderStatus,
) error {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}
	query := `
		UPDATE orders
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`
	tag, err := tx.Exec(ctx, query, status, orderID)
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
	// DateFrom optionally restricts listing to orders created at/after this time.
	DateFrom *time.Time
	// DateTo optionally restricts listing to orders created at/before this time.
	DateTo *time.Time
	// Limit controls page size.
	Limit int
	// Offset controls page offset.
	Offset int
	// Cursor, when set, switches the query to keyset pagination on (created_at, id) DESC.
	Cursor *pagination.Keyset
}

// ListOrders fetches a filtered, paginated slice of orders from the database.
func (r *Repository) ListOrders(
	ctx context.Context,
	filter OrderListFilter,
) ([]model.Order, error) {
	qe, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}
	query := `
		SELECT o.id, o.customer_id, o.customer_email, o.shipping_address,
		       o.status, o.coupon_id, o.discount_amount, o.shipping_amount, o.tax_amount, o.cod_fee,
		       o.total_amount, o.currency, o.payment_method, o.tracking_number,
		       o.locale, o.carrier, o.tracking_url,
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

	if filter.DateFrom != nil {
		query += fmt.Sprintf(" AND o.created_at >= $%d", i)
		args = append(args, *filter.DateFrom)
		i++
	}

	if filter.DateTo != nil {
		query += fmt.Sprintf(" AND o.created_at <= $%d", i)
		args = append(args, *filter.DateTo)
		i++
	}

	// Keyset pagination: restrict to orders strictly older than the cursor and
	// over-fetch one row so the caller can detect whether more remain.
	keysetClause, keysetArgs := pagination.KeysetClause("o", filter.Cursor, i)
	query += keysetClause
	args = append(args, keysetArgs...)
	i += len(keysetArgs)

	query += pagination.KeysetOrderBy("o")
	query += fmt.Sprintf(" LIMIT $%d", i)
	args = append(args, filter.Limit+1)

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
		var paymentMethod, trackingNumber, locale, carrier, trackingURL *string

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
			&o.CODFee,
			&o.TotalAmount,
			&o.Currency,
			&paymentMethod,
			&trackingNumber,
			&locale,
			&carrier,
			&trackingURL,
			&o.CreatedAt,
			&o.UpdatedAt,
			&couponCode,
		); err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		o.CouponCode = couponCode
		o.PaymentMethod = derefString(paymentMethod)
		o.TrackingNumber = derefString(trackingNumber)
		o.Locale = derefString(locale)
		if o.Locale == "" {
			o.Locale = "en"
		}
		o.Carrier = carrier
		o.TrackingURL = trackingURL

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

		itemsMap, err := r.GetOrderItemsByOrderIDs(ctx, orderIDs)
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
	filter OrderListFilter,
) (int, error) {
	qe, err := database.Executor(ctx)
	if err != nil {
		return 0, err
	}
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

func (r *Repository) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	qe, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}
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
	err = qe.QueryRow(ctx, query).Scan(
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

	// Product count and daily sales feed the admin dashboard tiles/charts.
	// Propagating their errors (rather than silently returning zeroes) keeps
	// the dashboard honest: a transient DB failure surfaces as an error the
	// UI can retry instead of an all-zero "healthy" dashboard.
	if err := qe.QueryRow(ctx, `SELECT COUNT(*) FROM products WHERE deleted_at IS NULL`).Scan(&stats.TotalProducts); err != nil {
		return nil, fmt.Errorf("get total products: %w", err)
	}

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
	if err != nil {
		return nil, fmt.Errorf("get recent sales: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var ds DailySales
		if err := rows.Scan(&ds.Date, &ds.Sales, &ds.Orders); err != nil {
			return nil, fmt.Errorf("scan recent sales: %w", err)
		}
		stats.RecentSales = append(stats.RecentSales, ds)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate recent sales: %w", err)
	}

	return stats, nil
}

// PaymentListFilter narrows a payments listing.
type PaymentListFilter struct {
	// Status optionally restricts listing to a specific payment status.
	Status *string
	// DateFrom optionally restricts listing to payments created at/after this time.
	DateFrom *time.Time
	// DateTo optionally restricts listing to payments created at/before this time.
	DateTo *time.Time
	// Limit controls page size.
	Limit int
	// Cursor, when set, switches the query to keyset pagination on (created_at, id) DESC.
	Cursor *pagination.Keyset
}

func (r *Repository) ListPayments(ctx context.Context, filter PaymentListFilter) ([]PaymentItem, *pagination.CursorPage, error) {
	qe, err := database.Executor(ctx)
	if err != nil {
		return nil, nil, err
	}
	query := `
		SELECT p.id, p.order_id, p.amount, p.currency, p.status, p.provider, p.provider_transaction_id,
		       COALESCE(o.customer_email, 'customer@example.com'), p.created_at, p.updated_at
		FROM payments p
		LEFT JOIN orders o ON p.order_id = o.id
		WHERE 1=1
	`
	args := []any{}
	i := 1

	if filter.Status != nil && *filter.Status != "" {
		query += fmt.Sprintf(" AND p.status = $%d::payment_status", i)
		args = append(args, *filter.Status)
		i++
	}

	if filter.DateFrom != nil {
		query += fmt.Sprintf(" AND p.created_at >= $%d", i)
		args = append(args, *filter.DateFrom)
		i++
	}

	if filter.DateTo != nil {
		query += fmt.Sprintf(" AND p.created_at <= $%d", i)
		args = append(args, *filter.DateTo)
		i++
	}

	keysetClause, keysetArgs := pagination.KeysetClause("p", filter.Cursor, i)
	query += keysetClause
	args = append(args, keysetArgs...)
	i += len(keysetArgs)

	query += pagination.KeysetOrderBy("p")
	query += fmt.Sprintf(" LIMIT $%d", i)
	args = append(args, filter.Limit+1)

	rows, err := qe.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("list payments: %w", err)
	}
	defer rows.Close()

	var payments []PaymentItem
	for rows.Next() {
		var pi PaymentItem
		if err := rows.Scan(
			&pi.ID, &pi.OrderID, &pi.Amount, &pi.Currency, &pi.Status, &pi.Provider, &pi.ProviderTransactionID,
			&pi.CustomerEmail, &pi.CreatedAt, &pi.UpdatedAt,
		); err != nil {
			return nil, nil, fmt.Errorf("scan payment: %w", err)
		}
		payments = append(payments, pi)
	}

	payments, page := pagination.NewCursorPageFrom(payments, filter.Limit, func(p PaymentItem) pagination.Keyset {
		return pagination.Keyset{CreatedAt: p.CreatedAt, ID: p.ID}
	})
	return payments, page, nil
}

func (r *Repository) UpdatePaymentStatus(ctx context.Context, id uuid.UUID, status string) error {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}
	query := `UPDATE payments SET status = $1::payment_status, updated_at = NOW() WHERE id = $2`
	tag, err := tx.Exec(ctx, query, status, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// GetPaymentStatusForUpdate locks a payment row and returns its current status.
// Callers use it to validate state transitions (e.g. only captured payments may
// be refunded) while serializing concurrent updates on the same payment.
func (r *Repository) GetPaymentStatusForUpdate(ctx context.Context, id uuid.UUID) (string, error) {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return "", err
	}
	var status string
	err = tx.QueryRow(ctx, `SELECT status FROM payments WHERE id = $1 FOR UPDATE`, id).Scan(&status)
	if err != nil {
		return "", fmt.Errorf("get payment status for update: %w", err)
	}
	return status, nil
}

// GetPaymentOrderIDForUpdate locks a payment row and returns its owning order
// id, so a status transition (e.g. capture) can act on the order's inventory
// reservations within the same transaction.
func (r *Repository) GetPaymentOrderIDForUpdate(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	var orderID uuid.UUID
	err = tx.QueryRow(ctx, `SELECT order_id FROM payments WHERE id = $1 FOR UPDATE`, id).Scan(&orderID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("get payment order id for update: %w", err)
	}
	return orderID, nil
}

func (r *Repository) UpdateOrderTracking(
	ctx context.Context,
	orderID uuid.UUID,
	trackingNumber string,
	carrier *string,
	trackingURL *string,
) error {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}
	query := `UPDATE orders SET tracking_number = $1, carrier = $2, tracking_url = $3, updated_at = NOW() WHERE id = $4`
	tag, err := tx.Exec(ctx, query, trackingNumber, carrier, trackingURL, orderID)
	if err != nil {
		return fmt.Errorf("update order tracking: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// CreatePayment records a payment transaction for an order on the active transaction.
func (r *Repository) CreatePayment(ctx context.Context, p *model.Payment) error {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}
	query := `
		INSERT INTO payments (id, order_id, amount, currency, status, provider, provider_transaction_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
	`
	if _, err := tx.Exec(
		ctx,
		query,
		p.ID,
		p.OrderID,
		p.Amount,
		p.Currency,
		p.Status,
		p.Provider,
		p.ProviderTransactionID,
	); err != nil {
		return fmt.Errorf("create payment: %w", err)
	}
	return nil
}

// RefundOrFailPaymentsForOrder closes every still-open payment of a cancelled
// order: captured funds become refunded, anything else becomes failed.
func (r *Repository) RefundOrFailPaymentsForOrder(ctx context.Context, orderID uuid.UUID) error {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}
	query := `
		UPDATE payments
		SET status = CASE
			WHEN status = 'captured'::payment_status THEN 'refunded'::payment_status
			ELSE 'failed'::payment_status
		END,
		updated_at = now()
		WHERE order_id = $1 AND status NOT IN ('refunded'::payment_status, 'failed'::payment_status)
	`
	if _, err := tx.Exec(ctx, query, orderID); err != nil {
		return fmt.Errorf("refund or fail payments for order: %w", err)
	}
	return nil
}

// FailPendingPaymentsForOrder marks never-captured payments as failed. It is
// used when an unpaid order's stock hold lapses, so stale payments no longer
// look receivable.
func (r *Repository) FailPendingPaymentsForOrder(ctx context.Context, orderID uuid.UUID) error {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}
	query := `
		UPDATE payments
		SET status = 'failed'::payment_status, updated_at = now()
		WHERE order_id = $1 AND status IN ('pending'::payment_status, 'authorized'::payment_status)
	`
	if _, err := tx.Exec(ctx, query, orderID); err != nil {
		return fmt.Errorf("fail pending payments for order: %w", err)
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

// scanOrders consumes order rows from an already-open *pgx.Rows and hydrates
// shipping/coupon fields, mirroring the per-method scan loops above.
func scanOrders(rows pgx.Rows) ([]model.Order, error) {
	defer rows.Close()

	orders := make([]model.Order, 0)
	for rows.Next() {
		var o model.Order
		var shippingBytes []byte
		var couponCode *string
		var paymentMethod, trackingNumber, locale, carrier, trackingURL *string

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
			&o.CODFee,
			&o.TotalAmount,
			&o.Currency,
			&paymentMethod,
			&trackingNumber,
			&locale,
			&carrier,
			&trackingURL,
			&o.CreatedAt,
			&o.UpdatedAt,
			&couponCode,
		); err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		o.CouponCode = couponCode
		o.PaymentMethod = derefString(paymentMethod)
		o.TrackingNumber = derefString(trackingNumber)
		o.Locale = derefString(locale)
		if o.Locale == "" {
			o.Locale = "en"
		}
		o.Carrier = carrier
		o.TrackingURL = trackingURL

		if err := json.Unmarshal(shippingBytes, &o.ShippingAddress); err != nil {
			return nil, fmt.Errorf("unmarshal shipping address: %w", err)
		}

		orders = append(orders, o)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate orders: %w", err)
	}
	return orders, nil
}
