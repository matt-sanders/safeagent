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
			return printMounts(cmd.OutOrStdout(), mountRows(prof.Mounts))
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
			m, err := readMount(cmd, deps, host, container)
			if err != nil {
				return err
			}
			m, err = store.AddMount(prof.ID, m)
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "Added mount %s -> %s to profile %s.\n", m.HostPath, m.ContainerPath, prof.Name)
			return nil
		},
	}
	addMountFlags(c, &host, &container)
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

// addMountFlags registers the --host and --container flags of `mounts add`.
func addMountFlags(c *cobra.Command, host, container *string) {
	c.Flags().StringVar(host, "host", "", "Host path to mount (must exist; relative to the current directory)")
	c.Flags().StringVar(container, "container", "", "Absolute path inside the container")
}

// readMount builds the Mount for `mounts add` from the --host and --container
// values, prompting for whichever flag wasn't passed. A relative host path is
// resolved against the current directory.
func readMount(cmd *cobra.Command, deps Deps, host, container string) (profile.Mount, error) {
	fmt.Fprintf(cmd.OutOrStdout(), "Container home directory: /home/developer\n")

	var fields []huh.Field
	if !cmd.Flags().Changed("host") {
		fields = append(fields, huh.NewInput().Title("Host path").Value(&host))
	}
	if !cmd.Flags().Changed("container") {
		fields = append(fields, huh.NewInput().Title("Container path").Value(&container))
	}
	if len(fields) > 0 {
		if err := runForm(cmd, fields...); err != nil {
			return profile.Mount{}, err
		}
	}
	if host != "" && !filepath.IsAbs(host) {
		cwd, err := deps.cwd()
		if err != nil {
			return profile.Mount{}, err
		}
		host = filepath.Join(cwd, host)
	}
	return profile.Mount{HostPath: host, ContainerPath: container}, nil
}

// mountRows converts stored Mounts to printMounts rows.
func mountRows(mounts []profile.Mount) [][]string {
	rows := make([][]string, len(mounts))
	for i, m := range mounts {
		rows[i] = []string{m.HostPath, m.ContainerPath}
	}
	return rows
}

// printMounts prints rows of (host path, container path) as the Mounts table
// shared by `profile mounts ls`, `project mounts ls` and `project inspect`.
func printMounts(w io.Writer, rows [][]string) error {
	return table.Print(w, "No mounts.", []string{"HOST PATH", "CONTAINER PATH"}, rows)
}
