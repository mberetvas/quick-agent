package version

import "testing"

func TestString(t *testing.T) {
	got := String()
	if got == "" {
		t.Fatal("String() returned empty string")
	}
	if got != Version {
		t.Errorf("String() = %q, want %q", got, Version)
	}
}
