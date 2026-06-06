package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/yourname/clipboard-tui/internal/config"
	"github.com/yourname/clipboard-tui/internal/daemon"
)

var daemonStop bool

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Start the background clipboard polling daemon",
	Long:  "Runs in the foreground, polls the clipboard, listens for the hotkey, and spawns the TUI.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithEnv(configPath)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		if cmd.Flags().Changed("log-level") {
			cfg.Logging.Level = logLevel
		}

		if daemonStop {
			if err := daemon.Stop(cfg.Daemon.PIDFile); err != nil {
				return err
			}
			fmt.Println("Daemon stop signal sent.")
			return nil
		}

		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		if err := daemon.Run(ctx, cfg); err != nil && err != context.Canceled {
			return err
		}
		return nil
	},
}

func init() {
	daemonCmd.Flags().BoolVar(&daemonStop, "stop", false, "Signal the running daemon to stop")
}
