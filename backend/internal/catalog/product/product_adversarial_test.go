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

type mockQETx struct {
	*mockQE
}

func (m *mockQETx) Begin(context.Context) (pgx.Tx, error) { return m, nil }
func (m *mockQETx) Commit(context.Context) error { return nil }
func (m *mockQETx) Rollback(context.Context) error { return nil }
func (m *mockQETx) CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error) {
	return 0, nil
}
func (m *mockQETx) SendBatch(context.Context, *pgx.Batch) pgx.BatchResults { return nil }
func (m *mockQETx) Prepare(context.Context, string, string) (*pgconn.StatementDescription, error) {
	return nil, nil
}
func (m *mockQETx) Conn() *pgx.Conn { return nil }
func (m *mockQETx) LargeObjects() pgx.LargeObjects { return pgx.LargeObjects{} }

type mockConnectionTrackingRunner struct {
	withDBCalls int
	withTxCalls int
	qe          *mockQE
}

func (m *mockConnectionTrackingRunner) WithExecutor(ctx context.Context, fn func(database.QueryExecutor) error) error {
	m.withDBCalls++
	return fn(m.qe)
}

func (m *mockConnectionTrackingRunner) WithTx(ctx context.Context, fn func(context.Context, *database.Tx) error) error {
	m.withTxCalls++
	return fn(ctx, database.NewTx(&mockQETx{mockQE: m.qe}))
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

func (f *fakeVariantRepoForInvariants) ListMediaByVariantIDs(ctx context.Context, qe database.QueryExecutor, variantIDs []uuid.UUID) (map[uuid.UUID][]*model.VariantMedia, error) {
	return nil, nil
}

func TestAdversarial_PublishProduct_SingleWithTx(t *testing.T) {
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

	// Invariant check: Exactly one WithTx call, zero WithExecutor calls
	assert.Equal(t, 0, runner.withDBCalls, "PublishProduct must NOT use WithExecutor")
	assert.Equal(t, 1, runner.withTxCalls, "PublishProduct must use exactly one WithTx")
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

	// Invariant check: Exactly one WithTx call, zero WithExecutor calls
	assert.Equal(t, 0, runner.withDBCalls, "SetDefaultVariant must NOT use WithExecutor")
	assert.Equal(t, 1, runner.withTxCalls, "SetDefaultVariant must use exactly one WithTx")
}
