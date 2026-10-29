package cmd

import (
	"fmt"
	"io"
	"path/filepath"

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
			prof, err := profile.NewStore(deps.ConfigRoot).FindByName(args[0])
			if err != nil {
				return err
			}
			rows := make([][]string, len(prof.Mounts))
			for i, m := range prof.Mounts {
				rows[i] = []string{m.HostPath, m.ContainerPath}
			}
			return printMounts(cmd.OutOrStdout(), rows)
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
			prof, err := store.FindByName(args[0])
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Profile folder (for reference): %s\n", store.Dir(prof.ID))

			var fields []huh.Field
			if !cmd.Flags().Changed("host") {
				fields = append(fields, huh.NewInput().Title("Host path").Value(&host))
			}
			if !cmd.Flags().Changed("container") {
				fields = append(fields, huh.NewInput().Title("Container path").Value(&container))
			}
			if len(fields) > 0 {
				if err := runForm(cmd, fields...); err != nil {
					return err
				}
			}
			if host != "" && !filepath.IsAbs(host) {
				cwd, err := deps.cwd()
				if err != nil {
					return err
				}
				host = filepath.Join(cwd, host)
			}

			m, err := store.AddMount(prof.ID, profile.Mount{HostPath: host, ContainerPath: container})
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "Added mount %s -> %s to profile %s.\n", m.HostPath, m.ContainerPath, prof.Name)
			return nil
		},
	}
	c.Flags().StringVar(&host, "host", "", "Host path to mount (must exist; relative to the current directory)")
	c.Flags().StringVar(&container, "container", "", "Absolute path inside the container")
	return c
}

func newProfileMountsRmCommand(deps Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "rm <name> <container-path>",
		Short: "Remove a Profile's Mount by its container path",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			store := profile.NewStore(deps.ConfigRoot)
			prof, err := store.FindByName(args[0])
			if err != nil {
				return err
			}
			if err := store.RemoveMount(prof.ID, args[1]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Removed mount %s from profile %s.\n", args[1], prof.Name)
			return nil
		},
	}
}

// printMounts prints rows of (host path, container path) as the Mounts table
// shared by `profile mounts ls` and `project inspect`.
func printMounts(w io.Writer, rows [][]string) error {
	return table.Print(w, "No mounts.", []string{"HOST PATH", "CONTAINER PATH"}, rows)
}
