package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var profilesRmCmd = &cobra.Command{
	Use:   "rm [profile-id]",
	Short: "Remove a profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		if err := profileService.Remove(id); err != nil {
			return err
		}
		fmt.Printf("Profile %q removed.\n", id)
		return nil
	},
}

func init() {
	profilesCmd.AddCommand(profilesRmCmd)
}
