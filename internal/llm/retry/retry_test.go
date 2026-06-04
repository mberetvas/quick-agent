package retry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yourname/clipboard-tui/internal/config"
)

func TestDoHTTP_retries_on_503(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := config.LLMConfig{
		RetryAttempts: 3,
		RetryBackoff:  []int{10},
	}

	resp, err := DoHTTP(context.Background(), cfg, func() (*http.Response, error) {
		return http.Get(server.URL)
	})
	if err != nil {
		t.Fatalf("DoHTTP: %v", err)
	}
	defer resp.Body.Close()
	if calls.Load() != 2 {
		t.Errorf("expected 2 calls, got %d", calls.Load())
	}
}

func TestDoHTTP_no_retry_on_4xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	cfg := config.LLMConfig{RetryAttempts: 3, RetryBackoff: []int{10}}
	resp, err := DoHTTP(context.Background(), cfg, func() (*http.Response, error) {
		return http.Get(server.URL)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d", resp.StatusCode)
	}
}

func TestDelay(t *testing.T) {
	if got := Delay([]int{100, 200}, 1); got != 200*time.Millisecond {
		t.Errorf("got %v", got)
	}
}
