package cmd

import (
	"fmt"
	"strings"

	"charm.land/huh/v2"
	"github.com/spf13/cobra"
)

var profilesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		var name string
		var nodeVersion string
		var bunVersion string

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
				huh.NewInput().
					Title("Bun version (leave empty to skip)").
					Value(&bunVersion),
			),
		)

		if err := form.Run(); err != nil {
			return fmt.Errorf("form cancelled: %w", err)
		}

		p, err := profileService.Create(strings.TrimSpace(name), strings.TrimSpace(nodeVersion), strings.TrimSpace(bunVersion))
		if err != nil {
			return err
		}

		fmt.Printf("Profile %q created.\n", p.Name)
		return nil
	},
}

func init() {
	profilesCmd.AddCommand(profilesCreateCmd)
}
