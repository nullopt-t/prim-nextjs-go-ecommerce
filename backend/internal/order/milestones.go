package order

import "github.com/nullopt-t/prim-backend/internal/model"

// MilestonePercent is the monotonic fulfillment percent of a milestone.
// Single source of truth shared by derivation, the milestone state machine,
// and the API response (see docs/order-milestones.md).
var MilestonePercent = map[model.OrderMilestone]int{
	model.OrderMilestonePreTransit:          0,
	model.OrderMilestonePickedAtWarehouse:   5,
	model.OrderMilestoneArrivedRegionalHub:  10,
	model.OrderMilestoneArrivedLocalDepot:   25,
	model.OrderMilestoneOutForDelivery:      50,
	model.OrderMilestoneDelivered:           100,
}

// DeriveMilestone maps the order lifecycle status (plus tracking presence)
// onto the granular milestone a client renders. It never regresses: every
// lifecycle status maps to a milestone at or after the previous status's.
//
//	shipped without a tracking number -> arrived_regional_hub (10%)
//	shipped with a tracking number    -> out_for_delivery (50%)
//
// arrived_local_depot is reserved for future carrier scan events and is not
// emitted by derivation today.
func DeriveMilestone(status model.OrderStatus, hasTracking bool) model.OrderMilestone {
	switch status {
	case model.OrderStatusProcessing:
		return model.OrderMilestonePickedAtWarehouse
	case model.OrderStatusShipped:
		if hasTracking {
			return model.OrderMilestoneOutForDelivery
		}
		return model.OrderMilestoneArrivedRegionalHub
	case model.OrderStatusDelivered:
		return model.OrderMilestoneDelivered
	default:
		// pending, paid, cancelled, refunded, unknown: no fulfillment progress.
		return model.OrderMilestonePreTransit
	}
}

// CanTransitionMilestone enforces forward-only milestone motion: a milestone
// may only advance to a strictly later percent (or stay). Delivered (100%) is
// therefore terminal, and regressions are always rejected.
func CanTransitionMilestone(from, to model.OrderMilestone) bool {
	if from == to {
		return true
	}
	fromPct, fromOK := MilestonePercent[from]
	toPct, toOK := MilestonePercent[to]
	if !fromOK || !toOK {
		return false
	}
	return toPct > fromPct
}
