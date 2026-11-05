package checkout

import (
	"testing"

	"github.com/google/uuid"
)

func TestComputeVendorEarning(t *testing.T) {
	notAttributed := uuid.Nil
	tests := []struct {
		name          string
		itemTotal     int64
		vendorID      *uuid.UUID
		commissionBps int
		want          int64
	}{
		{
			name:          "platform line earns nothing without a vendor",
			itemTotal:     12_345,
			vendorID:      nil,
			commissionBps: 1000,
			want:          0,
		},
		{
			name:          "vendor line keeps the commission share",
			itemTotal:     12_345,
			vendorID:      &notAttributed,
			commissionBps: 1000,
			want:          12_345 - 12_345*1000/10000,
		},
		{
			name:          "zero commission transfers the full line to the vendor",
			itemTotal:     9_999,
			vendorID:      &notAttributed,
			commissionBps: 0,
			want:          9_999,
		},
		{
			name:          "full commission takes the whole line",
			itemTotal:     7_500,
			vendorID:      &notAttributed,
			commissionBps: 10_000,
			want:          0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := computeVendorEarning(tt.itemTotal, tt.vendorID, tt.commissionBps); got != tt.want {
				t.Fatalf("computeVendorEarning() = %d, want %d", got, tt.want)
			}
		})
	}
}
