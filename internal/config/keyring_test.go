package config

import (
	"testing"

	"github.com/99designs/keyring"
)

func TestKeyringStoreAndRetrieve(t *testing.T) {
	// Initialize with an in-memory/array keyring to ensure we don't write to the real OS keyring during tests
	mockKeyring := keyring.NewArrayKeyring(nil)
	oldKr := kr
	kr = mockKeyring
	defer func() {
		kr = oldKr
	}()

	backend := "openrouter"
	secret := "sk-or-v1-some-random-api-key"

	// 1. Get when missing
	val, err := GetAPIKey(backend)
	if err == nil {
		t.Fatalf("expected error getting missing key, got value: %s", val)
	}

	// 2. Set key
	err = SaveAPIKey(backend, secret)
	if err != nil {
		t.Fatalf("failed to save API key: %v", err)
	}

	// 3. Get key
	val, err = GetAPIKey(backend)
	if err != nil {
		t.Fatalf("failed to retrieve API key: %v", err)
	}
	if val != secret {
		t.Errorf("expected secret '%s', got '%s'", secret, val)
	}

	// 4. Delete key
	err = DeleteAPIKey(backend)
	if err != nil {
		t.Fatalf("failed to delete API key: %v", err)
	}

	// 5. Get after delete
	val, err = GetAPIKey(backend)
	if err == nil {
		t.Fatalf("expected error getting deleted key, got value: %s", val)
	}
}

func TestDeleteMissingKey(t *testing.T) {
	mockKeyring := keyring.NewArrayKeyring(nil)
	oldKr := kr
	kr = mockKeyring
	defer func() {
		kr = oldKr
	}()

	err := DeleteAPIKey("nonexistent")
	if err == nil {
		t.Fatal("expected error when deleting nonexistent key")
	}
}
