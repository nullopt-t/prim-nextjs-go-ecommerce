package order_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/order"
	"github.com/nullopt-t/prim-backend/internal/setting"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/nullopt-t/prim-backend/pkg/log"
)

type lifecycleRow struct{ order model.Order }

func (r lifecycleRow) Scan(dest ...any) error {
	if len(dest) < 17 {
		return fmt.Errorf("lifecycleRow: expected 17 destinations, got %d", len(dest))
	}

	*dest[0].(*uuid.UUID) = r.order.ID
	if p, ok := dest[1].(**uuid.UUID); ok {
		*p = r.order.CustomerID
	}
	*dest[2].(*string) = r.order.CustomerEmail

	shipping := []byte("{}")
	if r.order.ShippingAddress != nil {
		var err error
		shipping, err = json.Marshal(r.order.ShippingAddress)
		if err != nil {
			return err
		}
	}
	*dest[3].(*[]byte) = shipping
	*dest[4].(*model.OrderStatus) = r.order.Status
	if p, ok := dest[5].(**uuid.UUID); ok {
		*p = r.order.CouponID
	}
	*dest[6].(*int64) = r.order.DiscountAmount
	*dest[7].(*int64) = r.order.ShippingAmount
	*dest[8].(*int64) = r.order.TaxAmount
	*dest[9].(*int64) = r.order.CODFee
	*dest[10].(*int64) = r.order.TotalAmount
	*dest[11].(*string) = r.order.Currency

	paymentMethod := r.order.PaymentMethod
	if p, ok := dest[12].(**string); ok {
		*p = &paymentMethod
	}
	trackingNumber := r.order.TrackingNumber
	if p, ok := dest[13].(**string); ok {
		*p = &trackingNumber
	}
	locale := r.order.Locale
	if p, ok := dest[14].(**string); ok {
		*p = &locale
	}
	if p, ok := dest[15].(**string); ok {
		*p = r.order.Carrier
	}
	if p, ok := dest[16].(**string); ok {
		*p = r.order.TrackingURL
	}
	*dest[17].(*time.Time) = r.order.CreatedAt
	*dest[18].(*time.Time) = r.order.UpdatedAt
	if p, ok := dest[19].(**string); ok {
		*p = nil
	}
	return nil
}

type emptyRows struct{}

func (emptyRows) Close()                                       {}
func (emptyRows) Err() error                                   { return nil }
func (emptyRows) CommandTag() pgconn.CommandTag                { return pgconn.CommandTag{} }
func (emptyRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (emptyRows) Next() bool                                   { return false }
func (emptyRows) Scan(...any) error                            { return nil }
func (emptyRows) Values() ([]any, error)                       { return nil, nil }
func (emptyRows) RawValues() [][]byte                          { return nil }
func (emptyRows) Conn() *pgx.Conn                              { return nil }

type lifecycleExecutor struct {
	order   model.Order
	execErr error
}

func (e lifecycleExecutor) ExecContext(context.Context, string, ...interface{}) (pgconn.CommandTag, error) {
	if e.execErr != nil {
		return pgconn.CommandTag{}, e.execErr
	}
	return pgconn.NewCommandTag("UPDATE 1"), nil
}

func (e lifecycleExecutor) QueryContext(context.Context, string, ...interface{}) (pgx.Rows, error) {
	return emptyRows{}, nil
}

func (e lifecycleExecutor) QueryRowContext(context.Context, string, ...interface{}) pgx.Row {
	return lifecycleRow{order: e.order}
}

func (e lifecycleExecutor) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return e.ExecContext(ctx, sql, args...)
}
func (e lifecycleExecutor) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return e.QueryContext(ctx, sql, args...)
}
func (e lifecycleExecutor) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return e.QueryRowContext(ctx, sql, args...)
}

func (e lifecycleExecutor) Begin(context.Context) (pgx.Tx, error) { return e, nil }
func (e lifecycleExecutor) Commit(context.Context) error          { return nil }
func (e lifecycleExecutor) Rollback(context.Context) error        { return nil }
func (e lifecycleExecutor) CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error) {
	return 0, nil
}
func (e lifecycleExecutor) SendBatch(context.Context, *pgx.Batch) pgx.BatchResults { return nil }
func (e lifecycleExecutor) Prepare(context.Context, string, string) (*pgconn.StatementDescription, error) {
	return nil, nil
}
func (e lifecycleExecutor) Conn() *pgx.Conn                { return nil }
func (e lifecycleExecutor) LargeObjects() pgx.LargeObjects { return pgx.LargeObjects{} }

