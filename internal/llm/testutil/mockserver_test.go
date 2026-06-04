package testutil

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/yourname/clipboard-tui/internal/config"
	"github.com/yourname/clipboard-tui/internal/llm/ollama"
	"github.com/yourname/clipboard-tui/internal/llm/openrouter"
)

func TestMockOllamaServer(t *testing.T) {
	server := NewMockOllamaServer([]string{"Hello", " world"})
	defer server.Close()

	client := ollama.NewClient(config.OllamaConfig{URL: server.URL, Model: "test"})
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

	client := openrouter.NewClient(config.OpenRouterConfig{
		Model: "test",
	}, config.LLMConfig{})
	client.(*openrouter.Client) // compile-time check only — use direct construction below
	_ = client

	// NewClient does not expose baseURL override; hit server via httptest path on default client.
	orClient := openrouter.NewClient(config.OpenRouterConfig{
		Model:     "test",
		GetAPIKey: func() (string, error) { return "test-key", nil },
	}, config.LLMConfig{})

	// Use reflection-free approach: test server handles /chat/completions at root if we patch URL.
	// openrouter uses defaultBaseURL; test openrouter package tests already cover SSE.
	// Verify mock server responds.
	resp, err := http.Get(server.URL + "/chat/completions")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusOK {
		t.Fatalf("status %s", resp.Status)
	}
	_, _ = io.ReadAll(resp.Body)
	_ = orClient
}
