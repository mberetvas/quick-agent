package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yourname/clipboard-tui/internal/config"
	"golang.org/x/term"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Print the resolved configuration as JSON",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithEnv(configPath)
		if err != nil {
			return err
		}

		if cmd.Flags().Changed("log-level") {
			cfg.Logging.Level = logLevel
		}

		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return err
		}

		fmt.Println(string(data))
		return nil
	},
}

var setKeyCmd = &cobra.Command{
	Use:   "set-key <backend>",
	Short: "Securely store the API key for a backend in the system keyring",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		backend := strings.ToLower(args[0])
		if backend != "openrouter" && backend != "ollama" {
			return fmt.Errorf("invalid backend '%s': must be openrouter or ollama", backend)
		}

		var secret string
		var err error

		// Reads the secret from stdin or secure prompt
		if term.IsTerminal(int(os.Stdin.Fd())) {
			fmt.Printf("Enter API key for %s: ", backend)
			byteSecret, err := term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Println() // Print newline after user enters password
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}
			secret = strings.TrimSpace(string(byteSecret))
		} else {
			// Read from piped standard input
			reader := bufio.NewReader(os.Stdin)
			input, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read from standard input: %w", err)
			}
			secret = strings.TrimSpace(input)
		}

		if secret == "" {
			return errors.New("cannot set an empty API key")
		}

		err = config.SaveAPIKey(backend, secret)
		if err != nil {
			return fmt.Errorf("failed to save API key to keyring: %w", err)
		}

		fmt.Printf("Successfully saved API key for %s to system keyring\n", backend)
		return nil
	},
}

var getKeyCmd = &cobra.Command{
	Use:   "get-key <backend>",
	Short: "Retrieve the API key for a backend from the system keyring",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		backend := strings.ToLower(args[0])
		if backend != "openrouter" && backend != "ollama" {
			return fmt.Errorf("invalid backend '%s': must be openrouter or ollama", backend)
		}

		val, err := config.GetAPIKey(backend)
		if err != nil {
			fmt.Printf("API key not set for backend '%s'\n", backend)
			return nil
		}

		fmt.Println(val)
		return nil
	},
}

func init() {
	configCmd.AddCommand(showCmd)
	configCmd.AddCommand(setKeyCmd)
	configCmd.AddCommand(getKeyCmd)
}
