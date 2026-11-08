package order

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/nullopt-t/prim-backend/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeTxRunner struct {
	exec database.QueryExecutor
}

func (f *fakeTxRunner) WithTx(ctx context.Context, fn func(context.Context) error) error {
	tx := database.NewTx(&fakePgxTx{exec: f.exec})
	return fn(database.WithTx(ctx, tx))
}
func (f *fakeTxRunner) WithExecutor(ctx context.Context, fn func(context.Context) error) error {
	return fn(database.WithExecutor(ctx, f.exec))
}
func (f *fakeTxRunner) WithRetry(ctx context.Context, fn func(context.Context) error, n int) error {
	return f.WithTx(ctx, fn)
}

type fakePgxTx struct {
	exec database.QueryExecutor
}

func (f *fakePgxTx) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if f.exec != nil {
		return f.exec.ExecContext(ctx, sql, args...)
	}
	return pgconn.NewCommandTag("UPDATE 1"), nil
}
func (f *fakePgxTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if f.exec != nil {
		return f.exec.QueryContext(ctx, sql, args...)
	}
	return nil, nil
}
func (f *fakePgxTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if f.exec != nil {
		return f.exec.QueryRowContext(ctx, sql, args...)
	}
	return nil
}
func (f *fakePgxTx) Begin(ctx context.Context) (pgx.Tx, error) { return f, nil }
func (f *fakePgxTx) Commit(ctx context.Context) error          { return nil }
func (f *fakePgxTx) Rollback(ctx context.Context) error        { return nil }
func (f *fakePgxTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return 0, nil
}
func (f *fakePgxTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults { return nil }
func (f *fakePgxTx) Prepare(ctx context.Context, name string, sql string) (*pgconn.StatementDescription, error) {
	return nil, nil
}
func (f *fakePgxTx) Conn() *pgx.Conn                { return nil }
func (f *fakePgxTx) LargeObjects() pgx.LargeObjects { return pgx.LargeObjects{} }

func TestOrderHandler_RequestOrderReturn_AuthAndValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.ErrorHandler(log.NewNoopLogger()))
	handler := &OrderHandler{}
	router.POST("/orders/:id/returns", handler.RequestOrderReturn)

	// Unauthenticated
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/orders/10000000-0000-0000-0000-000000000001/returns", bytes.NewBufferString(`{"reason":"defective"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Authenticated with invalid order ID
	routerAuth := gin.New()
	routerAuth.Use(func(c *gin.Context) {
		session := &model.UserSession{
			ID:     "sess_1",
			UserID: uuid.New(),
			Type:   model.SessionTypeAuthenticated,
		}
		c.Set("session", session)
		c.Next()
	})
	routerAuth.Use(middleware.ErrorHandler(log.NewNoopLogger()))
	routerAuth.POST("/orders/:id/returns", handler.RequestOrderReturn)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/orders/not-a-uuid/returns", bytes.NewBufferString(`{"reason":"defective"}`))
	req.Header.Set("Content-Type", "application/json")
	routerAuth.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Authenticated with empty reason
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/orders/10000000-0000-0000-0000-000000000001/returns", bytes.NewBufferString(`{"reason":""}`))
	req.Header.Set("Content-Type", "application/json")
	routerAuth.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

type fakeOrderReturnExecutor struct {
	order *model.Order
	items []model.OrderItem
	notes []string
}

func (e *fakeOrderReturnExecutor) ExecContext(_ context.Context, _ string, args ...any) (pgconn.CommandTag, error) {
	if len(args) >= 3 {
		if n, ok := args[2].(*string); ok && n != nil {
			e.notes = append(e.notes, *n)
		}
	}
	return pgconn.NewCommandTag("INSERT 0 1"), nil
}

type fakeReturnItemRows struct {
	items []model.OrderItem
	idx   int
}

