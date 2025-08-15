package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "botcli",
	Short: "Discord Bot CLI - Create and manage your modular Discord bot",
	Long: `A CLI tool for creating modular Discord bots, similar to create-next-app.
Choose modules, configure settings, and generate a custom bot instance.`,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(buildCmd)
}