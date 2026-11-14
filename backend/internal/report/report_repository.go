package report

import (
	"context"
	"fmt"
	"time"

	"github.com/nullopt-t/prim-backend/pkg/database"
)

// ExportType identifies a report export dataset.
type ExportType string

const (
	// ExportOrders exports the order ledger.
	ExportOrders ExportType = "orders"
	// ExportPayments exports payment transactions.
	ExportPayments ExportType = "payments"
	// ExportCustomers exports registered accounts.
	ExportCustomers ExportType = "customers"
	// ExportInventory exports per-variant stock levels.
	ExportInventory ExportType = "inventory"
)

// ParseExportType validates and normalizes an export type string.
func ParseExportType(s string) (ExportType, bool) {
	switch ExportType(s) {
	case ExportOrders, ExportPayments, ExportCustomers, ExportInventory:
		return ExportType(s), true
	default:
		return "", false
	}
}

// maxExportRows caps every export so a report can never exhaust memory.
const maxExportRows = 10000

// ReportFilter narrows a report or export by creation time and, for orders and
// payments, by status.
type ReportFilter struct {
	From   *time.Time
	To     *time.Time
	Status string
}

// SalesTotals are the headline aggregates for a period.
type SalesTotals struct {
	Revenue           int64   `json:"revenue"`
	Orders            int     `json:"orders"`
	AverageOrderValue float64 `json:"averageOrderValue"`
	CancelledOrders   int     `json:"cancelledOrders"`
}

// SalesPoint is one bucket of the sales time series.
type SalesPoint struct {
	Period  string `json:"period"`
	Revenue int64  `json:"revenue"`
	Orders  int    `json:"orders"`
}

// SalesReport is the date-ranged sales summary.
type SalesReport struct {
	From    *time.Time   `json:"from"`
	To      *time.Time   `json:"to"`
	GroupBy string       `json:"groupBy"`
	Totals  SalesTotals  `json:"totals"`
	Series  []SalesPoint `json:"series"`
}

// Repository reads reporting aggregates and export datasets.
type Repository struct{}

// NewRepository constructs a report repository.
func NewRepository() *Repository { return &Repository{} }

// Sales computes totals and a day/month time series for the given period.
func (r *Repository) Sales(
	ctx context.Context,
	f ReportFilter,
	groupBy string,
) (*SalesReport, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}

	rep := &SalesReport{From: f.From, To: f.To, GroupBy: groupBy, Series: []SalesPoint{}}

	totalsQuery := `
		SELECT
			COALESCE(SUM(total_amount) FILTER (WHERE status != 'cancelled'), 0),
			COUNT(*) FILTER (WHERE status != 'cancelled'),
			COALESCE(AVG(total_amount) FILTER (WHERE status != 'cancelled'), 0),
			COUNT(*) FILTER (WHERE status = 'cancelled')
		FROM orders
		WHERE ($1::timestamptz IS NULL OR created_at >= $1)
		  AND ($2::timestamptz IS NULL OR created_at <= $2)
	`
	if err := executor.QueryRowContext(ctx, totalsQuery, f.From, f.To).Scan(
		&rep.Totals.Revenue,
		&rep.Totals.Orders,
		&rep.Totals.AverageOrderValue,
		&rep.Totals.CancelledOrders,
	); err != nil {
		return nil, fmt.Errorf("sales totals: %w", err)
	}

	seriesQuery := `
		SELECT to_char(date_trunc($3, created_at), 'YYYY-MM-DD'),
		       COALESCE(SUM(total_amount) FILTER (WHERE status != 'cancelled'), 0),
		       COUNT(*) FILTER (WHERE status != 'cancelled')
		FROM orders
		WHERE ($1::timestamptz IS NULL OR created_at >= $1)
		  AND ($2::timestamptz IS NULL OR created_at <= $2)
		GROUP BY 1
		ORDER BY 1
	`
	rows, err := executor.QueryContext(ctx, seriesQuery, f.From, f.To, groupBy)
	if err != nil {
		return nil, fmt.Errorf("sales series: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var p SalesPoint
		if err := rows.Scan(&p.Period, &p.Revenue, &p.Orders); err != nil {
			return nil, fmt.Errorf("scan sales point: %w", err)
		}
		rep.Series = append(rep.Series, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sales series: %w", err)
	}

	return rep, nil
}

// ExportDataset returns the CSV header and rows for the requested dataset.
func (r *Repository) ExportDataset(
	ctx context.Context,
	t ExportType,
	f ReportFilter,
) ([]string, [][]string, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, nil, err
	}

	switch t {
	case ExportOrders:
		return r.exportOrders(ctx, executor, f)
	case ExportPayments:
		return r.exportPayments(ctx, executor, f)
	case ExportCustomers:
		return r.exportCustomers(ctx, executor, f)
	case ExportInventory:
		return r.exportInventory(ctx, executor, f)
	default:
		return nil, nil, fmt.Errorf("unknown export type %q", t)
	}
}

