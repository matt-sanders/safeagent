package cmd

import (
	"errors"
	"fmt"
	"strconv"

	"safeagent/image"
	"safeagent/profile"
	"safeagent/project"
	"safeagent/stack"
	"safeagent/table"

	"github.com/spf13/cobra"
)

func newImageCommand(deps Deps) *cobra.Command {
	c := &cobra.Command{
		Use:   "image",
		Short: "Manage Images (built Stack and Profile pairs)",
		Args:  cobra.NoArgs,
	}
	c.AddCommand(newImageLsCommand(deps), newImageRmCommand(deps), newImageRebuildCommand(deps))
	return c
}

func newImageLsCommand(deps Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "ls",
		Short: "List built Images",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			imgs, err := image.ListExisting(deps.Docker)
			if err != nil {
				return err
			}
			n, err := loadNames(deps)
			if err != nil {
				return err
			}
			projects, err := project.NewStore(deps.ConfigRoot).List()
			if err != nil {
				return err
			}
			rows := make([][]string, len(imgs))
			for i, img := range imgs {
				stackName, profileName := n.stack(img.StackID), n.profile(img.ProfileID)
				rows[i] = []string{
					image.Name(stackName, profileName),
					stackName,
					profileName,
					strconv.Itoa(len(projectsUsing(projects, img))),
				}
			}
			return table.Print(cmd.OutOrStdout(), "No images.", []string{"IMAGE", "STACK", "PROFILE", "PROJECTS"}, rows)
		},
	}
}

func newImageRmCommand(deps Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "rm <stack>:<profile>",
		Short: "Remove a built Image (Projects using it rebuild it next session)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			st, p, err := resolveImage(deps, args[0])
			if err != nil {
				return err
			}
			name := image.Name(st.Name, p.Name)
			ref := image.Ref(st.ID, p.ID)
			exists, err := deps.Docker.ImageExists(ref)
			if err != nil {
				return err
			}
			if !exists {
				return fmt.Errorf("image %s is not built", name)
			}
			projects, err := project.NewStore(deps.ConfigRoot).List()
			if err != nil {
				return err
			}
			if err := deps.Docker.RemoveImage(ref); err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Removed image %s\n", name)
			affected := projectsUsing(projects, image.Existing{Ref: ref, StackID: st.ID, ProfileID: p.ID})
			if len(affected) > 0 {
				fmt.Fprintln(out, "These Projects will rebuild it at their next session:")
				for _, pr := range affected {
					fmt.Fprintf(out, "  %s\n", pr.Path)
				}
			}
			return nil
		},
	}
}

func newImageRebuildCommand(deps Deps) *cobra.Command {
	var all bool
	c := &cobra.Command{
		Use:   "rebuild [<stack>:<profile> | --all]",
		Short: "Rebuild one Image, or every built Image with --all",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if all && len(args) > 0 {
				return errors.New("pass an image or --all, not both")
			}
			if !all && len(args) == 0 {
				return errors.New("pass an image or --all (e.g. `safeagent image rebuild node22:work`)")
			}
			if all {
				return rebuildAllImages(cmd, deps)
			}
			st, p, err := resolveImage(deps, args[0])
			if err != nil {
				return err
			}
			if err := image.Build(deps.Docker, deps.ConfigRoot, st, p); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Rebuilt image %s\n", image.Name(st.Name, p.Name))
			return nil
		},
	}
	c.Flags().BoolVar(&all, "all", false, "Rebuild every built safeagent Image")
	return c
}

// rebuildAllImages rebuilds every existing safeagent Image. Images whose Stack
// or Profile no longer exists can't be rebuilt, so they're skipped with a note.
func rebuildAllImages(cmd *cobra.Command, deps Deps) error {
	imgs, err := image.ListExisting(deps.Docker)
	if err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	if len(imgs) == 0 {
		fmt.Fprintln(out, "No images to rebuild.")
		return nil
	}
	n, err := loadNames(deps)
	if err != nil {
		return err
	}
	stacks, profiles := stack.NewStore(deps.ConfigRoot), profile.NewStore(deps.ConfigRoot)
	for _, img := range imgs {
		name := image.Name(n.stack(img.StackID), n.profile(img.ProfileID))
		st, stErr := stacks.FindByID(img.StackID)
		p, pErr := profiles.FindByID(img.ProfileID)
		if stErr != nil || pErr != nil {
			fmt.Fprintf(out, "Skipped %s: its Stack or Profile no longer exists\n", name)
			continue
		}
		if err := image.Build(deps.Docker, deps.ConfigRoot, st, p); err != nil {
			return fmt.Errorf("rebuild %s: %w", name, err)
		}
		fmt.Fprintf(out, "Rebuilt image %s\n", name)
	}
	return nil
}

// resolveImage turns the CLI form "<stack>:<profile>" into its Stack and
// Profile records, erroring on a malformed reference or an unknown Name.
func resolveImage(deps Deps, s string) (stack.Stack, profile.Profile, error) {
	stackName, profileName, err := image.ParseName(s)
	if err != nil {
		return stack.Stack{}, profile.Profile{}, err
	}
	st, err := stack.NewStore(deps.ConfigRoot).FindByName(stackName)
	if err != nil {
		return stack.Stack{}, profile.Profile{}, err
	}
	p, err := profile.NewStore(deps.ConfigRoot).FindByName(profileName)
	if err != nil {
		return stack.Stack{}, profile.Profile{}, err
	}
	return st, p, nil
}

// projectsUsing returns the Projects whose Stack and Profile make up img.
func projectsUsing(projects []project.Project, img image.Existing) []project.Project {
	var out []project.Project
	for _, p := range projects {
		if p.StackID == img.StackID && p.ProfileID == img.ProfileID {
			out = append(out, p)
		}
	}
	return out
}
