package testutil

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/mberetvas/quick-agent/internal/config"
	"github.com/mberetvas/quick-agent/internal/llm/ollama"
)

func TestMockOllamaServer(t *testing.T) {
	server := NewMockOllamaServer([]string{"Hello", " world"})
	defer server.Close()

	client := ollama.NewClient(config.OllamaConfig{URL: server.URL, Model: "test"}, config.DefaultLLMConfig())
	tokens, errs, err := client.Generate(context.Background(), "hi")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	var got string
	for tok := range tokens {
		got += tok
	}
	for err := range errs {
		if err != nil {
			t.Fatalf("stream error: %v", err)
		}
	}
	if got != "Hello world" {
		t.Errorf("got %q, want %q", got, "Hello world")
	}
}

func TestMockOpenRouterServer(t *testing.T) {
	server := NewMockOpenRouterServer([]string{"Hi", " there"})
	defer server.Close()

	resp, err := http.Post(server.URL+"/chat/completions", "application/json", strings.NewReader(`{"model":"test","messages":[{"role":"user","content":"hi"}],"stream":true}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	s := string(body)
	if !strings.Contains(s, "Hi") || !strings.Contains(s, " there") {
		t.Errorf("unexpected SSE body: %s", s)
	}
}
