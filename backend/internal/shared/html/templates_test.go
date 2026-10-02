package html_test

import (
	"testing"

	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/html"
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
}
