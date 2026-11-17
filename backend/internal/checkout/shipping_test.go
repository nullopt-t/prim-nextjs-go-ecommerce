package checkout

import (
	"testing"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/setting"
)

func TestComputeShippingBreakdown(t *testing.T) {
	cfg := setting.DefaultConfig()
	a, b := uuid.New(), uuid.New()
	threshold := int64(10000)
	sellers := map[uuid.UUID]*model.Seller{
		a: {StoreName: "Vendor A", ShippingFee: 500, FreeShippingThreshold: &threshold, ShippingLeadTimeDays: 2},
		b: {StoreName: "Vendor B", ShippingFee: 700, ShippingLeadTimeDays: 6},
	}
	totals := map[uuid.UUID]int64{a: threshold, b: 3000}

	breakdown, total := computeShippingBreakdown(1000, totals, sellers, cfg)
	if total != 999+700 {
		t.Fatalf("total shipping = %d, want 1699 (platform 999 + vendor B 700)", total)
	}
	if len(breakdown) != 3 {
		t.Fatalf("breakdown has %d groups, want 3", len(breakdown))
	}

	byName := make(map[string]ShippingBreakdown, len(breakdown))
	for _, group := range breakdown {
		byName[group.SellerName] = group
	}
	if group := byName["Vendor A"]; group.ShippingFee != 0 || !group.IsFree || group.Subtotal != threshold || group.LeadTimeDays != 2 {
		t.Errorf("vendor A at its threshold = %+v, want free with 2-day lead", group)
	}
	if group := byName["Vendor B"]; group.ShippingFee != 700 || group.IsFree || group.LeadTimeDays != 6 {
		t.Errorf("vendor B without threshold = %+v, want flat 700 with 6-day lead", group)
	}
	platform := byName[cfg.StoreName]
	if platform.SellerID != nil || platform.ShippingFee != 999 || platform.Subtotal != 1000 {
		t.Errorf("platform group = %+v, want global rule on 1000 subtotal", platform)
	}
}

func TestComputeShippingBreakdownNoLines(t *testing.T) {
	breakdown, total := computeShippingBreakdown(0, nil, nil, setting.DefaultConfig())
	if total != 0 || len(breakdown) != 0 {
		t.Fatalf("empty cart = (%d, %+v), want (0, empty)", total, breakdown)
	}
}
