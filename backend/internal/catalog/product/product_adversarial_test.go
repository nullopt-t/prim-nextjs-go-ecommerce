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

func (m *mockQE) ExecContext(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	if m.execFn != nil {
		return m.execFn(ctx, sql, arguments...)
	}
	return pgconn.NewCommandTag("UPDATE 1"), nil
}

func (m *mockQE) QueryContext(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return nil, pgx.ErrNoRows
}

func (m *mockQE) QueryRowContext(ctx context.Context, sql string, args ...any) pgx.Row {
	return nil
}

type mockQETx struct {
	*mockQE
}

func (m *mockQETx) Begin(context.Context) (pgx.Tx, error) { return m, nil }
func (m *mockQETx) Commit(context.Context) error          { return nil }
func (m *mockQETx) Rollback(context.Context) error        { return nil }
func (m *mockQETx) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return m.ExecContext(ctx, sql, args...)
}
func (m *mockQETx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return m.QueryContext(ctx, sql, args...)
}
func (m *mockQETx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return m.QueryRowContext(ctx, sql, args...)
}
func (m *mockQETx) CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error) {
	return 0, nil
}
func (m *mockQETx) SendBatch(context.Context, *pgx.Batch) pgx.BatchResults { return nil }
func (m *mockQETx) Prepare(context.Context, string, string) (*pgconn.StatementDescription, error) {
	return nil, nil
}
func (m *mockQETx) Conn() *pgx.Conn                { return nil }
func (m *mockQETx) LargeObjects() pgx.LargeObjects { return pgx.LargeObjects{} }

type mockConnectionTrackingRunner struct {
	withDBCalls int
	withTxCalls int
	qe          *mockQE
}

func (m *mockConnectionTrackingRunner) WithExecutor(ctx context.Context, fn func(context.Context) error) error {
	m.withDBCalls++
	ctx = database.WithExecutor(ctx, m.qe)
	return fn(ctx)
}

func (m *mockConnectionTrackingRunner) WithTx(ctx context.Context, fn func(context.Context) error) error {
	m.withTxCalls++
	ctx = database.WithTx(ctx, database.NewTx(&mockQETx{mockQE: m.qe}))
	return fn(ctx)
}

func (m *mockConnectionTrackingRunner) WithRetry(ctx context.Context, fn func(context.Context) error, n int) error {
	return m.WithTx(ctx, fn)
}

type fakeVariantRepoForInvariants struct {
	countFunc func() (int, error)
	getFunc   func(id uuid.UUID) (*model.ProductVariant, error)
	listFunc  func() ([]*model.ProductVariant, error)
}

func (f *fakeVariantRepoForInvariants) CountByProductID(ctx context.Context, productID uuid.UUID) (int, error) {
	if f.countFunc != nil {
		return f.countFunc()
	}
	return 1, nil
}

func (f *fakeVariantRepoForInvariants) GetByID(ctx context.Context, id uuid.UUID) (*model.ProductVariant, error) {
	if f.getFunc != nil {
		return f.getFunc(id)
	}
	return nil, nil
}

func (f *fakeVariantRepoForInvariants) ListByProductID(ctx context.Context, opts variant.ListVariantOptions) (*pagination.PagedResult[model.ProductVariant], error) {
	if f.listFunc != nil {
		items, err := f.listFunc()
		if err != nil {
			return nil, err
		}
		return pagination.NewPagedResult(items, pagination.NewPage(1, 100, len(items))), nil
	}
	return &pagination.PagedResult[model.ProductVariant]{Items: nil}, nil
}

func (f *fakeVariantRepoForInvariants) ListMediaByVariantIDs(ctx context.Context, variantIDs []uuid.UUID) (map[uuid.UUID][]*model.VariantMedia, error) {
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
		listFunc: func() ([]*model.ProductVariant, error) {
			thumbID := uuid.New()
			return []*model.ProductVariant{{ID: uuid.New(), ThumbnailObjectID: &thumbID}}, nil
		},
	}

	svc := &ProductService{
		dbRunner:      runner,
		variantReader: vRepo,
		productRepo:   NewProductRepository(),
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
		dbRunner:      runner,
		variantReader: vRepo,
		productRepo:   NewProductRepository(),
	}

	// This call will fail inside WithTx at productRepo.GetByID (returning ErrNoRows from our mock query)
	// which tests that WithTx was entered, executed the query on tx, and rolled back/returned cleanly.
	err := svc.SetDefaultVariant(context.Background(), prodID, varID)
	require.Error(t, err)

	// Invariant check: Exactly one WithTx call, zero WithExecutor calls
	assert.Equal(t, 0, runner.withDBCalls, "SetDefaultVariant must NOT use WithExecutor")
	assert.Equal(t, 1, runner.withTxCalls, "SetDefaultVariant must use exactly one WithTx")
}
