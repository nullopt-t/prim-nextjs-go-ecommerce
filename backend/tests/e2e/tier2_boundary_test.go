package e2e

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

// =============================================================================
// Tier 2: Category 1 - Canonical Error Mapping Edge Cases (R5)
// =============================================================================

func TestTier2_Boundary_ErrNoRows_MapsTo404NotFound(t *testing.T) {
	testCases := []struct {
		name        string
		inputErr    error
		fallbackMsg string
	}{
		{"pgx.ErrNoRows", pgx.ErrNoRows, "product"},
		{"database.ErrNotFound", database.ErrNotFound, "variant"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mapped := CanonicalHandleError(tc.inputErr, tc.fallbackMsg)
			require.Error(t, mapped)

			var apiErr *apierr.APIError
			require.True(t, errors.As(mapped, &apiErr))
			assert.Equal(t, http.StatusNotFound, apiErr.Status)
			assert.Contains(t, apiErr.Error(), tc.fallbackMsg+" not found")
		})
	}
}

func TestTier2_Boundary_UniqueViolation_MapsTo409Conflict(t *testing.T) {
	mapped := CanonicalHandleError(database.ErrConflict, "product SKU")
	require.Error(t, mapped)

	var apiErr *apierr.APIError
	require.True(t, errors.As(mapped, &apiErr))
	assert.Equal(t, http.StatusConflict, apiErr.Status)
	assert.Contains(t, apiErr.Error(), "already exists or conflict occurred")
}

func TestTier2_Boundary_ForeignKeyViolation_MapsTo400BadRequest(t *testing.T) {
	mapped := CanonicalHandleError(database.ErrForeignKeyViolation, "order category")
	require.Error(t, mapped)

	var apiErr *apierr.APIError
	require.True(t, errors.As(mapped, &apiErr))
	assert.Equal(t, http.StatusBadRequest, apiErr.Status)
	assert.Contains(t, apiErr.Error(), "Referenced entity does not exist")
}

func TestTier2_Boundary_NotNullViolation_MapsTo400BadRequest(t *testing.T) {
	mapped := CanonicalHandleError(database.ErrNotNullViolation, "user email")
	require.Error(t, mapped)

	var apiErr *apierr.APIError
	require.True(t, errors.As(mapped, &apiErr))
	assert.Equal(t, http.StatusBadRequest, apiErr.Status)
	assert.Contains(t, apiErr.Error(), "Missing required field")
}

func TestTier2_Boundary_CheckViolation_MapsTo400BadRequest(t *testing.T) {
	mapped := CanonicalHandleError(database.ErrCheckViolation, "cart price")
	require.Error(t, mapped)

	var apiErr *apierr.APIError
	require.True(t, errors.As(mapped, &apiErr))
	assert.Equal(t, http.StatusBadRequest, apiErr.Status)
	assert.Contains(t, apiErr.Error(), "Constraint check violation")
}

func TestTier2_Boundary_ContextCanceled_MapsTo499ClientClosed(t *testing.T) {
	testCases := []struct {
		name     string
		inputErr error
	}{
		{"context.Canceled", context.Canceled},
		{"database.ErrCanceled", database.ErrCanceled},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mapped := CanonicalHandleError(tc.inputErr, "fetch inventory")
			require.Error(t, mapped)

			var apiErr *apierr.APIError
			require.True(t, errors.As(mapped, &apiErr))
			assert.Equal(t, 499, apiErr.Status)
			assert.Equal(t, apierr.CodeClientClosedRequest, apiErr.Code)
		})
	}
}

func TestTier2_Boundary_ContextDeadlineExceeded_MapsTo408Timeout(t *testing.T) {
	testCases := []struct {
		name     string
		inputErr error
	}{
		{"context.DeadlineExceeded", context.DeadlineExceeded},
		{"database.ErrTimeout", database.ErrTimeout},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mapped := CanonicalHandleError(tc.inputErr, "order query")
			require.Error(t, mapped)

			var apiErr *apierr.APIError
			require.True(t, errors.As(mapped, &apiErr))
			assert.Equal(t, http.StatusRequestTimeout, apiErr.Status)
			assert.Equal(t, "TIMEOUT", apiErr.Code)
		})
	}
}

