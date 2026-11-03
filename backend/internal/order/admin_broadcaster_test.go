package order

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/notifier"
	"github.com/nullopt-t/prim-backend/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type broadcasterTestLogger struct{}

func (broadcasterTestLogger) Info(string, ...log.Meta)  {}
func (broadcasterTestLogger) Warn(string, ...log.Meta)  {}
func (broadcasterTestLogger) Error(string, ...log.Meta) {}
func (broadcasterTestLogger) Debug(string, ...log.Meta) {}

func TestAdminBroadcaster_LocalDispatch(t *testing.T) {
	broadcaster := NewAdminBroadcaster(nil, broadcasterTestLogger{})
	defer broadcaster.Close()

	ch := broadcaster.Subscribe()

	ev := notifier.AdminOrderEvent{
		ID:            uuid.New(),
		OrderNumber:   "PRIM-TEST1234",
		CustomerEmail: "test@example.com",
		CustomerName:  "Test Customer",
		TotalAmount:   50000,
		Currency:      "EGP",
		Status:        "pending",
		CreatedAt:     time.Now(),
	}

	err := broadcaster.Publish(context.Background(), ev)
	require.NoError(t, err)

	select {
	case received := <-ch:
		require.Equal(t, "order.created", received.Type)
		require.NotNil(t, received.Order)
		assert.Equal(t, ev.ID, received.Order.ID)
		assert.Equal(t, ev.OrderNumber, received.Order.OrderNumber)
		assert.Equal(t, ev.TotalAmount, received.Order.TotalAmount)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for broadcast event")
	}

	broadcaster.Unsubscribe(ch)
}
