// Package testutil provides shared HTTP test servers for LLM client tests.
package testutil

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
)

// NewMockOllamaServer returns an httptest server that streams NDJSON like Ollama /api/generate.
func NewMockOllamaServer(responses []string) *httptest.Server {
	index := 0
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/tags":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"models":[]}`))
			return
		case "/api/generate":
			if index >= len(responses) {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/x-ndjson")
			for _, resp := range responses {
				line := fmt.Sprintf(`{"response":%q}`, resp)
				_, _ = w.Write([]byte(line + "\n"))
			}
			_, _ = w.Write([]byte(`{"done":true}` + "\n"))
			index = len(responses)
			return
		default:
			http.NotFound(w, r)
		}
	}))
}

// NewMockOpenRouterServer returns an httptest server that streams SSE like OpenRouter chat completions.
func NewMockOpenRouterServer(responses []string) *httptest.Server {
	index := 0
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" && r.URL.Path != "/api/v1/chat/completions" && !strings.HasSuffix(r.URL.Path, "/chat/completions") {
			if r.URL.Path == "/models" {
				w.WriteHeader(http.StatusOK)
				return
			}
			http.NotFound(w, r)
			return
		}
		if index >= len(responses) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		for _, resp := range responses {
			_, _ = fmt.Fprintf(w, "data: {\"choices\":[{\"delta\":{\"content\":%q}}]}\n\n", resp)
		}
		_, _ = fmt.Fprint(w, "data: [DONE]\n\n")
		index = len(responses)
	}))
}
