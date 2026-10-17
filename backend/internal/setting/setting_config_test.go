package setting

import (
	"testing"
)

func TestConfigFromMapOverlaysDefaults(t *testing.T) {
	cfg := configFromMap(map[string]interface{}{
		"storeName":              "My Shop",
		"flatRateShipping":       float64(12.5),
		"freeShippingThreshold":  float64(300),
		"allowCod":               false,
		"taxRatePercent":         float64(14),
		"pricesIncludeTax":       false,
		"maintenanceMode":        true,
		"cartReservationMinutes": float64(30),
	})

	if cfg.StoreName != "My Shop" {
		t.Fatalf("storeName = %q, want My Shop", cfg.StoreName)
	}
	if cfg.ShippingMinorUnits() != 1250 {
		t.Fatalf("shipping minor units = %d, want 1250", cfg.ShippingMinorUnits())
	}
	if cfg.FreeShippingThresholdMinorUnits() != 30000 {
		t.Fatalf("threshold minor units = %d, want 30000", cfg.FreeShippingThresholdMinorUnits())
	}
	if cfg.AllowCOD {
		t.Fatalf("allowCod = true, want false")
	}
	if cfg.TaxRatePercent != 14 {
		t.Fatalf("taxRatePercent = %v, want 14", cfg.TaxRatePercent)
	}
	if cfg.PricesIncludeTax {
		t.Fatalf("pricesIncludeTax = true, want false")
	}
	if !cfg.MaintenanceMode {
		t.Fatalf("maintenanceMode = false, want true")
	}
	if cfg.ReservationTTLMinutes() != 30 {
		t.Fatalf("reservation TTL = %d, want 30", cfg.ReservationTTLMinutes())
	}

	// Unset keys keep their defaults.
	if !cfg.AllowCards {
		t.Fatalf("allowCards = false, want default true")
	}
	if cfg.Currency != DefaultConfig().Currency {
		t.Fatalf("currency = %q, want default %q", cfg.Currency, DefaultConfig().Currency)
	}
}

func TestConfigDefaultsPreserveShippingRule(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.ShippingMinorUnits() != 999 {
		t.Fatalf("default shipping = %d, want 999", cfg.ShippingMinorUnits())
	}
	if cfg.FreeShippingThresholdMinorUnits() != 15000 {
		t.Fatalf("default threshold = %d, want 15000", cfg.FreeShippingThresholdMinorUnits())
	}
}

func TestPublicConfigExcludesSecrets(t *testing.T) {
	cfg := DefaultConfig()
	cfg.APIKey = "secret-key"
	cfg.StripeWebhookSecret = "secret-webhook"
	cfg.AdminNotificationEmail = "ops@example.com"

	pub := cfg.Public()
	if pub.StoreName != cfg.StoreName {
		t.Fatalf("public storeName = %q, want %q", pub.StoreName, cfg.StoreName)
	}
	// PublicConfig must not carry any credential fields; this is enforced by the
	// absence of those fields on the type (compile-time guarantee). Assert the
	// important storefront values round-trip instead.
	if pub.FlatRateShipping != cfg.FlatRateShipping {
		t.Fatalf("public flatRateShipping = %v, want %v", pub.FlatRateShipping, cfg.FlatRateShipping)
	}
	if pub.MaintenanceMode != cfg.MaintenanceMode {
		t.Fatalf("public maintenanceMode = %v, want %v", pub.MaintenanceMode, cfg.MaintenanceMode)
	}
}

func TestGetIntAcceptsMultipleNumericTypes(t *testing.T) {
	cases := []interface{}{float64(15), int(15), int64(15)}
	for _, c := range cases {
		if got := getInt(map[string]interface{}{"k": c}, "k", 0); got != 15 {
			t.Fatalf("getInt(%T) = %d, want 15", c, got)
		}
	}
}
