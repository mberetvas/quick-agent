package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yourname/clipboard-tui/internal/config"
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

func init() {
	configCmd.AddCommand(showCmd)
}
