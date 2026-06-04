// Package retry provides shared HTTP retry helpers for LLM clients.
package retry

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/yourname/clipboard-tui/internal/config"
)

// DoHTTP executes do up to RetryAttempts times with configured backoff on retryable failures.
func DoHTTP(ctx context.Context, llmCfg config.LLMConfig, do func() (*http.Response, error)) (*http.Response, error) {
	attempts := llmCfg.RetryAttempts
	if attempts < 1 {
		attempts = 1
	}
	backoffs := llmCfg.RetryBackoff

	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if attempt > 0 {
			delay := Delay(backoffs, attempt-1)
			if delay > 0 {
				timer := time.NewTimer(delay)
				select {
				case <-ctx.Done():
					timer.Stop()
					return nil, ctx.Err()
				case <-timer.C:
				}
			}
		}

		resp, err := do()
		if err != nil {
			lastErr = err
			if !IsRetryableError(err) || attempt == attempts-1 {
				return nil, err
			}
			continue
		}

		if IsRetryableStatus(resp.StatusCode) {
			lastErr = errors.New("retryable HTTP status: " + resp.Status)
			resp.Body.Close()
			if attempt == attempts-1 {
				return nil, lastErr
			}
			continue
		}

		return resp, nil
	}

	return nil, lastErr
}

// Delay returns the backoff duration for the given attempt index.
func Delay(backoffs []int, attempt int) time.Duration {
	if len(backoffs) == 0 {
		return 0
	}
	idx := attempt
	if idx >= len(backoffs) {
		idx = len(backoffs) - 1
	}
	return time.Duration(backoffs[idx]) * time.Millisecond
}

// IsRetryableStatus reports whether an HTTP status should be retried.
func IsRetryableStatus(code int) bool {
	return code == http.StatusTooManyRequests || code >= 500
}

// IsRetryableError reports whether a transport error should be retried.
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}
	return true
}
