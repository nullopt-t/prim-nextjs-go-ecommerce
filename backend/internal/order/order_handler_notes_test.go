package order

import "testing"

func TestCustomerHistoryNotes_StripsInternalDiagnostics(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want bool
	}{
		{"client intent", "payment confirmed via client intent=pi_3UcLqY2FuPaeAQGE0Eg5SMwf", false},
		{"webhook tx", "payment webhook payment_intent.succeeded tx=tx_123", false},
		{"creation status", "Order created with status pending", false},
		{"uppercase marker", "Payment WEBHOOK tx=abc", false},
		{"human note", "Cancelled by customer", true},
		{"tracking", "Tracking updated: TRACK-1", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := tc.in
			got := customerHistoryNotes(&in)
			if (got != nil) != tc.want {
				t.Fatalf("customerHistoryNotes(%q) = %v, want visible=%v", tc.in, got, tc.want)
			}
		})
	}
	if customerHistoryNotes(nil) != nil {
		t.Fatal("nil notes must stay nil")
	}
}
