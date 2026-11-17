package checkout_test

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/catalog/inventory"
	"github.com/nullopt-t/prim-backend/internal/catalog/product"
	"github.com/nullopt-t/prim-backend/internal/checkout"
)

func shippingInt64Ptr(v int64) *int64 { return &v }

// seedShippingSeller creates an active vendor with explicit per-vendor shipping
// terms. A nil threshold means the vendor never ships free.
func (s *CheckoutTestSuite) seedShippingSeller(fee int64, threshold *int64, leadDays int) uuid.UUID {
	ctx := context.Background()
	userID := uuid.New()
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO users (id, phone, full_name, role, phone_verified, status, country_code)
		VALUES ($1, $2, $3, 'seller', true, 'active', '+20')
	`, userID, "vendor_"+uuid.NewString()+"@example.com", "Vendor User")
	s.Require().NoError(err)

	sellerID := uuid.New()
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO sellers (
			id, user_id, store_name, slug, commission_rate_bps, status,
			shipping_fee, free_shipping_threshold, shipping_lead_time_days
		)
		VALUES ($1, $2, $3, $4, 1000, 'active', $5, $6, $7)
	`, sellerID, userID, "Vendor Store "+sellerID.String()[:8], "vendor-"+sellerID.String(), fee, threshold, leadDays)
	s.Require().NoError(err)
	return sellerID
}

// seedVendorProduct creates a published product owned by one seller.
func (s *CheckoutTestSuite) seedVendorProduct(sellerID uuid.UUID, title string) uuid.UUID {
	productID := uuid.New()
	_, err := s.db.ExecContext(context.Background(), `
		INSERT INTO products (id, category_id, slug, title, status, product_type, variation_axes, seller_id)
		VALUES ($1, $2, $3, $4, 'published', 'variable', ARRAY['color']::text[], $5)
	`, productID, s.categoryID, strings.ToLower(title)+"-"+uuid.NewString(), title, sellerID)
	s.Require().NoError(err)
	return productID
}

// seedProductVariant creates a priced, in-stock variant on an arbitrary product.
func (s *CheckoutTestSuite) seedProductVariant(productID uuid.UUID, price int64, stock int) uuid.UUID {
	p := price
	currency := defaultCurrency
	sku := "SKU-" + uuid.NewString()
	v, err := s.variantService.CreateVariant(context.Background(), &product.CreateVariantInput{
		ProductID:  productID,
		SKU:        &sku,
		Title:      "Shipping Widget",
		Price:      &p,
		Currency:   &currency,
		Attributes: map[string]any{"color": sku},
	})
	s.Require().NoError(err)
	if stock > 0 {
		_, err = s.inventoryService.AdjustStock(context.Background(), inventory.AdjustStockInput{
			VariantID: v.ID,
			Quantity:  stock,
			Reason:    "restock",
		})
		s.Require().NoError(err)
	}
	return v.ID
}

// TestCheckout_MultiVendorShipping pins per-vendor shipping: one vendor crosses
// its own free-shipping threshold, another without a threshold pays its flat
// fee, and platform items keep the global store rule. The order's shipping is
// the sum of all three groups and the response exposes every group.
func (s *CheckoutTestSuite) TestCheckout_MultiVendorShipping() {
	freeVendor := s.seedShippingSeller(500, shippingInt64Ptr(10000), 2)
	flatVendor := s.seedShippingSeller(700, nil, 6)

	freeProduct := s.seedVendorProduct(freeVendor, "Free Ship Vendor Product")
	flatProduct := s.seedVendorProduct(flatVendor, "Flat Ship Vendor Product")
	freeVariant := s.seedProductVariant(freeProduct, 6000, 10)
	flatVariant := s.seedProductVariant(flatProduct, 3000, 10)
	platformVariant := s.seedPricedVariant(1000, 10)

	buyerID, sessionID := s.seedUser()
	addressID := s.seedAddress(buyerID)
	s.addToCart(buyerID, freeVariant, 2)     // 12000 >= 10000 -> free
	s.addToCart(buyerID, flatVariant, 1)     // 3000, no threshold -> 700
	s.addToCart(buyerID, platformVariant, 1) // 1000 < 15000 -> global 999

	w := s.postCheckout(sessionID, checkout.CheckoutRequest{
		ShippingAddressID: addressID,
		PaymentMethod:     "cash_on_delivery",
	})
	s.Require().Equal(http.StatusOK, w.Code, w.Body.String())

	var resp dataEnvelope[checkout.CheckoutResponse]
	s.decodeSuccess(w, &resp)
	s.Require().Len(resp.Data.ShippingBreakdown, 3)

	bySeller := make(map[string]checkout.ShippingBreakdown, 3)
	for _, group := range resp.Data.ShippingBreakdown {
		key := "platform"
		if group.SellerID != nil {
			key = group.SellerID.String()
		}
		bySeller[key] = group
	}

	freeGroup, ok := bySeller[freeVendor.String()]
	s.Require().True(ok)
	s.Equal(int64(12000), freeGroup.Subtotal)
	s.Equal(int64(0), freeGroup.ShippingFee)
	s.True(freeGroup.IsFree)
	s.Equal(2, freeGroup.LeadTimeDays)
	s.NotEmpty(freeGroup.SellerName)

	flatGroup, ok := bySeller[flatVendor.String()]
	s.Require().True(ok)
	s.Equal(int64(3000), flatGroup.Subtotal)
	s.Equal(int64(700), flatGroup.ShippingFee)
	s.False(flatGroup.IsFree)
	s.Equal(6, flatGroup.LeadTimeDays)

	platformGroup, ok := bySeller["platform"]
	s.Require().True(ok)
	s.Nil(platformGroup.SellerID)
	s.Equal(int64(1000), platformGroup.Subtotal)
	s.Equal(shippingFee, platformGroup.ShippingFee)
	s.False(platformGroup.IsFree)

	orderID, err := uuid.Parse(resp.Data.OrderID)
	s.Require().NoError(err)
	order := s.fetchOrder(orderID)
	s.Equal(shippingFee+700, order.ShippingAmount, "order shipping is the sum of all vendor groups")
	s.Equal(int64(12000+3000+1000)+order.ShippingAmount, order.TotalAmount)
}
