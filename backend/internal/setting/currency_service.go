package setting

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

const (
	SettingKeyCurrencyRates = "currency_rates"
	currencyCacheTTL        = 10 * time.Second
)

// CurrencyInfo represents a supported currency, its display metadata, and
// conversion rates relative to the base anchor currency.
type CurrencyInfo struct {
	Code       string  `json:"code"`
	Name       string  `json:"name"`
	Symbol     string  `json:"symbol"`
	Flag       string  `json:"flag"`
	Decimals   int     `json:"decimals"`
	RateToEGP  float64 `json:"rateToEGP"`
	RateToBase float64 `json:"rateToBase"`
	IsBase     bool    `json:"isBase"`
	Enabled    bool    `json:"enabled"`
}

// DefaultCurrencies provides default static currency definitions and baseline
// exchange rates relative to the base anchor currency (EGP = 1.0).
var DefaultCurrencies = []CurrencyInfo{
	{Code: "EGP", Name: "Egyptian Pound", Symbol: "E£", Flag: "🇪🇬", Decimals: 2, RateToEGP: 1.0, RateToBase: 1.0, IsBase: true, Enabled: true},
	{Code: "USD", Name: "US Dollar", Symbol: "$", Flag: "🇺🇸", Decimals: 2, RateToEGP: 48.5, RateToBase: 48.5, IsBase: false, Enabled: true},
	{Code: "EUR", Name: "Euro", Symbol: "€", Flag: "🇪🇺", Decimals: 2, RateToEGP: 52.0, RateToBase: 52.0, IsBase: false, Enabled: true},
	{Code: "SAR", Name: "Saudi Riyal", Symbol: "ر.س", Flag: "🇸🇦", Decimals: 2, RateToEGP: 12.9, RateToBase: 12.9, IsBase: false, Enabled: true},
	{Code: "AED", Name: "UAE Dirham", Symbol: "د.إ", Flag: "🇦🇪", Decimals: 2, RateToEGP: 13.2, RateToBase: 13.2, IsBase: false, Enabled: true},
}

// CurrencyService coordinates currency metadata and exchange rate operations.
type CurrencyService struct {
	dr   database.Runner
	repo *Repository

	mu       sync.RWMutex
	cached   []CurrencyInfo
	cachedAt time.Time
}

// NewCurrencyService constructs a CurrencyService bound to a database runner and repository.
func NewCurrencyService(dr database.Runner, repo *Repository) *CurrencyService {
	return &CurrencyService{
		dr:   dr,
		repo: repo,
	}
}

// GetCurrencies returns supported currencies with live exchange rates overlaid on defaults.
func (s *CurrencyService) GetCurrencies(ctx context.Context) ([]CurrencyInfo, error) {
	if cached, ok := s.getCached(); ok {
		return cached, nil
	}

	currencies := make([]CurrencyInfo, len(DefaultCurrencies))
	copy(currencies, DefaultCurrencies)

	if s.dr == nil || s.repo == nil {
		return currencies, nil
	}

	var customRates map[string]float64
	err := s.dr.WithExecutor(ctx, func(ctx context.Context) error {
		items, err := s.repo.GetAll(ctx)
		if err != nil {
			return err
		}
		for _, item := range items {
			if item.Key == SettingKeyCurrencyRates {
				rawBytes, err := json.Marshal(item.Value)
				if err != nil {
					return nil
				}
				_ = json.Unmarshal(rawBytes, &customRates)
				break
			}
		}
		return nil
	})
	if err != nil {
		// Log or fall back safely to defaults
		return currencies, nil
	}

	if len(customRates) > 0 {
		for i := range currencies {
			code := currencies[i].Code
			if rate, ok := customRates[code]; ok && rate > 0 {
				currencies[i].RateToEGP = rate
				currencies[i].RateToBase = rate
			}
		}
	}

	s.setCached(currencies)
	return currencies, nil
}

// Convert converts an amount from one currency to another using the current exchange rates.
func (s *CurrencyService) Convert(ctx context.Context, amount float64, fromCode, toCode string) (float64, error) {
	from := strings.ToUpper(strings.TrimSpace(fromCode))
	to := strings.ToUpper(strings.TrimSpace(toCode))

	if from == "" || to == "" {
		return 0, apierr.BadRequestError("Both from and to currency codes are required")
	}
	if from == to {
		return amount, nil
	}

	currencies, err := s.GetCurrencies(ctx)
	if err != nil {
		return 0, fmt.Errorf("fetch currencies for conversion: %w", err)
	}

	var fromRate, toRate float64
	var foundFrom, foundTo bool

	for _, c := range currencies {
		if c.Code == from {
			fromRate = c.RateToBase
			foundFrom = true
		}
		if c.Code == to {
			toRate = c.RateToBase
			foundTo = true
		}
	}

	if !foundFrom {
		return 0, apierr.BadRequestError(fmt.Sprintf("Unsupported source currency: %s", from))
	}
	if !foundTo {
		return 0, apierr.BadRequestError(fmt.Sprintf("Unsupported target currency: %s", to))
	}
	if toRate <= 0 {
		return 0, apierr.InternalServerError("Target currency has invalid zero exchange rate")
	}

	amountInBase := amount * fromRate
	converted := amountInBase / toRate
	return converted, nil
}

// UpdateRates persists updated exchange rates in the admin settings store and invalidates cache.
func (s *CurrencyService) UpdateRates(ctx context.Context, rates map[string]float64) error {
	if len(rates) == 0 {
		return apierr.BadRequestError("Rates payload cannot be empty")
	}

	// Validate rates are positive
	sanitized := make(map[string]float64, len(rates))
	for code, rate := range rates {
		upper := strings.ToUpper(strings.TrimSpace(code))
		if rate <= 0 {
			return apierr.BadRequestError(fmt.Sprintf("Rate for %s must be greater than zero", upper))
		}
		sanitized[upper] = rate
	}

	if s.dr == nil || s.repo == nil {
		return apierr.InternalServerError("Persistence runner unavailable")
	}

	err := s.dr.WithTx(ctx, func(ctx context.Context) error {
		// Fetch existing rates if any, to merge
		var existing map[string]float64
		items, err := s.repo.GetAll(ctx)
		if err == nil {
			for _, item := range items {
				if item.Key == SettingKeyCurrencyRates {
					if rawBytes, mErr := json.Marshal(item.Value); mErr == nil {
						_ = json.Unmarshal(rawBytes, &existing)
					}
					break
				}
			}
		}

		if existing == nil {
			existing = make(map[string]float64)
		}
		for code, rate := range sanitized {
			existing[code] = rate
		}

		return s.repo.Upsert(ctx, SettingKeyCurrencyRates, existing)
	})
	if err != nil {
		return apierr.InternalServerError("Failed to update currency exchange rates").Wrap(err)
	}

	s.Invalidate()
	return nil
}

// Invalidate clears the in-memory cache of currency rates.
func (s *CurrencyService) Invalidate() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cached = nil
	s.cachedAt = time.Time{}
}

func (s *CurrencyService) getCached() ([]CurrencyInfo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.cached != nil && time.Since(s.cachedAt) < currencyCacheTTL {
		result := make([]CurrencyInfo, len(s.cached))
		copy(result, s.cached)
		return result, true
	}
	return nil, false
}

func (s *CurrencyService) setCached(currencies []CurrencyInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cached = make([]CurrencyInfo, len(currencies))
	copy(s.cached, currencies)
	s.cachedAt = time.Now()
}
