package checkout

import (
	"testing"

	"github.com/google/uuid"
)

func TestComputeSellerEarning(t *testing.T) {
	notAttributed := uuid.Nil
	tests := []struct {
		name          string
		itemTotal     int64
		sellerID      *uuid.UUID
		commissionBps int
		want          int64
	}{
		{
			name:          "platform line earns nothing without a seller",
			itemTotal:     12_345,
			sellerID:      nil,
			commissionBps: 1000,
			want:          0,
		},
		{
			name:          "seller line keeps the commission share",
			itemTotal:     12_345,
			sellerID:      &notAttributed,
			commissionBps: 1000,
			want:          12_345 - 12_345*1000/10000,
		},
		{
			name:          "zero commission transfers the full line to the seller",
			itemTotal:     9_999,
			sellerID:      &notAttributed,
			commissionBps: 0,
			want:          9_999,
		},
		{
			name:          "full commission takes the whole line",
			itemTotal:     7_500,
			sellerID:      &notAttributed,
			commissionBps: 10_000,
			want:          0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := computeSellerEarning(tt.itemTotal, tt.sellerID, tt.commissionBps); got != tt.want {
				t.Fatalf("computeSellerEarning() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestComputeSellerSplitPreservesLineTotal(t *testing.T) {
	sellerID := uuid.New()
	earning, commission := computeSellerSplit(5_998, &sellerID, 1000)
	if earning != 5_399 || commission != 599 {
		t.Fatalf("computeSellerSplit() = (%d, %d), want (5399, 599)", earning, commission)
	}
	if earning+commission != 5_998 {
		t.Fatalf("split does not preserve line total: %d + %d", earning, commission)
	}

	earning, commission = computeSellerSplit(5_998, nil, 1000)
	if earning != 0 || commission != 0 {
		t.Fatalf("platform line split = (%d, %d), want (0, 0)", earning, commission)
	}
}
