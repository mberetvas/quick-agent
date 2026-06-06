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

func TestMockOllamaServer_tags(t *testing.T) {
	server := NewMockOllamaServer(nil)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/tags")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %s", resp.Status)
	}
}

func TestMockOllamaServer_unknownPath(t *testing.T) {
	server := NewMockOllamaServer(nil)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/unknown")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %s, want 404", resp.Status)
	}
}

func TestMockOllamaServer_indexExhausted(t *testing.T) {
	server := NewMockOllamaServer([]string{"once"})
	defer server.Close()

	url := server.URL + "/api/generate"
	first, err := http.Post(url, "application/json", strings.NewReader(`{"model":"test"}`))
	if err != nil {
		t.Fatal(err)
	}
	first.Body.Close()

	second, err := http.Post(url, "application/json", strings.NewReader(`{"model":"test"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer second.Body.Close()
	if second.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %s, want 400", second.Status)
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

func TestMockOpenRouterServer_models(t *testing.T) {
	server := NewMockOpenRouterServer(nil)
	defer server.Close()

	resp, err := http.Get(server.URL + "/models")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %s", resp.Status)
	}
}

func TestMockOpenRouterServer_unknownPath(t *testing.T) {
	server := NewMockOpenRouterServer(nil)
	defer server.Close()

	resp, err := http.Get(server.URL + "/unknown")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %s, want 404", resp.Status)
	}
}

func TestMockOpenRouterServer_indexExhausted(t *testing.T) {
	server := NewMockOpenRouterServer([]string{"once"})
	defer server.Close()

	url := server.URL + "/chat/completions"
	body := strings.NewReader(`{"model":"test","messages":[{"role":"user","content":"hi"}],"stream":true}`)

	first, err := http.Post(url, "application/json", body)
	if err != nil {
		t.Fatal(err)
	}
	first.Body.Close()

	second, err := http.Post(url, "application/json", strings.NewReader(`{"model":"test","messages":[{"role":"user","content":"hi"}],"stream":true}`))
	if err != nil {
		t.Fatal(err)
	}
	defer second.Body.Close()
	if second.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %s, want 400", second.Status)
	}
}
