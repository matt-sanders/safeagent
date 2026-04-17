package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var profilesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "Manage profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		profiles, err := profileService.List()
		if err != nil {
			return err
		}

		if len(profiles) == 0 {
			fmt.Println("No profiles yet. Create one with: safe-claude profiles create")
			return nil
		}

		for i, p := range profiles {
			fmt.Print(profileService.FormatOption(p))
			if i < len(profiles)-1 {
				fmt.Println()
			}
		}
		fmt.Println()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(profilesCmd)
}
