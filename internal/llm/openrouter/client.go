package openrouter

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/mberetvas/quick-agent/internal/config"
	"github.com/mberetvas/quick-agent/internal/llm/retry"
)

const defaultBaseURL = "https://openrouter.ai/api/v1"

// Client implements llm.LLMClient for the OpenRouter API.
type Client struct {
	orCfg  config.OpenRouterConfig
	llmCfg config.LLMConfig

	baseURL    string
	httpClient *http.Client
	getAPIKey  func() (string, error)
}

// NewClient returns a Client configured from OpenRouter and LLM retry settings.
func NewClient(orCfg config.OpenRouterConfig, llmCfg config.LLMConfig) *Client {
	return &Client{
		orCfg:      orCfg,
		llmCfg:     llmCfg,
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{},
		getAPIKey:  orCfg.GetAPIKey,
	}
}

// chatRequest is the OpenRouter chat completions payload.
type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Stream      bool          `json:"stream"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// streamChunk parses a single SSE data line from OpenRouter.
type streamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// HealthCheck verifies the API key is set and the OpenRouter API is reachable.
func (c *Client) HealthCheck(ctx context.Context) error {
	key, err := c.getAPIKey()
	if err != nil {
		return err
	}
	if strings.TrimSpace(key) == "" {
		return errors.New("openrouter API key is empty")
	}

	url := c.baseURL + "/models"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+key)

	client := c.httpClient
	if client == nil {
		client = http.DefaultClient
	}
	timeout := time.Duration(c.orCfg.Timeout) * time.Second
	if timeout > 0 {
		client = &http.Client{
			Timeout:   timeout,
			Transport: c.httpClient.Transport,
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("openrouter healthcheck failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("openrouter healthcheck returned status %s", resp.Status)
	}
	return nil
}

// Generate streams completion tokens from OpenRouter via SSE.
func (c *Client) Generate(ctx context.Context, prompt string) (<-chan string, <-chan error, error) {
	key, err := c.getAPIKey()
	if err != nil {
		return nil, nil, err
	}

	payload := chatRequest{
		Model: c.orCfg.Model,
		Messages: []chatMessage{
			{Role: "user", Content: prompt},
		},
		Stream:      true,
		MaxTokens:   c.orCfg.MaxTokens,
		Temperature: c.orCfg.Temperature,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal openrouter request: %w", err)
	}

	url := c.baseURL + "/chat/completions"
	resp, err := c.postWithRetry(ctx, url, body, key)
	if err != nil {
		return nil, nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, nil, fmt.Errorf("openrouter returned status %s", resp.Status)
	}

	tokens := make(chan string, 100)
	errs := make(chan error, 1)

	go func() {
		defer func() {
			resp.Body.Close()
			close(tokens)
			close(errs)
		}()
		if err := c.readSSEStream(ctx, resp.Body, tokens, errs); err != nil {
			select {
			case errs <- err:
			default:
			}
		}
	}()

	return tokens, errs, nil
}

func (c *Client) postWithRetry(ctx context.Context, url string, body []byte, apiKey string) (*http.Response, error) {
	client := c.httpClient
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := retry.DoHTTP(ctx, c.llmCfg, func() (*http.Response, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("Accept", "text/event-stream")
		return client.Do(req)
	})
	if err != nil {
		return nil, fmt.Errorf("openrouter request failed: %w", err)
	}

	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		resp.Body.Close()
		return nil, fmt.Errorf("openrouter client error: %s", resp.Status)
	}

	return resp, nil
}

func (c *Client) readSSEStream(ctx context.Context, body io.Reader, tokens chan<- string, errs chan<- error) error {
	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return err
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			return nil
		}

		var chunk streamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if chunk.Error != nil && chunk.Error.Message != "" {
			return fmt.Errorf("openrouter stream error: %s", chunk.Error.Message)
		}
		if len(chunk.Choices) == 0 {
			continue
		}
		content := chunk.Choices[0].Delta.Content
		if content == "" {
			continue
		}

		select {
		case tokens <- content:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}