type lifecycleRunner struct {
	order         model.Order
	execErr       error
	expiredOrders []uuid.UUID
}

func (r lifecycleRunner) WithTx(ctx context.Context, fn func(context.Context) error) error {
	tx := database.NewTx(lifecycleExecutor{order: r.order, execErr: r.execErr})
	ctx = database.WithTx(ctx, tx)
	return fn(ctx)
}

func (r lifecycleRunner) WithExecutor(ctx context.Context, fn func(context.Context) error) error {
	exec := lifecycleExecutor{order: r.order, execErr: r.execErr}
	return fn(database.WithExecutor(ctx, exec))
}

func lifecycleService(r lifecycleRunner, settings *fakeSettingsProvider, notifier *fakeOrderNotifier, userReader ...*fakeUserReader) *order.OrderService {
	if settings == nil {
		settings = &fakeSettingsProvider{}
	}
	if notifier == nil {
		notifier = &fakeOrderNotifier{}
	}
	var ur order.UserReader = &fakeUserReader{}
	if len(userReader) > 0 && userReader[0] != nil {
		ur = userReader[0]
	}
	return order.NewService(r, order.NewRepository(), log.NewNoopLogger(), &fakeInventoryReader{orderIDs: r.expiredOrders}, &fakeInventoryWriter{}, ur, settings, notifier)
}

func TestUpdateOrderStatus_ShippedPublishesNotification(t *testing.T) {
	orderID := uuid.New()
	customerID := uuid.New()
	carrier := "DHL"
	trackingURL := "https://dhl.com/TRACK-1"
	notifier := &fakeOrderNotifier{}
	userReader := &fakeUserReader{
		users: map[uuid.UUID]*model.User{
			customerID: {
				ID:       customerID,
				FullName: "Jane Doe",
			},
		},
	}
	svc := lifecycleService(lifecycleRunner{order: model.Order{
		ID:             orderID,
		CustomerID:     &customerID,
		CustomerEmail:  "buyer@example.com",
		Status:         model.OrderStatusProcessing,
		TrackingNumber: "TRACK-1",
		Carrier:        &carrier,
		TrackingURL:    &trackingURL,
		Locale:         "ar",
	}}, nil, notifier, userReader)

	require.NoError(t, svc.UpdateOrderStatus(context.Background(), orderID, model.OrderStatusShipped, nil))
	require.Len(t, notifier.shipped, 1)
	assert.Equal(t, orderID, notifier.shipped[0].OrderID)
	assert.Equal(t, "TRACK-1", notifier.shipped[0].TrackingNumber)
	assert.Equal(t, "Jane Doe", notifier.shipped[0].CustomerName)
	assert.Equal(t, "DHL", notifier.shipped[0].Carrier)
	assert.Equal(t, "https://dhl.com/TRACK-1", notifier.shipped[0].TrackingURL)
	assert.Equal(t, "ar", notifier.shipped[0].Locale)
}

func TestUpdateOrderStatus_ShippedHonorsNotificationSetting(t *testing.T) {
	cfg := setting.DefaultConfig()
	cfg.NotifyOrderShipped = false
	notifier := &fakeOrderNotifier{}
	orderID := uuid.New()
	svc := lifecycleService(lifecycleRunner{order: model.Order{
		ID:            orderID,
		CustomerEmail: "buyer@example.com",
		Status:        model.OrderStatusProcessing,
	}}, &fakeSettingsProvider{config: &cfg}, notifier)

	require.NoError(t, svc.UpdateOrderStatus(context.Background(), orderID, model.OrderStatusShipped, nil))
	assert.Empty(t, notifier.shipped)
}

