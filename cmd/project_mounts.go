package cmd

import (
	"fmt"

	"safeagent/project"

	"github.com/spf13/cobra"
)

func newProjectMountsCommand(deps Deps) *cobra.Command {
	c := &cobra.Command{
		Use:   "mounts",
		Short: "Manage a Project's Mounts (host paths shown in its Sessions)",
		Args:  cobra.NoArgs,
	}
	c.AddCommand(newProjectMountsLsCommand(deps), newProjectMountsAddCommand(deps), newProjectMountsRmCommand(deps))
	return c
}

func newProjectMountsLsCommand(deps Deps) *cobra.Command {
	var path string
	c := &cobra.Command{
		Use:   "ls",
		Short: "List the Project's Mounts",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := resolveProject(deps, path)
			if err != nil {
				return err
			}
			return printMounts(cmd.OutOrStdout(), mountRows(p.Mounts))
		},
	}
	addProjectFlag(c, &path)
	return c
}

func newProjectMountsAddCommand(deps Deps) *cobra.Command {
	var path, host, container string
	c := &cobra.Command{
		Use:   "add",
		Short: "Add a Mount to the Project",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := resolveProject(deps, path)
			if err != nil {
				return err
			}
			m, err := readMount(cmd, deps, host, container)
			if err != nil {
				return err
			}
			m, err = project.NewStore(deps.ConfigRoot).AddMount(p.Path, m)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Added mount %s -> %s to project %s.\n", m.HostPath, m.ContainerPath, p.Path)
			return nil
		},
	}
	addProjectFlag(c, &path)
	addMountFlags(c, &host, &container)
	return c
}

func newProjectMountsRmCommand(deps Deps) *cobra.Command {
	var path string
	c := &cobra.Command{
		Use:   "rm <container-path>",
		Short: "Remove a Project's Mount by its container path",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := resolveProject(deps, path)
			if err != nil {
				return err
			}
			if err := project.NewStore(deps.ConfigRoot).RemoveMount(p.Path, args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Removed mount %s from project %s.\n", args[0], p.Path)
			return nil
		},
	}
	addProjectFlag(c, &path)
	return c
}
