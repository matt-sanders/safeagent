// cmd/use_profile.go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var useProfileCmd = &cobra.Command{
	Use:   "use-profile [profile-name]",
	Short: "Change or set the profile for the current project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("you called use-profile")
	},
}

func init() {
	rootCmd.AddCommand(useProfileCmd)
}
