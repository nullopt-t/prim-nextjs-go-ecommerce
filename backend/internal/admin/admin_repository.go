package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

type DashboardStats struct {
	TotalRevenue      int64   `json:"totalRevenue"`
	TotalOrders       int     `json:"totalOrders"`
	TotalCustomers    int     `json:"totalCustomers"`
	TotalProducts     int     `json:"totalProducts"`
	AverageOrderValue float64 `json:"averageOrderValue"`
	PendingOrders     int     `json:"pendingOrders"`
	ProcessingOrders  int     `json:"processingOrders"`
	ShippedOrders     int     `json:"shippedOrders"`
	DeliveredOrders   int     `json:"deliveredOrders"`
	CanceledOrders    int     `json:"canceledOrders"`
	RecentSales       []DailySales `json:"recentSales"`
}

type DailySales struct {
	Date   string `json:"date"`
	Sales  int64  `json:"sales"`
	Orders int    `json:"orders"`
}

type PromotionItem struct {
	ID             uuid.UUID  `json:"id"`
	Code           string     `json:"code"`
	DiscountType   string     `json:"discountType"`
	DiscountValue  int64      `json:"discountValue"`
	Currency       *string    `json:"currency"`
	MinCartAmount  *int64     `json:"minCartAmount"`
	UsageLimit     *int       `json:"usageLimit"`
	PerUserLimit   *int       `json:"perUserLimit"`
	StartsAt       *time.Time `json:"startsAt"`
	ExpiresAt      *time.Time `json:"expiresAt"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
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

type MediaItem struct {
	ID          uuid.UUID `json:"id"`
	Bucket      string    `json:"bucket"`
	ObjectKey   string    `json:"objectKey"`
	ContentType string    `json:"contentType"`
	FileSize    int64     `json:"fileSize"`
	Status      string    `json:"status"`
	URL         string    `json:"url"`
	CreatedAt   time.Time `json:"createdAt"`
}

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

type InventoryStockItem struct {
	VariantID         uuid.UUID `json:"variantId"`
	ProductID         uuid.UUID `json:"productId"`
	ProductName       string    `json:"productName"`
	SKU               string    `json:"sku"`
	VariantTitle      string    `json:"variantTitle"`
	OnHandQuantity    int       `json:"onHandQuantity"`
	ReservedQuantity  int       `json:"reservedQuantity"`
	AvailableQuantity int       `json:"availableQuantity"`
	IsInStock         bool      `json:"isInStock"`
	Price             int64     `json:"price"`
	Currency          string    `json:"currency"`
	Thumbnail         string    `json:"thumbnail"`
}

type InventoryLedgerItem struct {
	ID           uuid.UUID `json:"id"`
	VariantID    uuid.UUID `json:"variantId"`
	ProductName  string    `json:"productName"`
	VariantTitle string    `json:"variantTitle"`
	SKU          string    `json:"sku"`
	Quantity     int       `json:"quantity"`
	Reason       string    `json:"reason"`
	CreatedAt    time.Time `json:"createdAt"`
}

type AdminRepository struct{}

func NewRepository() *AdminRepository {
	return &AdminRepository{}
}

func (r *AdminRepository) GetDashboardStats(ctx context.Context, qe database.QueryExecutor) (*DashboardStats, error) {
	stats := &DashboardStats{}

	// Aggregates from orders
	query := `
		SELECT 
			COALESCE(SUM(total_amount), 0),
			COUNT(*),
			COUNT(DISTINCT customer_email),
			COALESCE(AVG(total_amount), 0),
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

	// Product count
	_ = qe.QueryRow(ctx, `SELECT COUNT(*) FROM products WHERE deleted_at IS NULL`).Scan(&stats.TotalProducts)

	// Daily sales last 14 days
	salesQuery := `
		SELECT 
			TO_CHAR(created_at, 'YYYY-MM-DD') as day,
			COALESCE(SUM(total_amount), 0) as daily_sales,
			COUNT(*) as daily_orders
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

func (r *AdminRepository) ListCoupons(ctx context.Context, qe database.QueryExecutor) ([]PromotionItem, error) {
	query := `
		SELECT id, code, discount_type, discount_value, currency, min_cart_amount,
		       usage_limit, per_user_limit, starts_at, expires_at, status, created_at, updated_at
		FROM coupons
		ORDER BY created_at DESC
	`
	rows, err := qe.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list coupons: %w", err)
	}
	defer rows.Close()

	var list []PromotionItem
	for rows.Next() {
		var p PromotionItem
		if err := rows.Scan(
			&p.ID, &p.Code, &p.DiscountType, &p.DiscountValue, &p.Currency, &p.MinCartAmount,
			&p.UsageLimit, &p.PerUserLimit, &p.StartsAt, &p.ExpiresAt, &p.Status, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan coupon: %w", err)
		}
		list = append(list, p)
	}
	return list, nil
}

func (r *AdminRepository) CreateCoupon(ctx context.Context, qe database.QueryExecutor, p *PromotionItem) error {
	query := `
		INSERT INTO coupons (id, code, discount_type, discount_value, currency, min_cart_amount, usage_limit, per_user_limit, starts_at, expires_at, status, created_at, updated_at)
		VALUES ($1, $2, $3::discount_type, $4, $5, $6, $7, $8, $9, $10, $11::coupon_status, NOW(), NOW())
		RETURNING created_at, updated_at
	`
	return qe.QueryRow(ctx, query,
		p.ID, p.Code, p.DiscountType, p.DiscountValue, p.Currency, p.MinCartAmount,
		p.UsageLimit, p.PerUserLimit, p.StartsAt, p.ExpiresAt, p.Status,
	).Scan(&p.CreatedAt, &p.UpdatedAt)
}

func (r *AdminRepository) UpdateCouponStatus(ctx context.Context, qe database.QueryExecutor, id uuid.UUID, status string) error {
	query := `UPDATE coupons SET status = $1::coupon_status, updated_at = NOW() WHERE id = $2`
	_, err := qe.Exec(ctx, query, status, id)
	return err
}

func (r *AdminRepository) DeleteCoupon(ctx context.Context, qe database.QueryExecutor, id uuid.UUID) error {
	query := `DELETE FROM coupons WHERE id = $1`
	_, err := qe.Exec(ctx, query, id)
	return err
}

func (r *AdminRepository) ListPayments(ctx context.Context, qe database.QueryExecutor, limit, offset int) ([]PaymentItem, int, error) {
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

func (r *AdminRepository) UpdatePaymentStatus(ctx context.Context, qe database.QueryExecutor, id uuid.UUID, status string) error {
	query := `UPDATE payments SET status = $1::payment_status, updated_at = NOW() WHERE id = $2`
	_, err := qe.Exec(ctx, query, status, id)
	return err
}

func (r *AdminRepository) ListMedia(ctx context.Context, qe database.QueryExecutor, limit, offset int) ([]MediaItem, int, error) {
	var total int
	_ = qe.QueryRow(ctx, `SELECT COUNT(*) FROM storage_objects WHERE deleted_at IS NULL`).Scan(&total)

	query := `
		SELECT id, bucket, object_key, content_type, file_size, status, created_at
		FROM storage_objects
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := qe.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list media: %w", err)
	}
	defer rows.Close()

	var list []MediaItem
	for rows.Next() {
		var mi MediaItem
		if err := rows.Scan(
			&mi.ID, &mi.Bucket, &mi.ObjectKey, &mi.ContentType, &mi.FileSize, &mi.Status, &mi.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan media: %w", err)
		}
		mi.URL = fmt.Sprintf("http://localhost:9000/%s/%s", mi.Bucket, mi.ObjectKey)
		list = append(list, mi)
	}
	return list, total, nil
}

