package checkout

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/catalog/coupon"
	"github.com/nullopt-t/prim-backend/internal/catalog/inventory"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/notifier"
	"github.com/nullopt-t/prim-backend/internal/order"
	"github.com/nullopt-t/prim-backend/internal/setting"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/nullopt-t/prim-backend/pkg/log"
	"github.com/nullopt-t/prim-backend/pkg/payment"
	"github.com/nullopt-t/prim-backend/pkg/storage"
)

type unitTestLogger struct{}

func (u unitTestLogger) Debug(string, ...log.Meta) {}
func (u unitTestLogger) Info(string, ...log.Meta)  {}
func (u unitTestLogger) Warn(string, ...log.Meta)  {}
func (u unitTestLogger) Error(string, ...log.Meta) {}

type unitTestRunner struct{}

func (u unitTestRunner) WithTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

func (u unitTestRunner) WithExecutor(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

func (u unitTestRunner) WithRetry(ctx context.Context, fn func(context.Context) error, _ int) error {
	return fn(ctx)
}

type unitTestSettings struct {
	cfg setting.Config
}

func (s unitTestSettings) GetConfig(_ context.Context) (setting.Config, error) {
	return s.cfg, nil
}

type unitTestOrderNotifier struct{}

func (u unitTestOrderNotifier) NotifyOrderPlaced(_ context.Context, _ notifier.OrderPlacedPayload) error {
	return nil
}

type unitTestAdminNotifier struct{}

func (u unitTestAdminNotifier) Publish(_ context.Context, _ notifier.AdminOrderEvent) error {
	return nil
}

type unitTestStockWriter struct {
	committed []uuid.UUID
}

func (s *unitTestStockWriter) HoldStock(_ context.Context, _ inventory.ReserveStockInput) (*model.InventoryReservation, error) {
	return nil, nil
}

func (s *unitTestStockWriter) CommitReservationsForOrder(_ context.Context, orderID uuid.UUID) error {
	s.committed = append(s.committed, orderID)
	return nil
}

func (s *unitTestStockWriter) ReleaseReservationsForOrder(_ context.Context, _ uuid.UUID) (int64, error) {
	return 1, nil
}

type unitTestOrderWriter struct {
	orders         map[uuid.UUID]*model.Order
	paymentsByTx   map[string]*model.Payment
	pendingPayment map[uuid.UUID]*model.Payment
	updatedPayment map[uuid.UUID]model.PaymentStatus
	orderStatuses  map[uuid.UUID]model.OrderStatus
	historyNotes   []string
}

func newUnitTestOrderWriter() *unitTestOrderWriter {
	return &unitTestOrderWriter{
		orders:         make(map[uuid.UUID]*model.Order),
		paymentsByTx:   make(map[string]*model.Payment),
		pendingPayment: make(map[uuid.UUID]*model.Payment),
		updatedPayment: make(map[uuid.UUID]model.PaymentStatus),
		orderStatuses:  make(map[uuid.UUID]model.OrderStatus),
	}
}

func (w *unitTestOrderWriter) PlaceOrder(_ context.Context, _ *order.CreateOrderInput) (*model.Order, error) {
	return nil, nil
}

func (w *unitTestOrderWriter) RecordPayment(_ context.Context, in *order.CreatePaymentInput) (*model.Payment, error) {
	p := &model.Payment{
		ID:                    uuid.New(),
		OrderID:               in.OrderID,
		Amount:                in.Amount,
		Currency:              in.Currency,
		Status:                in.Status,
		Type:                  in.Type,
		Provider:              in.Provider,
		ProviderTransactionID: in.ProviderTransactionID,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}
	if in.ProviderTransactionID != nil {
		w.paymentsByTx[*in.ProviderTransactionID] = p
	}
	return p, nil
}

func (w *unitTestOrderWriter) GetOrderForUpdate(_ context.Context, orderID uuid.UUID) (*model.Order, error) {
	ord, ok := w.orders[orderID]
	if !ok {
		return nil, database.ErrNotFound
	}
	return ord, nil
}

func (w *unitTestOrderWriter) FindPaymentByTransactionID(_ context.Context, transactionID string) (*model.Payment, error) {
	p, ok := w.paymentsByTx[transactionID]
	if !ok {
		return nil, database.ErrNotFound
	}
	return p, nil
}

func (w *unitTestOrderWriter) FindPendingPaymentForOrderForUpdate(_ context.Context, orderID uuid.UUID) (*model.Payment, error) {
	p, ok := w.pendingPayment[orderID]
	if !ok {
		return nil, database.ErrNotFound
	}
	return p, nil
}

func (w *unitTestOrderWriter) UpdatePaymentWebhook(_ context.Context, paymentID uuid.UUID, status model.PaymentStatus, transactionID string) error {
	w.updatedPayment[paymentID] = status
	for _, p := range w.pendingPayment {
		if p.ID == paymentID {
			p.Status = status
			p.ProviderTransactionID = &transactionID
			w.paymentsByTx[transactionID] = p
		}
	}
	for _, p := range w.paymentsByTx {
		if p.ID == paymentID {
			p.Status = status
			p.ProviderTransactionID = &transactionID
			w.paymentsByTx[transactionID] = p
		}
	}
	return nil
}

func (w *unitTestOrderWriter) UpdateOrderStatus(_ context.Context, orderID uuid.UUID, status model.OrderStatus, notes *string) error {
	w.orderStatuses[orderID] = status
	if ord, ok := w.orders[orderID]; ok {
		ord.Status = status
	}
	if notes != nil {
		w.historyNotes = append(w.historyNotes, *notes)
	}
	return nil
}

func (w *unitTestOrderWriter) AddOrderStatusHistory(_ context.Context, _ uuid.UUID, _ model.OrderStatus, notes *string) error {
	if notes != nil {
		w.historyNotes = append(w.historyNotes, *notes)
	}
	return nil
}

// Stubs for remaining unused readers/writers in webhook test
type stubCartReader struct{}

func (stubCartReader) LockCartByUserID(context.Context, uuid.UUID) error { return nil }
func (stubCartReader) GetCartByUserID(context.Context, uuid.UUID) (*model.Cart, error) {
	return nil, nil
}
func (stubCartReader) GetCartBySessionID(context.Context, string) (*model.Cart, error) {
	return nil, nil
}

type stubCartWriter struct{}

func (stubCartWriter) ClearItems(context.Context, uuid.UUID) error { return nil }

type stubUserReader struct{}

func (stubUserReader) GetByID(context.Context, uuid.UUID) (*model.User, error) { return nil, nil }

type stubAddressReader struct{}

func (stubAddressReader) GetByIDAndUser(context.Context, uuid.UUID, uuid.UUID) (*model.UserAddress, error) {
	return nil, nil
}

type stubCouponService struct{}

func (stubCouponService) GetCodeByID(context.Context, uuid.UUID) (string, error) { return "", nil }
func (stubCouponService) GetSellerIDByCode(context.Context, string) (*uuid.UUID, error) {
	return nil, nil
}
func (stubCouponService) ApplyByCode(context.Context, string, int64, *uuid.UUID) (*coupon.AppliedCoupon, error) {
	return nil, nil
}
func (stubCouponService) RecordRedemption(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, int64, string) error {
	return nil
}

type stubVariantReader struct{}

func (stubVariantReader) GetByIDForUpdate(context.Context, uuid.UUID) (*model.ProductVariant, error) {
	return nil, nil
}
func (stubVariantReader) GetByID(context.Context, uuid.UUID) (*model.ProductVariant, error) {
	return nil, nil
}

type stubProductReader struct{}

func (stubProductReader) GetByID(context.Context, uuid.UUID) (*model.Product, error) {
	return nil, nil
}
func (stubProductReader) GetBundleComposition(context.Context, uuid.UUID) (*model.BundleComposition, error) {
	return nil, database.ErrNotFound
}

type stubSellerReader struct{}

func (stubSellerReader) GetByID(context.Context, uuid.UUID) (*model.Seller, error) {
	return nil, nil
}

type stubObjectReader struct{}

func (stubObjectReader) GetByID(context.Context, uuid.UUID) (*model.Object, error) {
	return nil, nil
}

func setupTestService(orderWriter *unitTestOrderWriter, stockWriter *unitTestStockWriter, webhookSecret string, opts ...CheckoutOption) *CheckoutService {
	settings := unitTestSettings{
		cfg: setting.Config{
			StripeWebhookSecret: webhookSecret,
		},
	}
	return NewService(
		stubCartReader{},
		stubCartWriter{},
		orderWriter,
		stockWriter,
		stubUserReader{},
		stubAddressReader{},
		stubCouponService{},
		stubVariantReader{},
		stubProductReader{},
		stubSellerReader{},
		stubObjectReader{},
		storage.NoopProvider{},
		settings,
		unitTestRunner{},
		unitTestOrderNotifier{},
		unitTestAdminNotifier{},
		unitTestLogger{},
		opts...,
	)
}

func TestHandlePaymentWebhook_Success(t *testing.T) {
	secret := "whsec_test_secret_123"
	orderID := uuid.New()
	txID := "ch_test_charge_1"
	event := "payment_intent.succeeded"

	orderWriter := newUnitTestOrderWriter()
	orderWriter.orders[orderID] = &model.Order{
		ID:            orderID,
		Status:        model.OrderStatusPending,
		TotalAmount:   25000,
		Currency:      "EGP",
		PaymentMethod: "stripe",
	}
	pendingPaymentID := uuid.New()
	orderWriter.pendingPayment[orderID] = &model.Payment{
		ID:      pendingPaymentID,
		OrderID: orderID,
		Status:  model.StatusPending,
	}

	stockWriter := &unitTestStockWriter{}
	svc := setupTestService(orderWriter, stockWriter, secret)

	sig := payment.ComputeWebhookSignature(secret, event, orderID.String(), txID)

	res, err := svc.HandlePaymentWebhook(context.Background(), &PaymentWebhookInput{
		EventType:     event,
		OrderID:       orderID,
		TransactionID: txID,
		Signature:     sig,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Duplicate {
		t.Fatalf("expected duplicate=false on first delivery")
	}
	if res.OrderStatus != string(model.OrderStatusProcessing) {
		t.Errorf("expected order status %s, got %s", model.OrderStatusProcessing, res.OrderStatus)
	}
	if res.PaymentStatus != string(model.StatusCaptured) {
		t.Errorf("expected payment status %s, got %s", model.StatusCaptured, res.PaymentStatus)
	}

	// Verify order status updated in repo
	if orderWriter.orderStatuses[orderID] != model.OrderStatusProcessing {
		t.Errorf("orderWriter order status not updated to Processing")
	}
	// Verify stock committed
	if len(stockWriter.committed) != 1 || stockWriter.committed[0] != orderID {
		t.Errorf("expected stock commit for order %v, got %v", orderID, stockWriter.committed)
	}
	// Verify payment updated
	if orderWriter.updatedPayment[pendingPaymentID] != model.StatusCaptured {
		t.Errorf("expected payment updated to captured, got %v", orderWriter.updatedPayment[pendingPaymentID])
	}
}

func TestHandlePaymentWebhook_IdempotentDuplicate(t *testing.T) {
	secret := "whsec_test_secret_123"
	orderID := uuid.New()
	txID := "ch_test_charge_1"
	event := "payment_intent.succeeded"

	orderWriter := newUnitTestOrderWriter()
	orderWriter.orders[orderID] = &model.Order{
		ID:     orderID,
		Status: model.OrderStatusProcessing,
	}
	orderWriter.paymentsByTx[txID] = &model.Payment{
		ID:                    uuid.New(),
		OrderID:               orderID,
		Status:                model.StatusCaptured,
		ProviderTransactionID: &txID,
	}

	stockWriter := &unitTestStockWriter{}
	svc := setupTestService(orderWriter, stockWriter, secret)

	sig := payment.ComputeWebhookSignature(secret, event, orderID.String(), txID)

	res, err := svc.HandlePaymentWebhook(context.Background(), &PaymentWebhookInput{
		EventType:     event,
		OrderID:       orderID,
		TransactionID: txID,
		Signature:     sig,
	})
	if err != nil {
		t.Fatalf("unexpected error on idempotent retry: %v", err)
	}
	if !res.Duplicate {
		t.Fatalf("expected duplicate=true on replay")
	}
	// Stock must NOT be committed a second time
	if len(stockWriter.committed) != 0 {
		t.Fatalf("stock was re-committed on duplicate webhook!")
	}
}

func TestHandlePaymentWebhook_BadSignature(t *testing.T) {
	secret := "whsec_test_secret_123"
	orderID := uuid.New()
	txID := "ch_test_charge_1"
	event := "payment_intent.succeeded"

	orderWriter := newUnitTestOrderWriter()
	stockWriter := &unitTestStockWriter{}
	svc := setupTestService(orderWriter, stockWriter, secret)

	_, err := svc.HandlePaymentWebhook(context.Background(), &PaymentWebhookInput{
		EventType:     event,
		OrderID:       orderID,
		TransactionID: txID,
		Signature:     "invalid_sig_abc",
	})
	if err == nil {
		t.Fatalf("expected error on invalid signature, got nil")
	}
}

func TestHandlePaymentWebhook_MissingSecret(t *testing.T) {
	orderID := uuid.New()
	txID := "ch_test_charge_1"
	event := "payment_intent.succeeded"

	orderWriter := newUnitTestOrderWriter()
	stockWriter := &unitTestStockWriter{}
	svc := setupTestService(orderWriter, stockWriter, "") // empty secret

	_, err := svc.HandlePaymentWebhook(context.Background(), &PaymentWebhookInput{
		EventType:     event,
		OrderID:       orderID,
		TransactionID: txID,
		Signature:     "sig",
	})
	if err == nil {
		t.Fatalf("expected error on missing secret, got nil")
	}
}

func TestHandlePaymentWebhook_UnknownEventIgnored(t *testing.T) {
	secret := "whsec_test_secret_123"
	orderID := uuid.New()
	txID := "ch_test_1"

	orderWriter := newUnitTestOrderWriter()
	stockWriter := &unitTestStockWriter{}
	svc := setupTestService(orderWriter, stockWriter, secret)

	sig := payment.ComputeWebhookSignature(secret, "customer.created", orderID.String(), txID)

	res, err := svc.HandlePaymentWebhook(context.Background(), &PaymentWebhookInput{
		EventType:     "customer.created",
		OrderID:       orderID,
		TransactionID: txID,
		Signature:     sig,
	})
	if err != nil {
		t.Fatalf("unexpected error on unhandled event: %v", err)
	}
	if res.OrderStatus != "ignored" {
		t.Errorf("expected order status ignored, got %s", res.OrderStatus)
	}
}

func TestHandlePaymentWebhook_FailedPayment(t *testing.T) {
	secret := "whsec_test_secret_123"
	orderID := uuid.New()
	txID := "ch_test_failed_1"
	event := "payment_intent.failed"

	orderWriter := newUnitTestOrderWriter()
	orderWriter.orders[orderID] = &model.Order{
		ID:            orderID,
		Status:        model.OrderStatusPending,
		TotalAmount:   10000,
		Currency:      "EGP",
		PaymentMethod: "stripe",
	}
	pendingPaymentID := uuid.New()
	orderWriter.pendingPayment[orderID] = &model.Payment{
		ID:      pendingPaymentID,
		OrderID: orderID,
		Status:  model.StatusPending,
	}

	stockWriter := &unitTestStockWriter{}
	svc := setupTestService(orderWriter, stockWriter, secret)

	sig := payment.ComputeWebhookSignature(secret, event, orderID.String(), txID)

	res, err := svc.HandlePaymentWebhook(context.Background(), &PaymentWebhookInput{
		EventType:     event,
		OrderID:       orderID,
		TransactionID: txID,
		Signature:     sig,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.PaymentStatus != string(model.StatusFailed) {
		t.Errorf("expected payment status failed, got %s", res.PaymentStatus)
	}
	// Stock must NOT be committed for failed payment
	if len(stockWriter.committed) != 0 {
		t.Fatalf("stock was committed for failed payment!")
	}
	// Payment status must be updated to failed
	if orderWriter.updatedPayment[pendingPaymentID] != model.StatusFailed {
		t.Errorf("expected payment updated to failed, got %v", orderWriter.updatedPayment[pendingPaymentID])
	}
}

type customPaymentGatewayMock struct {
	called bool
	intent *payment.PaymentIntentResult
}

func (c *customPaymentGatewayMock) VerifyWebhook(_ context.Context, payload payment.WebhookPayload) (*payment.WebhookResult, error) {
	c.called = true
	return &payment.WebhookResult{
		EventType:     "custom.paid",
		OrderID:       payload.OrderID,
		TransactionID: payload.TransactionID,
		Succeeded:     true,
	}, nil
}

func (c *customPaymentGatewayMock) CreatePaymentIntent(_ context.Context, input payment.PaymentIntentInput) (*payment.PaymentIntentResult, error) {
	return &payment.PaymentIntentResult{
		ID:           "pi_custom_" + input.OrderID,
		ClientSecret: "pi_custom_" + input.OrderID + "_secret",
		Status:       "requires_payment_method",
	}, nil
}

func (c *customPaymentGatewayMock) RetrievePaymentIntent(_ context.Context, intentID string) (*payment.PaymentIntentResult, error) {
	if c.intent != nil {
		return c.intent, nil
	}
	return &payment.PaymentIntentResult{
		ID:     intentID,
		Status: "succeeded",
	}, nil
}

func TestHandlePaymentWebhook_WithCustomGatewayAdapter(t *testing.T) {
	orderID := uuid.New()
	txID := "tx_custom_99"

	orderWriter := newUnitTestOrderWriter()
	orderWriter.orders[orderID] = &model.Order{
		ID:            orderID,
		Status:        model.OrderStatusPending,
		TotalAmount:   5000,
		Currency:      "USD",
		PaymentMethod: "custom_gw",
	}

	stockWriter := &unitTestStockWriter{}
	customGW := &customPaymentGatewayMock{}

	svc := setupTestService(
		orderWriter,
		stockWriter,
		"some_secret",
		WithPaymentGateway(customGW),
	)

	res, err := svc.HandlePaymentWebhook(context.Background(), &PaymentWebhookInput{
		EventType:     "custom.paid",
		OrderID:       orderID,
		TransactionID: txID,
		Signature:     "any_sig",
	})
	if err != nil {
		t.Fatalf("unexpected error with custom adapter: %v", err)
	}
	if !customGW.called {
		t.Errorf("expected custom payment gateway VerifyWebhook to be invoked!")
	}
	if res.OrderStatus != string(model.OrderStatusProcessing) {
		t.Errorf("expected order status Processing, got %s", res.OrderStatus)
	}
}

func TestHandlePaymentWebhook_UnlinkedEvent_Returns200(t *testing.T) {
	secret := "whsec_test_secret_key"
	orderWriter := newUnitTestOrderWriter()
	stockWriter := &unitTestStockWriter{}
	svc := setupTestService(orderWriter, stockWriter, secret)

	// Valid Stripe signature and timestamp, but no order_id metadata
	now := time.Now().Unix()
	rawBody := []byte(`{"id":"evt_unlinked","type":"payment_intent.succeeded","data":{"object":{"id":"pi_unlinked"}}}`)
	sig := payment.ComputeStripeSignature(secret, now, rawBody)
	header := "t=" + strconv.FormatInt(now, 10) + ",v1=" + sig

	res, err := svc.HandlePaymentWebhook(context.Background(), &PaymentWebhookInput{
		Signature:  header,
		RawPayload: rawBody,
	})
	if err != nil {
		t.Fatalf("expected nil error on unlinked event, got: %v", err)
	}
	if res.OrderStatus != "ignored" && res.OrderStatus != "unlinked" {
		t.Errorf("expected ignored or unlinked status, got %s", res.OrderStatus)
	}
}

func TestHandlePaymentWebhook_PendingPaymentFulfillment(t *testing.T) {
	secret := "whsec_test_secret_key"
	orderID := uuid.New()
	piID := "pi_checkout_created_123"

	orderWriter := newUnitTestOrderWriter()
	orderWriter.orders[orderID] = &model.Order{
		ID:            orderID,
		Status:        model.OrderStatusPending,
		TotalAmount:   8500,
		Currency:      "EGP",
		PaymentMethod: "card",
	}
	// Simulate checkout recording a pending payment row with provider_transaction_id set to piID
	paymentRow := &model.Payment{
		ID:                    uuid.New(),
		OrderID:               orderID,
		Amount:                8500,
		Currency:              "EGP",
		Status:                model.StatusPending,
		Type:                  "card",
		Provider:              "stripe",
		ProviderTransactionID: &piID,
	}
	orderWriter.paymentsByTx[piID] = paymentRow
	orderWriter.pendingPayment[orderID] = paymentRow

	stockWriter := &unitTestStockWriter{}
	svc := setupTestService(orderWriter, stockWriter, secret)

	now := time.Now().Unix()
	rawBody := []byte(fmt.Sprintf(`{
		"id": "evt_test",
		"type": "payment_intent.succeeded",
		"data": {
			"object": {
				"id": "%s",
				"metadata": {
					"order_id": "%s"
				}
			}
		}
	}`, piID, orderID.String()))
	sig := payment.ComputeStripeSignature(secret, now, rawBody)
	header := "t=" + strconv.FormatInt(now, 10) + ",v1=" + sig

	res, err := svc.HandlePaymentWebhook(context.Background(), &PaymentWebhookInput{
		Signature:  header,
		RawPayload: rawBody,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Duplicate {
		t.Errorf("expected Duplicate to be false on initial payment fulfillment, got true")
	}
	if res.OrderStatus != string(model.OrderStatusProcessing) {
		t.Errorf("expected order status Processing, got %s", res.OrderStatus)
	}
	if res.PaymentStatus != string(model.StatusCaptured) {
		t.Errorf("expected payment status Captured, got %s", res.PaymentStatus)
	}
	if len(stockWriter.committed) == 0 {
		t.Errorf("expected stock reservations to be committed")
	}

	// Delivering the exact same webhook a second time must report Duplicate: true
	res2, err := svc.HandlePaymentWebhook(context.Background(), &PaymentWebhookInput{
		Signature:  header,
		RawPayload: rawBody,
	})
	if err != nil {
		t.Fatalf("unexpected error on replay: %v", err)
	}
	if !res2.Duplicate {
		t.Errorf("expected Duplicate to be true on replay, got false")
	}
}

func TestHandlePaymentWebhook_FailedThenSucceededRetry(t *testing.T) {
	secret := "whsec_test_secret_key"
	orderID := uuid.New()
	piID := "pi_retry_test_456"

	orderWriter := newUnitTestOrderWriter()
	orderWriter.orders[orderID] = &model.Order{
		ID:            orderID,
		Status:        model.OrderStatusPending,
		TotalAmount:   6000,
		Currency:      "EGP",
		PaymentMethod: "card",
	}
	paymentRow := &model.Payment{
		ID:                    uuid.New(),
		OrderID:               orderID,
		Amount:                6000,
		Currency:              "EGP",
		Status:                model.StatusPending,
		Type:                  "card",
		Provider:              "stripe",
		ProviderTransactionID: &piID,
	}
	orderWriter.paymentsByTx[piID] = paymentRow
	orderWriter.pendingPayment[orderID] = paymentRow

	stockWriter := &unitTestStockWriter{}
	svc := setupTestService(orderWriter, stockWriter, secret)

	now := time.Now().Unix()
	// Step 1: Payment failed event
	failBody := []byte(fmt.Sprintf(`{
		"id": "evt_fail",
		"type": "payment_intent.payment_failed",
		"data": {
			"object": {
				"id": "%s",
				"metadata": {
					"order_id": "%s"
				}
			}
		}
	}`, piID, orderID.String()))
	failSig := payment.ComputeStripeSignature(secret, now, failBody)
	failHeader := "t=" + strconv.FormatInt(now, 10) + ",v1=" + failSig

	resFail, err := svc.HandlePaymentWebhook(context.Background(), &PaymentWebhookInput{
		Signature:  failHeader,
		RawPayload: failBody,
	})
	if err != nil {
		t.Fatalf("unexpected error on failure webhook: %v", err)
	}
	if resFail.PaymentStatus != string(model.StatusFailed) {
		t.Errorf("expected payment status Failed, got %s", resFail.PaymentStatus)
	}
	if orderWriter.orders[orderID].Status != model.OrderStatusPending {
		t.Errorf("order status should remain Pending after failed attempt to allow retry")
	}

	// Step 2: Customer retried card on same intent, success event arrives
	succBody := []byte(fmt.Sprintf(`{
		"id": "evt_succ",
		"type": "payment_intent.succeeded",
		"data": {
			"object": {
				"id": "%s",
				"metadata": {
					"order_id": "%s"
				}
			}
		}
	}`, piID, orderID.String()))
	succSig := payment.ComputeStripeSignature(secret, now, succBody)
	succHeader := "t=" + strconv.FormatInt(now, 10) + ",v1=" + succSig

	resSucc, err := svc.HandlePaymentWebhook(context.Background(), &PaymentWebhookInput{
		Signature:  succHeader,
		RawPayload: succBody,
	})
	if err != nil {
		t.Fatalf("unexpected error on retry success: %v", err)
	}
	if resSucc.Duplicate {
		t.Errorf("expected retry success not to be marked duplicate")
	}
	if resSucc.PaymentStatus != string(model.StatusCaptured) {
		t.Errorf("expected payment status Captured on retry success, got %s", resSucc.PaymentStatus)
	}
	if resSucc.OrderStatus != string(model.OrderStatusProcessing) {
		t.Errorf("expected order status Processing, got %s", resSucc.OrderStatus)
	}
}

func TestHandlePaymentWebhook_CancelledOrder_ReturnsDuplicateNot400(t *testing.T) {
	secret := "whsec_test_secret_key"
	orderID := uuid.New()
	piID := "pi_cancelled_order_789"

	orderWriter := newUnitTestOrderWriter()
	orderWriter.orders[orderID] = &model.Order{
		ID:            orderID,
		Status:        model.OrderStatusCancelled,
		TotalAmount:   4000,
		Currency:      "EGP",
		PaymentMethod: "card",
	}

	stockWriter := &unitTestStockWriter{}
	svc := setupTestService(orderWriter, stockWriter, secret)

	now := time.Now().Unix()
	body := []byte(fmt.Sprintf(`{
		"id": "evt_cancel_succ",
		"type": "payment_intent.succeeded",
		"data": {
			"object": {
				"id": "%s",
				"metadata": {
					"order_id": "%s"
				}
			}
		}
	}`, piID, orderID.String()))
	sig := payment.ComputeStripeSignature(secret, now, body)
	header := "t=" + strconv.FormatInt(now, 10) + ",v1=" + sig

	res, err := svc.HandlePaymentWebhook(context.Background(), &PaymentWebhookInput{
		Signature:  header,
		RawPayload: body,
	})
	if err != nil {
		t.Fatalf("cancelled order webhook must return 200 OK not error to stop retry storms, got: %v", err)
	}
	if !res.Duplicate {
		t.Errorf("expected duplicate=true for webhook on cancelled order, got %v", res.Duplicate)
	}
}

func TestConfirmPayment_Success(t *testing.T) {
	customerID := uuid.New()
	orderID := uuid.New()
	piID := "pi_test_confirm_123"

	orderWriter := newUnitTestOrderWriter()
	orderWriter.orders[orderID] = &model.Order{
		ID:            orderID,
		CustomerID:    &customerID,
		Status:        model.OrderStatusPending,
		TotalAmount:   4000,
		Currency:      "EGP",
		PaymentMethod: "card",
	}
	orderWriter.pendingPayment[orderID] = &model.Payment{
		ID:                    uuid.New(),
		OrderID:               orderID,
		Amount:                4000,
		Currency:              "EGP",
		Status:                model.StatusPending,
		ProviderTransactionID: &piID,
	}

	stockWriter := &unitTestStockWriter{}
	customGW := &customPaymentGatewayMock{intent: &payment.PaymentIntentResult{
		ID:       piID,
		Status:   "succeeded",
		Amount:   4000,
		Currency: "egp",
		Metadata: map[string]string{"order_id": orderID.String()},
	}}
	svc := setupTestService(orderWriter, stockWriter, "secret", WithPaymentGateway(customGW))

	res, err := svc.ConfirmPayment(context.Background(), &ConfirmPaymentInput{
		CustomerID:      customerID,
		OrderID:         orderID,
		PaymentIntentID: piID,
	})
	if err != nil {
		t.Fatalf("unexpected confirm error: %v", err)
	}

	if res.OrderStatus != string(model.OrderStatusProcessing) {
		t.Errorf("expected order status processing, got %s", res.OrderStatus)
	}
	if len(stockWriter.committed) == 0 || stockWriter.committed[0] != orderID {
		t.Errorf("expected stock reservations to be committed on confirm")
	}
	if orderWriter.orders[orderID].Status != model.OrderStatusProcessing {
		t.Errorf("expected order to be processing, got %s", orderWriter.orders[orderID].Status)
	}
}

func TestConfirmPayment_MockMode(t *testing.T) {
	customerID := uuid.New()
	orderID := uuid.New()
	piID := "pi_mock_" + orderID.String()

	orderWriter := newUnitTestOrderWriter()
	orderWriter.orders[orderID] = &model.Order{
		ID:            orderID,
		CustomerID:    &customerID,
		Status:        model.OrderStatusPending,
		TotalAmount:   4000,
		Currency:      "EGP",
		PaymentMethod: "card",
	}
	orderWriter.pendingPayment[orderID] = &model.Payment{
		ID:                    uuid.New(),
		OrderID:               orderID,
		Amount:                4000,
		Currency:              "EGP",
		Status:                model.StatusPending,
		ProviderTransactionID: &piID,
	}

	stockWriter := &unitTestStockWriter{}
	customGW := &customPaymentGatewayMock{intent: &payment.PaymentIntentResult{
		ID:       piID,
		Status:   "succeeded",
		Amount:   4000,
		Currency: "egp",
		Metadata: map[string]string{"order_id": orderID.String()},
	}}
	svc := setupTestService(orderWriter, stockWriter, "secret", WithPaymentGateway(customGW))

	res, err := svc.ConfirmPayment(context.Background(), &ConfirmPaymentInput{
		CustomerID:      customerID,
		OrderID:         orderID,
		PaymentIntentID: piID,
	})
	if err != nil {
		t.Fatalf("unexpected mock confirm error: %v", err)
	}

	if res.OrderStatus != string(model.OrderStatusProcessing) {
		t.Errorf("expected order status processing, got %s", res.OrderStatus)
	}
	if len(stockWriter.committed) == 0 || stockWriter.committed[0] != orderID {
		t.Errorf("expected stock reservations to be committed in mock confirm")
	}
}

func TestConfirmPayment_RejectsIntentFromAnotherOrder(t *testing.T) {
	customerID := uuid.New()
	orderID := uuid.New()
	ownPI := "pi_own_123"

	orderWriter := newUnitTestOrderWriter()
	orderWriter.orders[orderID] = &model.Order{
		ID:            orderID,
		CustomerID:    &customerID,
		Status:        model.OrderStatusPending,
		TotalAmount:   4000,
		Currency:      "EGP",
		PaymentMethod: "card",
	}
	orderWriter.pendingPayment[orderID] = &model.Payment{
		ID:                    uuid.New(),
		OrderID:               orderID,
		Amount:                4000,
		Currency:              "EGP",
		Status:                model.StatusPending,
		ProviderTransactionID: &ownPI,
	}

	stockWriter := &unitTestStockWriter{}
	// The metadata says this order, but the id is not the one stored for it:
	// a paid intent of another order must not confirm this one.
	customGW := &customPaymentGatewayMock{intent: &payment.PaymentIntentResult{
		ID:       "pi_paid_elsewhere",
		Status:   "succeeded",
		Amount:   4000,
		Currency: "egp",
		Metadata: map[string]string{"order_id": orderID.String()},
	}}
	svc := setupTestService(orderWriter, stockWriter, "secret", WithPaymentGateway(customGW))

	if _, err := svc.ConfirmPayment(context.Background(), &ConfirmPaymentInput{
		CustomerID:      customerID,
		OrderID:         orderID,
		PaymentIntentID: "pi_paid_elsewhere",
	}); err == nil {
		t.Fatal("expected replay of a foreign intent to be rejected")
	}
	if orderWriter.orders[orderID].Status != model.OrderStatusPending {
		t.Errorf("expected order to stay pending, got %s", orderWriter.orders[orderID].Status)
	}
	if len(stockWriter.committed) != 0 {
		t.Errorf("expected no stock commit, got %v", stockWriter.committed)
	}
}

func TestConfirmPayment_RejectsProcessingIntent(t *testing.T) {
	customerID := uuid.New()
	orderID := uuid.New()
	piID := "pi_processing_123"

	orderWriter := newUnitTestOrderWriter()
	orderWriter.orders[orderID] = &model.Order{
		ID:            orderID,
		CustomerID:    &customerID,
		Status:        model.OrderStatusPending,
		TotalAmount:   4000,
		Currency:      "EGP",
		PaymentMethod: "card",
	}
	orderWriter.pendingPayment[orderID] = &model.Payment{
		ID:                    uuid.New(),
		OrderID:               orderID,
		Amount:                4000,
		Currency:              "EGP",
		Status:                model.StatusPending,
		ProviderTransactionID: &piID,
	}

	stockWriter := &unitTestStockWriter{}
	customGW := &customPaymentGatewayMock{intent: &payment.PaymentIntentResult{
		ID:       piID,
		Status:   "processing",
		Amount:   4000,
		Currency: "egp",
		Metadata: map[string]string{"order_id": orderID.String()},
	}}
	svc := setupTestService(orderWriter, stockWriter, "secret", WithPaymentGateway(customGW))

	if _, err := svc.ConfirmPayment(context.Background(), &ConfirmPaymentInput{
		CustomerID:      customerID,
		OrderID:         orderID,
		PaymentIntentID: piID,
	}); err == nil {
		t.Fatal("expected a processing intent to be rejected")
	}
	if orderWriter.orders[orderID].Status != model.OrderStatusPending {
		t.Errorf("expected order to stay pending, got %s", orderWriter.orders[orderID].Status)
	}
	if len(stockWriter.committed) != 0 {
		t.Errorf("expected no stock commit before settlement, got %v", stockWriter.committed)
	}
}

func TestConfirmPayment_CancelledOrderRecordsLateCapture(t *testing.T) {
	customerID := uuid.New()
	orderID := uuid.New()
	piID := "pi_late_123"

	orderWriter := newUnitTestOrderWriter()
	orderWriter.orders[orderID] = &model.Order{
		ID:            orderID,
		CustomerID:    &customerID,
		Status:        model.OrderStatusCancelled,
		TotalAmount:   4000,
		Currency:      "EGP",
		PaymentMethod: "card",
	}
	latePayment := &model.Payment{
		ID:                    uuid.New(),
		OrderID:               orderID,
		Amount:                4000,
		Currency:              "EGP",
		Status:                model.StatusFailed,
		ProviderTransactionID: &piID,
	}
	orderWriter.paymentsByTx[piID] = latePayment

	stockWriter := &unitTestStockWriter{}
	customGW := &customPaymentGatewayMock{intent: &payment.PaymentIntentResult{
		ID:       piID,
		Status:   "succeeded",
		Amount:   4000,
		Currency: "egp",
		Metadata: map[string]string{"order_id": orderID.String()},
	}}
	svc := setupTestService(orderWriter, stockWriter, "secret", WithPaymentGateway(customGW))

	if _, err := svc.ConfirmPayment(context.Background(), &ConfirmPaymentInput{
		CustomerID:      customerID,
		OrderID:         orderID,
		PaymentIntentID: piID,
	}); err == nil {
		t.Fatal("expected late capture to be reported as a refund-required conflict")
	}
	if orderWriter.updatedPayment[latePayment.ID] != model.StatusCaptured {
		t.Errorf("expected the captured payment to be recorded, got %s", orderWriter.updatedPayment[latePayment.ID])
	}
	found := false
	for _, note := range orderWriter.historyNotes {
		if strings.Contains(note, "refund required") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a refund-required history note, got %v", orderWriter.historyNotes)
	}
	if len(stockWriter.committed) != 0 {
		t.Errorf("expected no stock commit for a cancelled order, got %v", stockWriter.committed)
	}
}
