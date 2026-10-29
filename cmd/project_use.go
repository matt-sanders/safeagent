package cmd

import (
	"fmt"

	"safeagent/profile"
	"safeagent/project"
	"safeagent/stack"

	"github.com/spf13/cobra"
)

func newProjectUseProfileCommand(deps Deps) *cobra.Command {
	var path string
	c := &cobra.Command{
		Use:   "use-profile <name>",
		Short: "Switch the Project's Profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := resolveProject(deps, path)
			if err != nil {
				return err
			}
			pr, err := profile.NewStore(deps.ConfigRoot).FindByName(args[0])
			if err != nil {
				return err
			}
			p.ProfileID = pr.ID
			if err := project.NewStore(deps.ConfigRoot).Update(p); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Project %s now uses profile %s.\n", p.Path, pr.Name)
			return nil
		},
	}
	addProjectFlag(c, &path)
	return c
}

func newProjectUseStackCommand(deps Deps) *cobra.Command {
	var path string
	c := &cobra.Command{
		Use:   "use-stack <name>",
		Short: "Switch the Project's Stack",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := resolveProject(deps, path)
			if err != nil {
				return err
			}
			st, err := stack.NewStore(deps.ConfigRoot).FindByName(args[0])
			if err != nil {
				return err
			}
			p.StackID = st.ID
			if err := project.NewStore(deps.ConfigRoot).Update(p); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Project %s now uses stack %s.\n", p.Path, st.Name)
			return nil
		},
	}
	addProjectFlag(c, &path)
	return c
}
