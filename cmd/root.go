package cmd

import (
	"fmt"
	"os"

	"safe-claude/config"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "safe-claude",
	Short: "Run Claude Code safely inside Docker",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("could not determine home directory: %w", err)
		}
		_, err = config.EnsureConfigDir(homeDir)
		if err != nil {
			return err
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		startSession()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
