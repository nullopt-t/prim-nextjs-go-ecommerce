package model_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
)

func TestBundleCompositionAvailableQuantity(t *testing.T) {
	first := uuid.New()
	second := uuid.New()
	missing := uuid.New()

	stockOf := func(quantities map[uuid.UUID]int) map[uuid.UUID]*model.InventoryStock {
		stock := make(map[uuid.UUID]*model.InventoryStock, len(quantities))
		for variantID, quantity := range quantities {
			stock[variantID] = &model.InventoryStock{
				VariantID:         variantID,
				AvailableQuantity: quantity,
				IsInStock:         quantity > 0,
			}
		}
		return stock
	}

	tests := []struct {
		name   string
		bundle *model.BundleComposition
		stock  map[uuid.UUID]int
		want   int
	}{
		{
			name: "divides component stock by kit quantity",
			bundle: &model.BundleComposition{Items: []model.BundleComponent{
				{StockVariantID: &first, Quantity: 2},
			}},
			stock: map[uuid.UUID]int{first: 5},
			want:  2,
		},
		{
			name: "a kit is limited by its scarcest component",
			bundle: &model.BundleComposition{Items: []model.BundleComponent{
				{StockVariantID: &first, Quantity: 1},
				{StockVariantID: &second, Quantity: 3},
			}},
			stock: map[uuid.UUID]int{first: 8, second: 7},
			want:  2,
		},
		{
			name: "a component without stock records makes the kit unsellable",
			bundle: &model.BundleComposition{Items: []model.BundleComponent{
				{StockVariantID: &first, Quantity: 1},
				{StockVariantID: &second, Quantity: 1},
			}},
			stock: map[uuid.UUID]int{first: 4},
			want:  0,
		},
		{
			name: "a component with no resolved variant makes the kit unsellable",
			bundle: &model.BundleComposition{Items: []model.BundleComponent{
				{StockVariantID: &first, Quantity: 1},
				{StockVariantID: nil, Quantity: 1},
			}},
			stock: map[uuid.UUID]int{first: 4},
			want:  0,
		},
		{
			name: "a component short of one kit zeroes the kit",
			bundle: &model.BundleComposition{Items: []model.BundleComponent{
				{StockVariantID: &first, Quantity: 2},
				{StockVariantID: &second, Quantity: 1},
			}},
			stock: map[uuid.UUID]int{first: 6, second: 0},
			want:  0,
		},
		{
			name:   "an unknown stock variant zeroes the kit",
			bundle: &model.BundleComposition{Items: []model.BundleComponent{{StockVariantID: &missing, Quantity: 1}}},
			stock:  map[uuid.UUID]int{first: 4},
			want:   0,
		},
		{
			name:   "an empty kit is unsellable",
			bundle: &model.BundleComposition{},
			want:   0,
		},
		{
			name:   "a nil kit is unsellable",
			bundle: nil,
			want:   0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.bundle.AvailableQuantity(stockOf(tc.stock)); got != tc.want {
				t.Fatalf("AvailableQuantity() = %d, want %d", got, tc.want)
			}
		})
	}
}
