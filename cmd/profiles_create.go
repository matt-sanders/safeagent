// cmd/profiles_create.go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var profilesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new profile",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("you called profiles create")
	},
}

func init() {
	profilesCmd.AddCommand(profilesCreateCmd)
}
