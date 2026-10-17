package setting

import (
	"encoding/json"
	"math"
)

// Config is the typed projection of the administrative settings that actually
// drive system behavior: checkout pricing, storefront identity, maintenance
// state, and system toggles. It is built by overlaying the persisted key/value
// settings on top of DefaultConfig so that unset keys keep sane defaults.
type Config struct {
	// Store identity & localization
	StoreName           string `json:"storeName"`
	StoreEmail          string `json:"storeEmail"`
	StorePhone          string `json:"storePhone"`
	StoreAddress        string `json:"storeAddress"`
	Currency            string `json:"currency"`
	Timezone            string `json:"timezone"`
	DefaultLocale       string `json:"defaultLocale"`
	AnnouncementEnabled bool   `json:"announcementEnabled"`
	AnnouncementText    string `json:"announcementText"`

	// Shipping & delivery (amounts are in major currency units, e.g. EGP)
	FlatRateShipping      float64 `json:"flatRateShipping"`
	FreeShippingThreshold float64 `json:"freeShippingThreshold"`
	EstimatedDeliveryDays string  `json:"estimatedDeliveryDays"`

	// Payments & taxes (amounts are in major currency units)
	AllowCOD               bool    `json:"allowCod"`
	CODFee                 float64 `json:"codFee"`
	AllowCards             bool    `json:"allowCards"`
	TaxRatePercent         float64 `json:"taxRatePercent"`
	PricesIncludeTax       bool    `json:"pricesIncludeTax"`
	CartReservationMinutes int     `json:"cartReservationMinutes"`

	// Notifications
	NotifyOrderConfirmation bool   `json:"notifyOrderConfirmation"`
	NotifyOrderShipped      bool   `json:"notifyOrderShipped"`
	AdminNotificationEmail  string `json:"adminNotificationEmail"`
	NotifyLowStock          bool   `json:"notifyLowStock"`

	// Security & integrations
	APIKey              string `json:"apiKey"`
	StripeWebhookSecret string `json:"stripeWebhookSecret"`

	// System toggles
	DisableCache           bool `json:"disableCache"`
	EnableDevOtpBypass     bool `json:"enableDevOtpBypass"`
	MockEmailNotifications bool `json:"mockEmailNotifications"`

	// Maintenance
	MaintenanceMode    bool   `json:"maintenanceMode"`
	MaintenanceMessage string `json:"maintenanceMessage"`
}

// DefaultConfig returns the effective settings used when no override is stored.
// Shipping defaults mirror the historical hardcoded checkout rule (9.99 flat,
// free at/above 150.00) so existing behavior is preserved until an admin changes
// the values from the settings page.
func DefaultConfig() Config {
	return Config{
		StoreName:           "PRIM Superstore",
		StoreEmail:          "support@prim.com",
		StorePhone:          "+20 100 000 0000",
		StoreAddress:        "Smart Village, Building B12, Giza, Egypt",
		Currency:            "EGP",
		Timezone:            "Africa/Cairo",
		DefaultLocale:       "ar",
		AnnouncementEnabled: true,
		AnnouncementText:    "Welcome to PRIM! Enjoy free delivery on qualifying orders.",

		FlatRateShipping:      9.99,
		FreeShippingThreshold: 150,
		EstimatedDeliveryDays: "2 - 4 business days",

		AllowCOD:               true,
		CODFee:                 0,
		AllowCards:             true,
		TaxRatePercent:         0,
		PricesIncludeTax:       true,
		CartReservationMinutes: 15,

		NotifyOrderConfirmation: true,
		NotifyOrderShipped:      true,
		AdminNotificationEmail:  "operations@prim.com",
		NotifyLowStock:          true,

		DisableCache:           false,
		EnableDevOtpBypass:     false,
		MockEmailNotifications: false,

		MaintenanceMode:    false,
		MaintenanceMessage: "PRIM is currently undergoing scheduled maintenance to upgrade our warehouse systems. We will be back shortly!",
	}
}

// ShippingMinorUnits converts the flat shipping rate to minor currency units.
func (c Config) ShippingMinorUnits() int64 { return toMinorUnits(c.FlatRateShipping) }

// FreeShippingThresholdMinorUnits converts the free-shipping threshold to minor currency units.
func (c Config) FreeShippingThresholdMinorUnits() int64 { return toMinorUnits(c.FreeShippingThreshold) }

// CODFeeMinorUnits converts the cash-on-delivery fee to minor currency units.
func (c Config) CODFeeMinorUnits() int64 { return toMinorUnits(c.CODFee) }

// ReservationTTL returns the configured cart reservation duration in minutes,
// falling back to 15 minutes when unset.
func (c Config) ReservationTTLMinutes() int {
	if c.CartReservationMinutes <= 0 {
		return 15
	}
	return c.CartReservationMinutes
}

// PublicConfig is the non-sensitive subset of the settings safe to expose to the
// unauthenticated storefront (no API keys, webhook secrets, or internal alerts).
type PublicConfig struct {
	StoreName             string  `json:"storeName"`
	StoreEmail            string  `json:"storeEmail"`
	StorePhone            string  `json:"storePhone"`
	StoreAddress          string  `json:"storeAddress"`
	Currency              string  `json:"currency"`
	Timezone              string  `json:"timezone"`
	DefaultLocale         string  `json:"defaultLocale"`
	AnnouncementEnabled   bool    `json:"announcementEnabled"`
	AnnouncementText      string  `json:"announcementText"`
	FlatRateShipping      float64 `json:"flatRateShipping"`
	FreeShippingThreshold float64 `json:"freeShippingThreshold"`
	EstimatedDeliveryDays string  `json:"estimatedDeliveryDays"`
	AllowCOD              bool    `json:"allowCod"`
	CODFee                float64 `json:"codFee"`
	AllowCards            bool    `json:"allowCards"`
	TaxRatePercent        float64 `json:"taxRatePercent"`
	PricesIncludeTax      bool    `json:"pricesIncludeTax"`
	MaintenanceMode       bool    `json:"maintenanceMode"`
	MaintenanceMessage    string  `json:"maintenanceMessage"`
}