func (r *Repository) exportOrders(ctx context.Context, qe database.QueryExecutor, f ReportFilter) ([]string, [][]string, error) {
	header := []string{"created_at", "order_id", "customer_email", "status", "total_amount", "currency", "payment_method", "tracking_number"}
	query := `
		SELECT o.created_at, o.id, o.customer_email, o.status, o.total_amount, o.currency,
		       COALESCE(o.payment_method, ''), COALESCE(o.tracking_number, '')
		FROM orders o
		WHERE ($1::timestamptz IS NULL OR o.created_at >= $1)
		  AND ($2::timestamptz IS NULL OR o.created_at <= $2)
		  AND ($3::text IS NULL OR o.status = $3::order_status)
		ORDER BY o.created_at DESC
		LIMIT $4
	`
	var status *string
	if f.Status != "" {
		status = &f.Status
	}
	rows, err := qe.QueryContext(ctx, query, f.From, f.To, status, maxExportRows)
	if err != nil {
		return nil, nil, fmt.Errorf("export orders: %w", err)
	}
	defer rows.Close()

	out := [][]string{}
	for rows.Next() {
		var (
			createdAt                   time.Time
			id                          string
			email, st, currency, pm, tn string
			total                       int64
		)
		if err := rows.Scan(&createdAt, &id, &email, &st, &total, &currency, &pm, &tn); err != nil {
			return nil, nil, fmt.Errorf("scan order export: %w", err)
		}
		out = append(out, []string{
			createdAt.Format(time.RFC3339), id, email, st, money(total), currency, pm, tn,
		})
	}
	return header, out, rows.Err()
}

func (r *Repository) exportPayments(ctx context.Context, qe database.QueryExecutor, f ReportFilter) ([]string, [][]string, error) {
	header := []string{"created_at", "payment_id", "order_id", "customer_email", "amount", "currency", "status", "provider", "provider_transaction_id"}
	query := `
		SELECT p.created_at, p.id, p.order_id, COALESCE(o.customer_email, ''), p.amount, p.currency,
		       p.status, p.provider, COALESCE(p.provider_transaction_id, '')
		FROM payments p
		LEFT JOIN orders o ON p.order_id = o.id
		WHERE ($1::timestamptz IS NULL OR p.created_at >= $1)
		  AND ($2::timestamptz IS NULL OR p.created_at <= $2)
		  AND ($3::text IS NULL OR p.status = $3::payment_status)
		ORDER BY p.created_at DESC
		LIMIT $4
	`
	var status *string
	if f.Status != "" {
		status = &f.Status
	}
	rows, err := qe.QueryContext(ctx, query, f.From, f.To, status, maxExportRows)
	if err != nil {
		return nil, nil, fmt.Errorf("export payments: %w", err)
	}
	defer rows.Close()

	out := [][]string{}
	for rows.Next() {
		var (
			createdAt                    time.Time
			id, orderID, email           string
			amount                       int64
			currency, st, provider, txid string
		)
		if err := rows.Scan(&createdAt, &id, &orderID, &email, &amount, &currency, &st, &provider, &txid); err != nil {
			return nil, nil, fmt.Errorf("scan payment export: %w", err)
		}
		out = append(out, []string{
			createdAt.Format(time.RFC3339), id, orderID, email, money(amount), currency, st, provider, txid,
		})
	}
	return header, out, rows.Err()
}

func (r *Repository) exportCustomers(ctx context.Context, qe database.QueryExecutor, f ReportFilter) ([]string, [][]string, error) {
	header := []string{"created_at", "user_id", "phone", "full_name", "role", "status"}
	query := `
		SELECT u.created_at, u.id, u.phone, COALESCE(u.full_name, ''), u.role, u.status
		FROM users u
		WHERE u.deleted_at IS NULL
		  AND ($1::timestamptz IS NULL OR u.created_at >= $1)
		  AND ($2::timestamptz IS NULL OR u.created_at <= $2)
		ORDER BY u.created_at DESC
		LIMIT $3
	`
	rows, err := qe.QueryContext(ctx, query, f.From, f.To, maxExportRows)
	if err != nil {
		return nil, nil, fmt.Errorf("export customers: %w", err)
	}
	defer rows.Close()

	out := [][]string{}
	for rows.Next() {
		var (
			createdAt       time.Time
			id, phone, name string
			role, status    string
		)
		if err := rows.Scan(&createdAt, &id, &phone, &name, &role, &status); err != nil {
			return nil, nil, fmt.Errorf("scan customer export: %w", err)
		}
		out = append(out, []string{createdAt.Format(time.RFC3339), id, phone, name, role, status})
	}
	return header, out, rows.Err()
}

func (r *Repository) exportInventory(ctx context.Context, qe database.QueryExecutor, f ReportFilter) ([]string, [][]string, error) {
	header := []string{"variant_id", "product_title", "sku", "on_hand", "reserved", "available"}
	query := `
		SELECT pv.id, p.title, pv.sku,
		       COALESCE((SELECT SUM(quantity) FROM inventory_ledgers l WHERE l.variant_id = pv.id), 0),
		       COALESCE((SELECT SUM(quantity) FROM inventory_reservations rr
		                 WHERE rr.variant_id = pv.id AND rr.released_at IS NULL AND rr.expires_at > now()), 0)
		FROM product_variants pv
		JOIN products p ON pv.product_id = p.id
		WHERE pv.deleted_at IS NULL AND p.deleted_at IS NULL
		ORDER BY p.title, pv.sku
		LIMIT $1
	`
	rows, err := qe.QueryContext(ctx, query, maxExportRows)
	if err != nil {
		return nil, nil, fmt.Errorf("export inventory: %w", err)
	}
	defer rows.Close()

	out := [][]string{}
	for rows.Next() {
		var (
			id, title, sku   string
			onHand, reserved int
		)
		if err := rows.Scan(&id, &title, &sku, &onHand, &reserved); err != nil {
			return nil, nil, fmt.Errorf("scan inventory export: %w", err)
		}
		out = append(out, []string{id, title, sku, fmt.Sprint(onHand), fmt.Sprint(reserved), fmt.Sprint(onHand - reserved)})
	}
	return header, out, rows.Err()
}

// money renders minor units as a decimal string with two fraction digits.
func money(minor int64) string {
	return fmt.Sprintf("%.2f", float64(minor)/100.0)
}
