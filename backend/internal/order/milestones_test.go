package order_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/order"
)

// TestDeriveMilestone pins the status -> milestone derivation contract
// (docs/order-milestones.md).
func TestDeriveMilestone(t *testing.T) {
	cases := []struct {
		name        string
		status      model.OrderStatus
		hasTracking bool
		want        model.OrderMilestone
		wantPercent int
	}{
		{"pending", model.OrderStatusPending, false, model.OrderMilestonePreTransit, 0},
		{"paid", model.OrderStatusPaid, false, model.OrderMilestonePreTransit, 0},
		{"processing", model.OrderStatusProcessing, false, model.OrderMilestonePickedAtWarehouse, 5},
		{"shipped without tracking", model.OrderStatusShipped, false, model.OrderMilestoneArrivedRegionalHub, 10},
		{"shipped with tracking", model.OrderStatusShipped, true, model.OrderMilestoneOutForDelivery, 50},
		{"delivered", model.OrderStatusDelivered, true, model.OrderMilestoneDelivered, 100},
		{"cancelled", model.OrderStatusCancelled, false, model.OrderMilestonePreTransit, 0},
		{"refunded", model.OrderStatusRefunded, true, model.OrderMilestonePreTransit, 0},
		{"unknown", model.OrderStatus("weird"), false, model.OrderMilestonePreTransit, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := order.DeriveMilestone(tc.status, tc.hasTracking)
			assert.Equal(t, tc.want, got)
			assert.Equal(t, tc.wantPercent, order.MilestonePercent[got])
		})
	}
}

// TestMilestonePercentMonotonic proves the percent table strictly increases
// along the milestone chain, so CanTransitionMilestone can never skip or
// regress silently.
func TestMilestonePercentMonotonic(t *testing.T) {
	chain := []model.OrderMilestone{
		model.OrderMilestonePreTransit,
		model.OrderMilestonePickedAtWarehouse,
		model.OrderMilestoneArrivedRegionalHub,
		model.OrderMilestoneArrivedLocalDepot,
		model.OrderMilestoneOutForDelivery,
		model.OrderMilestoneDelivered,
	}
	prev := -1
	for _, m := range chain {
		pct, ok := order.MilestonePercent[m]
		require.True(t, ok, "milestone %s missing from percent table", m)
		assert.Greater(t, pct, prev, "percent for %s must exceed previous", m)
		prev = pct
	}
	assert.Equal(t, 100, prev)
}

// TestCanTransitionMilestone pins forward-only motion, same-status idempotency,
// delivered terminality, and rejection of unknown milestones.
func TestCanTransitionMilestone(t *testing.T) {
	assert.True(t, order.CanTransitionMilestone(model.OrderMilestonePreTransit, model.OrderMilestonePreTransit))
	assert.True(t, order.CanTransitionMilestone(model.OrderMilestonePickedAtWarehouse, model.OrderMilestoneOutForDelivery))
	assert.True(t, order.CanTransitionMilestone(model.OrderMilestoneOutForDelivery, model.OrderMilestoneDelivered))
	assert.False(t, order.CanTransitionMilestone(model.OrderMilestoneOutForDelivery, model.OrderMilestonePickedAtWarehouse), "regression must be rejected")
	assert.False(t, order.CanTransitionMilestone(model.OrderMilestoneDelivered, model.OrderMilestonePreTransit), "delivered is terminal")
	assert.False(t, order.CanTransitionMilestone(model.OrderMilestone("bogus"), model.OrderMilestoneDelivered), "unknown from must be rejected")
	assert.False(t, order.CanTransitionMilestone(model.OrderMilestonePreTransit, model.OrderMilestone("bogus")), "unknown to must be rejected")
}
