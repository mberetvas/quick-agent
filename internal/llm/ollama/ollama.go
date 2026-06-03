package ollama

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/yourname/clipboard-tui/internal/config"
)

// OllamaClient implements the llm.LLMClient interface for the local Ollama API.
type OllamaClient struct {
	cfg config.OllamaConfig
}

// NewClient returns an initialized OllamaClient using the configured parameters.
func NewClient(cfg config.OllamaConfig) *OllamaClient {
	if cfg.URL == "" {
		cfg.URL = "http://localhost:11434"
	}
	return &OllamaClient{cfg: cfg}
}

// Request represents the payload for the Ollama /api/generate endpoint.
type Request struct {
	Model   string  `json:"model"`
	Prompt  string  `json:"prompt"`
	Stream  bool    `json:"stream"`
	Options Options `json:"options,omitempty"`
}

// Options represents model hyper-parameters.
type Options struct {
	NumPredict  int     `json:"num_predict,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
}

// Response represents a single streaming chunk from the Ollama API (/api/generate).
type Response struct {
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
	Response  string    `json:"response"`
	Done      bool      `json:"done"`
}

// HealthCheck checks if Ollama service is reachable.
func (oc *OllamaClient) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s/api/tags", oc.cfg.URL) // standard tags endpoint to verify up
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	client := &http.Client{
		Timeout: time.Duration(oc.cfg.Timeout) * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Ollama healthcheck failed (is it running?): %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Ollama healthcheck returned non-OK status: %s", resp.Status)
	}

	return nil
}

// Generate streams model response tokens back step-by-step.
func (oc *OllamaClient) Generate(ctx context.Context, prompt string) (<-chan string, error) {
	url := fmt.Sprintf("%s/api/generate", oc.cfg.URL)

	payload := Request{
		Model:  oc.cfg.Model,
		Prompt: prompt,
		Stream: true,
		Options: Options{
			NumPredict:  oc.cfg.MaxTokens,
			Temperature: oc.cfg.Temperature,
		},
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		// Note: Don't set overall client.Timeout when streaming, relies on context cancellation
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Ollama generate request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("Ollama returned raw error status: %s", resp.Status)
	}

	tokens := make(chan string, 100)

	go func() {
		defer func() {
			resp.Body.Close()
			close(tokens)
		}()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
				line := scanner.Bytes()
				if len(line) == 0 {
					continue
				}

				var chunk Response
				if err := json.Unmarshal(line, &chunk); err != nil {
					// Output parsing errors can be ignored during streaming, or we can exit early.
					// For a trace bullet, we skip invalid lines.
					continue
				}

				if chunk.Response != "" {
					select {
					case tokens <- chunk.Response:
					case <-ctx.Done():
						return
					}
				}

				if chunk.Done {
					return
				}
			}
		}
	}()

	return tokens, nil
}
