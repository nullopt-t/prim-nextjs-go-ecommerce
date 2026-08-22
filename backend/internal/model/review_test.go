package model_test

import (
	"testing"

	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestParseReviewStatus(t *testing.T) {
	tests := []struct {
		input    string
		expected model.ReviewStatus
		hasError bool
	}{
		{"pending", model.ReviewStatusPending, false},
		{"approved", model.ReviewStatusApproved, false},
		{"rejected", model.ReviewStatusRejected, false},
		{"invalid", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			res, err := model.ParseReviewStatus(tt.input)
			if tt.hasError {
				assert.Error(t, err)
				assert.Equal(t, model.ErrInvalidReviewStatus, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, res)
				assert.Equal(t, tt.input, res.String())
			}
		})
	}
}
