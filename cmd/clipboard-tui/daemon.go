package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Start the background clipboard polling daemon",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("daemon: not implemented")
	},
}
