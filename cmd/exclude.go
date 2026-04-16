// cmd/exclude.go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var excludeCmd = &cobra.Command{
	Use:   "exclude [path]",
	Short: "Exclude a directory or file from the project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("you called exclude")
	},
}

func init() {
	rootCmd.AddCommand(excludeCmd)
}
