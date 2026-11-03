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
			proj, err := resolveProject(deps, path)
			if err != nil {
				return err
			}
			prof, err := profile.NewStore(deps.ConfigRoot).FindByName(args[0])
			if err != nil {
				return err
			}
			proj.ProfileID = prof.ID
			if err := project.NewStore(deps.ConfigRoot).Update(proj); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Project %s now uses profile %s.\n", proj.Path, prof.Name)
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
			proj, err := resolveProject(deps, path)
			if err != nil {
				return err
			}
			st, err := stack.NewStore(deps.ConfigRoot).FindByName(args[0])
			if err != nil {
				return err
			}
			proj.StackID = st.ID
			if err := project.NewStore(deps.ConfigRoot).Update(proj); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Project %s now uses stack %s.\n", proj.Path, st.Name)
			return nil
		},
	}
	addProjectFlag(c, &path)
	return c
}
