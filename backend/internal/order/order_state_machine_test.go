package order_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/order"
)

// TestCanTransitionOrderStatus pins every pair of the order lifecycle state
// machine, including terminal states and same-status idempotency.
func TestCanTransitionOrderStatus(t *testing.T) {
	statuses := []model.OrderStatus{
		model.OrderStatusPending,
		model.OrderStatusPaid,
		model.OrderStatusProcessing,
		model.OrderStatusShipped,
		model.OrderStatusDelivered,
		model.OrderStatusCancelled,
		model.OrderStatusRefunded,
	}

	allowed := map[model.OrderStatus][]model.OrderStatus{
		model.OrderStatusPending: {
			model.OrderStatusPaid,
			model.OrderStatusProcessing,
			model.OrderStatusCancelled,
		},
		model.OrderStatusPaid: {
			model.OrderStatusProcessing,
			model.OrderStatusShipped,
			model.OrderStatusRefunded,
			model.OrderStatusCancelled,
		},
		model.OrderStatusProcessing: {
			model.OrderStatusShipped,
			model.OrderStatusRefunded,
			model.OrderStatusCancelled,
		},
		model.OrderStatusShipped: {
			model.OrderStatusDelivered,
			model.OrderStatusRefunded,
		},
		model.OrderStatusDelivered: {
			model.OrderStatusRefunded,
		},
		model.OrderStatusCancelled: {},
		model.OrderStatusRefunded:  {},
	}

	for _, current := range statuses {
		for _, next := range statuses {
			want := current == next
			if !want {
				for _, candidate := range allowed[current] {
					if candidate == next {
						want = true
						break
					}
				}
			}

			t.Run(string(current)+"_to_"+string(next), func(t *testing.T) {
				assert.Equal(t, want, order.CanTransitionOrderStatus(current, next),
					"transition %s -> %s", current, next)
			})
		}
	}
}
