package config

import (
	"errors"
	"fmt"

	"github.com/99designs/keyring"
)

const (
	ServiceName = "quick-agent"
)

// kr is the package-level Keyring instance. It can be overridden in tests.
var kr keyring.Keyring

func init() {
	var err error
	kr, err = keyring.Open(keyring.Config{
		ServiceName: ServiceName,
	})
	if err != nil {
		// Fallback to array keyring if native setup fails (e.g. headless CI environments)
		kr = keyring.NewArrayKeyring(nil)
	}
}

// GetAPIKeyKey constructs the keyring key for a given backend
func GetAPIKeyKey(backend string) string {
	return fmt.Sprintf("%s_api_key", backend)
}

// SaveAPIKey stores the API key in system keyring for the specified backend
func SaveAPIKey(backend string, key string) error {
	if kr == nil {
		return errors.New("keyring not initialized")
	}
	return kr.Set(keyring.Item{
		Key:  GetAPIKeyKey(backend),
		Data: []byte(key),
	})
}

// GetAPIKey retrieves the API key from system keyring for the specified backend
func GetAPIKey(backend string) (string, error) {
	if kr == nil {
		return "", errors.New("keyring not initialized")
	}
	item, err := kr.Get(GetAPIKeyKey(backend))
	if err != nil {
		// Use standard library check if needed or check error strings
		if errors.Is(err, keyring.ErrKeyNotFound) {
			return "", fmt.Errorf("API key not found in system keyring for backend '%s'", backend)
		}
		return "", err
	}
	return string(item.Data), nil
}

// DeleteAPIKey removes the API key from system keyring for the specified backend
func DeleteAPIKey(backend string) error {
	if kr == nil {
		return errors.New("keyring not initialized")
	}
	key := GetAPIKeyKey(backend)
	_, err := kr.Get(key)
	if err != nil {
		return fmt.Errorf("API key not found in system keyring for backend '%s'", backend)
	}
	return kr.Remove(key)
}

// GetAPIKey retrieves the API key for the current active backend in Config
func (cfg *Config) GetAPIKey() (string, error) {
	return GetAPIKey(cfg.Backend)
}

// GetAPIKey retrieves the API key specifically for OpenRouter
func (o OpenRouterConfig) GetAPIKey() (string, error) {
	return GetAPIKey("openrouter")
}