func (r *fakeReturnItemRows) Close()                                       {}
func (r *fakeReturnItemRows) Err() error                                   { return nil }
func (r *fakeReturnItemRows) CommandTag() pgconn.CommandTag                { return pgconn.CommandTag{} }
func (r *fakeReturnItemRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (r *fakeReturnItemRows) Next() bool {
	if r.idx < len(r.items) {
		r.idx++
		return true
	}
	return false
}
func (r *fakeReturnItemRows) Scan(dest ...any) error {
	it := r.items[r.idx-1]
	*dest[0].(*uuid.UUID) = it.ID
	*dest[1].(*uuid.UUID) = it.OrderID
	*dest[2].(*uuid.UUID) = it.VariantID
	*dest[3].(*int) = it.Quantity
	*dest[4].(*int64) = it.PriceAtPurchase
	*dest[5].(*string) = it.ProductSnapshot
	return nil
}
func (r *fakeReturnItemRows) Values() ([]any, error) { return nil, nil }
func (r *fakeReturnItemRows) RawValues() [][]byte    { return nil }
func (r *fakeReturnItemRows) Conn() *pgx.Conn        { return nil }

func (e *fakeOrderReturnExecutor) QueryContext(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
	return &fakeReturnItemRows{items: e.items}, nil
}

type fakeOrderRow struct {
	order *model.Order
}

func (r *fakeOrderRow) Scan(dest ...any) error {
	if r.order == nil {
		return pgx.ErrNoRows
	}
	now := time.Now()
	dummy := "dummy"

	*dest[0].(*uuid.UUID) = r.order.ID
	*dest[1].(**uuid.UUID) = r.order.CustomerID
	*dest[2].(*string) = r.order.CustomerEmail
	*dest[3].(*[]byte) = []byte(`{}`)
	*dest[4].(*model.OrderStatus) = r.order.Status
	*dest[5].(**uuid.UUID) = r.order.CouponID
	*dest[6].(*int64) = r.order.DiscountAmount
	*dest[7].(*int64) = r.order.ShippingAmount
	*dest[8].(*int64) = r.order.TaxAmount
	*dest[9].(*int64) = r.order.CODFee
	*dest[10].(*int64) = r.order.TotalAmount
	*dest[11].(*string) = r.order.Currency
	*dest[12].(**string) = &dummy // paymentMethod
	*dest[13].(**string) = &dummy // trackingNumber
	*dest[14].(**string) = &dummy // locale
	*dest[15].(**string) = &dummy // carrier
	*dest[16].(**string) = &dummy // trackingURL
	*dest[17].(*time.Time) = now  // created_at
	*dest[18].(*time.Time) = now  // updated_at
	*dest[19].(**string) = nil    // coupon code
	return nil
}

func (e *fakeOrderReturnExecutor) QueryRowContext(_ context.Context, _ string, _ ...any) pgx.Row {
	return &fakeOrderRow{order: e.order}
}

func TestOrderService_RequestOrderReturn(t *testing.T) {
	customerID := uuid.New()
	orderID := uuid.New()
	itemID := uuid.New()

	baseOrder := &model.Order{
		ID:         orderID,
		CustomerID: &customerID,
		Status:     model.OrderStatusDelivered,
	}

	orderItems := []model.OrderItem{
		{
			ID:       itemID,
			OrderID:  orderID,
			Quantity: 2,
		},
	}

	t.Run("successful return request on delivered order", func(t *testing.T) {
		exec := &fakeOrderReturnExecutor{
			order: baseOrder,
			items: orderItems,
		}
		repo := NewRepository()
		svc := &OrderService{
			qexecutor: &fakeTxRunner{exec: exec},
			orderRepo: repo,
		}

		err := svc.RequestOrderReturn(context.Background(), orderID, customerID, "Item wrong color", []ReturnItemInput{
			{ItemID: itemID, Quantity: 1},
		})
		require.NoError(t, err)
		require.Len(t, exec.notes, 1)
		assert.Contains(t, exec.notes[0], "Return requested")
		assert.Contains(t, exec.notes[0], "Item wrong color")
	})

	t.Run("rejected if order not delivered", func(t *testing.T) {
		shippedOrder := &model.Order{
			ID:         orderID,
			CustomerID: &customerID,
			Status:     model.OrderStatusShipped,
		}
		exec := &fakeOrderReturnExecutor{
			order: shippedOrder,
			items: orderItems,
		}
		repo := NewRepository()
		svc := &OrderService{
			qexecutor: &fakeTxRunner{exec: exec},
			orderRepo: repo,
		}

		err := svc.RequestOrderReturn(context.Background(), orderID, customerID, "Want return", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Return requests can only be submitted for delivered orders")
	})

	t.Run("rejected if not order owner", func(t *testing.T) {
		otherUser := uuid.New()
		exec := &fakeOrderReturnExecutor{
			order: baseOrder,
			items: orderItems,
		}
		repo := NewRepository()
		svc := &OrderService{
			qexecutor: &fakeTxRunner{exec: exec},
			orderRepo: repo,
		}

		err := svc.RequestOrderReturn(context.Background(), orderID, otherUser, "Want return", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not authorized")
	})

	t.Run("rejected if item does not belong to order", func(t *testing.T) {
		exec := &fakeOrderReturnExecutor{
			order: baseOrder,
			items: orderItems,
		}
		repo := NewRepository()
		svc := &OrderService{
			qexecutor: &fakeTxRunner{exec: exec},
			orderRepo: repo,
		}

		err := svc.RequestOrderReturn(context.Background(), orderID, customerID, "Want return", []ReturnItemInput{
			{ItemID: uuid.New(), Quantity: 1},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Item does not belong to this order")
	})

	t.Run("rejected if return quantity exceeds purchased quantity", func(t *testing.T) {
		exec := &fakeOrderReturnExecutor{
			order: baseOrder,
			items: orderItems,
		}
		repo := NewRepository()
		svc := &OrderService{
			qexecutor: &fakeTxRunner{exec: exec},
			orderRepo: repo,
		}

		err := svc.RequestOrderReturn(context.Background(), orderID, customerID, "Want return", []ReturnItemInput{
			{ItemID: itemID, Quantity: 5},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid return quantity")
	})
}
