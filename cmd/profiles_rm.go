// cmd/profiles_rm.go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var profilesRmCmd = &cobra.Command{
	Use:   "rm [profile-name]",
	Short: "Remove a profile",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("you called profiles rm")
	},
}

func init() {
	profilesCmd.AddCommand(profilesRmCmd)
}
