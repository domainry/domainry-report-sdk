package contract

import "testing"

func TestCanonicalJSONAndSafeExportFilename(t *testing.T) {
	left, err := CanonicalJSONSHA256(map[string]any{"report": "orders", "limit": 10})
	if err != nil {
		t.Fatal(err)
	}
	right, err := CanonicalJSONSHA256(map[string]any{"limit": 10, "report": "orders"})
	if err != nil || left != right || len(left) != 64 {
		t.Fatalf("canonical hashes left=%q right=%q err=%v", left, right, err)
	}
	if got := SafeExportFilename(" daily/orders ", "customer details"); got != "daily_orders-customer_details.csv" {
		t.Fatalf("safe filename = %q", got)
	}
}
