package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"safe-claude/profile"

	"charm.land/huh/v2"
	"github.com/spf13/cobra"
)

var profilesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("could not determine home directory: %w", err)
		}
		profilesPath := filepath.Join(homeDir, ".safe-claude", "profiles.json")

		var name string
		var nodeVersion string

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Profile name").
					Value(&name).
					Validate(func(s string) error {
						if strings.TrimSpace(s) == "" {
							return fmt.Errorf("name cannot be empty")
						}
						return nil
					}),
				huh.NewInput().
					Title("Node.js version").
					Value(&nodeVersion).
					Validate(func(s string) error {
						if strings.TrimSpace(s) == "" {
							return fmt.Errorf("node version cannot be empty")
						}
						return nil
					}),
			),
		)

		err = form.Run()
		if err != nil {
			return fmt.Errorf("form cancelled: %w", err)
		}

		p := profile.Profile{
			Name:        strings.TrimSpace(name),
			NodeVersion: strings.TrimSpace(nodeVersion),
		}

		if err := profile.Add(profilesPath, p); err != nil {
			return err
		}

		fmt.Printf("Profile %q created.\n", p.Name)
		return nil
	},
}

func init() {
	profilesCmd.AddCommand(profilesCreateCmd)
}
