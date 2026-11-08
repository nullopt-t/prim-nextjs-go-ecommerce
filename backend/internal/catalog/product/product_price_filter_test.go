package product

import (
	"context"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type captureQE struct {
	queries []string
	args    [][]any
}

func (c *captureQE) ExecContext(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	c.queries = append(c.queries, sql)
	c.args = append(c.args, arguments)
	return pgconn.NewCommandTag("OK 1"), nil
}

type emptyRowsStub struct{}

func (emptyRowsStub) Close()                                       {}
func (emptyRowsStub) Err() error                                   { return nil }
func (emptyRowsStub) CommandTag() pgconn.CommandTag                { return pgconn.CommandTag{} }
func (emptyRowsStub) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (emptyRowsStub) Next() bool                                   { return false }
func (emptyRowsStub) Scan(...any) error                            { return nil }
func (emptyRowsStub) Values() ([]any, error)                       { return nil, nil }
func (emptyRowsStub) RawValues() [][]byte                          { return nil }
func (emptyRowsStub) Conn() *pgx.Conn                              { return nil }

func (c *captureQE) QueryContext(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	c.queries = append(c.queries, sql)
	c.args = append(c.args, args)
	return emptyRowsStub{}, nil
}

type rowStub struct{}

func (rowStub) Scan(dest ...any) error {
	if len(dest) > 0 {
		if p, ok := dest[0].(*int); ok {
			*p = 1
		}
	}
	return nil
}

func (c *captureQE) QueryRowContext(ctx context.Context, sql string, args ...any) pgx.Row {
	c.queries = append(c.queries, sql)
	c.args = append(c.args, args)
	return rowStub{}
}

func TestProductList_PriceFilterAndSortSQL(t *testing.T) {
	repo := NewProductRepository()
	qe := &captureQE{}

	minPrice := int64(1500)
	maxPrice := int64(9900)

	q := &ListProductsQuery{
		MinPrice: &minPrice,
		MaxPrice: &maxPrice,
	}
	q.RawSort = []string{"price"}

	_, err := repo.List(database.WithExecutor(context.Background(), qe), q, false)
	require.NoError(t, err)

	require.GreaterOrEqual(t, len(qe.queries), 2)
	countSQL := qe.queries[0]
	selectSQL := qe.queries[1]

	// Verify countQuery joins product_variants and checks min/max price
	assert.Contains(t, countSQL, "LEFT JOIN product_variants pv ON p.default_variant_id = pv.id")
	assert.Contains(t, countSQL, "COALESCE(pv.price, 0) >= $")
	assert.Contains(t, countSQL, "COALESCE(pv.price, 0) <= $")

	// Verify selectSQL joins product_variants and sorts by price ASC
	assert.Contains(t, selectSQL, "COALESCE(pv.price, 0) >= $")
	assert.Contains(t, selectSQL, "COALESCE(pv.price, 0) <= $")
	assert.Contains(t, selectSQL, "ORDER BY COALESCE(pv.price, 0) ASC")

	// Verify arguments contain min and max prices
	require.NotEmpty(t, qe.args)
	allArgs := qe.args[0]
	assert.Contains(t, allArgs, minPrice)
	assert.Contains(t, allArgs, maxPrice)
}

func TestProductList_PriceSortDesc(t *testing.T) {
	repo := NewProductRepository()
	qe := &captureQE{}

	q := &ListProductsQuery{}
	q.RawSort = []string{"-price"}

	_, err := repo.List(database.WithExecutor(context.Background(), qe), q, false)
	require.NoError(t, err)

	require.GreaterOrEqual(t, len(qe.queries), 2)
	selectSQL := qe.queries[1]
	assert.True(t, strings.Contains(selectSQL, "ORDER BY COALESCE(pv.price, 0) DESC"))
}

func TestProductList_AttributeFilterSQL(t *testing.T) {
	repo := NewProductRepository()
	qe := &captureQE{}

	q := &ListProductsQuery{
		Attribute: []string{"color:black,size:XL"},
	}
	q.Process(pagination.QueryOptions{})

	_, err := repo.List(database.WithExecutor(context.Background(), qe), q, false)
	require.NoError(t, err)

	require.GreaterOrEqual(t, len(qe.queries), 2)
	countSQL := qe.queries[0]
	selectSQL := qe.queries[1]

	expectedClause := "EXISTS (SELECT 1 FROM product_variants v_attr WHERE v_attr.product_id = p.id AND v_attr.deleted_at IS NULL AND v_attr.attributes->>$1 = $2 AND v_attr.attributes->>$3 = $4)"
	assert.Contains(t, countSQL, expectedClause)
	assert.Contains(t, selectSQL, expectedClause)

	allArgs := qe.args[0]
	assert.Contains(t, allArgs, "color")
	assert.Contains(t, allArgs, "black")
	assert.Contains(t, allArgs, "size")
	assert.Contains(t, allArgs, "XL")
}
