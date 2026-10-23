package product

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nullopt-t/prim-backend/internal/catalog/product/variant"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockQE struct {
	execFn func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

func (m *mockQE) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	if m.execFn != nil {
		return m.execFn(ctx, sql, arguments...)
	}
	return pgconn.NewCommandTag("UPDATE 1"), nil
}

func (m *mockQE) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return nil, pgx.ErrNoRows
}

func (m *mockQE) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return nil
}

type mockConnectionTrackingRunner struct {
	withDBCalls int
	withTxCalls int
	qe          *mockQE
}

func (m *mockConnectionTrackingRunner) WithDB(ctx context.Context, fn func(database.QueryExecutor) error) error {
	m.withDBCalls++
	return fn(m.qe)
}

func (m *mockConnectionTrackingRunner) WithTx(ctx context.Context, fn func(context.Context, database.QueryExecutor) error) error {
	m.withTxCalls++
	return fn(ctx, m.qe)
}

type fakeVariantRepoForInvariants struct {
	countFunc func() (int, error)
	getFunc   func(id uuid.UUID) (*model.ProductVariant, error)
}

func (f *fakeVariantRepoForInvariants) CountByProductID(ctx context.Context, qe database.QueryExecutor, productID uuid.UUID) (int, error) {
	if f.countFunc != nil {
		return f.countFunc()
	}
	return 1, nil
}

func (f *fakeVariantRepoForInvariants) GetByID(ctx context.Context, qe database.QueryExecutor, id uuid.UUID) (*model.ProductVariant, error) {
	if f.getFunc != nil {
		return f.getFunc(id)
	}
	return nil, nil
}

func (f *fakeVariantRepoForInvariants) ListByProductID(ctx context.Context, qe database.QueryExecutor, opts variant.ListVariantOptions) (*pagination.PagedResult[model.ProductVariant], error) {
	return nil, nil
}

func TestAdversarial_PublishProduct_SingleWithDB(t *testing.T) {
	qe := &mockQE{}
	runner := &mockConnectionTrackingRunner{qe: qe}
	prodID := uuid.New()

	vRepo := &fakeVariantRepoForInvariants{
		countFunc: func() (int, error) {
			return 2, nil
		},
	}

	svc := &ProductService{
		dbRunner:    runner,
		variantRepo: vRepo,
		productRepo: NewProductRepository(),
	}

	err := svc.PublishProduct(context.Background(), prodID)
	require.NoError(t, err)

	// Invariant check: Exactly one WithDB call, zero WithTx calls
	assert.Equal(t, 1, runner.withDBCalls, "PublishProduct must use exactly one WithDB")
	assert.Equal(t, 0, runner.withTxCalls, "PublishProduct must NOT use WithTx")
}

func TestAdversarial_SetDefaultVariant_SingleWithTx(t *testing.T) {
	prodID := uuid.New()
	varID := uuid.New()

	qe := &mockQE{}
	runner := &mockConnectionTrackingRunner{qe: qe}

	vRepo := &fakeVariantRepoForInvariants{
		getFunc: func(id uuid.UUID) (*model.ProductVariant, error) {
			return &model.ProductVariant{
				ID:        varID,
				ProductID: prodID,
			}, nil
		},
	}

	svc := &ProductService{
		dbRunner:    runner,
		variantRepo: vRepo,
		productRepo: NewProductRepository(),
	}

	// This call will fail inside WithTx at productRepo.GetByID (returning ErrNoRows from our mock query)
	// which tests that WithTx was entered, executed the query on tx, and rolled back/returned cleanly.
	err := svc.SetDefaultVariant(context.Background(), prodID, varID)
	require.Error(t, err)

	// Invariant check: Exactly one WithTx call, zero WithDB calls
	assert.Equal(t, 0, runner.withDBCalls, "SetDefaultVariant must NOT use WithDB")
	assert.Equal(t, 1, runner.withTxCalls, "SetDefaultVariant must use exactly one WithTx")
}
