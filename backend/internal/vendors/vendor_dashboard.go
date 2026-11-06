package vendors

import (
	"context"

	"github.com/google/uuid"
)

// lowStockThreshold is the salable-quantity ceiling at which a variant is
// reported as running low on the merchant dashboard.
const lowStockThreshold = 5

// Dashboard is the merchant home screen: money position plus catalog health.
type Dashboard struct {
	// AvailableBalance is settled earnings minus outstanding payout claims.
	AvailableBalance int64 `json:"availableBalance"`
	// PendingEarnings is earned on paid/processing/shipped orders, not yet payout-eligible.
	PendingEarnings int64 `json:"pendingEarnings"`
	// LifetimeEarnings is the settled (delivered) earnings total.
	LifetimeEarnings int64 `json:"lifetimeEarnings"`
	// OutstandingPayouts is the sum of non-rejected payout claims.
	OutstandingPayouts int64 `json:"outstandingPayouts"`
	// GrossRevenue is the line-total revenue of paid-or-later orders.
	GrossRevenue int64 `json:"grossRevenue"`
	// OrderCount is the number of non-cancelled orders containing the vendor's items.
	OrderCount int `json:"orderCount"`
	// ProductCount is the number of live vendor products.
	ProductCount int `json:"productCount"`
	// PublishedProductCount is the subset visible in the storefront.
	PublishedProductCount int `json:"publishedProductCount"`
	// DraftProductCount is the subset not yet published.
	DraftProductCount int `json:"draftProductCount"`
	// LowStockVariantCount is the number of variants at/below the low-stock threshold.
	LowStockVariantCount int `json:"lowStockVariantCount"`
}

// GetDashboard assembles the merchant dashboard from the order, catalog, and
// payout projections. All reads share one executor; the balance formula is the
// same invariant RequestPayout enforces under lock.
func (s *Service) GetDashboard(ctx context.Context, userID uuid.UUID) (*Dashboard, error) {
	var dashboard *Dashboard
	err := s.runner.WithExecutor(ctx, func(ctx context.Context) error {
		vendor, err := s.requireActiveVendor(ctx, userID)
		if err != nil {
			return err
		}

		sales, err := s.settlement.GetVendorSalesStats(ctx, vendor.ID)
		if err != nil {
			return s.handleError(err, "Failed to compute sales statistics")
		}

		catalog, err := s.products.GetVendorProductStats(ctx, vendor.ID, lowStockThreshold)
		if err != nil {
			return s.handleError(err, "Failed to compute catalog statistics")
		}

		outstanding, err := s.repo.GetOutstandingPayoutTotal(ctx, vendor.ID)
		if err != nil {
			return s.handleError(err, "Failed to compute outstanding payouts")
		}

		dashboard = &Dashboard{
			AvailableBalance:      sales.SettledEarnings - outstanding,
			PendingEarnings:       sales.PendingEarnings,
			LifetimeEarnings:      sales.SettledEarnings,
			OutstandingPayouts:    outstanding,
			GrossRevenue:          sales.GrossRevenue,
			OrderCount:            sales.OrderCount,
			ProductCount:          catalog.Total,
			PublishedProductCount: catalog.Published,
			DraftProductCount:     catalog.Draft,
			LowStockVariantCount:  catalog.LowStock,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return dashboard, nil
}
