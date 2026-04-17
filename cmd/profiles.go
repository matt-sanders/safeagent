package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"safe-claude/profile"

	"github.com/spf13/cobra"
)

var profilesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "Manage profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("could not determine home directory: %w", err)
		}
		profilesPath := filepath.Join(homeDir, ".safe-claude", "profiles.json")

		profiles, err := profile.Load(profilesPath)
		if err != nil {
			return err
		}

		if len(profiles) == 0 {
			fmt.Println("No profiles yet. Create one with: safe-claude profiles create")
			return nil
		}

		for i, p := range profiles {
			fmt.Print(profile.FormatOption(p))
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
