package cmd

import (
	"fmt"
	"strings"

	"charm.land/huh/v2"
	"github.com/spf13/cobra"

	"safeagent/auth"
)

var authCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new auth identity",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var name string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Identity name").
					Value(&name).
					Validate(func(s string) error {
						return auth.ValidateName(strings.TrimSpace(s))
					}),
			),
		)
		if err := form.Run(); err != nil {
			return fmt.Errorf("form cancelled: %w", err)
		}

		name = strings.TrimSpace(name)
		if err := authService.Create(name); err != nil {
			return err
		}
		fmt.Printf("Identity %q created. It will prompt for login on the next session that uses it.\n", name)
		return nil
	},
}

func init() {
	authCmd.AddCommand(authCreateCmd)
}
