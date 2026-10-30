package notifier

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/shared/html"
	"github.com/nullopt-t/prim-backend/internal/shared/job"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeMailer struct {
	to      []string
	subject string
	body    string
	err     error
}

func (m *fakeMailer) SendHTML(to []string, subject, body string) error {
	m.to = to
	m.subject = subject
	m.body = body
	return m.err
}

type fakeNotificationStore struct {
	notification *model.UserNotification
	err          error
}

func (s *fakeNotificationStore) CreateNotification(_ context.Context, n *model.UserNotification) error {
	s.notification = n
	return s.err
}

func orderPlacedEvent(t *testing.T, locale string) *job.NotificationEvent {
	t.Helper()

	data, err := json.Marshal(OrderPlacedPayload{
		OrderID:       uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		OrderNumber:   "PRIM-1234",
		CustomerEmail: "buyer@example.com",
		CustomerName:  "Buyer Name",
		TotalAmount:   2598,
		Currency:      "EGP",
		PaymentMethod: "cash_on_delivery",
		Items: []OrderItemSummary{{
			Title:     "Wireless Headphones",
			SKU:       "WH-1",
			Quantity:  2,
			UnitPrice: 1299,
			LineTotal: 2598,
		}},
		ShippingAddress: OrderAddressSummary{
			Street:  "1 Main Street",
			City:    "Cairo",
			Country: "EG",
		},
	})
	require.NoError(t, err)

	return &job.NotificationEvent{
		EventID:   uuid.New(),
		Type:      "email",
		Category:  "order",
		Recipient: "buyer@example.com",
		Locale:    locale,
		Template:  "order_placed",
		Data:      data,
	}
}

func TestHandleEmailEvent_OrderPlacedEnglish(t *testing.T) {
	renderer, err := html.NewRenderer()
	require.NoError(t, err)
	mailer := &fakeMailer{}
	handler := NewEventHandler(renderer, mailer, &fakeNotificationStore{}, testLogger{}, "https://prim.test")

	err = handler.HandleEmailEvent(context.Background(), orderPlacedEvent(t, "en"))
	require.NoError(t, err)

	assert.Equal(t, []string{"buyer@example.com"}, mailer.to)
	assert.Equal(t, "Order Confirmation - PRIM", mailer.subject)
	assert.Contains(t, mailer.body, "PRIM-1234")
	assert.Contains(t, mailer.body, "Buyer Name")
	assert.Contains(t, mailer.body, "Wireless Headphones")
	assert.Contains(t, mailer.body, "25.98 EGP")
	assert.Contains(t, mailer.body, "https://prim.test/user/orders/11111111-1111-1111-1111-111111111111")
	assert.Contains(t, mailer.body, `dir="ltr"`)
}

func TestHandleEmailEvent_OrderPlacedArabic(t *testing.T) {
	renderer, err := html.NewRenderer()
	require.NoError(t, err)
	mailer := &fakeMailer{}
	handler := NewEventHandler(renderer, mailer, &fakeNotificationStore{}, testLogger{}, "https://prim.test")

	err = handler.HandleEmailEvent(context.Background(), orderPlacedEvent(t, "ar"))
	require.NoError(t, err)

	assert.Equal(t, "تأكيد طلبك - بريم", mailer.subject)
	assert.Contains(t, mailer.body, "PRIM-1234")
	assert.Contains(t, mailer.body, "Wireless Headphones")
	assert.Contains(t, mailer.body, "الدفع عند الاستلام")
	assert.Contains(t, mailer.body, `dir="rtl"`)
}

func TestHandleEmailEvent_OrderShippedEnglish(t *testing.T) {
	renderer, err := html.NewRenderer()
	require.NoError(t, err)
	mailer := &fakeMailer{}
	handler := NewEventHandler(renderer, mailer, &fakeNotificationStore{}, testLogger{}, "https://prim.test")

	data, err := json.Marshal(OrderShippedPayload{
		OrderID:        uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		OrderNumber:    "PRIM-5678",
		CustomerEmail:  "buyer@example.com",
		CustomerName:   "Buyer Name",
		Carrier:        "Aramex",
		TrackingNumber: "TRACK-999",
		TrackingURL:    "https://carrier.example/track/TRACK-999",
	})
	require.NoError(t, err)

	err = handler.HandleEmailEvent(context.Background(), &job.NotificationEvent{
		EventID:   uuid.New(),
		Type:      "email",
		Category:  "order",
		Recipient: "buyer@example.com",
		Locale:    "en",
		Template:  "order_shipped",
		Data:      data,
	})
	require.NoError(t, err)

	assert.Equal(t, "Your Order Has Shipped! - PRIM", mailer.subject)
	assert.Contains(t, mailer.body, "PRIM-5678")
	assert.Contains(t, mailer.body, "Aramex")
	assert.Contains(t, mailer.body, "TRACK-999")
	assert.Contains(t, mailer.body, "https://carrier.example/track/TRACK-999")
}

func TestHandleEmailEvent_OrderCancelledArabic(t *testing.T) {
	renderer, err := html.NewRenderer()
	require.NoError(t, err)
	mailer := &fakeMailer{}
	handler := NewEventHandler(renderer, mailer, &fakeNotificationStore{}, testLogger{}, "https://prim.test")

	data, err := json.Marshal(OrderCancelledPayload{
		OrderID:       uuid.MustParse("33333333-3333-3333-3333-333333333333"),
		OrderNumber:   "PRIM-9012",
		CustomerEmail: "buyer@example.com",
		CustomerName:  "Buyer Name",
		Reason:        "Out of stock",
	})
	require.NoError(t, err)

	err = handler.HandleEmailEvent(context.Background(), &job.NotificationEvent{
		EventID:   uuid.New(),
		Type:      "email",
		Category:  "order",
		Recipient: "buyer@example.com",
		Locale:    "ar",
		Template:  "order_cancelled",
		Data:      data,
	})
	require.NoError(t, err)

	assert.Equal(t, "تم إلغاء الطلب - بريم", mailer.subject)
	assert.Contains(t, mailer.body, "PRIM-9012")
	assert.Contains(t, mailer.body, "Out of stock")
	assert.Contains(t, mailer.body, `dir="rtl"`)
}

func TestHandleInAppEvent_PersistsOrderNotification(t *testing.T) {
	renderer, err := html.NewRenderer()
	require.NoError(t, err)
	store := &fakeNotificationStore{}
	handler := NewEventHandler(renderer, &fakeMailer{}, store, testLogger{}, "https://prim.test")
	customerID := uuid.New()
	actionURL := "/user/orders/" + uuid.NewString()

	data, err := json.Marshal(SendInAppInput{
		UserID:    customerID,
		Title:     map[string]string{"en": "Order confirmed", "ar": "تم تأكيد طلبك"},
		Message:   map[string]string{"en": "Your order was placed.", "ar": "تم إنشاء طلبك."},
		Category:  "order",
		ActionURL: &actionURL,
	})
	require.NoError(t, err)

	err = handler.HandleInAppEvent(context.Background(), &job.NotificationEvent{
		EventID: uuid.New(),
		Type:    "in_app",
		Data:    data,
	})
	require.NoError(t, err)
	require.NotNil(t, store.notification)

	assert.Equal(t, customerID, store.notification.UserID)
	assert.Equal(t, "order", store.notification.Category)
	require.NotNil(t, store.notification.ActionURL)
	assert.Equal(t, actionURL, *store.notification.ActionURL)
	assert.Equal(t, "تم تأكيد طلبك", store.notification.LocalizedTitle("ar"))
	assert.Equal(t, "Your order was placed.", store.notification.LocalizedMessage("en"))
}
