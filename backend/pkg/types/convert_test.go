package types_test

import (
	"testing"

	"github.com/nullopt-t/prim-backend/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestBoolFromPtr(t *testing.T) {
	assert.False(t, types.BoolFromPtr(nil))

	trueVal := true
	assert.True(t, types.BoolFromPtr(&trueVal))

	falseVal := false
	assert.False(t, types.BoolFromPtr(&falseVal), "BoolFromPtr should return false for a pointer to false")
}

func TestBoolValueFromPtr(t *testing.T) {
	assert.True(t, types.BoolValueFromPtr(nil, true))
	assert.False(t, types.BoolValueFromPtr(nil, false))

	trueVal := true
	assert.True(t, types.BoolValueFromPtr(&trueVal, false))

	falseVal := false
	assert.False(t, types.BoolValueFromPtr(&falseVal, true))
}

func TestHasValue(t *testing.T) {
	var p *int
	assert.False(t, types.HasValue(p))

	v := 42
	assert.True(t, types.HasValue(&v))
}
