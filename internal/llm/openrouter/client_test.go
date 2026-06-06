package openrouter

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mberetvas/quick-agent/internal/config"
)

func testClient(serverURL string, llmCfg config.LLMConfig) *Client {
	c := NewClient(config.OpenRouterConfig{
		Model:       "mistralai/mistral-7b-instruct",
		Timeout:     2,
		MaxTokens:   100,
		Temperature: 0.7,
	}, llmCfg)
	c.baseURL = serverURL + "/api/v1"
	c.getAPIKey = func() (string, error) { return "test-api-key", nil }
	return c
}

func TestHealthCheck_success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/models" {
			t.Errorf("expected /api/v1/models, got %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-api-key" {
			t.Errorf("expected Bearer test-api-key, got %q", got)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := testClient(server.URL, config.DefaultLLMConfig())
	if err := client.HealthCheck(context.Background()); err != nil {
		t.Fatalf("HealthCheck: %v", err)
	}
}

func TestHealthCheck_missing_api_key(t *testing.T) {
	client := NewClient(config.DefaultOpenRouterConfig(), config.DefaultLLMConfig())
	client.getAPIKey = func() (string, error) {
		return "", fmt.Errorf("API key not found in system keyring for backend 'openrouter'")
	}
	err := client.HealthCheck(context.Background())
	if err == nil {
		t.Fatal("expected error when API key missing")
	}
}

func TestHealthCheck_endpoint_error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := testClient(server.URL, config.DefaultLLMConfig())
	err := client.HealthCheck(context.Background())
	if err == nil {
		t.Fatal("expected healthcheck error on 5xx")
	}
}

func TestGenerateStream_success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/chat/completions" {
			t.Errorf("expected /api/v1/chat/completions, got %s", r.URL.Path)
		}
		if r.Header.Get("Accept") != "text/event-stream" {
			t.Errorf("expected Accept: text/event-stream")
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"Hello\"}}]}\n\n")
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\" world\"}}]}\n\n")
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	client := testClient(server.URL, config.LLMConfig{RetryAttempts: 1})
	tokens, errs, err := client.Generate(context.Background(), "Hi")
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

func TestGenerateStream_context_cancel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("expected flusher")
		}
		for i := 0; i < 50; i++ {
			fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"x\"}}]}\n\n")
			flusher.Flush()
			time.Sleep(20 * time.Millisecond)
		}
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	client := testClient(server.URL, config.LLMConfig{RetryAttempts: 1})
	ctx, cancel := context.WithCancel(context.Background())
	tokens, errs, err := client.Generate(ctx, "Hi")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	<-tokens
	cancel()

	deadline := time.After(2 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("timed out waiting for stream to end after cancel")
		case _, ok := <-tokens:
			if !ok {
				goto drained
			}
		case err, ok := <-errs:
			if ok && err != nil && err != context.Canceled {
				t.Fatalf("unexpected error: %v", err)
			}
			if !ok {
				goto drained
			}
		}
	}
drained:
}

func TestRetry_on_5xx_then_success(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := attempts.Add(1)
		if n < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"ok\"}}]}\n\n")
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	client := testClient(server.URL, config.LLMConfig{
		RetryAttempts: 3,
		RetryBackoff:  []int{10, 20},
	})
	tokens, errs, err := client.Generate(context.Background(), "test")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for range tokens {
	}
	for err := range errs {
		if err != nil {
			t.Fatalf("stream error: %v", err)
		}
	}
	if attempts.Load() != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts.Load())
	}
}

func TestNoRetry_on_4xx(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := testClient(server.URL, config.LLMConfig{
		RetryAttempts: 3,
		RetryBackoff:  []int{10},
	})
	_, _, err := client.Generate(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error on 401")
	}
	if attempts.Load() != 1 {
		t.Errorf("expected 1 attempt on 4xx, got %d", attempts.Load())
	}
}

func TestRetry_on_429(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := attempts.Add(1)
		if n < 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"ok\"}}]}\n\n")
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	client := testClient(server.URL, config.LLMConfig{
		RetryAttempts: 3,
		RetryBackoff:  []int{10},
	})
	tokens, errs, err := client.Generate(context.Background(), "test")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for range tokens {
	}
	for err := range errs {
		if err != nil {
			t.Fatalf("stream error: %v", err)
		}
	}
	if attempts.Load() != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts.Load())
	}
}

func TestNoRetry_on_context_cancel(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := testClient(server.URL, config.LLMConfig{
		RetryAttempts: 3,
		RetryBackoff:  []int{100, 200},
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err := client.Generate(ctx, "test")
	if err == nil {
		t.Fatal("expected error when context already canceled")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
	if attempts.Load() != 0 {
		t.Errorf("expected no HTTP attempts after canceled context, got %d", attempts.Load())
	}
}
