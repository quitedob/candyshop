package order

import "testing"

func TestValidateOrderStatusTransition_PartialFulfillment(t *testing.T) {
	cases := []struct {
		from, to string
		wantErr  bool
	}{
		{OrderStatusProduction, OrderStatusPartiallyShipped, false},
		{OrderStatusPartiallyShipped, OrderStatusPartiallyDelivered, false},
		{OrderStatusPartiallyDelivered, OrderStatusDelivered, false},
		{OrderStatusProduction, OrderStatusPartiallyDelivered, true},
	}
	for _, tc := range cases {
		err := ValidateOrderStatusTransition(tc.from, tc.to)
		if tc.wantErr && err == nil {
			t.Fatalf("expected error for %s -> %s", tc.from, tc.to)
		}
		if !tc.wantErr && err != nil {
			t.Fatalf("unexpected error for %s -> %s: %v", tc.from, tc.to, err)
		}
	}
}
