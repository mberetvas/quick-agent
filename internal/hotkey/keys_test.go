package hotkey

import (
	"reflect"
	"testing"
)

func TestNormalizeKeys_optionToAlt(t *testing.T) {
	got, err := normalizeKeys("cmd", "option", "v")
	if err != nil {
		t.Fatalf("normalizeKeys: %v", err)
	}
	want := []string{"cmd", "alt", "v"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("normalizeKeys = %v, want %v", got, want)
	}
}

func TestNormalizeKeys_rejectsEmpty(t *testing.T) {
	if _, err := normalizeKeys(); err == nil {
		t.Fatal("expected error for empty combination")
	}
	if _, err := normalizeKeys("ctrl", ""); err == nil {
		t.Fatal("expected error for empty key token")
	}
}