func TestTier2_Boundary_AlreadyWrappedAPIError_PreservesIdentity(t *testing.T) {
	original := apierr.New(http.StatusPaymentRequired, "Payment failed").WithCode("PAYMENT_FAILED")

	// Passing an already typed APIError through handleError must return it untouched
	mapped := CanonicalHandleError(original, "process payment")
	require.Error(t, mapped)

	var apiErr *apierr.APIError
	require.True(t, errors.As(mapped, &apiErr))
	assert.Equal(t, http.StatusPaymentRequired, apiErr.Status)
	assert.Equal(t, "PAYMENT_FAILED", apiErr.Code)
	assert.Equal(t, "Payment failed", apiErr.Error())
}

func TestTier2_Boundary_UnexpectedDriverError_SanitizesInternalDetails(t *testing.T) {
	leakyRawDriverErr := errors.New("pq: syntax error at or near \"SELECT\" in relation public.orders")

	mapped := CanonicalHandleError(leakyRawDriverErr, "load customer orders")
	require.Error(t, mapped)

	var apiErr *apierr.APIError
	require.True(t, errors.As(mapped, &apiErr))
	assert.Equal(t, http.StatusInternalServerError, apiErr.Status)
	assert.Equal(t, apierr.CodeInternalError, apiErr.Code)

	// Ensure no SQL or database table internals are exposed to client
	assert.False(t, strings.Contains(apiErr.Error(), "SELECT"), "Must not leak SQL syntax")
	assert.False(t, strings.Contains(apiErr.Error(), "public.orders"), "Must not leak table name")
	assert.Contains(t, apiErr.Error(), "Failed to load customer orders")
}

// =============================================================================
// Tier 2: Category 2 - Connection Invariants & Concurrency Corners (R2)
// =============================================================================

func TestTier2_Boundary_MaxConnsOne_ExecutesReadFlowWithoutPoolStarvation(t *testing.T) {
	runner := NewMockRunner(1) // Pool limit = 1
	repo := NewMockProductRepository()
	variantRepo := NewMockVariantRepository()

	prodID := uuid.New()
	repo.AddProduct(&model.Product{ID: prodID, Title: "Primary Item"})
	variantRepo.AddVariant(&model.ProductVariant{
		ID:        uuid.New(),
		ProductID: prodID,
		SKU:       "SKU-1",
	})

	// Rule 2: Execute multi-entity read on the held qe without acquiring second connection
	err := runner.WithExecutor(context.Background(), func(db database.QueryExecutor) error {
		p, err := repo.GetByID(context.Background(), db, prodID)
		if err != nil {
			return err
		}
		count, err := variantRepo.CountByProductID(context.Background(), db, p.ID)
		if err != nil {
			return err
		}
		assert.Equal(t, 1, count)
		return nil
	})

	require.NoError(t, err)
	assert.False(t, runner.ConnectionHopDetected(), "Must not detect any connection hopping")
	withDB, withTx, _, _ := runner.Stats()
	assert.Equal(t, 1, withDB)
	assert.Equal(t, 0, withTx)
}

func TestTier2_Boundary_MaxConnsOne_ExecutesWriteFlowWithoutPoolStarvation(t *testing.T) {
	runner := NewMockRunner(1)
	orderRepo := NewMockOrderRepository()
	orderID := uuid.New()

	// Validate + mutate + reload strictly on the same tx
	err := runner.WithTx(context.Background(), func(ctx context.Context, tx database.QueryExecutor) error {
		// Step 1: Validate / insert
		if err := orderRepo.CreateOrder(ctx, tx, &model.Order{ID: orderID, Status: model.OrderStatusPending}); err != nil {
			return err
		}
		// Step 2: Mutate status
		if err := orderRepo.UpdateOrderStatus(ctx, tx, orderID, model.OrderStatusProcessing); err != nil {
			return err
		}
		// Step 3: Reload on same tx
		reloaded, err := orderRepo.GetOrderByID(ctx, tx, orderID)
		if err != nil {
			return err
		}
		assert.Equal(t, model.OrderStatusProcessing, reloaded.Status)
		return nil
	})

	require.NoError(t, err)
	assert.False(t, runner.ConnectionHopDetected())
	_, withTx, commits, rollbacks := runner.Stats()
	assert.Equal(t, 1, withTx)
	assert.Equal(t, 1, commits)
	assert.Equal(t, 0, rollbacks)
}

