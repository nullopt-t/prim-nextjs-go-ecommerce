package order_test

import (
	"context"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/catalog/inventory"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

// fakeInventoryWriter is the in-memory stand-in for the inventory domain's
// participating writer. It records what the order use cases ask for so tests
// can assert that stock mutations are routed through the owning service and
// share the caller's executor, instead of being written straight to a table.
type fakeInventoryWriter struct {
	adjustments  []inventory.AdjustStockInput
	commitCalls  int
	releaseCalls int
	released     int64
	err          error
}

func (f *fakeInventoryWriter) ApplyStockAdjustment(
	_ context.Context,
	_ *database.Tx,
	in inventory.AdjustStockInput,
) error {
	f.adjustments = append(f.adjustments, in)
	return f.err
}

func (f *fakeInventoryWriter) ReleaseReservationsForOrder(
	_ context.Context,
	_ *database.Tx,
	_ uuid.UUID,
) (int64, error) {
	f.releaseCalls++
	return f.released, f.err
}

func (f *fakeInventoryWriter) CommitReservationsForOrder(
	_ context.Context,
	_ *database.Tx,
	_ uuid.UUID,
) error {
	f.commitCalls++
	return f.err
}
