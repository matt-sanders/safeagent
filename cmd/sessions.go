package cmd

import "github.com/spf13/cobra"

var sessionsCmd = &cobra.Command{
	Use:   "sessions",
	Short: "Manage running sessions",
}

func init() {
	rootCmd.AddCommand(sessionsCmd)
}