func TestTier2_Boundary_ConnectionHopDetection_TriggersPoolExhaustion(t *testing.T) {
	runner := NewMockRunner(1) // Pool limit = 1

	// Simulate an anti-pattern: inner helper calls WithExecutor while WithTx holds a connection
	err := runner.WithTx(context.Background(), func(ctx context.Context, tx database.QueryExecutor) error {
		// Anti-pattern: inner WithExecutor without passing held connection
		return runner.WithExecutor(ctx, func(nestedDB database.QueryExecutor) error {
			return nil
		})
	})

	require.Error(t, err)
	assert.True(t, runner.ConnectionHopDetected(), "Must detect connection hopping violation")
	assert.Contains(t, err.Error(), "pool connection limit exceeded")
}

func TestTier2_Boundary_TxRollbackOnCallbackError(t *testing.T) {
	runner := NewMockRunner(1)
	orderRepo := NewMockOrderRepository()
	orderID := uuid.New()

	err := runner.WithTx(context.Background(), func(ctx context.Context, tx database.QueryExecutor) error {
		_ = orderRepo.CreateOrder(ctx, tx, &model.Order{ID: orderID, TotalAmount: 500})
		// Fail deliberately inside transaction
		return errors.New("simulated business rule rejection")
	})

	require.Error(t, err)
	_, withTx, commits, rollbacks := runner.Stats()
	assert.Equal(t, 1, withTx)
	assert.Equal(t, 0, commits, "Must not commit on error")
	assert.Equal(t, 1, rollbacks, "Must rollback on error")
}

func TestTier2_Boundary_TxRollbackOnPanic_AndPropagatesPanic(t *testing.T) {
	runner := NewMockRunner(1)

	defer func() {
		p := recover()
		require.NotNil(t, p, "Panic must propagate after transaction rollback")
		assert.Equal(t, "unexpected panic during tx execution", p)

		_, _, commits, rollbacks := runner.Stats()
		assert.Equal(t, 0, commits)
		assert.Equal(t, 1, rollbacks, "Rollback must execute during panic recovery")
	}()

	_ = runner.WithTx(context.Background(), func(ctx context.Context, tx database.QueryExecutor) error {
		panic("unexpected panic during tx execution")
	})
}

func TestTier2_Boundary_OptionalLookup_IgnoresErrNoRows_WithoutAbortingTx(t *testing.T) {
	runner := NewMockRunner(1)
	objRepo := NewMockObjectRepository()
	prodRepo := NewMockProductRepository()

	prodID := uuid.New()
	missingLogoID := uuid.New()

	// Optional lookup pattern from SKILL.md: treat ErrNoRows as "absent" and continue
	err := runner.WithTx(context.Background(), func(ctx context.Context, tx database.QueryExecutor) error {
		prod := &model.Product{ID: prodID, Title: "Logo-less Product"}
		prodRepo.AddProduct(prod)

		// Optional logo lookup
		logo, err := objRepo.GetByID(ctx, tx, missingLogoID)
		if err != nil && !errors.Is(err, database.ErrNotFound) {
			return CanonicalHandleError(err, "logo")
		}
		if logo != nil {
			t.Fatal("Expected logo to be absent")
		}

		// Transaction continues successfully
		return nil
	})

	require.NoError(t, err)
	_, _, commits, rollbacks := runner.Stats()
	assert.Equal(t, 1, commits)
	assert.Equal(t, 0, rollbacks)
}