func TestUpdateOrderStatus_CancelledPublishesReasonAndName(t *testing.T) {
	orderID := uuid.New()
	customerID := uuid.New()
	notifier := &fakeOrderNotifier{}
	userReader := &fakeUserReader{
		users: map[uuid.UUID]*model.User{
			customerID: {
				ID:       customerID,
				FullName: "Jane Doe",
			},
		},
	}
	svc := lifecycleService(lifecycleRunner{order: model.Order{
		ID:            orderID,
		CustomerID:    &customerID,
		CustomerEmail: "buyer@example.com",
		Status:        model.OrderStatusPending,
		Locale:        "ar",
	}}, nil, notifier, userReader)

	reason := "Out of stock"
	require.NoError(t, svc.UpdateOrderStatus(context.Background(), orderID, model.OrderStatusCancelled, &reason))
	require.Len(t, notifier.cancelled, 1)
	assert.Equal(t, reason, notifier.cancelled[0].Reason)
	assert.Equal(t, "Jane Doe", notifier.cancelled[0].CustomerName)
	assert.Equal(t, "ar", notifier.cancelled[0].Locale)
}

func TestUpdateOrderStatus_RollbackDoesNotNotify(t *testing.T) {
	notifier := &fakeOrderNotifier{}
	orderID := uuid.New()
	svc := lifecycleService(lifecycleRunner{
		order: model.Order{
			ID:            orderID,
			CustomerEmail: "buyer@example.com",
			Status:        model.OrderStatusProcessing,
		},
		execErr: errors.New("update failed"),
	}, nil, notifier)

	require.Error(t, svc.UpdateOrderStatus(context.Background(), orderID, model.OrderStatusShipped, nil))
	assert.Empty(t, notifier.shipped)
}

func TestUpdateOrderTracking_DoesNotDuplicateShippedNotification(t *testing.T) {
	orderID := uuid.New()
	notifier := &fakeOrderNotifier{}
	svc := lifecycleService(lifecycleRunner{order: model.Order{
		ID:             orderID,
		CustomerEmail:  "buyer@example.com",
		Status:         model.OrderStatusShipped,
		TrackingNumber: "OLD",
	}}, nil, notifier)

	carrier := "Aramex"
	url := "https://track.aramex.com/NEW"
	require.NoError(t, svc.UpdateOrderTracking(context.Background(), orderID, order.UpdateOrderTrackingInput{
		TrackingNumber: "NEW",
		Carrier:        &carrier,
		TrackingURL:    &url,
	}))
	// Tracking updates persist metadata without sending duplicate "Order Shipped" emails.
	assert.Empty(t, notifier.shipped)
}

func TestCancelOrder_PublishesAfterCommit(t *testing.T) {
	orderID := uuid.New()
	customerID := uuid.New()
	notifier := &fakeOrderNotifier{}
	userReader := &fakeUserReader{
		users: map[uuid.UUID]*model.User{
			customerID: {
				ID:       customerID,
				FullName: "Jane Doe",
			},
		},
	}
	svc := lifecycleService(lifecycleRunner{order: model.Order{
		ID:            orderID,
		CustomerID:    &customerID,
		CustomerEmail: "buyer@example.com",
		Status:        model.OrderStatusPending,
	}}, nil, notifier, userReader)

	require.NoError(t, svc.CancelOrder(context.Background(), orderID, customerID))
	require.Len(t, notifier.cancelled, 1)
	assert.Equal(t, orderID, notifier.cancelled[0].OrderID)
	assert.Equal(t, "Cancelled by customer", notifier.cancelled[0].Reason)
	assert.Equal(t, "Jane Doe", notifier.cancelled[0].CustomerName)
}

func TestReapExpiredReservations_PublishesCancellationNotification(t *testing.T) {
	orderID := uuid.New()
	customerID := uuid.New()
	notifier := &fakeOrderNotifier{}
	userReader := &fakeUserReader{
		users: map[uuid.UUID]*model.User{
			customerID: {
				ID:       customerID,
				FullName: "Jane Doe",
			},
		},
	}
	svc := lifecycleService(lifecycleRunner{
		order: model.Order{
			ID:            orderID,
			CustomerID:    &customerID,
			CustomerEmail: "buyer@example.com",
			Status:        model.OrderStatusPending,
			Locale:        "ar",
		},
		expiredOrders: []uuid.UUID{orderID},
	}, nil, notifier, userReader)

	cancelled, err := svc.ReapExpiredReservations(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, cancelled)
	require.Len(t, notifier.cancelled, 1)
	assert.Equal(t, orderID, notifier.cancelled[0].OrderID)
	assert.Equal(t, "Jane Doe", notifier.cancelled[0].CustomerName)
	assert.Equal(t, "ar", notifier.cancelled[0].Locale)
	assert.Contains(t, notifier.cancelled[0].Reason, "reservation expired")
}
