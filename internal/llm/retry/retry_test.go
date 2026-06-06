package retry

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mberetvas/quick-agent/internal/config"
)

func TestIsRetryableError(t *testing.T) {
	timeoutErr := &net.DNSError{IsTimeout: true}
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		{name: "context canceled", err: context.Canceled, want: false},
		{name: "deadline exceeded", err: context.DeadlineExceeded, want: false},
		{name: "net timeout", err: timeoutErr, want: true},
		{name: "generic", err: errors.New("connection reset"), want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRetryableError(tt.err); got != tt.want {
				t.Errorf("IsRetryableError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsRetryableStatus(t *testing.T) {
	tests := []struct {
		code int
		want bool
	}{
		{code: http.StatusTooManyRequests, want: true},
		{code: http.StatusInternalServerError, want: true},
		{code: http.StatusServiceUnavailable, want: true},
		{code: http.StatusBadRequest, want: false},
		{code: http.StatusNotFound, want: false},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d", tt.code), func(t *testing.T) {
			if got := IsRetryableStatus(tt.code); got != tt.want {
				t.Errorf("IsRetryableStatus(%d) = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}

func TestDelay(t *testing.T) {
	tests := []struct {
		name     string
		backoffs []int
		attempt  int
		want     time.Duration
	}{
		{name: "empty backoffs", backoffs: nil, attempt: 0, want: 0},
		{name: "first entry", backoffs: []int{100, 200}, attempt: 0, want: 100 * time.Millisecond},
		{name: "second entry", backoffs: []int{100, 200}, attempt: 1, want: 200 * time.Millisecond},
		{name: "past end uses last", backoffs: []int{100, 200}, attempt: 5, want: 200 * time.Millisecond},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Delay(tt.backoffs, tt.attempt); got != tt.want {
				t.Errorf("Delay() = %v, want %v", got, tt.want)
			}
		})
	}
}

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

func TestDoHTTP_retries_on_transport_error(t *testing.T) {
	var calls atomic.Int32
	cfg := config.LLMConfig{RetryAttempts: 3, RetryBackoff: []int{1}}

	resp, err := DoHTTP(context.Background(), cfg, func() (*http.Response, error) {
		n := calls.Add(1)
		if n == 1 {
			return nil, &net.DNSError{IsTimeout: true, Err: "timeout", Name: "example.com"}
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       http.NoBody,
		}, nil
	})
	if err != nil {
		t.Fatalf("DoHTTP: %v", err)
	}
	defer resp.Body.Close()
	if calls.Load() != 2 {
		t.Errorf("expected 2 calls, got %d", calls.Load())
	}
}

func TestDoHTTP_context_cancel_during_backoff(t *testing.T) {
	cfg := config.LLMConfig{RetryAttempts: 3, RetryBackoff: []int{5000}}
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		_, _ = DoHTTP(ctx, cfg, func() (*http.Response, error) {
			return nil, &net.DNSError{IsTimeout: true, Err: "timeout", Name: "example.com"}
		})
		close(done)
	}()

	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("DoHTTP did not exit after context cancel")
	}
}

func TestDoHTTP_exhausts_retries_on_503(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	cfg := config.LLMConfig{RetryAttempts: 2, RetryBackoff: []int{1}}
	_, err := DoHTTP(context.Background(), cfg, func() (*http.Response, error) {
		return http.Get(server.URL)
	})
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if calls.Load() != 2 {
		t.Errorf("expected 2 calls, got %d", calls.Load())
	}
}
