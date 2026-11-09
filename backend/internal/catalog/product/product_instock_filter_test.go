package product

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Regression: the storefront sends ?inStock=true but the backend silently
// dropped it (no binding), so the server returned out-of-stock products and
// the client had to hide them in memory with wrong totals.
func TestListProductsQuery_InStockBinding(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, err := http.NewRequest(http.MethodGet, "/api/v1/products?inStock=true", nil)
	require.NoError(t, err)
	c.Request = req

	var q ListProductsQuery
	require.NoError(t, c.ShouldBindQuery(&q))
	q.Process(pagination.QueryOptions{})

	require.NotNil(t, q.InStock)
	assert.True(t, *q.InStock)
}

func TestProductList_InStockFilterSQL(t *testing.T) {
	repo := NewProductRepository()
	qe := &captureQE{}

	inStock := true
	q := &ListProductsQuery{InStock: &inStock}

	_, err := repo.List(database.WithExecutor(context.Background(), qe), q, false)
	require.NoError(t, err)

	require.GreaterOrEqual(t, len(qe.queries), 2)
	assert.Contains(t, qe.queries[0], "inventory_ledgers")
	assert.Contains(t, qe.queries[0], "inventory_reservations")
	assert.Contains(t, qe.queries[1], "inventory_ledgers")
	assert.Contains(t, qe.queries[1], "inventory_reservations")
}
