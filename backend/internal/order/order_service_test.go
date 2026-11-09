package order

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/notifier"
	"github.com/nullopt-t/prim-backend/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeOrderFulfillExecutor struct {
	order         *model.Order
	items         []model.OrderItem
	notes         []string
	statusUpdates []model.OrderStatus
}

func (e *fakeOrderFulfillExecutor) ExecContext(_ context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if strings.Contains(sql, "UPDATE order_items") {
		status := args[0].(model.FulfillmentStatus)
		var trackingNumber, carrier *string
		if args[1] != nil {
			trackingNumber = args[1].(*string)
		}
		if args[2] != nil {
			carrier = args[2].(*string)
		}
		targetIDs := args[5].([]uuid.UUID)
		targetMap := make(map[uuid.UUID]bool, len(targetIDs))
		for _, id := range targetIDs {
			targetMap[id] = true
		}

		now := time.Now()
		for i := range e.items {
			if targetMap[e.items[i].ID] {
				e.items[i].FulfillmentStatus = status
				if trackingNumber != nil {
					e.items[i].TrackingNumber = trackingNumber
				}
				if carrier != nil {
					e.items[i].Carrier = carrier
				}
				if e.items[i].ShippedAt == nil {
					e.items[i].ShippedAt = &now
				}
			}
		}
		return pgconn.NewCommandTag(fmt.Sprintf("UPDATE %d", len(targetIDs))), nil
	}

	if strings.Contains(sql, "UPDATE orders") && strings.Contains(sql, "status = $1") {
		newStatus := args[0].(model.OrderStatus)
		if e.order != nil {
			e.order.Status = newStatus
		}
		e.statusUpdates = append(e.statusUpdates, newStatus)
		return pgconn.NewCommandTag("UPDATE 1"), nil
	}

	if strings.Contains(sql, "UPDATE orders") && strings.Contains(sql, "tracking_number = $1") {
		if e.order != nil {
			e.order.TrackingNumber = args[0].(string)
		}
		return pgconn.NewCommandTag("UPDATE 1"), nil
	}

	if strings.Contains(sql, "INSERT INTO order_status_history") {
		if len(args) > 2 && args[2] != nil {
			if n, ok := args[2].(*string); ok && n != nil {
				e.notes = append(e.notes, *n)
			}
		}
		return pgconn.NewCommandTag("INSERT 1"), nil
	}

	return pgconn.NewCommandTag("OK 1"), nil
}

func (e *fakeOrderFulfillExecutor) QueryContext(_ context.Context, sql string, _ ...any) (pgx.Rows, error) {
	if strings.Contains(sql, "FROM order_items") {
		return &fakeFulfillItemRows{items: e.items}, nil
	}
	if strings.Contains(sql, "FROM order_status_history") {
		return &fakeHistoryRows{}, nil
	}
	return nil, nil
}

func (e *fakeOrderFulfillExecutor) QueryRowContext(_ context.Context, sql string, args ...any) pgx.Row {
	if strings.Contains(sql, "SELECT EXISTS") {
		sellerID := args[1].(uuid.UUID)
		owned := false
		for _, it := range e.items {
			if it.SellerID != nil && *it.SellerID == sellerID {
				owned = true
				break
			}
		}
		return &fakeExistsRow{exists: owned}
	}
	return &fakeOrderRow{order: e.order}
}

type fakeExistsRow struct {
	exists bool
}

func (r *fakeExistsRow) Scan(dest ...any) error {
	*dest[0].(*bool) = r.exists
	return nil
}

type fakeHistoryRows struct{}

