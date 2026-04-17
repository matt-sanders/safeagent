package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"safe-claude/profile"

	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"
)

var (
	nameStyle   = lipgloss.NewStyle().Bold(true)
	detailStyle = lipgloss.NewStyle().Faint(true)
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
			fmt.Println(nameStyle.Render(p.Name))
			fmt.Println(detailStyle.Render(fmt.Sprintf("  id: %s  node: %s", p.ID, p.NodeVersion)))
			if i < len(profiles)-1 {
				fmt.Println()
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(profilesCmd)
}
