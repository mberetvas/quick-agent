package action

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mberetvas/quick-agent/internal/config"
)

// mockClient is an in-process LLM client for testing — no network required.
type mockClient struct {
	tokens            []string
	healthErr         error
	generateErr       error
	blockUntilCtxDone bool
}

func (m *mockClient) HealthCheck(_ context.Context) error {
	return m.healthErr
}

func (m *mockClient) Generate(ctx context.Context, _ string) (<-chan string, <-chan error, error) {
	if m.generateErr != nil {
		return nil, nil, m.generateErr
	}
	if m.blockUntilCtxDone {
		tokenCh := make(chan string)
		errCh := make(chan error, 1)
		go func() {
			defer close(tokenCh)
			defer close(errCh)
			<-ctx.Done()
		}()
		return tokenCh, errCh, nil
	}
	tokenCh := make(chan string, len(m.tokens))
	errCh := make(chan error, 1)
	for _, t := range m.tokens {
		tokenCh <- t
	}
	close(tokenCh)
	close(errCh)
	return tokenCh, errCh, nil
}

func defaultCfg() *config.Config {
	return config.Default()
}

func TestRunWithClient_RefineStreaming(t *testing.T) {
	mc := &mockClient{tokens: []string{"Hello", " world", "!"}}
	cfg := defaultCfg()
	var buf bytes.Buffer
	result, err := RunWithClient(context.Background(), Options{
		Action: "refine",
		Text:   "hello world",
	}, &buf, mc, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "Hello world!"
	if result != want {
		t.Errorf("result = %q, want %q", result, want)
	}
	if buf.String() != want {
		t.Errorf("output = %q, want %q", buf.String(), want)
	}
}

func TestRunWithClient_VerboseOutput(t *testing.T) {
	mc := &mockClient{tokens: []string{"refined"}}
	cfg := defaultCfg()
	var buf bytes.Buffer
	_, err := RunWithClient(context.Background(), Options{
		Action:  "refine",
		Text:    "test text",
		Verbose: true,
	}, &buf, mc, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "[healthcheck] OK") {
		t.Errorf("expected '[healthcheck] OK' in verbose output, got: %s", out)
	}
	if !strings.Contains(out, "[template] refine") {
		t.Errorf("expected '[template] refine' in verbose output, got: %s", out)
	}
	if !strings.Contains(out, "[prompt]") {
		t.Errorf("expected '[prompt]' in verbose output, got: %s", out)
	}
}

func TestRunWithClient_SanitizationError(t *testing.T) {
	mc := &mockClient{}
	cfg := defaultCfg()
	tests := []struct {
		name string
		text string
	}{
		{"empty", ""},
		{"whitespace", "   \t\n  "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := RunWithClient(context.Background(), Options{
				Action: "refine",
				Text:   tt.text,
			}, &buf, mc, cfg)
			if err == nil {
				t.Fatal("expected error for empty/whitespace input")
			}
			if !strings.Contains(err.Error(), "empty") {
				t.Errorf("error = %v, want empty-related error", err)
			}
		})
	}
}

