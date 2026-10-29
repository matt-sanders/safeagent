package cmd

import (
	"fmt"
	"strconv"

	"safeagent/ident"
	"safeagent/profile"
	"safeagent/table"

	"charm.land/huh/v2"
	"github.com/spf13/cobra"
)

func newProfileCommand(deps Deps) *cobra.Command {
	c := &cobra.Command{
		Use:   "profile",
		Short: "Manage Profiles (the agent's identity)",
		Args:  cobra.NoArgs,
	}
	c.AddCommand(newProfileCreateCommand(deps), newProfileLsCommand(deps))
	return c
}

func newProfileCreateCommand(deps Deps) *cobra.Command {
	var name string
	c := &cobra.Command{
		Use:   "create",
		Short: "Create a Profile",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed("name") {
				err := huh.NewForm(huh.NewGroup(
					huh.NewInput().Title("Profile name").Value(&name).Validate(ident.ValidateName),
				)).WithInput(cmd.InOrStdin()).WithOutput(cmd.ErrOrStderr()).Run()
				if err != nil {
					return err
				}
			}
			p, err := profile.NewStore(deps.ConfigRoot).Create(name)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Created profile %s (%s).\n", p.Name, p.ID)
			return nil
		},
	}
	c.Flags().StringVar(&name, "name", "", "Profile name (lowercase letters, digits, '-' and '_')")
	return c
}

func newProfileLsCommand(deps Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "ls",
		Short: "List Profiles",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			profiles, err := profile.NewStore(deps.ConfigRoot).List()
			if err != nil {
				return err
			}
			rows := make([][]string, 0, len(profiles))
			for _, p := range profiles {
				// PROJECTS stays 0 until Projects exist.
				rows = append(rows, []string{p.Name, p.ID, strconv.Itoa(len(p.Mounts)), "0"})
			}
			return table.Print(cmd.OutOrStdout(), "No profiles.", []string{"NAME", "ID", "MOUNTS", "PROJECTS"}, rows)
		},
	}
}
