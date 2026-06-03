package hotkey

import (
	"errors"
	"strings"
)

// normalizeKeys lowercases key tokens, maps option→alt, and rejects empty input.
func normalizeKeys(keys ...string) ([]string, error) {
	if len(keys) == 0 {
		return nil, errors.New("empty key combination")
	}
	out := make([]string, len(keys))
	for i, k := range keys {
		k = strings.ToLower(strings.TrimSpace(k))
		if k == "" {
			return nil, errors.New("empty key")
		}
		if k == "option" {
			k = "alt"
		}
		out[i] = k
	}
	return out, nil
}
