package clipboard

import "testing"

func TestSystemClipboard_roundTrip(t *testing.T) {
	var cb SystemClipboard
	const want = "quick-agent clipboard test"

	if err := cb.Set(want); err != nil {
		t.Skipf("clipboard unavailable in this environment: %v", err)
	}

	got, err := cb.Get()
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got != want {
		t.Errorf("Get() = %q, want %q", got, want)
	}
}
