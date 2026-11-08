package product

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Regression: ?attribute=color:red&attribute=size:xl must not drop the first
// value (previous *string binding kept only one). Legacy comma-joined form
// and duplicate-key (last wins) behaviour are pinned too.
func TestParseAttributeFilters_RepeatedParams(t *testing.T) {
	got := parseAttributeFilters([]string{"color:red", "size:xl"})
	assert.Equal(t, map[string]string{"color": "red", "size": "xl"}, got)

	got = parseAttributeFilters([]string{"color:red,size:xl"})
	assert.Equal(t, map[string]string{"color": "red", "size": "xl"}, got)

	// Duplicate keys: last wins (single value per key by design).
	got = parseAttributeFilters([]string{"color:red", "color:blue"})
	assert.Equal(t, map[string]string{"color": "blue"}, got)

	// Value may itself contain a colon; only the first colon splits.
	got = parseAttributeFilters([]string{"time:12:30"})
	assert.Equal(t, map[string]string{"time": "12:30"}, got)

	// Malformed segments are ignored, never injected into SQL (parameters only).
	assert.Nil(t, parseAttributeFilters(nil))
	assert.Nil(t, parseAttributeFilters([]string{"  ", "nocolon", ":novalue", "nokey:"}))
}

func TestListProductsQuery_RepeatedAttributeParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, err := http.NewRequest(http.MethodGet, "/api/v1/products?attribute=color:red&attribute=size:xl", nil)
	require.NoError(t, err)
	c.Request = req

	var q ListProductsQuery
	require.NoError(t, c.ShouldBindQuery(&q))
	q.Process(pagination.QueryOptions{})

	assert.Equal(t, []string{"color:red", "size:xl"}, q.Attribute)
	assert.Equal(t, map[string]string{"color": "red", "size": "xl"}, q.AttributeFilters)
}
