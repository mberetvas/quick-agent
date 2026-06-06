package main

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/mberetvas/quick-agent/internal/config"
	"github.com/mberetvas/quick-agent/internal/llm"
	"github.com/mberetvas/quick-agent/internal/tui"
)

var (
	tuiText string
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the terminal user interface",
	Run: func(cmd *cobra.Command, args []string) {
		var text string

		// 1. Check if we have piped stdin
		stat, _ := os.Stdin.Stat()
		isPiped := (stat.Mode() & os.ModeCharDevice) == 0

		if isPiped {
			// Read all from stdin
			bytes, err := io.ReadAll(os.Stdin)
			if err == nil {
				text = string(bytes)
			}
		}

		// 2. Override with --text flag if specified
		if tuiText != "" {
			text = tuiText
		}

		// 3. Load config for TUI keybindings
		cfg, err := config.LoadWithEnv(configPath)
		if err != nil {
			fmt.Printf("Failed to load configuration: %v\n", err)
			os.Exit(1)
		}

		llmClient, err := llm.NewClientFromConfig(cfg)
		if err != nil {
			fmt.Printf("Failed to create LLM client: %v\n", err)
			os.Exit(1)
		}

		// 4. Initialize Bubble Tea Program
		m := tui.NewModel(text, cfg, llmClient)
		var p *tea.Program

		if isPiped {
			p = tea.NewProgram(m, tea.WithInputTTY())
		} else {
			p = tea.NewProgram(m)
		}

		if _, err := p.Run(); err != nil {
			fmt.Printf("Alas, there's been an error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	tuiCmd.Flags().StringVar(&tuiText, "text", "", "Text to display in TUI")
}
