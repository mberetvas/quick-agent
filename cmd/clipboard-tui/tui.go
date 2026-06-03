package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the terminal user interface",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("tui: not implemented")
	},
}