func (r *AdminRepository) ListAttributes(ctx context.Context, qe database.QueryExecutor) ([]AttributeItem, error) {
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

func (r *AdminRepository) CreateAttribute(ctx context.Context, qe database.QueryExecutor, ai *AttributeItem) error {
	query := `
		INSERT INTO category_attributes (id, category_id, key, label, value_type, allowed_values, is_filterable, sort_order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5::attribute_value_type, $6, $7, $8, NOW(), NOW())
		RETURNING created_at, updated_at
	`
	return qe.QueryRow(ctx, query,
		ai.ID, ai.CategoryID, ai.Key, ai.Label, ai.ValueType, ai.AllowedValues, ai.IsFilterable, ai.SortOrder,
	).Scan(&ai.CreatedAt, &ai.UpdatedAt)
}

func (r *AdminRepository) DeleteAttribute(ctx context.Context, qe database.QueryExecutor, id uuid.UUID) error {
	query := `DELETE FROM category_attributes WHERE id = $1`
	_, err := qe.Exec(ctx, query, id)
	return err
}

func (r *AdminRepository) ListInventoryOverview(ctx context.Context, qe database.QueryExecutor, limit, offset int) ([]InventoryStockItem, int, error) {
	var total int
	_ = qe.QueryRow(ctx, `SELECT COUNT(*) FROM product_variants WHERE deleted_at IS NULL`).Scan(&total)

	query := `
		SELECT 
			pv.id,
			pv.product_id,
			p.title,
			pv.sku,
			pv.title,
			COALESCE((SELECT SUM(quantity) FROM inventory_ledgers WHERE variant_id = pv.id), 0)::int AS on_hand,
			COALESCE((SELECT SUM(quantity) FROM inventory_reservations WHERE variant_id = pv.id AND released_at IS NULL AND expires_at > NOW()), 0)::int AS reserved,
			pv.price,
			pv.currency
		FROM product_variants pv
		JOIN products p ON pv.product_id = p.id
		WHERE pv.deleted_at IS NULL
		ORDER BY on_hand ASC, pv.created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := qe.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list inventory: %w", err)
	}
	defer rows.Close()

	var list []InventoryStockItem
	for rows.Next() {
		var item InventoryStockItem
		if err := rows.Scan(
			&item.VariantID,
			&item.ProductID,
			&item.ProductName,
			&item.SKU,
			&item.VariantTitle,
			&item.OnHandQuantity,
			&item.ReservedQuantity,
			&item.Price,
			&item.Currency,
		); err != nil {
			return nil, 0, fmt.Errorf("scan inventory item: %w", err)
		}
		item.AvailableQuantity = item.OnHandQuantity - item.ReservedQuantity
		if item.AvailableQuantity < 0 {
			item.AvailableQuantity = 0
		}
		item.IsInStock = item.AvailableQuantity > 0
		list = append(list, item)
	}
	return list, total, nil
}

func (r *AdminRepository) ListRecentLedgers(ctx context.Context, qe database.QueryExecutor, limit int, variantID *uuid.UUID, productID *uuid.UUID) ([]InventoryLedgerItem, error) {
	query := `
		SELECT il.id, il.variant_id, p.title, pv.title, pv.sku, il.quantity, il.reason, il.created_at
		FROM inventory_ledgers il
		JOIN product_variants pv ON il.variant_id = pv.id
		JOIN products p ON pv.product_id = p.id
		WHERE 1=1
	`
	args := []any{}
	i := 1

	if variantID != nil {
		query += fmt.Sprintf(" AND il.variant_id = $%d", i)
		args = append(args, *variantID)
		i++
	}

	if productID != nil {
		query += fmt.Sprintf(" AND p.id = $%d", i)
		args = append(args, *productID)
		i++
	}

	query += fmt.Sprintf(" ORDER BY il.created_at DESC LIMIT $%d", i)
	args = append(args, limit)

	rows, err := qe.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list recent ledgers: %w", err)
	}
	defer rows.Close()

	var list []InventoryLedgerItem
	for rows.Next() {
		var item InventoryLedgerItem
		if err := rows.Scan(
			&item.ID, &item.VariantID, &item.ProductName, &item.VariantTitle, &item.SKU, &item.Quantity, &item.Reason, &item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan ledger: %w", err)
		}
		list = append(list, item)
	}
	return list, nil
}
