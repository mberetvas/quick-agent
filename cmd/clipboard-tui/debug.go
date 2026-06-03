package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/yourname/clipboard-tui/internal/clipboard"
	"github.com/yourname/clipboard-tui/internal/config"
	"github.com/yourname/clipboard-tui/internal/hotkey"
	"github.com/yourname/clipboard-tui/internal/llm"
	"github.com/yourname/clipboard-tui/internal/llm/ollama"
	"github.com/yourname/clipboard-tui/internal/llm/openrouter"
	"github.com/yourname/clipboard-tui/internal/terminal"
)

var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Diagnostics and developer utilities",
}

var watchClipboardCmd = &cobra.Command{
	Use:   "watch-clipboard",
	Short: "Start the clipboard poller and stream any changes to stdout",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load the config (loads from CLI option or default path)
		cfg, err := config.LoadWithEnv(configPath)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		if cmd.Flags().Changed("log-level") {
			cfg.Logging.Level = logLevel
		}

		fmt.Printf("Starting clipboard watch (Poll Interval: %dms, Max Size: %d chars)...\n",
			cfg.Clipboard.PollIntervalMS, cfg.Clipboard.MaxSize)
		fmt.Println("Press Ctrl+C to stop.")

		sysCB := clipboard.SystemClipboard{}
		poller := clipboard.NewPoller(sysCB, cfg)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Wire OS signal listener for clean exit
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigs
			fmt.Println("\nStopping clipboard watch gracefully...")
			cancel()
		}()

		// Start poller
		poller.Start(ctx)

		// Loop consuming poller channels
		for {
			select {
			case <-ctx.Done():
				return nil
			case text, ok := <-poller.Changes():
				if !ok {
					return nil
				}
				fmt.Printf("\n--- CLIPBOARD CHANGE DETECTED ---\n%s\n---------------------------------\n", text)
			case err, ok := <-poller.Errors():
				if !ok {
					return nil
				}
				fmt.Fprintf(os.Stderr, "Poller Error: %v\n", err)
			}
		}
	},
}

var (
	debugLLMTemplate string
	debugLLMBackend  string
)

var testLLMCmd = &cobra.Command{
	Use:   "llm <text>",
	Short: "Send a sample text prompt directly to the LLM and stream the response",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithEnv(configPath)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		inputText := strings.Join(args, " ")

		backend := cfg.Backend
		if cmd.Flags().Changed("backend") {
			backend = debugLLMBackend
		}

		var client llm.LLMClient
		var healthTarget string
		switch backend {
		case "ollama":
			client = ollama.NewClient(cfg.Ollama)
			healthTarget = cfg.Ollama.URL
		case "openrouter":
			client = openrouter.NewClient(cfg.OpenRouter, cfg.LLM)
			healthTarget = "https://openrouter.ai/api/v1"
		default:
			return fmt.Errorf("unsupported backend: %s (use ollama or openrouter)", backend)
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Intercept OS Signals to allow Clean Exit
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigs
			fmt.Println("\nCanceling LLM session stream...")
			cancel()
		}()

		// 1. Healthcheck verification first
		fmt.Printf("Performing healthcheck on backend '%s' (%s)...\n", backend, healthTarget)
		if err := client.HealthCheck(ctx); err != nil {
			return fmt.Errorf("pre-generation LLM healthcheck failed: %w", err)
		}
		fmt.Println("Healthcheck OK!")

		// 2. Render Prompt
		registry := llm.NewPromptRegistry()
		template := registry.Get(debugLLMTemplate)
		renderedPrompt := template.Render(inputText)

		fmt.Printf("\nUsing Prompt Template: '%s'\n", template.Name)
		fmt.Println("--- RENDERED PROMPT ---")
		fmt.Println(renderedPrompt)
		fmt.Println("-----------------------")
		fmt.Println("Streaming response:")

		// 3. Generate stream
		tokens, errs, err := client.Generate(ctx, renderedPrompt)
		if err != nil {
			return fmt.Errorf("generation failed: %w", err)
		}

		for {
			select {
			case <-ctx.Done():
				return nil
			case token, ok := <-tokens:
				if !ok {
					fmt.Println() // final newline
					return nil
				}
				fmt.Print(token)
				// Flush stdout immediately for responsive token streams
				os.Stdout.Sync()
			case err, ok := <-errs:
				if ok && err != nil {
					fmt.Printf("\nStream Error: %v\n", err)
					return err
				}
			}
		}
	},
}

var debugHotkeyCmd = &cobra.Command{
	Use:   "hotkey",
	Short: "Listen for hotkey press and print confirmation",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithEnv(configPath)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		fmt.Printf("Listening for hotkey: %s+%s (Debounce: %dms)...\n",
			strings.Join(cfg.Hotkey.Modifiers, "+"),
			cfg.Hotkey.Key,
			cfg.Hotkey.DebounceMS)
		fmt.Println("Press Ctrl+C to stop.")

		listener := hotkey.NewListener(cfg.Hotkey)
		events := make(chan struct{}, 1)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Wire OS signal listener for clean exit
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(sigs)
		go func() {
			<-sigs
			fmt.Println("\nStopping hotkey listener gracefully...")
			cancel()
		}()

		errCh := make(chan error, 1)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := listener.Start(ctx, events); err != nil {
				errCh <- err
			}
		}()

		for {
			select {
			case err := <-errCh:
				return err
			case <-events:
				fmt.Println("Hotkey pressed!")
			case <-ctx.Done():
				wg.Wait()
				return nil
			}
		}
	},
}

var (
	spawnTerminalCommand string
	spawnTerminalEmulator string
)

var debugSpawnTerminalCmd = &cobra.Command{
	Use:   "spawn-terminal",
	Short: "Open a new terminal window running a shell command",
	RunE: func(cmd *cobra.Command, args []string) error {
		if spawnTerminalCommand == "" {
			return fmt.Errorf("--command is required")
		}

		cfg, err := config.LoadWithEnv(configPath)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		if spawnTerminalEmulator != "" {
			cfg.Terminal.Emulator = spawnTerminalEmulator
			if err := cfg.Validate(); err != nil {
				return fmt.Errorf("invalid --emulator: %w", err)
			}
		}

		inner := terminal.BuildDebugInnerCommand(spawnTerminalCommand)
		spawner := terminal.NewSpawner(cfg.Terminal)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		profileID, err := spawner.Spawn(ctx, inner)
		if err != nil {
			return fmt.Errorf("spawn terminal: %w", err)
		}

		fmt.Printf("Spawned using terminal profile: %s\n", profileID)
		return nil
	},
}

func init() {
	testLLMCmd.Flags().StringVar(&debugLLMTemplate, "template", "refine", "Select prompt template: refine, translate, summarize, explain, custom")
	testLLMCmd.Flags().StringVar(&debugLLMBackend, "backend", "", "Override config backend (ollama or openrouter)")
	debugSpawnTerminalCmd.Flags().StringVar(&spawnTerminalCommand, "command", "", "Shell one-liner to run in the new terminal (required)")
	debugSpawnTerminalCmd.Flags().StringVar(&spawnTerminalEmulator, "emulator", "", "Override terminal.emulator for this invocation")
	_ = debugSpawnTerminalCmd.MarkFlagRequired("command")

	debugCmd.AddCommand(watchClipboardCmd)
	debugCmd.AddCommand(testLLMCmd)
	debugCmd.AddCommand(debugHotkeyCmd)
	debugCmd.AddCommand(debugSpawnTerminalCmd)
	rootCmd.AddCommand(debugCmd)
}
