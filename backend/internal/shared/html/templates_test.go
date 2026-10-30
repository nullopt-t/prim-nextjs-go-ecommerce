package html_test

import (
	"testing"

	"github.com/nullopt-t/prim-backend/internal/shared/html"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderer_BilingualTemplates(t *testing.T) {
	renderer, err := html.NewRenderer()
	require.NoError(t, err)

	t.Run("Render English OTP template", func(t *testing.T) {
		out, err := renderer.Render("email-otp.en", map[string]any{"Code": "123456"})
		require.NoError(t, err)
		assert.Contains(t, out, "123456")
		assert.Contains(t, out, "Verify Your Email")
		assert.Contains(t, out, `dir="ltr"`)
	})

	t.Run("Render Arabic OTP template", func(t *testing.T) {
		out, err := renderer.Render("email-otp.ar", map[string]any{"Code": "987654"})
		require.NoError(t, err)
		assert.Contains(t, out, "987654")
		assert.Contains(t, out, "تأكيد بريدك الإلكتروني")
		assert.Contains(t, out, `dir="rtl"`)
	})

	t.Run("Render English Welcome template", func(t *testing.T) {
		out, err := renderer.Render("welcome.en", map[string]any{"ClientURL": "https://prim.store"})
		require.NoError(t, err)
		assert.Contains(t, out, "https://prim.store")
		assert.Contains(t, out, "Welcome to Prim!")
	})

	t.Run("Render Arabic Welcome template", func(t *testing.T) {
		out, err := renderer.Render("welcome.ar", map[string]any{"ClientURL": "https://prim.store"})
		require.NoError(t, err)
		assert.Contains(t, out, "https://prim.store")
		assert.Contains(t, out, "مرحبًا بك في بريم!")
	})

	t.Run("Fallback to .en when template lacks locale suffix", func(t *testing.T) {
		out, err := renderer.Render("email-otp", map[string]any{"Code": "555555"})
		require.NoError(t, err)
		assert.Contains(t, out, "555555")
	})

	orderData := map[string]any{
		"CustomerName":    "Buyer Name",
		"OrderNumber":     "PRIM-1234",
		"PaymentMethod":   "cash_on_delivery",
		"ShippingAddress": "1 Main Street, Cairo, EG",
		"OrderURL":        "https://prim.store/user/orders/11111111-1111-1111-1111-111111111111",
		"Total":           "25.98 EGP",
		"Items": []map[string]any{{
			"Title":     "Wireless Headphones",
			"Quantity":  2,
			"UnitPrice": "12.99 EGP",
			"LineTotal": "25.98 EGP",
		}},
	}

	t.Run("Render English order confirmation template", func(t *testing.T) {
		out, err := renderer.Render("order-confirmation.en", orderData)
		require.NoError(t, err)
		assert.Contains(t, out, "PRIM-1234")
		assert.Contains(t, out, "Wireless Headphones")
		assert.Contains(t, out, "25.98 EGP")
		assert.Contains(t, out, "https://prim.store/user/orders/11111111-1111-1111-1111-111111111111")
		assert.Contains(t, out, `dir="ltr"`)
	})

	t.Run("Render Arabic order confirmation template", func(t *testing.T) {
		out, err := renderer.Render("order-confirmation.ar", orderData)
		require.NoError(t, err)
		assert.Contains(t, out, "PRIM-1234")
		assert.Contains(t, out, "Wireless Headphones")
		assert.Contains(t, out, "25.98 EGP")
		assert.Contains(t, out, `dir="rtl"`)
	})

	shippedData := map[string]any{
		"CustomerName":   "Buyer Name",
		"OrderNumber":    "PRIM-1234",
		"Carrier":        "Aramex",
		"TrackingNumber": "TRACK-999",
		"TrackingURL":    "https://carrier.example/track/TRACK-999",
		"OrderURL":       "https://prim.store/user/orders/11111111-1111-1111-1111-111111111111",
	}

	t.Run("Render English shipped template", func(t *testing.T) {
		out, err := renderer.Render("order-shipped.en", shippedData)
		require.NoError(t, err)
		assert.Contains(t, out, "PRIM-1234")
		assert.Contains(t, out, "Aramex")
		assert.Contains(t, out, "TRACK-999")
		assert.Contains(t, out, "https://carrier.example/track/TRACK-999")
		assert.Contains(t, out, `dir="ltr"`)
	})

	t.Run("Render Arabic shipped template", func(t *testing.T) {
		out, err := renderer.Render("order-shipped.ar", shippedData)
		require.NoError(t, err)
		assert.Contains(t, out, "PRIM-1234")
		assert.Contains(t, out, "TRACK-999")
		assert.Contains(t, out, `dir="rtl"`)
	})

	cancelledData := map[string]any{
		"CustomerName": "Buyer Name",
		"OrderNumber":  "PRIM-1234",
		"Reason":       "Customer changed their mind",
		"OrderURL":     "https://prim.store/user/orders/11111111-1111-1111-1111-111111111111",
	}

	t.Run("Render English cancelled template", func(t *testing.T) {
		out, err := renderer.Render("order-cancelled.en", cancelledData)
		require.NoError(t, err)
		assert.Contains(t, out, "PRIM-1234")
		assert.Contains(t, out, "Customer changed their mind")
		assert.Contains(t, out, `dir="ltr"`)
	})

	t.Run("Render Arabic cancelled template", func(t *testing.T) {
		out, err := renderer.Render("order-cancelled.ar", cancelledData)
		require.NoError(t, err)
		assert.Contains(t, out, "PRIM-1234")
		assert.Contains(t, out, "Customer changed their mind")
		assert.Contains(t, out, `dir="rtl"`)
	})
}
