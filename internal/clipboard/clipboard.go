package clipboard

import (
	"strings"
	"unicode/utf8"

	"github.com/atotto/clipboard"
)

// Clipboard defines the interface for interacting with system or mock clipboards.
type Clipboard interface {
	Get() (string, error)
	Set(text string) error
}

// SystemClipboard implements Clipboard using the OS-native clipboard.
type SystemClipboard struct{}

func (sc SystemClipboard) Get() (string, error) {
	return clipboard.ReadAll()
}

func (sc SystemClipboard) Set(text string) error {
	return clipboard.WriteAll(text)
}

// TruncateMarker is appended to the text when it is truncated due to exceeding MaxSize.
const TruncateMarker = "\n... [TRUNCATED]"

// Sanitize UTF-8 validates the input, drops invalid bytes/sequences,
// truncates to maxChars if exceeded, and returns empty if the input is whitespace-only.
func Sanitize(input string, maxChars int) string {
	// 1. Skip empty / whitespace-only
	if strings.TrimSpace(input) == "" {
		return ""
	}

	// 2. Validate and clean non-UTF-8 input
	var cleaned string
	if !utf8.ValidString(input) {
		// strings.ToValidUTF8 replaces invalid bytes with a replacement character (usually replacement rune)
		cleaned = strings.ToValidUTF8(input, string(utf8.RuneError))
	} else {
		cleaned = input
	}

	// 3. Truncate to maximum characters (runes) if necessary
	runes := []rune(cleaned)
	if len(runes) > maxChars {
		// Truncate and append a marker
		cleaned = string(runes[:maxChars]) + TruncateMarker
	}

	return cleaned
}
