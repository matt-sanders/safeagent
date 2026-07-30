package cmd

import (
	"fmt"
	"os"

	"charm.land/huh/v2"
	"github.com/spf13/cobra"

	"safeagent/auth"
)

var useAuthCmd = &cobra.Command{
	Use:   "use-auth",
	Short: "Change or set the auth identity for the current project",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("could not determine working directory: %w", err)
		}

		proj, _, err := projectService.FindOrCreateForCwd(cwd)
		if err != nil {
			return err
		}

		identities, err := authService.List()
		if err != nil {
			return err
		}

		currentName := auth.Resolve(proj.AuthName)
		options := make([]huh.Option[string], len(identities))
		for i, id := range identities {
			prefix := "  "
			if id.Name == currentName {
				prefix = "* "
			}
			status := "logged in"
			if !id.LoggedIn {
				status = "not logged in"
			}
			label := fmt.Sprintf("%s%s %s", prefix, id.Name, useAuthDetailStyle.Render("("+status+")"))
			options[i] = huh.NewOption(label, id.Name)
		}

		selected := currentName
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select an auth identity").
					Options(options...).
					Value(&selected),
			),
		)
		if err := form.Run(); err != nil {
			return nil
		}

		if err := projectService.SetAuth(proj.Path, selected); err != nil {
			return err
		}

		fmt.Printf("Now using identity %q\n", selected)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(useAuthCmd)
}
