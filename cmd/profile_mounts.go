package cmd

import (
	"fmt"

	"safeagent/profile"
	"safeagent/table"

	"charm.land/huh/v2"
	"github.com/spf13/cobra"
)

func newProfileMountsCommand(deps Deps) *cobra.Command {
	c := &cobra.Command{
		Use:   "mounts",
		Short: "Manage a Profile's Mounts (host paths shown in every Session)",
		Args:  cobra.NoArgs,
	}
	c.AddCommand(newProfileMountsLsCommand(deps), newProfileMountsAddCommand(deps), newProfileMountsRmCommand(deps))
	return c
}

func newProfileMountsLsCommand(deps Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "ls <name>",
		Short: "List a Profile's Mounts",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := profile.NewStore(deps.ConfigRoot).FindByName(args[0])
			if err != nil {
				return err
			}
			rows := make([][]string, 0, len(p.Mounts))
			for _, m := range p.Mounts {
				rows = append(rows, []string{m.HostPath, m.ContainerPath})
			}
			return table.Print(cmd.OutOrStdout(), "No mounts.", []string{"HOST PATH", "CONTAINER PATH"}, rows)
		},
	}
}

func newProfileMountsAddCommand(deps Deps) *cobra.Command {
	var host, container string
	c := &cobra.Command{
		Use:   "add <name>",
		Short: "Add a Mount to a Profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store := profile.NewStore(deps.ConfigRoot)
			p, err := store.FindByName(args[0])
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Profile folder (for reference): %s\n", store.Dir(p.ID))

			var fields []huh.Field
			if !cmd.Flags().Changed("host") {
				fields = append(fields, huh.NewInput().Title("Host path").Value(&host))
			}
			if !cmd.Flags().Changed("container") {
				fields = append(fields, huh.NewInput().Title("Container path").Value(&container))
			}
			if len(fields) > 0 {
				err := huh.NewForm(huh.NewGroup(fields...)).
					WithInput(cmd.InOrStdin()).WithOutput(cmd.ErrOrStderr()).Run()
				if err != nil {
					return err
				}
			}

			m, err := store.AddMount(p.Name, profile.Mount{HostPath: host, ContainerPath: container})
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "Added mount %s -> %s to profile %s.\n", m.HostPath, m.ContainerPath, p.Name)
			return nil
		},
	}
	c.Flags().StringVar(&host, "host", "", "Host path to mount (must exist)")
	c.Flags().StringVar(&container, "container", "", "Absolute path inside the container")
	return c
}

func newProfileMountsRmCommand(deps Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "rm <name> <container-path>",
		Short: "Remove a Profile's Mount by its container path",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := profile.NewStore(deps.ConfigRoot).RemoveMount(args[0], args[1]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Removed mount %s from profile %s.\n", args[1], args[0])
			return nil
		},
	}
}
