package cmd

import (
	"fmt"
	"strings"

	"safeagent/image"
	"safeagent/profile"
	"safeagent/project"
	"safeagent/stack"

	"github.com/spf13/cobra"
)

func newProfileRmCommand(deps Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "rm <name>",
		Short: "Remove a Profile no Project uses, with its folder and Images",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store := profile.NewStore(deps.ConfigRoot)
			p, err := store.FindByName(args[0])
			if err != nil {
				return err
			}
			if err := refuseIfUsed(deps, "profile", p.Name, func(pr project.Project) bool { return pr.ProfileID == p.ID }); err != nil {
				return err
			}
			if err := removeImages(cmd, deps, func() ([]string, error) { return image.ListForProfile(deps.Docker, p.ID) }); err != nil {
				return err
			}
			if err := store.Remove(p.ID); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Removed profile %s (%s).\n", p.Name, p.ID)
			return nil
		},
	}
}

func newStackRmCommand(deps Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "rm <name>",
		Short: "Remove a Stack no Project uses, with its folder and Images",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store := stack.NewStore(deps.ConfigRoot)
			st, err := store.FindByName(args[0])
			if err != nil {
				return err
			}
			if err := refuseIfUsed(deps, "stack", st.Name, func(pr project.Project) bool { return pr.StackID == st.ID }); err != nil {
				return err
			}
			if err := removeImages(cmd, deps, func() ([]string, error) { return image.ListForStack(deps.Docker, st.ID) }); err != nil {
				return err
			}
			if err := store.Remove(st.ID); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Removed stack %s (%s).\n", st.Name, st.ID)
			return nil
		},
	}
}

// refuseIfUsed errors, listing the Project paths, when any Project matches.
func refuseIfUsed(deps Deps, kind, name string, uses func(project.Project) bool) error {
	projects, err := project.NewStore(deps.ConfigRoot).List()
	if err != nil {
		return err
	}
	var paths []string
	for _, pr := range projects {
		if uses(pr) {
			paths = append(paths, "  "+pr.Path)
		}
	}
	if len(paths) == 0 {
		return nil
	}
	return fmt.Errorf("%s %q is used by %d project(s); switch or remove them first:\n%s",
		kind, name, len(paths), strings.Join(paths, "\n"))
}

// removeImages removes every Image the list function returns, reporting each.
func removeImages(cmd *cobra.Command, deps Deps, list func() ([]string, error)) error {
	refs, err := list()
	if err != nil {
		return err
	}
	for _, ref := range refs {
		if err := deps.Docker.RemoveImage(ref); err != nil {
			return fmt.Errorf("removing image %s: %w", ref, err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Removed image %s.\n", ref)
	}
	return nil
}
