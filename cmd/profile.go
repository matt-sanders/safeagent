package cmd

import (
	"fmt"
	"strconv"

	"safeagent/ident"
	"safeagent/profile"
	"safeagent/project"
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
	c.AddCommand(newProfileCreateCommand(deps), newProfileLsCommand(deps), newProfileRmCommand(deps), newProfileDockerfileCommand(deps), newProfileRebuildCommand(deps))
	c.AddCommand(newProfileMountsCommand(deps))
	return c
}

func newProfileCreateCommand(deps Deps) *cobra.Command {
	var name string
	c := &cobra.Command{
		Use:   "create",
		Short: "Create a Profile",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := createProfile(cmd, deps, name, !cmd.Flags().Changed("name"))
			return err
		},
	}
	c.Flags().StringVar(&name, "name", "", "Profile name (lowercase letters, digits, '-' and '_')")
	return c
}

// createProfile is the `profile create` flow: it asks for the Name when ask is
// true (otherwise uses name), creates the Profile and reports it. The Profile
// picker's "Create new…" option runs it too.
func createProfile(cmd *cobra.Command, deps Deps, name string, ask bool) (profile.Profile, error) {
	if ask {
		err := huh.NewForm(huh.NewGroup(
			huh.NewInput().Title("Profile name").Value(&name).Validate(ident.ValidateName),
		)).WithInput(cmd.InOrStdin()).WithOutput(cmd.ErrOrStderr()).Run()
		if err != nil {
			return profile.Profile{}, err
		}
	}
	p, err := profile.NewStore(deps.ConfigRoot).Create(name)
	if err != nil {
		return profile.Profile{}, err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Created profile %s (%s).\n", p.Name, p.ID)
	return p, nil
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
			counts, err := project.NewStore(deps.ConfigRoot).CountByProfile()
			if err != nil {
				return err
			}
			rows := make([][]string, 0, len(profiles))
			for _, p := range profiles {
				rows = append(rows, []string{p.Name, p.ID, strconv.Itoa(len(p.Mounts)), strconv.Itoa(counts[p.ID])})
			}
			return table.Print(cmd.OutOrStdout(), "No profiles.", []string{"NAME", "ID", "MOUNTS", "PROJECTS"}, rows)
		},
	}
}
