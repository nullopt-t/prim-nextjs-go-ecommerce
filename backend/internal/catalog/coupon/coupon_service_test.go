package coupon_test

import (
	"context"
	"errors"
	"testing"

	"github.com/nullopt-t/prim-backend/internal/catalog/coupon"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateCouponValidation(t *testing.T) {
	svc := coupon.NewService(nil, nil)
	ctx := context.Background()

	t.Run("empty code fails validation", func(t *testing.T) {
		_, err := svc.CreateCoupon(ctx, coupon.CreateCouponInput{
			Code:          "",
			DiscountType:  model.CouponDiscountTypePercentage,
			DiscountValue: 20,
		})
		require.Error(t, err)
		var apiErr *apierr.APIError
		require.True(t, errors.As(err, &apiErr))
		assert.Equal(t, apierr.CodeValidationFailed, apiErr.Code)
	})

	t.Run("invalid discount type fails validation", func(t *testing.T) {
		_, err := svc.CreateCoupon(ctx, coupon.CreateCouponInput{
			Code:          "TEST",
			DiscountType:  "bogus",
			DiscountValue: 20,
		})
		require.Error(t, err)
		var apiErr *apierr.APIError
		require.True(t, errors.As(err, &apiErr))
		assert.Equal(t, apierr.CodeValidationFailed, apiErr.Code)
	})

	t.Run("invalid percentage values fail validation", func(t *testing.T) {
		for _, val := range []int64{0, -5, 101, 200} {
			_, err := svc.CreateCoupon(ctx, coupon.CreateCouponInput{
				Code:          "TEST",
				DiscountType:  model.CouponDiscountTypePercentage,
				DiscountValue: val,
			})
			require.Error(t, err)
			var apiErr *apierr.APIError
			require.True(t, errors.As(err, &apiErr))
			assert.Equal(t, apierr.CodeValidationFailed, apiErr.Code)
		}
	})

	t.Run("invalid fixed amount value fails validation", func(t *testing.T) {
		for _, val := range []int64{0, -100} {
			_, err := svc.CreateCoupon(ctx, coupon.CreateCouponInput{
				Code:          "TEST",
				DiscountType:  model.CouponDiscountTypeFixedAmount,
				DiscountValue: val,
			})
			require.Error(t, err)
			var apiErr *apierr.APIError
			require.True(t, errors.As(err, &apiErr))
			assert.Equal(t, apierr.CodeValidationFailed, apiErr.Code)
		}
	})

	t.Run("invalid status update fails", func(t *testing.T) {
		err := svc.UpdateStatus(ctx, [16]byte{}, "invalid_status")
		require.Error(t, err)
		var apiErr *apierr.APIError
		require.True(t, errors.As(err, &apiErr))
		assert.Equal(t, apierr.CodeValidationFailed, apiErr.Code)
	})

	t.Run("empty code in ValidateAndCalculate fails", func(t *testing.T) {
		_, err := svc.ValidateAndCalculate(ctx, "   ", 1000, nil)
		require.Error(t, err)
		var apiErr *apierr.APIError
		require.True(t, errors.As(err, &apiErr))
		assert.Equal(t, apierr.CodeValidationFailed, apiErr.Code)
	})
}
