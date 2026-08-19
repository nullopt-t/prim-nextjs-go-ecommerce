package model_test

import (
	"testing"

	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestParseInventoryReason(t *testing.T) {
	tests := []struct {
		input    string
		expected model.InventoryReason
		hasError bool
	}{
		{"restock", model.InventoryReasonRestock, false},
		{"sale", model.InventoryReasonSale, false},
		{"return", model.InventoryReasonReturn, false},
		{"adjustment", model.InventoryReasonAdjustment, false},
		{"reservation_release", model.InventoryReasonReservationRelease, false},
		{"invalid", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			res, err := model.ParseInventoryReason(tt.input)
			if tt.hasError {
				assert.Error(t, err)
				assert.Equal(t, model.ErrInvalidInventoryReason, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, res)
				assert.Equal(t, tt.input, res.String())
			}
		})
	}
}
