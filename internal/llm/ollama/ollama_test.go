package ollama

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourname/clipboard-tui/internal/config"
)

func TestOllamaHealthCheck(t *testing.T) {
	// Setup a local test HTTP server to mock Ollama /api/tags
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tags" {
			t.Errorf("expected healthcheck to hit /api/tags, got: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"models": []}`))
	}))
	defer server.Close()

	cfg := config.OllamaConfig{
		URL:     server.URL,
		Timeout: 2,
	}

	client := NewClient(cfg)
	err := client.HealthCheck(context.Background())
	if err != nil {
		t.Fatalf("expected healthcheck to pass, got error: %v", err)
	}
}

func TestOllamaHealthCheck_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cfg := config.OllamaConfig{
		URL:     server.URL,
		Timeout: 2,
	}

	client := NewClient(cfg)
	err := client.HealthCheck(context.Background())
	if err == nil {
		t.Fatal("expected healthcheck to fail on non-OK code, but got nil")
	}
}

func TestOllamaGenerateStream(t *testing.T) {
	// Setup mock streaming output format (NDJSON lines)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/generate" {
			t.Errorf("expected URL path /api/generate, got: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Send two model tokens and a done flag
		fmt.Fprintln(w, `{"model":"llama3:8b","response":"Hello","done":false}`)
		fmt.Fprintln(w, `{"model":"llama3:8b","response":" world!","done":false}`)
		fmt.Fprintln(w, `{"model":"llama3:8b","response":"","done":true}`)
	}))
	defer server.Close()

	cfg := config.OllamaConfig{
		URL:   server.URL,
		Model: "llama3:8b",
	}

	client := NewClient(cfg)
	tokens, err := client.Generate(context.Background(), "Greet me")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	var results []string
	for token := range tokens {
		results = append(results, token)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 tokens, got %v", results)
	}

	fullResponse := results[0] + results[1]
	if fullResponse != "Hello world!" {
		t.Errorf("expected 'Hello world!' response, got: %s", fullResponse)
	}
}
