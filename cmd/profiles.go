// cmd/profiles.go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var profilesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "Manage profiles",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("you called profiles")
	},
}

func init() {
	rootCmd.AddCommand(profilesCmd)
}
