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
			st, prof, err := resolveImage(deps, args[0])
			if err != nil {
				return err
			}
			name := image.Name(st.Name, prof.Name)
			ref := image.Ref(st.ID, prof.ID)
			exists, err := deps.Docker.ImageExists(ref)
			if err != nil {
				return err
			}
			projects, err := project.NewStore(deps.ConfigRoot).List()
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			header := "These Projects will rebuild it at their next session:"
			if exists {
				if err := deps.Docker.RemoveImage(ref); err != nil {
					return fmt.Errorf("removing image %s: %w", name, err)
				}
				fmt.Fprintf(out, "Removed image %s.\n", name)
			} else {
				fmt.Fprintf(out, "Image %s is not built; nothing to remove.\n", name)
				header = "These Projects will build it at their next session:"
			}
			affected := projectsUsing(projects, image.Existing{Ref: ref, StackID: st.ID, ProfileID: prof.ID})
			if len(affected) > 0 {
				fmt.Fprintln(out, header)
				for _, proj := range affected {
					fmt.Fprintf(out, "  %s\n", proj.Path)
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
		Short: "Rebuild one built Image, or every built Image with --all",
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
			st, prof, err := resolveImage(deps, args[0])
			if err != nil {
				return err
			}
			ref := image.Ref(st.ID, prof.ID)
			exists, err := deps.Docker.ImageExists(ref)
			if err != nil {
				return err
			}
			if !exists {
				return fmt.Errorf("image %s is not built; it will be built the next time a session needs it", image.Name(st.Name, prof.Name))
			}
			pairs := []image.Existing{{Ref: ref, StackID: st.ID, ProfileID: prof.ID}}
			return rebuildPairs(cmd, deps, pairs, "", func(image.Existing) (stack.Stack, profile.Profile, error) {
				return st, prof, nil
			})
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
	stacks, profiles := stack.NewStore(deps.ConfigRoot), profile.NewStore(deps.ConfigRoot)
	return rebuildPairs(cmd, deps, imgs, "No images to rebuild.", func(img image.Existing) (stack.Stack, profile.Profile, error) {
		st, err := stacks.FindByID(img.StackID)
		if err != nil {
			return stack.Stack{}, profile.Profile{}, err
		}
		prof, err := profiles.FindByID(img.ProfileID)
		return st, prof, err
	})
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
	prof, err := profile.NewStore(deps.ConfigRoot).FindByName(profileName)
	if err != nil {
		return stack.Stack{}, profile.Profile{}, err
	}
	return st, prof, nil
}

// projectsUsing returns the Projects whose Stack and Profile make up img.
func projectsUsing(projects []project.Project, img image.Existing) []project.Project {
	var out []project.Project
	for _, proj := range projects {
		if proj.StackID == img.StackID && proj.ProfileID == img.ProfileID {
			out = append(out, proj)
		}
	}
	return out
}
