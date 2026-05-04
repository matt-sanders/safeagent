package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var projectCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a project for the current directory without starting a session",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("could not determine working directory: %w", err)
		}

		existing, found, err := projectService.FindForCwd(cwd)
		if err != nil {
			return err
		}
		if found {
			if existing.Path == cwd {
				fmt.Printf("A project already exists at %s.\n", existing.Path)
			} else {
				fmt.Printf("A project already exists at %s, which covers this directory.\n", existing.Path)
			}
			return nil
		}

		proj, _, err := projectService.GetOrCreate(cwd)
		if err != nil {
			return err
		}
		fmt.Printf("Created project at %s.\n", proj.Path)

		if _, _, err := promptAndSetProfile(proj.Path); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	projectCmd.AddCommand(projectCreateCmd)
}