func TestRunWithClient_InputPrecedence(t *testing.T) {
	// Verifies that opts.Text (resolved by CLI from positional arg) is used directly.
	mc := &mockClient{tokens: []string{"refined result"}}
	cfg := defaultCfg()
	var buf bytes.Buffer
	result, err := RunWithClient(context.Background(), Options{
		Action: "refine",
		Text:   "arg text takes precedence",
	}, &buf, mc, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "refined result" {
		t.Errorf("result = %q, want 'refined result'", result)
	}
}

func TestRunWithClient_HealthcheckFail(t *testing.T) {
	mc := &mockClient{healthErr: errors.New("ollama not running")}
	cfg := defaultCfg()
	var buf bytes.Buffer
	_, err := RunWithClient(context.Background(), Options{
		Action: "refine",
		Text:   "some text",
	}, &buf, mc, cfg)
	if err == nil {
		t.Fatal("expected healthcheck error")
	}
	if !strings.Contains(err.Error(), "healthcheck failed") {
		t.Errorf("error = %v, want 'healthcheck failed'", err)
	}
}

func TestRunWithClient_ContextCancelled(t *testing.T) {
	// blockUntilCtxDone mock blocks until ctx is cancelled, ensuring the stream
	// is in-flight when cancellation fires — no race condition.
	mc := &mockClient{blockUntilCtxDone: true}
	cfg := defaultCfg()
	ctx, cancel := context.WithCancel(context.Background())

	var buf bytes.Buffer
	done := make(chan struct{})
	var runErr error
	go func() {
		defer close(done)
		_, runErr = RunWithClient(ctx, Options{
			Action: "refine",
			Text:   "test text",
		}, &buf, mc, cfg)
	}()

	cancel()
	<-done

	if !errors.Is(runErr, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", runErr)
	}
}

func TestRunWithClient_Truncation(t *testing.T) {
	mc := &mockClient{tokens: []string{"ok"}}
	cfg := defaultCfg()
	cfg.Clipboard.TruncateSize = 5
	var buf bytes.Buffer
	_, err := RunWithClient(context.Background(), Options{
		Action: "refine",
		Text:   "hello world this is long",
	}, &buf, mc, cfg)
	if err != nil {
		t.Fatalf("unexpected error for truncated input: %v", err)
	}
}

func TestRunWithClient_GenerateError(t *testing.T) {
	mc := &mockClient{generateErr: errors.New("backend unavailable")}
	cfg := defaultCfg()
	var buf bytes.Buffer
	_, err := RunWithClient(context.Background(), Options{
		Action: "refine",
		Text:   "some text",
	}, &buf, mc, cfg)
	if err == nil {
		t.Fatal("expected generate error")
	}
	if !strings.Contains(err.Error(), "generate failed") {
		t.Errorf("error = %v, want 'generate failed'", err)
	}
}

func TestRunWithClient_VerboseHealthcheckFail(t *testing.T) {
	mc := &mockClient{healthErr: errors.New("connection refused")}
	cfg := defaultCfg()
	var buf bytes.Buffer
	_, err := RunWithClient(context.Background(), Options{
		Action:  "refine",
		Text:    "some text",
		Verbose: true,
	}, &buf, mc, cfg)
	if err == nil {
		t.Fatal("expected healthcheck error")
	}
	if !strings.Contains(buf.String(), "[healthcheck] FAIL") {
		t.Errorf("expected '[healthcheck] FAIL' in verbose output, got: %s", buf.String())
	}
}

func TestRunWithClient_TranslateDefaultLanguage(t *testing.T) {
	mc := &mockClient{tokens: []string{"Ik ben een dokter"}}
	cfg := defaultCfg()
	cfg.Prompts.TranslateTargetLanguage = "Dutch"

	var buf bytes.Buffer
	_, err := RunWithClient(context.Background(), Options{
		Action:  "translate",
		Text:    "I am a doctor",
		Verbose: true,
	}, &buf, mc, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Dutch") {
		t.Errorf("expected config language 'Dutch' in prompt, got: %s", out)
	}
	if strings.Contains(out, "{{.Language}}") {
		t.Errorf("rendered prompt still contains {{.Language}} placeholder: %s", out)
	}
}

func TestRunWithClient_TranslateLanguageOverride(t *testing.T) {
	mc := &mockClient{tokens: []string{"Bonjour"}}
	cfg := defaultCfg()
	cfg.Prompts.TranslateTargetLanguage = "English"

	var buf bytes.Buffer
	_, err := RunWithClient(context.Background(), Options{
		Action:   "translate",
		Text:     "Hello",
		Language: "French",
		Verbose:  true,
	}, &buf, mc, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "French") {
		t.Errorf("expected overridden language 'French' in prompt, got: %s", out)
	}
	if strings.Contains(out, "English") && !strings.Contains(out, "French") {
		t.Errorf("config language 'English' used instead of override 'French': %s", out)
	}
	if strings.Contains(out, "{{.Language}}") {
		t.Errorf("rendered prompt still contains {{.Language}} placeholder: %s", out)
	}
}

func TestRunWithClient_CopyMock(t *testing.T) {
	mc := &mockClient{tokens: []string{"refined", " result"}}
	cfg := defaultCfg()

	var written string
	mockWriter := func(s string) error {
		written = s
		return nil
	}

	var buf bytes.Buffer
	result, err := RunWithClient(context.Background(), Options{
		Action:          "refine",
		Text:            "some text",
		Copy:            true,
		ClipboardWriter: mockWriter,
	}, &buf, mc, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if written != result {
		t.Errorf("clipboard got %q, want %q", written, result)
	}
	if written != "refined result" {
		t.Errorf("clipboard content = %q, want 'refined result'", written)
	}
}

func TestRunWithClient_CopyWriteError(t *testing.T) {
	mc := &mockClient{tokens: []string{"output"}}
	cfg := defaultCfg()

	mockWriter := func(_ string) error {
		return errors.New("clipboard unavailable")
	}

	var buf bytes.Buffer
	_, err := RunWithClient(context.Background(), Options{
		Action:          "refine",
		Text:            "some text",
		Copy:            true,
		ClipboardWriter: mockWriter,
	}, &buf, mc, cfg)
	if err == nil {
		t.Fatal("expected clipboard write error")
	}
	if !strings.Contains(err.Error(), "clipboard write failed") {
		t.Errorf("error = %v, want 'clipboard write failed'", err)
	}
}

func TestRunWithClient_Summarize(t *testing.T) {
	mc := &mockClient{tokens: []string{"• Point one\n", "• Point two\n"}}
	cfg := defaultCfg()
	var buf bytes.Buffer
	result, err := RunWithClient(context.Background(), Options{
		Action: "summarize",
		Text:   "Long article text here",
	}, &buf, mc, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "• Point one\n• Point two\n" {
		t.Errorf("result = %q", result)
	}
}

func TestRunWithClient_Explain(t *testing.T) {
	mc := &mockClient{tokens: []string{"This code does X"}}
	cfg := defaultCfg()
	var buf bytes.Buffer
	result, err := RunWithClient(context.Background(), Options{
		Action: "explain",
		Text:   "func main() {}",
	}, &buf, mc, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "This code does X" {
		t.Errorf("result = %q, want 'This code does X'", result)
	}
}