// Public projects the config down to the storefront-safe subset.
func (c Config) Public() PublicConfig {
	return PublicConfig{
		StoreName:             c.StoreName,
		StoreEmail:            c.StoreEmail,
		StorePhone:            c.StorePhone,
		StoreAddress:          c.StoreAddress,
		Currency:              c.Currency,
		Timezone:              c.Timezone,
		DefaultLocale:         c.DefaultLocale,
		AnnouncementEnabled:   c.AnnouncementEnabled,
		AnnouncementText:      c.AnnouncementText,
		FlatRateShipping:      c.FlatRateShipping,
		FreeShippingThreshold: c.FreeShippingThreshold,
		EstimatedDeliveryDays: c.EstimatedDeliveryDays,
		AllowCOD:              c.AllowCOD,
		CODFee:                c.CODFee,
		AllowCards:            c.AllowCards,
		TaxRatePercent:        c.TaxRatePercent,
		PricesIncludeTax:      c.PricesIncludeTax,
		MaintenanceMode:       c.MaintenanceMode,
		MaintenanceMessage:    c.MaintenanceMessage,
	}
}

// configFromMap overlays persisted key/value settings onto the defaults.
func configFromMap(m map[string]interface{}) Config {
	c := DefaultConfig()

	c.StoreName = getString(m, "storeName", c.StoreName)
	c.StoreEmail = getString(m, "storeEmail", c.StoreEmail)
	c.StorePhone = getString(m, "storePhone", c.StorePhone)
	c.StoreAddress = getString(m, "storeAddress", c.StoreAddress)
	c.Currency = getString(m, "currency", c.Currency)
	c.Timezone = getString(m, "timezone", c.Timezone)
	c.DefaultLocale = getString(m, "defaultLocale", c.DefaultLocale)
	c.AnnouncementEnabled = getBool(m, "announcementEnabled", c.AnnouncementEnabled)
	c.AnnouncementText = getString(m, "announcementText", c.AnnouncementText)

	c.FlatRateShipping = getFloat(m, "flatRateShipping", c.FlatRateShipping)
	c.FreeShippingThreshold = getFloat(m, "freeShippingThreshold", c.FreeShippingThreshold)
	c.EstimatedDeliveryDays = getString(m, "estimatedDeliveryDays", c.EstimatedDeliveryDays)

	c.AllowCOD = getBool(m, "allowCod", c.AllowCOD)
	c.CODFee = getFloat(m, "codFee", c.CODFee)
	c.AllowCards = getBool(m, "allowCards", c.AllowCards)
	c.TaxRatePercent = getFloat(m, "taxRatePercent", c.TaxRatePercent)
	c.PricesIncludeTax = getBool(m, "pricesIncludeTax", c.PricesIncludeTax)
	c.CartReservationMinutes = getInt(m, "cartReservationMinutes", c.CartReservationMinutes)

	c.NotifyOrderConfirmation = getBool(m, "notifyOrderConfirmation", c.NotifyOrderConfirmation)
	c.NotifyOrderShipped = getBool(m, "notifyOrderShipped", c.NotifyOrderShipped)
	c.AdminNotificationEmail = getString(m, "adminNotificationEmail", c.AdminNotificationEmail)
	c.NotifyLowStock = getBool(m, "notifyLowStock", c.NotifyLowStock)

	c.APIKey = getString(m, "apiKey", c.APIKey)
	c.StripeWebhookSecret = getString(m, "stripeWebhookSecret", c.StripeWebhookSecret)

	c.DisableCache = getBool(m, "disableCache", c.DisableCache)
	c.EnableDevOtpBypass = getBool(m, "enableDevOtpBypass", c.EnableDevOtpBypass)
	c.MockEmailNotifications = getBool(m, "mockEmailNotifications", c.MockEmailNotifications)

	c.MaintenanceMode = getBool(m, "maintenanceMode", c.MaintenanceMode)
	c.MaintenanceMessage = getString(m, "maintenanceMessage", c.MaintenanceMessage)

	return c
}

// toMinorUnits converts a major-unit amount (e.g. 9.99 EGP) into minor units (999).
func toMinorUnits(v float64) int64 {
	return int64(math.Round(v * 100))
}

func getString(m map[string]interface{}, key, def string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return def
	}
	if s, ok := v.(string); ok {
		return s
	}
	return def
}

func getFloat(m map[string]interface{}, key string, def float64) float64 {
	v, ok := m[key]
	if !ok || v == nil {
		return def
	}
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case json.Number:
		if f, err := n.Float64(); err == nil {
			return f
		}
	}
	return def
}

func getBool(m map[string]interface{}, key string, def bool) bool {
	v, ok := m[key]
	if !ok || v == nil {
		return def
	}
	if b, ok := v.(bool); ok {
		return b
	}
	return def
}

func getInt(m map[string]interface{}, key string, def int) int {
	v, ok := m[key]
	if !ok || v == nil {
		return def
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case float32:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	case json.Number:
		if i, err := n.Int64(); err == nil {
			return int(i)
		}
	}
	return def
}
