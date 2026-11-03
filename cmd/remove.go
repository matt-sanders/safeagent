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
			prof, err := store.FindByName(args[0])
			if err != nil {
				return err
			}
			if err := refuseIfUsed(deps, "profile", prof.Name, func(proj project.Project) bool { return proj.ProfileID == prof.ID }); err != nil {
				return err
			}
			imgs, err := image.ExistingWithProfile(deps.Docker, prof.ID)
			if err != nil {
				return err
			}
			if err := removeImages(cmd, deps, imgs); err != nil {
				return err
			}
			if err := store.Remove(prof.ID); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Removed profile %s (%s).\n", prof.Name, prof.ID)
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
			if err := refuseIfUsed(deps, "stack", st.Name, func(proj project.Project) bool { return proj.StackID == st.ID }); err != nil {
				return err
			}
			imgs, err := image.ExistingWithStack(deps.Docker, st.ID)
			if err != nil {
				return err
			}
			if err := removeImages(cmd, deps, imgs); err != nil {
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
	for _, proj := range projects {
		if uses(proj) {
			paths = append(paths, "  "+proj.Path)
		}
	}
	if len(paths) == 0 {
		return nil
	}
	return fmt.Errorf("%s %q is used by %d project(s); switch or remove them first:\n%s",
		kind, name, len(paths), strings.Join(paths, "\n"))
}

// removeImages removes each of imgs, reporting it by Name (or raw ID where a
// Name doesn't resolve). Call it before removing the Profile or Stack record
// so its Name still resolves.
func removeImages(cmd *cobra.Command, deps Deps, imgs []image.Existing) error {
	if len(imgs) == 0 {
		return nil
	}
	n, err := loadNames(deps)
	if err != nil {
		return err
	}
	for _, img := range imgs {
		name := image.Name(n.stack(img.StackID), n.profile(img.ProfileID))
		if err := deps.Docker.RemoveImage(img.Ref); err != nil {
			return fmt.Errorf("removing image %s: %w", name, err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Removed image %s.\n", name)
	}
	return nil
}
