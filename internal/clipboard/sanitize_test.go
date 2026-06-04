package clipboard

import (
	"strings"
	"testing"
)

func TestSanitize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxChars int
		want     string
	}{
		{
			name:     "valid simple text",
			input:    "hello world",
			maxChars: 100,
			want:     "hello world",
		},
		{
			name:     "empty text",
			input:    "",
			maxChars: 100,
			want:     "",
		},
		{
			name:     "whitespace only",
			input:    "   \n \t  ",
			maxChars: 100,
			want:     "",
		},
		{
			name:     "non-UTF8 input cleaned",
			input:    "hello \xff world",
			maxChars: 100,
			want:     "hello \ufffd world",
		},
		{
			name:     "truncation with marker",
			input:    "abcdefghij",
			maxChars: 5,
			want:     "abcde" + TruncateMarker,
		},
		{
			name:     "exactly max length no truncation",
			input:    "abcde",
			maxChars: 5,
			want:     "abcde",
		},
		{
			name:     "unicode character aware truncation",
			input:    "世界你好!",
			maxChars: 2,
			want:     "世界" + TruncateMarker,
		},
		{
			name:     "large text truncated",
			input:    strings.Repeat("a", 150),
			maxChars: 10,
			want:     strings.Repeat("a", 10) + TruncateMarker,
		},
		{
			name:     "utf8 emoji preserved",
			input:    "Hello 世界 🌍",
			maxChars: 100,
			want:     "Hello 世界 🌍",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Sanitize(tt.input, tt.maxChars)
			if got != tt.want {
				t.Errorf("Sanitize() got = %q, want = %q", got, tt.want)
			}
		})
	}
}

func BenchmarkSanitize(b *testing.B) {
	input := strings.Repeat("hello world ", 1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Sanitize(input, 10000)
	}
}