func (r *fakeHistoryRows) Close()                                       {}
func (r *fakeHistoryRows) Err() error                                   { return nil }
func (r *fakeHistoryRows) CommandTag() pgconn.CommandTag                { return pgconn.CommandTag{} }
func (r *fakeHistoryRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (r *fakeHistoryRows) Next() bool                                   { return false }
func (r *fakeHistoryRows) Scan(dest ...any) error                       { return nil }
func (r *fakeHistoryRows) Values() ([]any, error)                       { return nil, nil }
func (r *fakeHistoryRows) RawValues() [][]byte                          { return nil }
func (r *fakeHistoryRows) Conn() *pgx.Conn                              { return nil }

type fakeFulfillItemRows struct {
	items []model.OrderItem
	idx   int
}

func (r *fakeFulfillItemRows) Close()                                       {}
func (r *fakeFulfillItemRows) Err() error                                   { return nil }
func (r *fakeFulfillItemRows) CommandTag() pgconn.CommandTag                { return pgconn.CommandTag{} }
func (r *fakeFulfillItemRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (r *fakeFulfillItemRows) Next() bool {
	if r.idx < len(r.items) {
		r.idx++
		return true
	}
	return false
}

func (r *fakeFulfillItemRows) Scan(dest ...any) error {
	it := r.items[r.idx-1]
	*dest[0].(*uuid.UUID) = it.ID
	*dest[1].(*uuid.UUID) = it.OrderID
	*dest[2].(*uuid.UUID) = it.VariantID
	*dest[3].(*int) = it.Quantity
	*dest[4].(*int64) = it.PriceAtPurchase
	*dest[5].(*string) = it.ProductSnapshot
	*dest[6].(**uuid.UUID) = it.SellerID
	*dest[7].(*int64) = it.SellerEarning
	*dest[8].(*int) = it.CommissionRateBps
	*dest[9].(*int64) = it.CommissionAmount

	statusStr := string(it.FulfillmentStatus)
	if statusStr == "" {
		statusStr = string(model.FulfillmentStatusPending)
	}
	*dest[10].(*string) = statusStr
	*dest[11].(**string) = it.TrackingNumber
	*dest[12].(**string) = it.Carrier
	*dest[13].(**time.Time) = it.ShippedAt
	return nil
}

func (r *fakeFulfillItemRows) Values() ([]any, error) { return nil, nil }
func (r *fakeFulfillItemRows) RawValues() [][]byte    { return nil }
func (r *fakeFulfillItemRows) Conn() *pgx.Conn        { return nil }

type mockNotifier struct {
	shippedCalls int
	lastPayload  *notifier.OrderShippedPayload
}

func (m *mockNotifier) NotifyOrderShipped(_ context.Context, in notifier.OrderShippedPayload) error {
	m.shippedCalls++
	m.lastPayload = &in
	return nil
}

func (m *mockNotifier) NotifyOrderCancelled(_ context.Context, _ notifier.OrderCancelledPayload) error {
	return nil
}

func TestFulfillSellerOrderItems_SingleSeller_AllItemsShipped(t *testing.T) {
	sellerID := uuid.New()
	orderID := uuid.New()
	item1ID := uuid.New()
	item2ID := uuid.New()

	baseOrder := &model.Order{
		ID:        orderID,
		Status:    model.OrderStatusPending,
		Currency:  "EGP",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	items := []model.OrderItem{
		{
			ID:                item1ID,
			OrderID:           orderID,
			SellerID:          &sellerID,
			Quantity:          1,
			PriceAtPurchase:   5000,
			FulfillmentStatus: model.FulfillmentStatusPending,
		},
		{
			ID:                item2ID,
			OrderID:           orderID,
			SellerID:          &sellerID,
			Quantity:          2,
			PriceAtPurchase:   3000,
			FulfillmentStatus: model.FulfillmentStatusPending,
		},
	}

	exec := &fakeOrderFulfillExecutor{
		order: baseOrder,
		items: items,
	}
	notifierMock := &mockNotifier{}
	repo := NewRepository()
	svc := &OrderService{
		qexecutor:     &fakeTxRunner{exec: exec},
		orderRepo:     repo,
		logger:        log.NewNoopLogger(),
		orderNotifier: notifierMock,
	}

	input := FulfillSellerOrderInput{
		ItemIDs:        []uuid.UUID{item1ID, item2ID},
		TrackingNumber: "TRK-ALL-001",
		Carrier:        "Aramex",
	}

	res, err := svc.FulfillSellerOrderItems(context.Background(), sellerID, orderID, input)
	require.NoError(t, err)
	require.NotNil(t, res)

	// Verify parent order transitioned to Shipped because all items are fulfilled
	assert.Equal(t, model.OrderStatusShipped, exec.order.Status)
	assert.Contains(t, exec.statusUpdates, model.OrderStatusShipped)

	// Verify line items are marked shipped with tracking details
	for _, it := range exec.items {
		assert.Equal(t, model.FulfillmentStatusShipped, it.FulfillmentStatus)
		require.NotNil(t, it.TrackingNumber)
		assert.Equal(t, "TRK-ALL-001", *it.TrackingNumber)
		require.NotNil(t, it.Carrier)
		assert.Equal(t, "Aramex", *it.Carrier)
		assert.NotNil(t, it.ShippedAt)
	}

	// Verify status history audit
	require.NotEmpty(t, exec.notes)
	assert.Contains(t, exec.notes[0], "All seller items fulfilled")
	assert.Contains(t, exec.notes[0], "TRK-ALL-001")
	assert.Contains(t, exec.notes[0], "Aramex")
}

func TestFulfillSellerOrderItems_MultiSeller_PartialShipment(t *testing.T) {
	seller1ID := uuid.New()
	seller2ID := uuid.New()
	orderID := uuid.New()
	item1ID := uuid.New()
	item2ID := uuid.New()

	baseOrder := &model.Order{
		ID:        orderID,
		Status:    model.OrderStatusPending,
		Currency:  "EGP",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	items := []model.OrderItem{
		{
			ID:                item1ID,
			OrderID:           orderID,
			SellerID:          &seller1ID,
			Quantity:          1,
			PriceAtPurchase:   4000,
			FulfillmentStatus: model.FulfillmentStatusPending,
		},
		{
			ID:                item2ID,
			OrderID:           orderID,
			SellerID:          &seller2ID,
			Quantity:          1,
			PriceAtPurchase:   6000,
			FulfillmentStatus: model.FulfillmentStatusPending,
		},
	}

	exec := &fakeOrderFulfillExecutor{
		order: baseOrder,
		items: items,
	}
	repo := NewRepository()
	svc := &OrderService{
		qexecutor:     &fakeTxRunner{exec: exec},
		orderRepo:     repo,
		logger:        log.NewNoopLogger(),
		orderNotifier: &mockNotifier{},
	}

	// Step 1: Seller 1 fulfills Item 1 only
	input1 := FulfillSellerOrderInput{
		ItemIDs:        []uuid.UUID{item1ID},
		TrackingNumber: "TRK-SELLER1",
		Carrier:        "Bosta",
	}

	res1, err := svc.FulfillSellerOrderItems(context.Background(), seller1ID, orderID, input1)
	require.NoError(t, err)
	require.NotNil(t, res1)

	// Since only Item 1 is shipped, parent order status transitions to Processing (partial fulfillment)
	assert.Equal(t, model.OrderStatusProcessing, exec.order.Status)
	assert.Equal(t, model.FulfillmentStatusShipped, exec.items[0].FulfillmentStatus)
	assert.Equal(t, model.FulfillmentStatusPending, exec.items[1].FulfillmentStatus)

	// Step 2: Seller 2 fulfills Item 2
	input2 := FulfillSellerOrderInput{
		ItemIDs:        []uuid.UUID{item2ID},
		TrackingNumber: "TRK-SELLER2",
		Carrier:        "DHL",
	}

	res2, err := svc.FulfillSellerOrderItems(context.Background(), seller2ID, orderID, input2)
	require.NoError(t, err)
	require.NotNil(t, res2)

	// Now ALL items across both sellers are shipped -> parent order transitions to Shipped!
	assert.Equal(t, model.OrderStatusShipped, exec.order.Status)
	assert.Equal(t, model.FulfillmentStatusShipped, exec.items[1].FulfillmentStatus)
	assert.Equal(t, "TRK-SELLER2", *exec.items[1].TrackingNumber)
	assert.Equal(t, "DHL", *exec.items[1].Carrier)
}

func TestFulfillSellerOrderItems_CannotFulfillForeignItem(t *testing.T) {
	seller1ID := uuid.New()
	seller2ID := uuid.New()
	orderID := uuid.New()
	item2ID := uuid.New()

	baseOrder := &model.Order{
		ID:        orderID,
		Status:    model.OrderStatusPending,
		Currency:  "EGP",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	items := []model.OrderItem{
		{
			ID:                item2ID,
			OrderID:           orderID,
			SellerID:          &seller2ID, // owned by seller 2
			Quantity:          1,
			PriceAtPurchase:   3500,
			FulfillmentStatus: model.FulfillmentStatusPending,
		},
	}

	exec := &fakeOrderFulfillExecutor{
		order: baseOrder,
		items: items,
	}
	repo := NewRepository()
	svc := &OrderService{
		qexecutor:     &fakeTxRunner{exec: exec},
		orderRepo:     repo,
		logger:        log.NewNoopLogger(),
		orderNotifier: &mockNotifier{},
	}

	// Seller 1 attempts to fulfill Seller 2's item -> must be rejected with 403 Forbidden
	input := FulfillSellerOrderInput{
		ItemIDs:        []uuid.UUID{item2ID},
		TrackingNumber: "TRK-HACK",
		Carrier:        "Aramex",
	}

	_, err := svc.FulfillSellerOrderItems(context.Background(), seller1ID, orderID, input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not owned by this seller")
}

func TestFulfillSellerOrderItems_ItemNotInOrder(t *testing.T) {
	sellerID := uuid.New()
	orderID := uuid.New()
	randomItemID := uuid.New()

	baseOrder := &model.Order{
		ID:        orderID,
		Status:    model.OrderStatusPending,
		Currency:  "EGP",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	exec := &fakeOrderFulfillExecutor{
		order: baseOrder,
		items: []model.OrderItem{},
	}
	repo := NewRepository()
	svc := &OrderService{
		qexecutor:     &fakeTxRunner{exec: exec},
		orderRepo:     repo,
		logger:        log.NewNoopLogger(),
		orderNotifier: &mockNotifier{},
	}

	input := FulfillSellerOrderInput{
		ItemIDs:        []uuid.UUID{randomItemID},
		TrackingNumber: "TRK-001",
		Carrier:        "Aramex",
	}

	_, err := svc.FulfillSellerOrderItems(context.Background(), sellerID, orderID, input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not belong to this order")
}

func TestFulfillSellerOrderItems_CancelledOrder(t *testing.T) {
	sellerID := uuid.New()
	orderID := uuid.New()
	itemID := uuid.New()

	cancelledOrder := &model.Order{
		ID:        orderID,
		Status:    model.OrderStatusCancelled,
		Currency:  "EGP",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	exec := &fakeOrderFulfillExecutor{
		order: cancelledOrder,
		items: []model.OrderItem{
			{
				ID:                itemID,
				OrderID:           orderID,
				SellerID:          &sellerID,
				Quantity:          1,
				FulfillmentStatus: model.FulfillmentStatusPending,
			},
		},
	}
	repo := NewRepository()
	svc := &OrderService{
		qexecutor:     &fakeTxRunner{exec: exec},
		orderRepo:     repo,
		logger:        log.NewNoopLogger(),
		orderNotifier: &mockNotifier{},
	}

	input := FulfillSellerOrderInput{
		ItemIDs:        []uuid.UUID{itemID},
		TrackingNumber: "TRK-001",
		Carrier:        "Aramex",
	}

	_, err := svc.FulfillSellerOrderItems(context.Background(), sellerID, orderID, input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cancelled or refunded")
}

func TestFulfillSellerOrderItems_Validation(t *testing.T) {
	svc := &OrderService{}

	// Missing sellerID
	_, err := svc.FulfillSellerOrderItems(context.Background(), uuid.Nil, uuid.New(), FulfillSellerOrderInput{
		ItemIDs:        []uuid.UUID{uuid.New()},
		TrackingNumber: "123",
		Carrier:        "Aramex",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Seller ID is required")

	// Missing orderID
	_, err = svc.FulfillSellerOrderItems(context.Background(), uuid.New(), uuid.Nil, FulfillSellerOrderInput{
		ItemIDs:        []uuid.UUID{uuid.New()},
		TrackingNumber: "123",
		Carrier:        "Aramex",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Order ID is required")

	// Missing tracking number
	_, err = svc.FulfillSellerOrderItems(context.Background(), uuid.New(), uuid.New(), FulfillSellerOrderInput{
		ItemIDs:        []uuid.UUID{uuid.New()},
		TrackingNumber: "   ",
		Carrier:        "Aramex",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Tracking number is required")

	// Missing carrier
	_, err = svc.FulfillSellerOrderItems(context.Background(), uuid.New(), uuid.New(), FulfillSellerOrderInput{
		ItemIDs:        []uuid.UUID{uuid.New()},
		TrackingNumber: "123",
		Carrier:        "   ",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Carrier is required")
}

func TestFulfillSellerOrderItems_OmittedItemIDs_FulfillsAllSellerItems(t *testing.T) {
	sellerID := uuid.New()
	orderID := uuid.New()
	item1ID := uuid.New()
	item2ID := uuid.New()

	baseOrder := &model.Order{
		ID:        orderID,
		Status:    model.OrderStatusPending,
		Currency:  "EGP",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	items := []model.OrderItem{
		{
			ID:                item1ID,
			OrderID:           orderID,
			SellerID:          &sellerID,
			Quantity:          1,
			PriceAtPurchase:   2000,
			FulfillmentStatus: model.FulfillmentStatusPending,
		},
		{
			ID:                item2ID,
			OrderID:           orderID,
			SellerID:          &sellerID,
			Quantity:          1,
			PriceAtPurchase:   3000,
			FulfillmentStatus: model.FulfillmentStatusPending,
		},
	}

	exec := &fakeOrderFulfillExecutor{
		order: baseOrder,
		items: items,
	}
	repo := NewRepository()
	svc := &OrderService{
		qexecutor:     &fakeTxRunner{exec: exec},
		orderRepo:     repo,
		logger:        log.NewNoopLogger(),
		orderNotifier: &mockNotifier{},
	}

	// Omit ItemIDs -> automatically fulfills all items owned by sellerID
	input := FulfillSellerOrderInput{
		ItemIDs:        nil,
		TrackingNumber: "TRK-BULK",
		Carrier:        "Bosta",
	}

	res, err := svc.FulfillSellerOrderItems(context.Background(), sellerID, orderID, input)
	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, model.OrderStatusShipped, exec.order.Status)
	for _, it := range exec.items {
		assert.Equal(t, model.FulfillmentStatusShipped, it.FulfillmentStatus)
		assert.Equal(t, "TRK-BULK", *it.TrackingNumber)
		assert.Equal(t, "Bosta", *it.Carrier)
	}
}
