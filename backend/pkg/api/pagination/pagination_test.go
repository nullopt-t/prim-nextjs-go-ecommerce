package pagination_test

import (
	"math"
	"testing"

	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/stretchr/testify/assert"
)

func TestNewPage(t *testing.T) {
	p := pagination.NewPage(1, 10, 25)
	assert.Equal(t, 1, p.Page)
	assert.Equal(t, 10, p.PageSize)
	assert.Equal(t, 25, p.TotalItems)
	assert.Equal(t, 3, p.TotalPages)
	assert.False(t, p.HasPrevious)
	assert.True(t, p.HasNext)

	// Boundary edge case: 0 total items
	p0 := pagination.NewPage(2, 10, 0)
	assert.Equal(t, 0, p0.TotalPages)
	assert.False(t, p0.HasPrevious, "HasPrevious should be false if totalItems == 0")
	assert.False(t, p0.HasNext)

	// Overflow prevention near math.MaxInt
	pOverflow := pagination.NewPage(1, 10, math.MaxInt-5)
	assert.Greater(t, pOverflow.TotalPages, 0)
}

func TestListQuery_Process(t *testing.T) {
	var nilQ *pagination.ListQuery
	assert.Nil(t, nilQ.Process(pagination.QueryOptions{}))

	q := &pagination.ListQuery{
		Page:     2,
		PageSize: 20,
	}
	q.Process(pagination.QueryOptions{})
	assert.Equal(t, 20, q.Offset)

	// Huge page integer overflow check
	qHuge := &pagination.ListQuery{
		Page:     math.MaxInt,
		PageSize: 100,
	}
	qHuge.Process(pagination.QueryOptions{})
	assert.GreaterOrEqual(t, qHuge.Offset, 0, "Offset must not be negative on large page")
}
