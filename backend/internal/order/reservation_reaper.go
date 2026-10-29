package order

import (
	"context"
	"sync"
	"time"

	"github.com/nullopt-t/prim-backend/pkg/log"
)

// ReservationReaper periodically expires stale checkout holds by cancelling
// pending orders whose reservation TTL elapsed and returning their held stock
// to salable stock.
//
// Business rationale: a checkout places a short-lived hold on stock. If the
// customer never pays, the hold must be resolved or it would strand inventory
// indefinitely. Paid orders are handled by CommitReservationsForOrder
// (payment capture); this reaper handles the "never paid" branch.
//
// Safety: ReapExpiredReservations locks each order and only cancels it while it
// is still pending, so running the reaper from more than one process (or on
// overlapping ticks) cannot double-cancel or double-release.
type ReservationReaper struct {
	service  *OrderService
	logger   log.Logger
	interval time.Duration
}

// NewReservationReaper builds a reaper. A non-positive interval defaults to one
// minute.
func NewReservationReaper(service *OrderService, logger log.Logger, interval time.Duration) *ReservationReaper {
	if interval <= 0 {
		interval = time.Minute
	}
	return &ReservationReaper{
		service:  service,
		logger:   logger,
		interval: interval,
	}
}

// Start runs the reaper until ctx is cancelled, then returns. It is designed to
// be launched as a goroutine that shares the caller's WaitGroup.
func (r *ReservationReaper) Start(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	r.logger.Info("starting reservation reaper", log.Meta{"interval": r.interval.String()})

	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			r.logger.Info("reservation reaper stopped")
			return
		case <-ticker.C:
			cancelled, err := r.service.ReapExpiredReservations(ctx)
			if err != nil {
				r.logger.Error("reservation reap failed", log.Meta{"error": err.Error()})
				continue
			}
			if cancelled > 0 {
				r.logger.Info("expired stale reservations", log.Meta{"ordersCancelled": cancelled})
			}
		}
	}
}
