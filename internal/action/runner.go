// Package action provides an end-to-end action runner that executes LLM prompts
// outside the TUI, intended for CLI subcommands such as "refine".
package action

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/mberetvas/quick-agent/internal/clipboard"
	"github.com/mberetvas/quick-agent/internal/config"
	"github.com/mberetvas/quick-agent/internal/llm"
)

// Options configures a single action invocation.
type Options struct {
	Action          string // refine | translate | summarize | explain
	Text            string
	Language        string // translate only; empty → cfg.Prompts.TranslateTargetLanguage
	Copy            bool
	ClipboardWriter func(string) error // injectable for testing; Run() sets a default when nil and Copy is true
	Verbose         bool
	Backend         string
	ConfigPath      string
	LogLevel        string
}

// Run loads config, creates the LLM client from config, and executes the action.
func Run(ctx context.Context, opts Options, out io.Writer) (string, error) {
	cfg, err := config.LoadWithEnv(opts.ConfigPath)
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}
	if opts.Backend != "" {
		cfg.Backend = opts.Backend
	}
	client, err := llm.NewClientFromConfig(cfg)
	if err != nil {
		return "", err
	}
	if opts.Copy && opts.ClipboardWriter == nil {
		opts.ClipboardWriter = clipboard.SystemClipboard{}.Set
	}
	return RunWithClient(ctx, opts, out, client, cfg)
}

// RunWithClient is the testable core — accepts an injected LLM client and config.
func RunWithClient(ctx context.Context, opts Options, out io.Writer, client llm.LLMClient, cfg *config.Config) (string, error) {
	sanitized := clipboard.Sanitize(opts.Text, cfg.Clipboard.TruncateSize)
	if sanitized == "" {
		return "", fmt.Errorf("input is empty or whitespace-only")
	}

	if err := client.HealthCheck(ctx); err != nil {
		if opts.Verbose {
			fmt.Fprintf(out, "[healthcheck] FAIL: %v\n", err)
		}
		return "", fmt.Errorf("healthcheck failed: %w", err)
	}
	if opts.Verbose {
		fmt.Fprintln(out, "[healthcheck] OK")
	}

	reg := llm.NewPromptRegistry(cfg.Prompts)
	tmpl := reg.Get(opts.Action)

	var rendered string
	if opts.Action == "translate" {
		lang := opts.Language
		if lang == "" {
			lang = cfg.Prompts.TranslateTargetLanguage
		}
		rendered = tmpl.RenderWithOptions(sanitized, map[string]string{"Language": lang})
	} else {
		rendered = tmpl.Render(sanitized)
	}

	if opts.Verbose {
		fmt.Fprintf(out, "[template] %s\n", tmpl.Name)
		fmt.Fprintf(out, "[prompt] %s\n", rendered)
	}

	tokenCh, errCh, err := client.Generate(ctx, rendered)
	if err != nil {
		return "", fmt.Errorf("generate failed: %w", err)
	}

	var sb strings.Builder
	for tokenCh != nil || errCh != nil {
		select {
		case <-ctx.Done():
			return sb.String(), ctx.Err()
		case token, ok := <-tokenCh:
			if !ok {
				tokenCh = nil
				continue
			}
			sb.WriteString(token)
			fmt.Fprint(out, token)
		case streamErr, ok := <-errCh:
			if !ok {
				errCh = nil
				continue
			}
			if streamErr != nil {
				return sb.String(), streamErr
			}
		}
	}
	result := sb.String()
	if opts.Copy && opts.ClipboardWriter != nil {
		if err := opts.ClipboardWriter(result); err != nil {
			return result, fmt.Errorf("clipboard write failed: %w", err)
		}
	}
	return result, nil
}
